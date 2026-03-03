// Package runner executes Terragrunt commands across multiple modules in parallel.
package runner

import (
	"bytes"
	"context"
	"errors"
	"os/exec"
	"sync"

	"github.com/Ops-Talks/cultivator/internal/discovery"
)

const (
	CommandPlan    = "plan"
	CommandApply   = "apply"
	CommandDestroy = "destroy"
)

// Result holds the outcome of a single Terragrunt execution for one module.
type Result struct {
	Module   discovery.Module
	Command  string
	Stdout   string
	Stderr   string
	ExitCode int
	Error    error
}

// Options configures the behaviour of a Runner.Run call.
type Options struct {
	Parallelism        int
	NonInteractive     bool
	PlanDestroy        bool
	ApplyAutoApprove   bool
	DestroyAutoApprove bool
}

// Executor runs an external command in a given working directory and returns its
// stdout, stderr, exit code, and any execution error.
type Executor interface {
	Run(ctx context.Context, workDir string, command string, args []string, extraEnv []string) (string, string, int, error)
}

// DefaultExecutor is the production Executor that delegates to exec.CommandContext.
type DefaultExecutor struct{}

// Run implements Executor by running the command via exec.CommandContext in workDir.
func (de *DefaultExecutor) Run(ctx context.Context, workDir string, command string, args []string, extraEnv []string) (string, string, int, error) {
	// #nosec G204 -- command is controlled by the caller and expected to be terragrunt
	cmd := exec.CommandContext(ctx, command, args...)
	cmd.Dir = workDir

	if len(extraEnv) > 0 {
		cmd.Env = append(cmd.Environ(), extraEnv...)
	}

	outBuf := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}

	cmd.Stdout = outBuf
	cmd.Stderr = errBuf

	err := cmd.Run()
	exitCode := 0
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		exitCode = exitErr.ExitCode()
	}

	return outBuf.String(), errBuf.String(), exitCode, err
}

// Runner orchestrates parallel Terragrunt executions across multiple modules.
type Runner struct {
	executor Executor
}

// New returns a Runner using the DefaultExecutor.
func New() *Runner {
	return &Runner{
		executor: &DefaultExecutor{},
	}
}

// WithExecutor replaces the Runner's executor and returns the Runner for chaining.
// It is primarily used in tests to inject a mock Executor.
func (r *Runner) WithExecutor(executor Executor) *Runner {
	r.executor = executor
	return r
}

// Run executes the given Terragrunt command for each module concurrently, bounded
// by opts.Parallelism. It always returns the full result slice; individual errors
// are captured inside each Result.
func (r *Runner) Run(ctx context.Context, command string, modules []discovery.Module, opts Options) ([]Result, error) {
	if opts.Parallelism < 1 {
		opts.Parallelism = 1
	}

	results := make([]Result, len(modules))
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, opts.Parallelism)

	for i, module := range modules {
		idx, mod := i, module
		wg.Go(func() {
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			args := BuildArgs(command, opts)
			stdout, stderr, exitCode, err := r.executor.Run(ctx, mod.Path, "terragrunt", args, []string{})

			results[idx] = Result{
				Module:   mod,
				Command:  command,
				Stdout:   stdout,
				Stderr:   stderr,
				ExitCode: exitCode,
				Error:    err,
			}
		})
	}

	wg.Wait()
	return results, nil
}

// BuildArgs constructs the argument list for a terragrunt invocation based on
// the command and the provided Options.
func BuildArgs(command string, opts Options) []string {
	args := []string{command, "-no-color"}

	switch command {
	case CommandPlan:
		if opts.PlanDestroy {
			args = append(args, "-destroy")
		}
		if opts.NonInteractive {
			args = append(args, "-input=false")
		}
	case CommandApply:
		if opts.ApplyAutoApprove {
			args = append(args, "-auto-approve")
		}
		if opts.NonInteractive {
			args = append(args, "-input=false")
		}
	case CommandDestroy:
		if opts.DestroyAutoApprove {
			args = append(args, "-auto-approve")
		}
		if opts.NonInteractive {
			args = append(args, "-input=false")
		}
	}

	return args
}
