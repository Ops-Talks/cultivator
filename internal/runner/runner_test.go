package runner

import (
	"bytes"
	"context"
	"path/filepath"
	"sync"
	"testing"

	"github.com/Ops-Talks/cultivator/internal/discovery"
	"github.com/Ops-Talks/cultivator/internal/logging"
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

func TestBuildArgs(t *testing.T) {
	t.Parallel()

	args := BuildArgs(CommandPlan, Options{PlanDestroy: true, NonInteractive: true})
	expected := []string{"plan", "-no-color", "-destroy", "-input=false"}
	assertArgs(t, expected, args)

	args = BuildArgs(CommandApply, Options{ApplyAutoApprove: true})
	expected = []string{"apply", "-no-color", "-auto-approve"}
	assertArgs(t, expected, args)

	args = BuildArgs(CommandDestroy, Options{DestroyAutoApprove: true})
	expected = []string{"destroy", "-no-color", "-auto-approve"}
	assertArgs(t, expected, args)
}

func TestRunnerRun(t *testing.T) {
	t.Parallel()

	logger := logging.New("text", &bytes.Buffer{}, &bytes.Buffer{})
	fake := &fakeExecutor{}
	runner := New(logger).WithExecutor(fake)

	modules := []discovery.Module{
		{Path: filepath.Join("/tmp", "app1")},
		{Path: filepath.Join("/tmp", "app2")},
	}

	results, err := runner.Run(context.Background(), CommandPlan, modules, Options{Parallelism: 2})
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}

	fake.mu.Lock()
	defer fake.mu.Unlock()
	if len(fake.calls) != 2 {
		t.Fatalf("expected 2 executor calls, got %d", len(fake.calls))
	}
}

func assertArgs(t *testing.T, expected []string, actual []string) {
	t.Helper()
	if len(expected) != len(actual) {
		t.Fatalf("expected %d args, got %d", len(expected), len(actual))
	}
	for i, value := range expected {
		if actual[i] != value {
			t.Fatalf("expected arg %d to be %q, got %q", i, value, actual[i])
		}
	}
}
