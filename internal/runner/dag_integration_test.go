package runner

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/Ops-Talks/cultivator/internal/discovery"
)

type timedExecutor struct {
	mu       sync.Mutex
	starts   map[string]time.Time
	finishes map[string]time.Time
}

func (e *timedExecutor) Run(_ context.Context, workDir string, _ string, _ []string, _ []string) (string, string, int, error) {
	e.mu.Lock()
	e.starts[workDir] = time.Now()
	e.mu.Unlock()

	time.Sleep(100 * time.Millisecond)

	e.mu.Lock()
	e.finishes[workDir] = time.Now()
	e.mu.Unlock()

	return "ok", "", 0, nil
}

func Test_Run_RespectsDependencies(t *testing.T) {
	t.Parallel()

	executor := &timedExecutor{
		starts:   make(map[string]time.Time),
		finishes: make(map[string]time.Time),
	}
	r := New().WithExecutor(executor)

	vpc := discovery.Module{Path: "/vpc"}
	db := discovery.Module{Path: "/db", Dependencies: []string{"/vpc"}}
	app := discovery.Module{Path: "/app", Dependencies: []string{"/db"}}

	modules := []discovery.Module{app, db, vpc}

	results, err := r.Run(context.Background(), CommandPlan, modules, Options{Parallelism: 3})
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	for _, res := range results {
		if res.Duration < 100*time.Millisecond {
			t.Errorf("module %s duration = %v, want >= 100ms", res.Module.Path, res.Duration)
		}
	}

	executor.mu.Lock()
	defer executor.mu.Unlock()

	// DB should start after VPC finishes
	if executor.starts["/db"].Before(executor.finishes["/vpc"]) {
		t.Errorf("DB started before VPC finished: VPC finish at %v, DB start at %v",
			executor.finishes["/vpc"], executor.starts["/db"])
	}

	// App should start after DB finishes
	if executor.starts["/app"].Before(executor.finishes["/db"]) {
		t.Errorf("App started before DB finished: DB finish at %v, App start at %v",
			executor.finishes["/db"], executor.starts["/app"])
	}
}
