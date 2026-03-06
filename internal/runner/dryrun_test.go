package runner

import (
	"context"
	"strings"
	"testing"

	"github.com/Ops-Talks/cultivator/internal/discovery"
)

func Test_Run_DryRun(t *testing.T) {
	t.Parallel()

	fake := &fakeExecutor{}
	r := New().WithExecutor(fake)

	modules := []discovery.Module{
		{Path: "/tmp/app1"},
	}

	opts := Options{
		DryRun:         true,
		NonInteractive: true,
	}

	results, err := r.Run(context.Background(), CommandPlan, modules, opts)
	if err != nil {
		t.Fatalf("Run() unexpected error: %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}

	res := results[0]
	if !strings.HasPrefix(res.Stdout, "Dry run: terragrunt plan") {
		t.Errorf("unexpected stdout: %q", res.Stdout)
	}
	if !strings.Contains(res.Stdout, "-input=false") {
		t.Errorf("expected stdout to contain flags, got %q", res.Stdout)
	}
	if res.Duration <= 0 {
		t.Errorf("res.Duration = %v, want > 0", res.Duration)
	}

	fake.mu.Lock()
	defer fake.mu.Unlock()
	if len(fake.calls) != 0 {
		t.Errorf("expected 0 executor calls in dry-run mode, got %d", len(fake.calls))
	}
}
