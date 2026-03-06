package runner

import (
	"context"
	"path/filepath"
	"strings"
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

func Test_BuildArgs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		command string
		opts    Options
		want    []string
	}{
		{
			name:    "plan: no flags",
			command: CommandPlan,
			opts:    Options{},
			want:    []string{"plan", "-no-color"},
		},
		{
			name:    "plan: destroy flag",
			command: CommandPlan,
			opts:    Options{PlanDestroy: true},
			want:    []string{"plan", "-no-color", "-destroy"},
		},
		{
			name:    "plan: non-interactive",
			command: CommandPlan,
			opts:    Options{NonInteractive: true},
			want:    []string{"plan", "-no-color", "-input=false"},
		},
		{
			name:    "apply: no flags",
			command: CommandApply,
			opts:    Options{},
			want:    []string{"apply", "-no-color"},
		},
		{
			name:    "apply: auto-approve",
			command: CommandApply,
			opts:    Options{ApplyAutoApprove: true},
			want:    []string{"apply", "-no-color", "-auto-approve"},
		},
		{
			name:    "destroy: no flags",
			command: CommandDestroy,
			opts:    Options{},
			want:    []string{"destroy", "-no-color"},
		},
		{
			name:    "destroy: auto-approve",
			command: CommandDestroy,
			opts:    Options{DestroyAutoApprove: true},
			want:    []string{"destroy", "-no-color", "-auto-approve"},
		},
		{
			name:    "dry-run: plan",
			command: CommandPlan,
			opts:    Options{DryRun: true},
			want:    []string{"plan", "-no-color"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := BuildArgs(tc.command, tc.opts)
			compareSlices(t, tc.want, got)
		})
	}
}

func Test_Run(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		modules     []discovery.Module
		parallelism int
		command     string
		validate    func(*testing.T, []Result, *fakeExecutor)
	}{
		{
			name: "parallel execution and order preservation",
			modules: []discovery.Module{
				{Path: filepath.Join("/tmp", "alpha")},
				{Path: filepath.Join("/tmp", "beta")},
				{Path: filepath.Join("/tmp", "gamma")},
			},
			parallelism: 3,
			command:     CommandPlan,
			validate: func(t *testing.T, results []Result, fake *fakeExecutor) {
				if len(results) != 3 {
					t.Fatalf("got %d results, want 3", len(results))
				}
				// Verify order preservation
				paths := []string{"alpha", "beta", "gamma"}
				for i, want := range paths {
					if !strings.Contains(results[i].Module.Path, want) {
						t.Errorf("results[%d].Module.Path = %q, want to contain %q", i, results[i].Module.Path, want)
					}
				}
				fake.mu.Lock()
				defer fake.mu.Unlock()
				if len(fake.calls) != 3 {
					t.Errorf("expected 3 executor calls, got %d", len(fake.calls))
				}
			},
		},
		{
			name: "default parallelism (0 -> 1)",
			modules: []discovery.Module{
				{Path: filepath.Join("/tmp", "app1")},
			},
			parallelism: 0,
			command:     CommandPlan,
			validate: func(t *testing.T, results []Result, fake *fakeExecutor) {
				if len(results) != 1 {
					t.Fatalf("got %d results, want 1", len(results))
				}
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			fake := &fakeExecutor{}
			r := New().WithExecutor(fake)
			opts := Options{Parallelism: tc.parallelism}

			results, err := r.Run(context.Background(), tc.command, tc.modules, opts)
			if err != nil {
				t.Fatalf("Run() unexpected error: %v", err)
			}
			tc.validate(t, results, fake)
		})
	}
}

func Test_DefaultExecutor_Run(t *testing.T) {
	t.Parallel()

	ex := &DefaultExecutor{}

	tests := []struct {
		name     string
		script   string
		validate func(*testing.T, string, string, int, error)
	}{
		{
			name:   "stderr is merged into stdout",
			script: "echo hello >&2",
			validate: func(t *testing.T, stdout, stderr string, exitCode int, err error) {
				if stderr != "" {
					t.Errorf("stderr should be empty, got %q", stderr)
				}
				if !strings.Contains(stdout, "hello") {
					t.Errorf("stdout should contain 'hello', got %q", stdout)
				}
			},
		},
		{
			name:   "output is combined chronologically",
			script: "printf 'line1\\n'; sleep 0.1; printf 'line2\\n' >&2; sleep 0.1; printf 'line3\\n'",
			validate: func(t *testing.T, stdout, stderr string, exitCode int, err error) {
				lines := []string{"line1", "line2", "line3"}
				prev := -1
				for _, line := range lines {
					pos := strings.Index(stdout, line)
					if pos == -1 {
						t.Errorf("output missing %q", line)
						continue
					}
					if pos < prev {
						t.Errorf("%q appeared before previous line", line)
					}
					prev = pos
				}
			},
		},
		{
			name:   "non-zero exit code",
			script: "exit 1",
			validate: func(t *testing.T, stdout, stderr string, exitCode int, err error) {
				if exitCode != 1 {
					t.Errorf("expected exit code 1, got %d", exitCode)
				}
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			stdout, stderr, exitCode, err := ex.Run(
				context.Background(),
				t.TempDir(),
				"sh",
				[]string{"-c", tc.script},
				nil,
			)
			tc.validate(t, stdout, stderr, exitCode, err)
		})
	}
}

func compareSlices(t *testing.T, want, got []string) {
	t.Helper()
	if len(want) != len(got) {
		t.Fatalf("slice length mismatch: want %d (%v), got %d (%v)", len(want), want, len(got), got)
	}
	for i := range want {
		if want[i] != got[i] {
			t.Errorf("slice element mismatch at index %d: want %q, got %q", i, want[i], got[i])
		}
	}
}
