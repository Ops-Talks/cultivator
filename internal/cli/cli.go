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
	"strings"
	"syscall"

	"github.com/Ops-Talks/cultivator/internal/config"
	"github.com/Ops-Talks/cultivator/internal/discovery"
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

type VersionInfo struct {
	Version string
	Commit  string
	Date    string
}

func Run(args []string, version VersionInfo) int {
	if len(args) < 2 {
		printUsage()
		return 2
	}

	command := args[1]
	switch command {
	case cmdPlan, cmdApply, cmdDestroy:
		return runTerragruntCommand(args[2:], command)
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

func runTerragruntCommand(args []string, command string) int {
	state, code := parseTerragruntFlags(args, command)
	if code != 0 {
		return code
	}

	cfg, err := buildTerragruntConfig(state)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return 1
	}

	logger := logging.New(cfg.OutputFormat, os.Stdout, os.Stderr)
	ctx := withSignalContext(context.Background())

	modules, err := discovery.Discover(cfg.Root, discovery.Options{
		Env:     cfg.Env,
		Include: cfg.Include,
		Exclude: cfg.Exclude,
		Tags:    cfg.Tags,
	})
	if err != nil {
		logger.Error("module discovery failed", logging.Fields{"error": err.Error()})
		return 1
	}

	if len(modules) == 0 {
		logger.Info("no modules matched", logging.Fields{"root": cfg.Root})
		return 0
	}

	logger.Info("modules discovered", logging.Fields{"count": len(modules), "root": cfg.Root})

	r := runner.New(logger)
	runErr := runTerragruntModules(ctx, logger, r, command, cfg, modules)
	if runErr != nil {
		logger.Error("execution completed with errors", logging.Fields{"error": runErr.Error()})
		return 1
	}

	logger.Info("execution completed", logging.Fields{"modules": len(modules)})
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
	outputFormat            string
	nonInteractiveValue     bool
	nonInteractiveSet       bool
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
	outputFormat := fs.String("output-format", "", "output format: text or json")
	nonInteractive := newBoolFlag(fs, "non-interactive", "force non-interactive mode")

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

	// Capture positional argument (module path) if provided.
	// This supports usage like: cultivator plan cloudwatch/log-group/example [flags]
	// The path is normalized and treated as an include filter.
	if len(fs.Args()) > 0 {
		modulePath := fs.Args()[0]
		state.module = normalizePath(modulePath)
	}

	state.configPath = *configPath
	state.root = *root
	state.env = *env

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
	if outputFormat != nil {
		state.outputFormat = *outputFormat
	}
	if nonInteractive.set {
		state.nonInteractiveValue = nonInteractive.value
		state.nonInteractiveSet = true
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

	return state, 0
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
	if state.outputFormat != "" {
		flagOverrides.OutputFormat = &state.outputFormat
	}
	if state.nonInteractiveSet {
		value := state.nonInteractiveValue
		flagOverrides.NonInteractive = &value
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

func runTerragruntModules(ctx context.Context, logger *logging.Logger, r *runner.Runner, command string, cfg config.Config, modules []discovery.Module) error {
	switch command {
	case cmdPlan:
		planDestroy := false
		if val, ok := cfg.Plan["destroy"]; ok {
			if b, ok := val.(bool); ok {
				planDestroy = b
			}
		}
		results, _ := r.Run(ctx, runner.CommandPlan, modules, runner.Options{
			Parallelism:    cfg.Parallelism,
			NonInteractive: cfg.NonInteractive,
			PlanDestroy:    planDestroy,
		})
		return logExecutionResults(logger, results)
	case cmdApply:
		applyAutoApprove := false
		if val, ok := cfg.Apply["auto_approve"]; ok {
			if b, ok := val.(bool); ok {
				applyAutoApprove = b
			}
		}
		results, _ := r.Run(ctx, runner.CommandApply, modules, runner.Options{
			Parallelism:      cfg.Parallelism,
			NonInteractive:   cfg.NonInteractive,
			ApplyAutoApprove: applyAutoApprove,
		})
		return logExecutionResults(logger, results)
	case cmdDestroy:
		destroyAutoApprove := false
		if val, ok := cfg.Destroy["auto_approve"]; ok {
			if b, ok := val.(bool); ok {
				destroyAutoApprove = b
			}
		}
		results, _ := r.Run(ctx, runner.CommandDestroy, modules, runner.Options{
			Parallelism:        cfg.Parallelism,
			NonInteractive:     cfg.NonInteractive,
			DestroyAutoApprove: destroyAutoApprove,
		})
		return logExecutionResults(logger, results)
	default:
		return fmt.Errorf("unknown command: %s", command)
	}
}

// logExecutionResults processes execution results and logs detailed error information including stderr.
func logExecutionResults(logger *logging.Logger, results []runner.Result) error {
	hasErrors := false
	for _, result := range results {
		if result.Error != nil || result.ExitCode != 0 {
			hasErrors = true
			logger.Error(fmt.Sprintf("%s %s failed", result.Command, result.Module.Path), logging.Fields{
				"exit_code": result.ExitCode,
				"stderr":    result.Stderr,
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

	logger := logging.New("text", os.Stdout, os.Stderr)

	terragruntPath, err := execLookPath("terragrunt")
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

func execLookPath(name string) (string, error) {
	path, err := exec.LookPath(name)
	if err != nil {
		return "", err
	}
	return path, nil
}

func withSignalContext(parent context.Context) context.Context {
	ctx, cancel := signal.NotifyContext(parent, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-ctx.Done()
		cancel()
	}()
	return ctx
}

func printVersion(version VersionInfo) {
	fmt.Printf("cultivator %s (commit %s, built %s)\n", fallback(version.Version, "dev"), fallback(version.Commit, "unknown"), fallback(version.Date, "unknown"))
}

func fallback(value string, fallback string) string {
	if value == "" {
		return fallback
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

	var parsed int
	_, err := fmt.Sscanf(clean, "%d", &parsed)
	if err != nil {
		return 0, errors.New("invalid integer value")
	}
	return parsed, nil
}
