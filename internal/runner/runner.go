// Package runner executes Terragrunt commands across multiple modules in parallel.
package runner

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/Ops-Talks/cultivator/internal/dag"
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
	Duration time.Duration
}

// Options configures the behavior of a Runner.Run call.
type Options struct {
	Parallelism        int
	NonInteractive     bool
	DryRun             bool
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
// stdout and stderr are merged into a single chronologically-ordered stream via
// CombinedOutput, preserving the interleaved output produced by Terragrunt and
// the underlying IaC tool. The returned stderr string is always empty.
func (de *DefaultExecutor) Run(ctx context.Context, workDir string, command string, args []string, extraEnv []string) (string, string, int, error) {
	// #nosec G204 -- command is controlled by the caller and expected to be terragrunt
	cmd := exec.CommandContext(ctx, command, args...)
	cmd.Dir = workDir

	if len(extraEnv) > 0 {
		cmd.Env = append(cmd.Environ(), extraEnv...)
	}

	combined, err := cmd.CombinedOutput()
	exitCode := 0
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		exitCode = exitErr.ExitCode()
	}

	return string(combined), "", exitCode, err
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

// Run executes the given Terragrunt command for each module according to their dependencies,
// bounded by opts.Parallelism. It always returns the full result slice; individual errors
// are captured inside each Result.
func (r *Runner) Run(ctx context.Context, command string, modules []discovery.Module, opts Options) ([]Result, error) {
	if opts.Parallelism < 1 {
		opts.Parallelism = 1
	}

	// Build the dependency graph
	g := dag.New()
	pathMap := make(map[string]int) // path -> index in modules slice
	for i, mod := range modules {
		g.AddNode(mod.Path)
		pathMap[mod.Path] = i
	}

	for _, mod := range modules {
		for _, depPath := range mod.Dependencies {
			// Only add dependencies that are part of the current execution set
			if _, ok := pathMap[depPath]; ok {
				g.AddEdge(mod.Path, depPath)
			}
		}
	}

	// Detect cycles
	_, err := g.TopologicalSort()
	if err != nil {
		return nil, fmt.Errorf("dependency error: %w", err)
	}

	results := make([]Result, len(modules))
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, opts.Parallelism)

	// Keep track of finished modules to signal dependents
	finished := make(map[string]chan struct{})
	for _, mod := range modules {
		finished[mod.Path] = make(chan struct{})
	}

	for i, module := range modules {
		idx, mod := i, module
		wg.Go(func() {
			// Wait for all dependencies to finish
			deps := g.GetDependencies(mod.Path)
			for _, dep := range deps {
				select {
				case <-finished[dep]:
					// Dependency finished
				case <-ctx.Done():
					results[idx] = Result{
						Module:  mod,
						Command: command,
						Error:   ctx.Err(),
					}
					close(finished[mod.Path])
					return
				}
			}

			semaphore <- struct{}{}
			defer func() {
				<-semaphore
				close(finished[mod.Path])
			}()

			args := BuildArgs(command, opts)

			var stdout string
			var stderr string
			var exitCode int
			var execErr error

			start := time.Now()
			if opts.DryRun {
				stdout = "Dry run: terragrunt " + strings.Join(args, " ")
				exitCode = 0
			} else {
				stdout, stderr, exitCode, execErr = r.executor.Run(ctx, mod.Path, "terragrunt", args, []string{})
			}
			duration := time.Since(start)

			results[idx] = Result{
				Module:   mod,
				Command:  command,
				Stdout:   stdout,
				Stderr:   stderr,
				ExitCode: exitCode,
				Error:    execErr,
				Duration: duration,
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
