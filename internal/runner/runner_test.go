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

// Test_Run_resultsPreserveDiscoveryOrder verifies that Run stores each result
// in the slot matching the module's position in the input slice, regardless of
// the order in which goroutines complete. This guarantees that logExecutionResults
// always prints output in a deterministic, discovery-ordered sequence.
func Test_Run_resultsPreserveDiscoveryOrder(t *testing.T) {
	t.Parallel()

	fake := &fakeExecutor{}
	r := New().WithExecutor(fake)

	modules := []discovery.Module{
		{Path: filepath.Join("/tmp", "alpha")},
		{Path: filepath.Join("/tmp", "beta")},
		{Path: filepath.Join("/tmp", "gamma")},
	}

	results, err := r.Run(context.Background(), CommandPlan, modules, Options{Parallelism: 3})
	if err != nil {
		t.Fatalf("Run() unexpected error: %v", err)
	}
	if len(results) != 3 {
		t.Fatalf("Run() returned %d results, want 3", len(results))
	}

	// Each result must carry the module that corresponds to its position in
	// the input slice, not the order in which the goroutine happened to finish.
	for i, mod := range modules {
		if results[i].Module.Path != mod.Path {
			t.Errorf("results[%d].Module.Path = %q, want %q", i, results[i].Module.Path, mod.Path)
		}
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

// Test_DefaultExecutor_stderrAlwaysEmpty verifies that DefaultExecutor never
// returns content in Stderr; all output (stdout + stderr of the subprocess) is
// merged and returned as Stdout via CombinedOutput.
func Test_DefaultExecutor_stderrAlwaysEmpty(t *testing.T) {
	t.Parallel()

	ex := &DefaultExecutor{}
	_, stderr, _, _ := ex.Run(
		context.Background(),
		t.TempDir(),
		"sh",
		[]string{"-c", "echo hello >&2"},
		nil,
	)
	if stderr != "" {
		t.Errorf("Stderr must always be empty (output merged into Stdout); got %q", stderr)
	}
}

// Test_DefaultExecutor_combinesOutputChronologically verifies that writes to
// stdout and stderr from the subprocess appear in the combined Stdout field in
// the order they were produced, not stdout-block-first then stderr-block-after.
func Test_DefaultExecutor_combinesOutputChronologically(t *testing.T) {
	t.Parallel()

	ex := &DefaultExecutor{}

	// The script writes alternating lines to stdout and stderr.
	// CombinedOutput must preserve the write order via a single shared pipe.
	script := "printf 'line1\\n'; printf 'line2\\n' >&2; printf 'line3\\n'; printf 'line4\\n' >&2"
	stdout, stderr, exitCode, err := ex.Run(
		context.Background(),
		t.TempDir(),
		"sh",
		[]string{"-c", script},
		nil,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d", exitCode)
	}
	if stderr != "" {
		t.Errorf("Stderr must always be empty; got %q", stderr)
	}

	lines := []string{"line1", "line2", "line3", "line4"}
	for _, line := range lines {
		if !strings.Contains(stdout, line) {
			t.Errorf("combined output missing %q, got %q", line, stdout)
		}
	}

	// Verify that lines appear in the order they were written.
	prev := 0
	for _, line := range lines {
		pos := strings.Index(stdout, line)
		if pos < prev {
			t.Errorf("output not in chronological order: %q appears at %d, before previous line at %d; full output: %q",
				line, pos, prev, stdout)
		}
		prev = pos
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
