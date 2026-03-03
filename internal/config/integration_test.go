package config

import (
	"context"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/Ops-Talks/cultivator/internal/discovery"
	"github.com/Ops-Talks/cultivator/internal/runner"
)

func TestIntegration_ConfigDiscoveryRunner(t *testing.T) {
	t.Parallel()

	testDataDir := getIntTestDataDir(t)
	cfgFile := filepath.Join(testDataDir, "terragrunt-structure", ".cultivator.yaml")

	cfg, _, found, err := LoadFile(cfgFile)
	if err != nil {
		t.Fatalf("LoadFile failed: %v", err)
	}
	if !found {
		t.Fatalf("config file not found")
	}

	cfg = MergeConfig(DefaultConfig(), cfg)
	if err := Validate(cfg); err != nil {
		t.Fatalf("config validation failed: %v", err)
	}

	rootDir := filepath.Join(testDataDir, "terragrunt-structure")
	modules, err := discovery.Discover(rootDir, discovery.Options{})
	if err != nil {
		t.Fatalf("discovery failed: %v", err)
	}

	if len(modules) < 3 {
		t.Fatalf("expected at least 3 modules, got %d", len(modules))
	}

	r := runner.New()

	opts := runner.Options{Parallelism: cfg.Parallelism}
	results, err := r.Run(context.Background(), runner.CommandPlan, modules, opts)
	if err != nil {
		t.Fatalf("run failed: %v", err)
	}

	if len(results) != len(modules) {
		t.Fatalf("expected %d results, got %d", len(modules), len(results))
	}
}

func TestIntegration_DiscoveryWithFilters(t *testing.T) {
	t.Parallel()

	testDataDir := getIntTestDataDir(t)
	rootDir := filepath.Join(testDataDir, "terragrunt-structure")

	modules, err := discovery.Discover(rootDir, discovery.Options{Env: "prod"})
	if err != nil {
		t.Fatalf("discover failed: %v", err)
	}
	if len(modules) != 2 {
		t.Fatalf("expected 2 prod modules, got %d", len(modules))
	}

	modules, err = discovery.Discover(rootDir, discovery.Options{Tags: []string{"app"}})
	if err != nil {
		t.Fatalf("discover with tags failed: %v", err)
	}
	if len(modules) != 2 {
		t.Fatalf("expected 2 modules with app tag, got %d", len(modules))
	}
}

func TestIntegration_ConfigOverrides(t *testing.T) {
	t.Parallel()

	testDataDir := getIntTestDataDir(t)
	cfgFile := filepath.Join(testDataDir, "terragrunt-structure", ".cultivator.yaml")

	cfg, _, _, err := LoadFile(cfgFile)
	if err != nil {
		t.Fatalf("LoadFile failed: %v", err)
	}

	parallelism := 4
	nonInteractive := true

	cfg = ApplyOverrides(cfg, Overrides{
		Parallelism:    &parallelism,
		NonInteractive: &nonInteractive,
	})

	if cfg.Parallelism != 4 || !cfg.NonInteractive {
		t.Error("overrides not applied correctly")
	}

	if err := Validate(cfg); err != nil {
		t.Fatalf("validation failed: %v", err)
	}
}

func getIntTestDataDir(t testing.TB) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("could not get file path")
	}
	projectRoot := filepath.Join(filepath.Dir(file), "..", "..")
	return filepath.Join(projectRoot, "testdata")
}
