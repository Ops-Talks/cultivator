package config

import (
	"os"
	"path/filepath"
	"testing"
)

// Test ApplyOverrides with partial overrides
func TestApplyOverrides_Partial(t *testing.T) {
	t.Parallel()

	cfg := DefaultConfig()
	cfg.Exclude = []string{"test"}
	cfg.Tags = []string{"prod"}

	// Only override one field
	overrides := Overrides{Exclude: []string{"dev"}, ExcludeSet: true}
	cfg = ApplyOverrides(cfg, overrides)

	if len(cfg.Exclude) != 1 || cfg.Exclude[0] != "dev" {
		t.Errorf("exclude not overridden, got %v", cfg.Exclude)
	}
	// Tags should remain unchanged
	if len(cfg.Tags) != 1 || cfg.Tags[0] != prodEnv {
		t.Errorf("tags should not change, got %v", cfg.Tags)
	}
}

// Test ApplyOverrides with boolean pointers
func TestApplyOverrides_Booleans(t *testing.T) {
	t.Parallel()

	trueBool := true

	cfg := DefaultConfig()
	overrides := Overrides{
		PlanDestroy:        &trueBool,
		ApplyAutoApprove:   &trueBool,
		DestroyAutoApprove: &trueBool,
	}

	cfg = ApplyOverrides(cfg, overrides)

	if !cfg.Plan.Destroy {
		t.Error("plan.Destroy should be true")
	}
	if !cfg.Apply.AutoApprove {
		t.Error("apply.AutoApprove should be true")
	}
	if !cfg.Destroy.AutoApprove {
		t.Error("destroy.AutoApprove should be true")
	}
}

// Test DefaultConfig returns correct values
func TestDefaultConfig_Values(t *testing.T) {
	t.Parallel()

	cfg := DefaultConfig()

	if cfg.Root == "" {
		t.Error("root should have default value")
	}
	if cfg.Parallelism < 1 {
		t.Errorf("parallelism should be at least 1, got %d", cfg.Parallelism)
	}
}

// Test MergeConfig basic merge
func TestMergeConfig_Basic(t *testing.T) {
	t.Parallel()

	base := Config{
		Root:    "/root",
		Exclude: []string{"dev"},
		Tags:    []string{"prod"},
	}

	override := Config{
		Root:    "/new",
		Exclude: []string{"test"},
	}

	result := MergeConfig(base, override)

	if result.Root != "/new" {
		t.Errorf("root should be overridden, got %q", result.Root)
	}
	if len(result.Exclude) != 1 || result.Exclude[0] != "test" {
		t.Errorf("exclude should be merged, got %v", result.Exclude)
	}
}

// Test Validate with invalid parallelism
func TestValidate_ParallelismZero(t *testing.T) {
	t.Parallel()

	cfg := DefaultConfig()
	cfg.Parallelism = 0

	err := Validate(cfg)
	if err == nil {
		t.Error("validate should error for parallelism=0")
	}
}

// Test LoadFile with valid YAML
func TestLoadFile_Valid(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "config.yml")
	content := []byte(`root: /test
parallelism: 4
exclude: [dev, test]`)

	if err := os.WriteFile(filePath, content, 0o600); err != nil {
		t.Fatalf("write file: %v", err)
	}

	cfg, extra, found, err := LoadFile(filePath)
	if err != nil {
		t.Errorf("load file failed: %v", err)
	}
	if !found {
		t.Error("should have found config")
	}
	if cfg.Root != "/test" {
		t.Errorf("root should be /test, got %q", cfg.Root)
	}
	if cfg.Parallelism != 4 {
		t.Errorf("parallelism should be 4, got %d", cfg.Parallelism)
	}
	if extra == nil {
		t.Error("extra should not be nil")
	}
}

// Test ParseInt valid cases
func TestParseInt_Valid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input    string
		expected int
	}{
		{"5", 5},
		{"0", 0},
		{"100", 100},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, err := ParseInt(tt.input)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, result)
			}
		})
	}
}

// Test ParseInt invalid cases
func TestParseInt_Invalid(t *testing.T) {
	t.Parallel()

	tests := []string{"abc", "3.14"}
	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			result, err := ParseInt(input)
			if err == nil {
				t.Errorf("expected error for invalid input %q, got %d", input, result)
			}
		})
	}
}

// Test ParseBool true cases
func TestParseBool_True(t *testing.T) {
	t.Parallel()

	trueValues := []string{"true", "True", "TRUE", "yes", "Yes", "YES", "1"}
	for _, val := range trueValues {
		t.Run(val, func(t *testing.T) {
			result := ParseBool(val)
			if !result {
				t.Errorf("%q should be parsed as true", val)
			}
		})
	}
}

// Test ParseBool false cases
func TestParseBool_False(t *testing.T) {
	t.Parallel()

	falseValues := []string{"false", "False", "FALSE", "no", "No", "NO", "0", ""}
	for _, val := range falseValues {
		t.Run(val, func(t *testing.T) {
			result := ParseBool(val)
			if result {
				t.Errorf("%q should be parsed as false", val)
			}
		})
	}
}

// Test LoadEnv with environment variables
func TestLoadEnv_WithVars(t *testing.T) {
	// Cannot use t.Parallel() with t.Setenv()

	// Set environment variables
	t.Setenv("TEST_ROOT", "/env/root")
	t.Setenv("TEST_PARALLELISM", "8")

	cfg := LoadEnv("TEST")
	if cfg.Root != "/env/root" {
		t.Errorf("root from env should be /env/root, got %q", cfg.Root)
	}
	if cfg.Parallelism != 8 {
		t.Errorf("parallelism from env should be 8, got %d", cfg.Parallelism)
	}
}

// Test LoadFile with missing file (returns defaults)
func TestLoadFile_Missing(t *testing.T) {
	t.Parallel()

	cfg, _, found, err := LoadFile("/nonexistent/path/config.yml")
	if err != nil {
		t.Errorf("should not error for missing file: %v", err)
	}
	if found {
		t.Error("should not find nonexistent file")
	}
	if cfg.Root == "" {
		t.Error("should return default config")
	}
}

func TestLoadFile_EmptyPath(t *testing.T) {
	t.Parallel()

	cfg, extra, found, err := LoadFile("   ")
	if err != nil {
		t.Fatalf("expected no error for empty path, got %v", err)
	}
	if found {
		t.Fatal("expected found=false for empty path")
	}
	if extra == nil {
		t.Fatal("expected non-nil extra map")
	}
	if cfg.Root == "" {
		t.Fatal("expected default config for empty path")
	}
}

func TestLoadFile_ReadError(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	_, _, found, err := LoadFile(dir)
	if err == nil {
		t.Fatal("expected read error when path points to directory")
	}
	if found {
		t.Fatal("expected found=false on read error")
	}
}

func TestLoadFile_ParseAndDecodeErrors(t *testing.T) {
	t.Parallel()

	t.Run("parse error", func(t *testing.T) {
		t.Parallel()

		tempDir := t.TempDir()
		filePath := filepath.Join(tempDir, "invalid-yaml.yml")
		content := []byte("root: [")

		if err := os.WriteFile(filePath, content, 0o600); err != nil {
			t.Fatalf("write file: %v", err)
		}

		_, _, _, err := LoadFile(filePath)
		if err == nil {
			t.Fatal("expected parse error")
		}
	})

	t.Run("decode error", func(t *testing.T) {
		t.Parallel()

		tempDir := t.TempDir()
		filePath := filepath.Join(tempDir, "decode-error.yml")
		content := []byte("parallelism: nope")

		if err := os.WriteFile(filePath, content, 0o600); err != nil {
			t.Fatalf("write file: %v", err)
		}

		_, _, _, err := LoadFile(filePath)
		if err == nil {
			t.Fatal("expected decode error")
		}
	})
}

func TestLoadFile_ExtraKeys(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "extra.yml")
	content := []byte("root: /tmp\ncustom_flag: true\ncustom_number: 10\n")

	if err := os.WriteFile(filePath, content, 0o600); err != nil {
		t.Fatalf("write file: %v", err)
	}

	_, extra, found, err := LoadFile(filePath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !found {
		t.Fatal("expected found=true")
	}
	if len(extra) != 2 {
		t.Fatalf("expected 2 extra keys, got %d", len(extra))
	}
	if extra["custom_flag"] != true {
		t.Fatalf("expected custom_flag=true, got %#v", extra["custom_flag"])
	}
}

func TestMergeConfig_Guards(t *testing.T) {
	t.Parallel()

	const baseRoot = "/base"
	base := DefaultConfig()
	base.Root = baseRoot
	base.Env = prodEnv
	base.Parallelism = 4

	override := Config{
		Root:        ".",
		Parallelism: 4,
	}

	result := MergeConfig(base, override)
	if result.Root != "/base" {
		t.Fatalf("expected root to remain /base, got %q", result.Root)
	}
	if result.Parallelism != 4 {
		t.Fatalf("expected parallelism to remain 4, got %d", result.Parallelism)
	}
}

func TestValidate_RootBranches(t *testing.T) {
	t.Parallel()

	t.Run("empty root", func(t *testing.T) {
		t.Parallel()

		cfg := DefaultConfig()
		cfg.Root = ""
		if err := Validate(cfg); err == nil {
			t.Fatal("expected error for empty root")
		}
	})

	t.Run("relative root missing", func(t *testing.T) {
		t.Parallel()

		cfg := DefaultConfig()
		cfg.Root = filepath.Join("does-not-exist", "for-test")
		if err := Validate(cfg); err == nil {
			t.Fatal("expected error for missing relative root")
		}
	})

	t.Run("absolute root does not stat", func(t *testing.T) {
		t.Parallel()

		cfg := DefaultConfig()
		cfg.Root = filepath.Join(t.TempDir(), "missing-abs-dir")
		if !filepath.IsAbs(cfg.Root) {
			t.Fatal("test root should be absolute")
		}
		if err := Validate(cfg); err != nil {
			t.Fatalf("expected no error for absolute root branch, got %v", err)
		}
	})
}

func TestMergeConfig_AllFields(t *testing.T) {
	t.Parallel()

	base := DefaultConfig()
	base.Root = "/base"
	base.Env = "dev"
	base.Include = []string{"old-inc"}
	base.Exclude = []string{"old-exc"}
	base.Tags = []string{"old-tag"}
	base.Parallelism = 2
	base.NonInteractive = false
	base.Plan.Destroy = false
	base.Apply.AutoApprove = false
	base.Destroy.AutoApprove = false

	override := Config{
		Root:           "/override",
		Env:            "prod",
		Include:        []string{"new-inc"},
		Exclude:        []string{"new-exc"},
		Tags:           []string{"new-tag"},
		Parallelism:    8,
		NonInteractive: true,
		Plan:           PlanConfig{Destroy: true},
		Apply:          ApplyConfig{AutoApprove: true},
		Destroy:        DestroyConfig{AutoApprove: true},
	}

	result := MergeConfig(base, override)

	if result.Root != "/override" || result.Env != "prod" {
		t.Fatalf("expected root/env overridden, got root=%q env=%q", result.Root, result.Env)
	}
	if len(result.Include) != 1 || result.Include[0] != "new-inc" {
		t.Fatalf("expected include override, got %v", result.Include)
	}
	if len(result.Exclude) != 1 || result.Exclude[0] != "new-exc" {
		t.Fatalf("expected exclude override, got %v", result.Exclude)
	}
	if len(result.Tags) != 1 || result.Tags[0] != "new-tag" {
		t.Fatalf("expected tags override, got %v", result.Tags)
	}
	if result.Parallelism != 8 || !result.NonInteractive {
		t.Fatalf("expected parallelism/nonInteractive overridden, got %d/%v", result.Parallelism, result.NonInteractive)
	}
	if !result.Plan.Destroy || !result.Apply.AutoApprove || !result.Destroy.AutoApprove {
		t.Fatal("expected plan/apply/destroy fields merged with true values")
	}
}

func TestApplyOverrides_SetsSubConfigFields(t *testing.T) {
	t.Parallel()

	trueVal := true
	cfg := Config{}
	override := Overrides{
		PlanDestroy:        &trueVal,
		ApplyAutoApprove:   &trueVal,
		DestroyAutoApprove: &trueVal,
	}

	result := ApplyOverrides(cfg, override)

	if !result.Plan.Destroy {
		t.Fatal("expected plan.Destroy to be true")
	}
	if !result.Apply.AutoApprove {
		t.Fatal("expected apply.AutoApprove to be true")
	}
	if !result.Destroy.AutoApprove {
		t.Fatal("expected destroy.AutoApprove to be true")
	}
}
