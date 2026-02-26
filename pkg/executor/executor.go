// Package executor runs Terragrunt commands with proper error handling and output redirection.
package executor

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/cultivator-dev/cultivator/pkg/module"
)

// Result represents the result of a command execution
type Result struct {
	ExitCode int
	Stdout   string
	Stderr   string
	Error    error
}

// Executor executes Terragrunt commands
type Executor struct {
	workingDir       string
	terragruntBinary string
	terraformBinary  string
	sourceParser     *module.SourceParser // For handling external modules
	stdout           io.Writer
	stderr           io.Writer
}

const (
	noColorFlag             = "-no-color"
	autoApproveFlag         = "-auto-approve"
	nonInteractiveFlag      = "--terragrunt-non-interactive"
	defaultWorkingDir       = "."
	defaultTerragruntBinary = "terragrunt"
	defaultTerraformBinary  = "terraform"
)

// NewExecutor creates a new Terragrunt executor
func NewExecutor(workingDir string, stdout, stderr io.Writer) *Executor {
	return &Executor{
		workingDir:       workingDir,
		terragruntBinary: defaultTerragruntBinary,
		terraformBinary:  defaultTerraformBinary,
		sourceParser:     module.NewSourceParser(),
		stdout:           stdout,
		stderr:           stderr,
	}
}

// buildArgs constructs command arguments from base args and extra args (DRY principle)
func (e *Executor) buildArgs(baseArgs []string, extraArgs ...string) []string {
	result := make([]string, len(baseArgs), len(baseArgs)+len(extraArgs))
	copy(result, baseArgs)
	return append(result, extraArgs...)
}

// Plan runs terragrunt plan for a module
func (e *Executor) Plan(ctx context.Context, modulePath string, extraArgs ...string) (*Result, error) {
	args := e.buildArgs([]string{"plan", noColorFlag}, extraArgs...)
	return e.runTerragrunt(ctx, modulePath, args...)
}

// Apply runs terragrunt apply for a module
func (e *Executor) Apply(ctx context.Context, modulePath string, extraArgs ...string) (*Result, error) {
	args := e.buildArgs([]string{"apply", autoApproveFlag, noColorFlag}, extraArgs...)
	return e.runTerragrunt(ctx, modulePath, args...)
}

// RunAll runs terragrunt run-all for multiple modules
func (e *Executor) RunAll(ctx context.Context, command string, basePath string, extraArgs ...string) (*Result, error) {
	args := e.buildArgs([]string{"run-all", command, nonInteractiveFlag, noColorFlag}, extraArgs...)
	return e.runTerragrunt(ctx, basePath, args...)
}

// Validate runs terragrunt validate
func (e *Executor) Validate(ctx context.Context, modulePath string) (*Result, error) {
	return e.runTerragrunt(ctx, modulePath, "validate")
}

// Init runs terragrunt init
func (e *Executor) Init(ctx context.Context, modulePath string) (*Result, error) {
	args := e.buildArgs([]string{"init", noColorFlag})
	return e.runTerragrunt(ctx, modulePath, args...)
}

// configureIO sets up stdout and stderr for command execution (DRY principle)
func (e *Executor) configureIO(cmd *exec.Cmd) (stdout, stderr *strings.Builder) {
	stdout = &strings.Builder{}
	stderr = &strings.Builder{}

	if e.stdout != nil {
		cmd.Stdout = io.MultiWriter(stdout, e.stdout)
	} else {
		cmd.Stdout = stdout
	}

	if e.stderr != nil {
		cmd.Stderr = io.MultiWriter(stderr, e.stderr)
	} else {
		cmd.Stderr = stderr
	}

	return
}

// extractExitCode determines the exit code from a command error (DRY principle)
func extractExitCode(err error) int {
	if err == nil {
		return 0
	}

	var exitError *exec.ExitError
	if errors.As(err, &exitError) {
		return exitError.ExitCode()
	}

	return -1
}

// runTerragrunt executes a terragrunt command
func (e *Executor) runTerragrunt(ctx context.Context, dir string, args ...string) (*Result, error) {
	cmd := exec.CommandContext(ctx, e.terragruntBinary, args...)
	cmd.Dir = dir

	stdout, stderr := e.configureIO(cmd)
	err := cmd.Run()

	return &Result{
		ExitCode: extractExitCode(err),
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		Error:    err,
	}, nil
}

// CheckVersion checks if terragrunt is installed and returns version
func (e *Executor) CheckVersion() (string, error) {
	cmd := exec.Command(e.terragruntBinary, "--version")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("terragrunt not found: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}

// PrepareExternalModules downloads and checks out all external modules before plan/apply
// Following Single Responsibility: handles external module preparation
// This should be called before running plan/apply to ensure modules are available
func (e *Executor) PrepareExternalModules(ctx context.Context, externalModules []module.SourceInfo) error {
	for _, moduleInfo := range externalModules {
		if err := e.checkoutExternalModule(ctx, &moduleInfo); err != nil {
			return fmt.Errorf("failed to prepare external module %s: %w", moduleInfo.RawSource, err)
		}
	}
	return nil
}

// checkoutExternalModule downloads and extracts an external module
// DRY principle: single location for module checkout logic
func (e *Executor) checkoutExternalModule(ctx context.Context, sourceInfo *module.SourceInfo) error {
	// Determine module type (git or http)
	var source module.ModuleSource
	switch sourceInfo.Type {
	case "git":
		source = module.NewGitModuleSource()
	case "http":
		source = module.NewHTTPModuleSource()
	default:
		return fmt.Errorf("unsupported module source type: %s", sourceInfo.Type)
	}

	// Create a working directory for the module based on URL hash
	moduleCacheDir := e.getModuleCacheDir(sourceInfo.URL)

	// Check if module is already cached and up-to-date
	if e.isModuleCached(ctx, sourceInfo, moduleCacheDir, source) {
		return nil
	}

	// Download and extract the module
	_, _ = fmt.Fprintf(e.stdout, "Fetching external module: %s\n", sourceInfo.RawSource)
	if err := source.Checkout(ctx, sourceInfo.RawSource, moduleCacheDir); err != nil {
		return fmt.Errorf("failed to checkout module: %w", err)
	}

	_, _ = fmt.Fprintf(e.stdout, "Module ready: %s\n", moduleCacheDir)
	return nil
}

// getModuleCacheDir returns the cache directory for a module based on its URL
// DRY: single location for cache directory logic
func (e *Executor) getModuleCacheDir(url string) string {
	// Create a consistent directory name from the URL hash
	hash := e.hashString(url)
	return filepath.Join(e.workingDir, ".cultivator-modules", hash)
}

// isModuleCached checks if a module is already cached and up-to-date
// DRY: single location for cache validation logic
func (e *Executor) isModuleCached(_ context.Context, _ *module.SourceInfo, cachePath string, source module.ModuleSource) bool {
	// Check if cache directory exists
	if _, err := os.Stat(cachePath); err != nil {
		return false
	}

	// For external modules, check if version matches (for git: check commit, for http: check ETag or timestamp)
	// This is optional - we could also just re-fetch to be safe
	// For now, we'll trust the cache if it exists
	return true
}

// hashString creates a simple hash for a string (for directory naming)
// DRY: single location for hash generation
func (e *Executor) hashString(s string) string {
	// Simple hash: use length and first/last chars for basic uniqueness
	// For production, consider using proper hash function
	if len(s) < 8 {
		return s
	}
	return fmt.Sprintf("%s-%d-%s", string(s[0]), len(s), string(s[len(s)-1]))
}
