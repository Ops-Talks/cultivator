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

	cfg = DefaultConfig()
	if err := Validate(cfg); err != nil {
		t.Fatalf("expected default config to be valid, got: %v", err)
	}
}

func TestLoadEnv(t *testing.T) {
	const prefix = "CULTIVATOR"

	t.Setenv(prefix+"_ROOT", "environments")
	t.Setenv(prefix+"_ENV", "staging")
	t.Setenv(prefix+"_INCLUDE", "app,db")
	t.Setenv(prefix+"_EXCLUDE", "legacy;tmp")
	t.Setenv(prefix+"_TAGS", "frontend,backend")
	t.Setenv(prefix+"_PARALLELISM", "8")
	t.Setenv(prefix+"_OUTPUT_FORMAT", jsonFormat)
	t.Setenv(prefix+"_NON_INTERACTIVE", "true")

	cfg := LoadEnv(prefix)

	if cfg.Root != "environments" {
		t.Errorf("expected root environments, got %q", cfg.Root)
	}
	if cfg.Env != "staging" {
		t.Errorf("expected env staging, got %q", cfg.Env)
	}
	if len(cfg.Include) != 2 || cfg.Include[0] != "app" || cfg.Include[1] != "db" {
		t.Errorf("unexpected include: %v", cfg.Include)
	}
	if len(cfg.Exclude) != 2 || cfg.Exclude[0] != "legacy" || cfg.Exclude[1] != "tmp" {
		t.Errorf("unexpected exclude: %v", cfg.Exclude)
	}
	if len(cfg.Tags) != 2 || cfg.Tags[0] != "frontend" || cfg.Tags[1] != "backend" {
		t.Errorf("unexpected tags: %v", cfg.Tags)
	}
	if cfg.Parallelism != 8 {
		t.Errorf("expected parallelism 8, got %d", cfg.Parallelism)
	}
	if cfg.OutputFormat != jsonFormat {
		t.Errorf("expected output format json, got %q", cfg.OutputFormat)
	}
	if !cfg.NonInteractive {
		t.Error("expected non-interactive true")
	}
}

func TestLoadEnv_invalidParallelism(t *testing.T) {
	const prefix = "CULTIVATOR"
	t.Setenv(prefix+"_PARALLELISM", "not-a-number")

	cfg := LoadEnv(prefix)

	if cfg.Parallelism < 1 {
		t.Errorf("expected parallelism to fallback to default (>= 1), got %d", cfg.Parallelism)
	}
}

func TestLoadFile_missingFile(t *testing.T) {
	t.Parallel()

	cfg, _, found, err := LoadFile("/nonexistent/path/.cultivator.yaml")
	if err != nil {
		t.Fatalf("expected no error for missing file, got: %v", err)
	}
	if found {
		t.Error("expected found=false for missing file")
	}
	if cfg.Parallelism < 1 {
		t.Errorf("expected default parallelism, got %d", cfg.Parallelism)
	}
}

func TestLoadFile_emptyPath(t *testing.T) {
	t.Parallel()

	_, _, found, err := LoadFile("")
	if err != nil {
		t.Fatalf("expected no error for empty path, got: %v", err)
	}
	if found {
		t.Error("expected found=false for empty path")
	}
}

func TestMergeConfig_extraFields(t *testing.T) {
	t.Parallel()

	base := DefaultConfig()
	override := DefaultConfig()
	override.Plan = map[string]interface{}{"destroy": true}
	override.Apply = map[string]interface{}{"auto_approve": true}

	merged := MergeConfig(base, override)

	if v, ok := merged.Plan["destroy"]; !ok || v != true {
		t.Errorf("expected plan.destroy=true, got %v", merged.Plan)
	}
	if v, ok := merged.Apply["auto_approve"]; !ok || v != true {
		t.Errorf("expected apply.auto_approve=true, got %v", merged.Apply)
	}
}
