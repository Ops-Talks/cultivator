// Package cli provides the command-line interface for cultivator.
package cli

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/Ops-Talks/cultivator/internal/config"
	"github.com/Ops-Talks/cultivator/internal/discovery"
	"github.com/Ops-Talks/cultivator/internal/git"
	"github.com/Ops-Talks/cultivator/internal/logging"
	"github.com/Ops-Talks/cultivator/internal/runner"
)

const (
	cmdPlan    = "plan"
	cmdApply   = "apply"
	cmdDestroy = "destroy"
	cmdVersion = "version"
	cmdDoctor  = "doctor"
)

// VersionInfo holds build-time version metadata for the cultivator binary.
type VersionInfo struct {
	Version string
	Commit  string
	Date    string
}

// Run is the entry point for the cultivator CLI. It dispatches subcommands and
// returns an exit code suitable for os.Exit.
func Run(args []string, version VersionInfo) int {
	if len(args) < 2 {
		printUsage()
		return 2
	}

	command := args[1]
	switch command {
	case cmdPlan, cmdApply, cmdDestroy:
		return runTerragruntCommand(args[2:], command, runner.New())
	case cmdVersion:
		printVersion(version)
		return 0
	case cmdDoctor:
		return runDoctor(args[2:])
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", command)
		printUsage()
		return 2
	}
}

func runTerragruntCommand(args []string, command string, r *runner.Runner) int {
	state, code := parseTerragruntFlags(args, command)
	if code != 0 {
		return code
	}

	cfg, err := buildTerragruntConfig(state)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return 1
	}

	logger := logging.New(logLevelFromEnv(), os.Stdout, os.Stderr)
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	modules, err := discovery.Discover(cfg.Root, discovery.Options{
		Env:     cfg.Env,
		Include: cfg.Include,
		Exclude: cfg.Exclude,
		Tags:    cfg.Tags,
		Logger:  logger,
	})
	if err != nil {
		logger.Error("module discovery failed", logging.Fields{"error": err.Error()})
		return 1
	}

	if cfg.ChangedOnly {
		if !git.IsGitRepo(cfg.Root) {
			logger.Error("not a git repository, --changed-only is not supported", logging.Fields{"root": cfg.Root})
			return 1
		}

		changedFiles, err := git.GetChangedFiles(ctx, cfg.Root, cfg.BaseRef)
		if err != nil {
			logger.Error("failed to get changed files", logging.Fields{"error": err.Error(), "base": cfg.BaseRef})
			return 1
		}

		var filtered []discovery.Module
		for _, mod := range modules {
			hasChanges := false
			for _, changedFile := range changedFiles {
				// Normalize both for comparison
				modPath := filepath.Clean(mod.Path)
				changePath := filepath.Clean(changedFile)

				if strings.HasPrefix(changePath, modPath) {
					hasChanges = true
					break
				}
			}
			if hasChanges {
				filtered = append(filtered, mod)
			}
		}
		modules = filtered
	}

	if len(modules) == 0 {
		logger.Info("no modules matched", logging.Fields{"root": cfg.Root, "changed_only": cfg.ChangedOnly})
		return 0
	}

	logger.Info("modules discovered", logging.Fields{"count": len(modules), "root": cfg.Root})

	startTime := time.Now()
	results, runErr := runTerragruntModules(ctx, logger, r, command, cfg, modules)
	duration := time.Since(startTime)

	// Log summary table at the end
	if len(results) > 0 {
		var rows []logging.SummaryRow
		for _, res := range results {
			status := "SUCCESS"
			notes := ""
			if res.Error != nil || res.ExitCode != 0 {
				status = "FAILURE"
				if res.Error != nil {
					notes = res.Error.Error()
				} else {
					notes = fmt.Sprintf("exit code %d", res.ExitCode)
				}
			}
			rows = append(rows, logging.SummaryRow{
				Module:   res.Module.Path,
				Command:  res.Command,
				Status:   status,
				Duration: res.Duration.String(),
				Notes:    notes,
			})
		}
		logger.LogSummaryTable(rows, duration.String())
	}

	if runErr != nil {
		logger.Error("execution completed with errors", logging.Fields{
			"error":    runErr.Error(),
			"duration": duration.String(),
		})
		return 1
	}

	logger.Info("execution completed", logging.Fields{
		"modules":  len(modules),
		"duration": duration.String(),
	})
	return 0
}

type terragruntFlagState struct {
	configPath              string
	root                    string
	env                     string
	module                  string // specific module path from positional argument (e.g., "cloudwatch/log-group/example")
	includeValues           []string
	includeSet              bool
	excludeValues           []string
	excludeSet              bool
	tagsValues              []string
	tagsSet                 bool
	parallelismValue        int
	parallelismSet          bool
	nonInteractiveValue     bool
	nonInteractiveSet       bool
	dryRunValue             bool
	dryRunSet               bool
	changedOnlyValue        bool
	changedOnlySet          bool
	baseRefValue            string
	baseRefSet              bool
	planDestroyValue        bool
	planDestroySet          bool
	applyAutoApproveValue   bool
	applyAutoApproveSet     bool
	destroyAutoApproveValue bool
	destroyAutoApproveSet   bool
}

func parseTerragruntFlags(args []string, command string) (terragruntFlagState, int) {
	state := terragruntFlagState{}
	fs := flag.NewFlagSet(command, flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	configPath := fs.String("config", "", "path to config file")
	root := fs.String("root", "", "root directory for terragrunt modules")
	env := fs.String("env", "", "environment filter")
	include := newStringSliceFlag(fs, "include", "relative paths to include")
	exclude := newStringSliceFlag(fs, "exclude", "relative paths to exclude")
	tags := newStringSliceFlag(fs, "tags", "tag filters")
	parallelism := newIntFlag(fs, "parallelism", "max parallel executions")
	nonInteractive := newBoolFlag(fs, "non-interactive", "force non-interactive mode")
	dryRun := newBoolFlag(fs, "dry-run", "don't execute terragrunt commands")
	changedOnly := newBoolFlag(fs, "changed-only", "only execute modules with changed files")
	baseRef := fs.String("base", "", "git base reference for --changed-only")

	var planDestroy *boolFlag
	var applyAutoApprove *boolFlag
	var destroyAutoApprove *boolFlag

	switch command {
	case cmdPlan:
		planDestroy = newBoolFlag(fs, "destroy", "plan destroy")
	case cmdApply:
		applyAutoApprove = newBoolFlag(fs, "auto-approve", "auto approve apply")
	case cmdDestroy:
		destroyAutoApprove = newBoolFlag(fs, "auto-approve", "auto approve destroy")
	}

	if err := fs.Parse(args); err != nil {
		return state, 2
	}

	state.configPath = *configPath
	state.root = *root
	state.env = *env

	populateFlagState(&state, fs, include, exclude, tags, parallelism, nonInteractive, dryRun, changedOnly, baseRef, planDestroy, applyAutoApprove, destroyAutoApprove, command)

	return state, 0
}

func populateFlagState(state *terragruntFlagState, fs *flag.FlagSet, include, exclude, tags *stringSliceFlag, parallelism *intFlag, nonInteractive, dryRun, changedOnly *boolFlag, baseRef *string, planDestroy, applyAutoApprove, destroyAutoApprove *boolFlag, command string) {
	// Capture positional argument (module path) if provided.
	if len(fs.Args()) > 0 {
		state.module = normalizePath(fs.Args()[0])
	}

	// Process include/exclude/tags filters first
	if include.set {
		state.includeValues = include.values
		state.includeSet = true
	}
	if exclude.set {
		state.excludeValues = exclude.values
		state.excludeSet = true
	}
	if tags.set {
		state.tagsValues = tags.values
		state.tagsSet = true
	}

	// If a specific module path is provided, add it to include filters
	if state.module != "" {
		if state.includeSet {
			state.includeValues = append(state.includeValues, state.module)
		} else {
			state.includeValues = []string{state.module}
			state.includeSet = true
		}
	}

	if parallelism.set {
		state.parallelismValue = parallelism.value
		state.parallelismSet = true
	}
	if nonInteractive.set {
		state.nonInteractiveValue = nonInteractive.value
		state.nonInteractiveSet = true
	}
	if dryRun.set {
		state.dryRunValue = dryRun.value
		state.dryRunSet = true
	}
	if changedOnly.set {
		state.changedOnlyValue = changedOnly.value
		state.changedOnlySet = true
	}
	if baseRef != nil && *baseRef != "" {
		state.baseRefValue = *baseRef
		state.baseRefSet = true
	}

	switch command {
	case cmdPlan:
		if planDestroy != nil && planDestroy.set {
			state.planDestroyValue = planDestroy.value
			state.planDestroySet = true
		}
	case cmdApply:
		if applyAutoApprove != nil && applyAutoApprove.set {
			state.applyAutoApproveValue = applyAutoApprove.value
			state.applyAutoApproveSet = true
		}
	case cmdDestroy:
		if destroyAutoApprove != nil && destroyAutoApprove.set {
			state.destroyAutoApproveValue = destroyAutoApprove.value
			state.destroyAutoApproveSet = true
		}
	}
}

func buildTerragruntConfig(state terragruntFlagState) (config.Config, error) {
	cfg := config.DefaultConfig()
	fileCfg, _, found, err := config.LoadFile(state.configPath)
	if err != nil {
		return cfg, err
	}
	if found {
		cfg = config.MergeConfig(cfg, fileCfg)
	}

	envCfg := config.LoadEnv("CULTIVATOR")
	cfg = config.MergeConfig(cfg, envCfg)
	cfg = config.ApplyOverrides(cfg, buildOverrides(state))

	if !filepath.IsAbs(cfg.Root) {
		wd, err := os.Getwd()
		if err != nil {
			return cfg, fmt.Errorf("failed to get working directory: %w", err)
		}
		cfg.Root = filepath.Join(wd, cfg.Root)
	}

	if err := config.Validate(cfg); err != nil {
		return cfg, err
	}

	return cfg, nil
}

func buildOverrides(state terragruntFlagState) config.Overrides {
	flagOverrides := config.Overrides{}
	if state.root != "" {
		flagOverrides.Root = &state.root
	}
	if state.env != "" {
		flagOverrides.Env = &state.env
	}
	if state.includeSet {
		flagOverrides.Include = state.includeValues
		flagOverrides.IncludeSet = true
	}
	if state.excludeSet {
		flagOverrides.Exclude = state.excludeValues
		flagOverrides.ExcludeSet = true
	}
	if state.tagsSet {
		flagOverrides.Tags = state.tagsValues
		flagOverrides.TagsSet = true
	}
	if state.parallelismSet {
		value := state.parallelismValue
		flagOverrides.Parallelism = &value
	}
	if state.nonInteractiveSet {
		value := state.nonInteractiveValue
		flagOverrides.NonInteractive = &value
	}
	if state.dryRunSet {
		value := state.dryRunValue
		flagOverrides.DryRun = &value
	}
	if state.changedOnlySet {
		value := state.changedOnlyValue
		flagOverrides.ChangedOnly = &value
	}
	if state.baseRefSet {
		value := state.baseRefValue
		flagOverrides.BaseRef = &value
	}
	if state.planDestroySet {
		value := state.planDestroyValue
		flagOverrides.PlanDestroy = &value
	}
	if state.applyAutoApproveSet {
		value := state.applyAutoApproveValue
		flagOverrides.ApplyAutoAppr = &value
	}
	if state.destroyAutoApproveSet {
		value := state.destroyAutoApproveValue
		flagOverrides.DestroyAutoApr = &value
	}

	return flagOverrides
}

func runTerragruntModules(ctx context.Context, logger *logging.Logger, r *runner.Runner, command string, cfg config.Config, modules []discovery.Module) ([]runner.Result, error) {
	switch command {
	case cmdPlan:
		planDestroy := false
		if val, ok := cfg.Plan["destroy"]; ok {
			if b, ok := val.(bool); ok {
				planDestroy = b
			}
		}
		results, err := r.Run(ctx, runner.CommandPlan, modules, runner.Options{
			Parallelism:    cfg.Parallelism,
			NonInteractive: cfg.NonInteractive,
			DryRun:         cfg.DryRun,
			PlanDestroy:    planDestroy,
		})
		if err != nil {
			return nil, err
		}
		return results, logExecutionResults(logger, results)
	case cmdApply:
		applyAutoApprove := false
		if val, ok := cfg.Apply["auto_approve"]; ok {
			if b, ok := val.(bool); ok {
				applyAutoApprove = b
			}
		}
		results, err := r.Run(ctx, runner.CommandApply, modules, runner.Options{
			Parallelism:      cfg.Parallelism,
			NonInteractive:   cfg.NonInteractive,
			DryRun:           cfg.DryRun,
			ApplyAutoApprove: applyAutoApprove,
		})
		if err != nil {
			return nil, err
		}
		return results, logExecutionResults(logger, results)
	case cmdDestroy:
		destroyAutoApprove := false
		if val, ok := cfg.Destroy["auto_approve"]; ok {
			if b, ok := val.(bool); ok {
				destroyAutoApprove = b
			}
		}
		results, err := r.Run(ctx, runner.CommandDestroy, modules, runner.Options{
			Parallelism:        cfg.Parallelism,
			NonInteractive:     cfg.NonInteractive,
			DryRun:             cfg.DryRun,
			DestroyAutoApprove: destroyAutoApprove,
		})
		if err != nil {
			return nil, err
		}
		return results, logExecutionResults(logger, results)
	default:
		return nil, fmt.Errorf("unknown command: %s", command)
	}
}

// logExecutionResults processes execution results and displays complete terragrunt output.
// Each module's output is prefixed with a header ("=== command: module ===") so that
// CI pipelines can parse and attribute output per module when posting PR/MR comments.
//
// stdout and stderr are merged by the runner into a single chronologically-ordered
// stream (result.Stdout), preserving the interleaved output produced by Terragrunt
// and the underlying IaC tool.
func logExecutionResults(logger *logging.Logger, results []runner.Result) error {
	hasErrors := false
	for _, result := range results {
		logger.Output(fmt.Sprintf("=== %s: %s ===", result.Command, result.Module.Path))

		if result.Stdout != "" {
			logger.Output(result.Stdout)
		}

		if result.Error != nil || result.ExitCode != 0 {
			hasErrors = true
			fields := logging.Fields{
				"exit_code": result.ExitCode,
				"duration":  result.Duration.String(),
			}
			if result.Error != nil {
				fields["error"] = result.Error.Error()
			}
			logger.Error(fmt.Sprintf("%s %s failed", result.Command, result.Module.Path), fields)
		} else {
			logger.Info(fmt.Sprintf("%s %s", result.Command, result.Module.Path), logging.Fields{
				"exit_code": result.ExitCode,
				"duration":  result.Duration.String(),
			})
		}
	}
	if hasErrors {
		return fmt.Errorf("execution failed for one or more modules")
	}
	return nil
}

func runDoctor(args []string) int {
	fs := flag.NewFlagSet("doctor", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	configPath := fs.String("config", "", "path to config file")
	root := fs.String("root", "", "root directory for terragrunt modules")
	if err := fs.Parse(args); err != nil {
		return 2
	}

	cfg := config.DefaultConfig()
	fileCfg, _, found, err := config.LoadFile(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return 1
	}
	if found {
		cfg = config.MergeConfig(cfg, fileCfg)
	}
	configFilePath := ""
	if found {
		configFilePath = *configPath
	}

	if root != nil && *root != "" {
		cfg.Root = *root
	}

	if !filepath.IsAbs(cfg.Root) {
		wd, err := os.Getwd()
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to get working directory: %v\n", err)
			return 1
		}
		cfg.Root = filepath.Join(wd, cfg.Root)
	}

	logger := logging.New(logLevelFromEnv(), os.Stdout, os.Stderr)

	terragruntPath, err := exec.LookPath("terragrunt")
	if err != nil {
		logger.Error("terragrunt not found in PATH", logging.Fields{"error": err.Error()})
		return 1
	}

	logger.Info("terragrunt found", logging.Fields{"path": terragruntPath})

	if configFilePath != "" {
		logger.Info("config file loaded", logging.Fields{"path": configFilePath})
	} else {
		logger.Info("config file not found", logging.Fields{"path": "(default)"})
	}

	if cfg.Root == "" {
		logger.Error("root directory not set", logging.Fields{})
		return 1
	}

	absRoot, err := filepath.Abs(cfg.Root)
	if err != nil {
		logger.Error("failed to resolve root", logging.Fields{"error": err.Error()})
		return 1
	}

	info, err := os.Stat(absRoot)
	if err != nil {
		logger.Error("root directory check failed", logging.Fields{"error": err.Error(), "root": absRoot})
		return 1
	}
	if !info.IsDir() {
		logger.Error("root path is not a directory", logging.Fields{"root": absRoot})
		return 1
	}

	logger.Info("root directory ok", logging.Fields{"root": absRoot})
	return 0
}

func printVersion(version VersionInfo) {
	fmt.Printf("cultivator %s (commit %s, built %s)\n", fallback(version.Version, "dev"), fallback(version.Commit, "unknown"), fallback(version.Date, "unknown"))
}

func fallback(value, def string) string {
	if value == "" {
		return def
	}
	return value
}

func printUsage() {
	usage := []string{
		fmt.Sprintf("cultivator %s [flags]", cmdPlan),
		fmt.Sprintf("cultivator %s [flags]", cmdApply),
		fmt.Sprintf("cultivator %s [flags]", cmdDestroy),
		fmt.Sprintf("cultivator %s", cmdVersion),
		fmt.Sprintf("cultivator %s", cmdDoctor),
	}

	fmt.Fprintf(os.Stderr, "Usage:\n  %s\n", strings.Join(usage, "\n  "))
}

type stringSliceFlag struct {
	values []string
	set    bool
}

func newStringSliceFlag(fs *flag.FlagSet, name string, usage string) *stringSliceFlag {
	flag := &stringSliceFlag{}
	fs.Var(flag, name, usage)
	return flag
}

func (s *stringSliceFlag) String() string {
	return strings.Join(s.values, ",")
}

func (s *stringSliceFlag) Set(value string) error {
	s.values = append(s.values, splitList(value)...)
	s.set = true
	return nil
}

type boolFlag struct {
	value bool
	set   bool
}

func newBoolFlag(fs *flag.FlagSet, name string, usage string) *boolFlag {
	flag := &boolFlag{}
	fs.Var(flag, name, usage)
	return flag
}

func (b *boolFlag) String() string {
	return fmt.Sprintf("%t", b.value)
}

func (b *boolFlag) Set(value string) error {
	parsed, err := parseBool(value)
	if err != nil {
		return err
	}
	b.value = parsed
	b.set = true
	return nil
}

type intFlag struct {
	value int
	set   bool
}

func newIntFlag(fs *flag.FlagSet, name string, usage string) *intFlag {
	flag := &intFlag{}
	fs.Var(flag, name, usage)
	return flag
}

func (i *intFlag) String() string {
	return fmt.Sprintf("%d", i.value)
}

func (i *intFlag) Set(value string) error {
	parsed, err := parseInt(value)
	if err != nil {
		return err
	}
	i.value = parsed
	i.set = true
	return nil
}

// normalizePath normalizes a module path by removing leading ./ and trailing /terragrunt.hcl.
// It handles both Unix-style (/) and platform-specific separators.
// Examples: "cloudwatch/log-group/example/terragrunt.hcl" → "cloudwatch/log-group/example"
//
//	"./cloudwatch/log-group/example" → "cloudwatch/log-group/example"
//
// logLevelFromEnv reads the CULTIVATOR_LOG_LEVEL environment variable and returns
// the corresponding logging.Level. Defaults to LevelInfo if the variable is unset
// or contains an unrecognized value.
func logLevelFromEnv() logging.Level {
	v := os.Getenv("CULTIVATOR_LOG_LEVEL")
	if v == "" {
		return logging.LevelInfo
	}

	level, err := logging.ParseLevel(v)
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: %v; defaulting to info\n", err)
		return logging.LevelInfo
	}

	return level
}

func normalizePath(path string) string {
	path = strings.TrimSpace(path)
	if path == "" {
		return ""
	}

	// Remove leading ./
	if strings.HasPrefix(path, "."+string(filepath.Separator)) {
		path = path[2:]
	}

	// Remove trailing /terragrunt.hcl
	if strings.HasSuffix(path, "/terragrunt.hcl") {
		path = path[:len(path)-len("/terragrunt.hcl")]
	} else if strings.HasSuffix(path, string(filepath.Separator)+"terragrunt.hcl") {
		path = path[:len(path)-len(string(filepath.Separator)+"terragrunt.hcl")]
	}

	return path
}

func splitList(value string) []string {
	parts := strings.FieldsFunc(value, func(r rune) bool {
		return r == ',' || r == ';'
	})

	items := make([]string, 0, len(parts))
	for _, part := range parts {
		clean := strings.TrimSpace(part)
		if clean == "" {
			continue
		}
		items = append(items, clean)
	}

	return items
}

func parseBool(value string) (bool, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "true", "1", "t", "yes", "y":
		return true, nil
	case "false", "0", "f", "no", "n":
		return false, nil
	default:
		return false, errors.New("invalid boolean value")
	}
}

func parseInt(value string) (int, error) {
	clean := strings.TrimSpace(value)
	if clean == "" {
		return 0, errors.New("invalid integer value")
	}

	parsed, err := strconv.Atoi(clean)
	if err != nil {
		return 0, errors.New("invalid integer value")
	}
	return parsed, nil
}
