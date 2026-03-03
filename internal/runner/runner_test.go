package runner

import (
	"context"
	"path/filepath"
	"sync"
	"testing"

	"github.com/Ops-Talks/cultivator/internal/discovery"
)

type fakeExecutor struct {
	mu    sync.Mutex
	calls []call
}

type call struct {
	workDir string
	args    []string
}

func (f *fakeExecutor) Run(_ context.Context, workDir string, _ string, args []string, _ []string) (string, string, int, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls = append(f.calls, call{workDir: workDir, args: args})
	return "ok", "", 0, nil
}

func Test_BuildArgs_plan(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		opts Options
		want []string
	}{
		{
			name: "no flags",
			opts: Options{},
			want: []string{"plan", "-no-color"},
		},
		{
			name: "destroy flag",
			opts: Options{PlanDestroy: true},
			want: []string{"plan", "-no-color", "-destroy"},
		},
		{
			name: "non-interactive",
			opts: Options{NonInteractive: true},
			want: []string{"plan", "-no-color", "-input=false"},
		},
		{
			name: "destroy and non-interactive",
			opts: Options{PlanDestroy: true, NonInteractive: true},
			want: []string{"plan", "-no-color", "-destroy", "-input=false"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assertArgs(t, tc.want, BuildArgs(CommandPlan, tc.opts))
		})
	}
}

func Test_BuildArgs_apply(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		opts Options
		want []string
	}{
		{
			name: "no flags",
			opts: Options{},
			want: []string{"apply", "-no-color"},
		},
		{
			name: "auto-approve",
			opts: Options{ApplyAutoApprove: true},
			want: []string{"apply", "-no-color", "-auto-approve"},
		},
		{
			name: "auto-approve and non-interactive",
			opts: Options{ApplyAutoApprove: true, NonInteractive: true},
			want: []string{"apply", "-no-color", "-auto-approve", "-input=false"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assertArgs(t, tc.want, BuildArgs(CommandApply, tc.opts))
		})
	}
}

func Test_BuildArgs_destroy(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		opts Options
		want []string
	}{
		{
			name: "no flags",
			opts: Options{},
			want: []string{"destroy", "-no-color"},
		},
		{
			name: "auto-approve",
			opts: Options{DestroyAutoApprove: true},
			want: []string{"destroy", "-no-color", "-auto-approve"},
		},
		{
			name: "auto-approve and non-interactive",
			opts: Options{DestroyAutoApprove: true, NonInteractive: true},
			want: []string{"destroy", "-no-color", "-auto-approve", "-input=false"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assertArgs(t, tc.want, BuildArgs(CommandDestroy, tc.opts))
		})
	}
}

func Test_Run_parallelExecution(t *testing.T) {
	t.Parallel()

	fake := &fakeExecutor{}
	r := New().WithExecutor(fake)

	modules := []discovery.Module{
		{Path: filepath.Join("/tmp", "app1")},
		{Path: filepath.Join("/tmp", "app2")},
	}

	results, err := r.Run(context.Background(), CommandPlan, modules, Options{Parallelism: 2})
	if err != nil {
		t.Fatalf("Run() unexpected error: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("Run() returned %d results, want 2", len(results))
	}

	fake.mu.Lock()
	defer fake.mu.Unlock()
	if len(fake.calls) != 2 {
		t.Fatalf("expected 2 executor calls, got %d", len(fake.calls))
	}
}

func Test_Run_defaultParallelism(t *testing.T) {
	t.Parallel()

	fake := &fakeExecutor{}
	r := New().WithExecutor(fake)

	modules := []discovery.Module{
		{Path: filepath.Join("/tmp", "app1")},
	}

	// Parallelism 0 should default to 1 without panic.
	results, err := r.Run(context.Background(), CommandPlan, modules, Options{Parallelism: 0})
	if err != nil {
		t.Fatalf("Run() unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("Run() returned %d results, want 1", len(results))
	}
}

func assertArgs(t *testing.T, expected, actual []string) {
	t.Helper()
	if len(expected) != len(actual) {
		t.Fatalf("expected %d args %v, got %d args %v", len(expected), expected, len(actual), actual)
	}
	for i, want := range expected {
		if actual[i] != want {
			t.Fatalf("arg[%d]: want %q, got %q", i, want, actual[i])
		}
	}
}
