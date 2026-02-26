package runner

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"sync"

	"github.com/Ops-Talks/cultivator/internal/discovery"
	"github.com/Ops-Talks/cultivator/internal/logging"
)

const (
	CommandPlan    = "plan"
	CommandApply   = "apply"
	CommandDestroy = "destroy"
)

type Result struct {
	Module   discovery.Module
	Command  string
	Stdout   string
	Stderr   string
	ExitCode int
	Error    error
}

type Options struct {
	Parallelism        int
	NonInteractive     bool
	PlanDestroy        bool
	ApplyAutoApprove   bool
	DestroyAutoApprove bool
}

type Executor interface {
	Run(ctx context.Context, workDir string, command string, args []string, extraEnv []string) (string, string, int, error)
}

type DefaultExecutor struct{}

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

type Runner struct {
	logger   *logging.Logger
	executor Executor
}

func New(logger *logging.Logger) *Runner {
	return &Runner{
		logger:   logger,
		executor: &DefaultExecutor{},
	}
}

func (r *Runner) WithExecutor(executor Executor) *Runner {
	r.executor = executor
	return r
}

func (r *Runner) Run(ctx context.Context, command string, modules []discovery.Module, opts Options) ([]Result, error) {
	if opts.Parallelism < 1 {
		opts.Parallelism = 1
	}

	results := make([]Result, len(modules))
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, opts.Parallelism)

	for i, module := range modules {
		wg.Add(1)
		go func(idx int, mod discovery.Module) {
			defer wg.Done()

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

			if err == nil && exitCode == 0 {
				r.logger.Info(fmt.Sprintf("%s %s", command, mod.Path), logging.Fields{
					"exit_code": exitCode,
				})
			} else {
				r.logger.Error(fmt.Sprintf("%s %s failed", command, mod.Path), logging.Fields{
					"exit_code": exitCode,
					"error":     err,
				})
			}
		}(i, module)
	}

	wg.Wait()
	return results, nil
}

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
