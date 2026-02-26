package config

import (
	"os"
	"path/filepath"
	"testing"
)

const jsonFormat = "json"

func TestLoadFileAndMerge(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, ".cultivator.yaml")
	content := []byte("root: live\nparallelism: 3\noutput_format: json\nnon_interactive: true\nplan:\n  destroy: true\napply:\n  auto_approve: true\n")
	if err := os.WriteFile(configPath, content, 0o600); err != nil {
		t.Fatalf("write config file: %v", err)
	}

	cfg, _, found, err := LoadFile(configPath)
	if err != nil {
		t.Fatalf("load config file: %v", err)
	}
	if !found {
		t.Fatalf("expected config file to be found")
	}

	merged := MergeConfig(DefaultConfig(), cfg)
	if merged.Root != "live" {
		t.Fatalf("expected root to be live, got %q", merged.Root)
	}
	if merged.Parallelism != 3 {
		t.Fatalf("expected parallelism 3, got %d", merged.Parallelism)
	}
	if merged.OutputFormat != jsonFormat {
		t.Fatalf("expected output format json, got %q", merged.OutputFormat)
	}
	if !merged.NonInteractive {
		t.Fatalf("expected non-interactive true")
	}
}

func TestApplyOverrides(t *testing.T) {
	t.Parallel()

	cfg := DefaultConfig()
	root := "envs"
	format := jsonFormat
	parallelism := 5
	nonInteractive := true

	overrides := Overrides{
		Root:           &root,
		OutputFormat:   &format,
		Parallelism:    &parallelism,
		NonInteractive: &nonInteractive,
		Include:        []string{"prod"},
		IncludeSet:     true,
	}

	cfg = ApplyOverrides(cfg, overrides)
	if cfg.Root != "envs" {
		t.Fatalf("expected root envs, got %q", cfg.Root)
	}
	if cfg.OutputFormat != "json" {
		t.Fatalf("expected output format json, got %q", cfg.OutputFormat)
	}
	if cfg.Parallelism != 5 {
		t.Fatalf("expected parallelism 5, got %d", cfg.Parallelism)
	}
	if !cfg.NonInteractive {
		t.Fatalf("expected non-interactive true")
	}
	if len(cfg.Include) != 1 || cfg.Include[0] != "prod" {
		t.Fatalf("expected include to be set")
	}
}

func TestValidate(t *testing.T) {
	t.Parallel()

	cfg := DefaultConfig()
	cfg.OutputFormat = "xml"
	if err := Validate(cfg); err == nil {
		t.Fatalf("expected invalid output format error")
	}

	cfg = DefaultConfig()
	cfg.Parallelism = 0
	if err := Validate(cfg); err == nil {
		t.Fatalf("expected invalid parallelism error")
	}
}
