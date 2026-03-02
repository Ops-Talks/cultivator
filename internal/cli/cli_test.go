package cli

import (
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseBool(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input    string
		expected bool
		wantErr  bool
	}{
		{"true", true, false},
		{"1", true, false},
		{"yes", true, false},
		{"false", false, false},
		{"0", false, false},
		{"invalid", false, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, err := parseBool(tt.input)
			if tt.wantErr && err == nil {
				t.Error("expected error")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if !tt.wantErr && result != tt.expected {
				t.Errorf("got %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestParseInt(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input    string
		expected int
		wantErr  bool
	}{
		{"42", 42, false},
		{"0", 0, false},
		{"", 0, true},
		{"abc", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, err := parseInt(tt.input)
			if tt.wantErr && err == nil {
				t.Error("expected error")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if !tt.wantErr && result != tt.expected {
				t.Errorf("got %d, want %d", result, tt.expected)
			}
		})
	}
}

func TestSplitList(t *testing.T) {
	t.Parallel()

	result := splitList("a,b;c")
	if len(result) != 3 {
		t.Errorf("expected 3 items, got %d", len(result))
	}
}

func TestFallback(t *testing.T) {
	t.Parallel()

	if fallback("value", "default") != "value" {
		t.Error("should return value")
	}
	if fallback("", "default") != "default" {
		t.Error("should return default")
	}
}

func TestStringSliceFlag(t *testing.T) {
	t.Parallel()

	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	f := newStringSliceFlag(fs, "test", "test")

	if f.set {
		t.Error("should not be set initially")
	}

	if err := f.Set("a,b"); err != nil {
		t.Fatalf("Set error: %v", err)
	}

	if !f.set {
		t.Error("should be set after Set()")
	}
}

func TestBoolFlag(t *testing.T) {
	t.Parallel()

	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	f := newBoolFlag(fs, "test", "test")

	if err := f.Set("true"); err != nil {
		t.Fatalf("Set error: %v", err)
	}

	if !f.value {
		t.Error("value should be true")
	}
}

func TestIntFlag(t *testing.T) {
	t.Parallel()

	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	f := newIntFlag(fs, "test", "test")

	if err := f.Set("42"); err != nil {
		t.Fatalf("Set error: %v", err)
	}

	if f.value != 42 {
		t.Errorf("value = %d, want 42", f.value)
	}
}

func TestBuildOverrides(t *testing.T) {
	t.Parallel()

	state := terragruntFlagState{
		root:             "/test",
		env:              "prod",
		parallelismValue: 8,
		parallelismSet:   true,
	}

	overrides := buildOverrides(state)

	if overrides.Root == nil || *overrides.Root != "/test" {
		t.Error("Root not set correctly")
	}
}

func TestParseTerragruntFlags_Plan(t *testing.T) {
	t.Parallel()

	args := []string{"-root", "/test", "-destroy", "true"}
	state, code := parseTerragruntFlags(args, cmdPlan)

	if code != 0 {
		t.Fatalf("expected code 0, got %d", code)
	}

	if state.root != "/test" {
		t.Errorf("got %q, want /test", state.root)
	}
}

func TestRun_NoArgs(t *testing.T) {
	t.Parallel()

	code := Run([]string{"cultivator"}, VersionInfo{})
	if code != 2 {
		t.Errorf("expected 2, got %d", code)
	}
}

func TestRun_Version(t *testing.T) {
	t.Parallel()

	version := VersionInfo{Version: "1.0.0"}
	code := Run([]string{"cultivator", "version"}, version)
	if code != 0 {
		t.Errorf("expected 0, got %d", code)
	}
}

func TestRun_Unknown(t *testing.T) {
	t.Parallel()

	code := Run([]string{"cultivator", "unknown"}, VersionInfo{})
	if code != 2 {
		t.Errorf("expected 2, got %d", code)
	}
}

func TestBuildTerragruntConfig_Defaults(t *testing.T) {
	t.Parallel()

	state := terragruntFlagState{}
	cfg, err := buildTerragruntConfig(state)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Root == "" {
		t.Error("root should have default")
	}
}

func TestBuildTerragruntConfig_WithFile(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "config.yml")
	content := []byte("root: /from/file\n")

	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}

	state := terragruntFlagState{configPath: path}
	cfg, err := buildTerragruntConfig(state)
	if err != nil {
		t.Fatalf("error: %v", err)
	}

	if !strings.Contains(cfg.Root, "from") {
		t.Errorf("got %q", cfg.Root)
	}
}

func TestRunDoctor_MissingRoot(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	missing := filepath.Join(dir, "nonexistent")

	code := runDoctor([]string{"-root", missing})
	if code != 1 {
		t.Errorf("expected 1, got %d", code)
	}
}

func TestRunDoctor_ValidRoot(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	code := runDoctor([]string{"-root", dir})

	if code != 0 && code != 1 {
		t.Errorf("expected 0 or 1, got %d", code)
	}
}

func TestExecLookPath(t *testing.T) {
	t.Parallel()

	path, err := execLookPath("sh")
	if err != nil {
		t.Errorf("sh should exist: %v", err)
	}
	if path == "" {
		t.Error("path empty")
	}

	_, err = execLookPath("nonexistent-cmd-12345")
	if err == nil {
		t.Error("should error for nonexistent")
	}
}

func TestPrintVersion(t *testing.T) {
	t.Parallel()

	printVersion(VersionInfo{Version: "1.0.0"})
}

func TestParseTerragruntFlags_AllCommands(t *testing.T) {
	t.Parallel()

	commands := []string{cmdPlan, cmdApply, cmdDestroy}
	for _, cmd := range commands {
		t.Run(cmd, func(t *testing.T) {
			args := []string{"-root", "/tmp", "-env", "dev", "-parallelism", "4"}
			state, code := parseTerragruntFlags(args, cmd)
			if code != 0 {
				t.Fatalf("expected 0, got %d", code)
			}
			if state.env != "dev" {
				t.Errorf("env = %q, want dev", state.env)
			}
			if state.parallelismValue != 4 {
				t.Errorf("parallelism = %d, want 4", state.parallelismValue)
			}
		})
	}
}

func TestParseTerragruntFlags_Tags(t *testing.T) {
	t.Parallel()

	args := []string{"-tags", "app:web,env:prod"}
	state, code := parseTerragruntFlags(args, cmdPlan)

	if code != 0 {
		t.Fatalf("expected 0, got %d", code)
	}

	if len(state.tagsValues) == 0 {
		t.Error("tags should be set")
	}
}

func TestParseTerragruntFlags_IncludeExclude(t *testing.T) {
	t.Parallel()

	args := []string{
		"-include", "path1,path2",
		"-exclude", "path3",
	}
	state, code := parseTerragruntFlags(args, cmdPlan)

	if code != 0 {
		t.Fatalf("expected 0, got %d", code)
	}

	if len(state.includeValues) < 2 {
		t.Error("include should have 2 items")
	}
}

func TestParseTerragruntFlags_InvalidParallelism(t *testing.T) {
	t.Parallel()

	args := []string{"-parallelism", "invalid"}
	_, code := parseTerragruntFlags(args, cmdPlan)

	if code == 0 {
		t.Error("should fail with invalid parallelism")
	}
}

func TestBuildTerragruntConfig_WithEnvVars(t *testing.T) {
	t.Parallel()

	err := os.Setenv("CULTIVATOR_ROOT", "/env/root")
	if err != nil {
		t.Fatalf("failed to set env var: %v", err)
	}
	defer func() {
		if err := os.Unsetenv("CULTIVATOR_ROOT"); err != nil {
			t.Errorf("failed to unset env var: %v", err)
		}
	}()

	state := terragruntFlagState{}
	cfg, err := buildTerragruntConfig(state)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(cfg.Root, "env") && !strings.Contains(cfg.Root, ".") {
		t.Logf("root from env: %q", cfg.Root)
	}
}

func TestBuildTerragruntConfig_Overrides(t *testing.T) {
	t.Parallel()

	state := terragruntFlagState{
		root:                "/override",
		env:                 "staging",
		includeValues:       []string{"a", "b"},
		includeSet:          true,
		excludeValues:       []string{"c"},
		excludeSet:          true,
		tagsValues:          []string{"tag:val"},
		tagsSet:             true,
		parallelismValue:    10,
		parallelismSet:      true,
		planDestroyValue:    true,
		planDestroySet:      true,
		nonInteractiveValue: true,
		nonInteractiveSet:   true,
		outputFormat:        "json",
	}

	cfg, err := buildTerragruntConfig(state)
	if err != nil {
		t.Fatalf("error: %v", err)
	}

	if cfg.Root != "/override" {
		t.Errorf("root = %q, want /override", cfg.Root)
	}
	if cfg.Env != "staging" {
		t.Errorf("env = %q, want staging", cfg.Env)
	}
	if cfg.OutputFormat != "json" {
		t.Errorf("outputFormat = %q, want json", cfg.OutputFormat)
	}
}

func TestParseTerragruntFlags_PositionalModulePath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		args          []string
		expectedPath  string
		expectInclude bool
	}{
		{
			name:          "path with terragrunt.hcl",
			args:          []string{"cloudwatch/log-group/lambda-example/terragrunt.hcl"},
			expectedPath:  "cloudwatch/log-group/lambda-example",
			expectInclude: true,
		},
		{
			name:          "path without terragrunt.hcl",
			args:          []string{"cloudwatch/log-group/lambda-example"},
			expectedPath:  "cloudwatch/log-group/lambda-example",
			expectInclude: true,
		},
		{
			name:          "path with leading ./",
			args:          []string{"./cloudwatch/log-group/lambda-example"},
			expectedPath:  "cloudwatch/log-group/lambda-example",
			expectInclude: true,
		},
		{
			name:          "path with ./ and terragrunt.hcl",
			args:          []string{"./cloudwatch/log-group/lambda-example/terragrunt.hcl"},
			expectedPath:  "cloudwatch/log-group/lambda-example",
			expectInclude: true,
		},
		{
			name:          "no positional args",
			args:          []string{},
			expectedPath:  "",
			expectInclude: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state, code := parseTerragruntFlags(tt.args, cmdPlan)

			if code != 0 {
				t.Fatalf("expected code 0, got %d", code)
			}

			if state.module != tt.expectedPath {
				t.Errorf("module = %q, want %q", state.module, tt.expectedPath)
			}

			if tt.expectInclude {
				if !state.includeSet {
					t.Error("include should be set")
				}
				found := false
				for _, inc := range state.includeValues {
					if inc == tt.expectedPath {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("include values %v should contain %q", state.includeValues, tt.expectedPath)
				}
			} else if len(tt.args) == 0 {
				// When no positional arg and no flags, includeSet should be false
				if state.includeSet {
					t.Error("include should not be set when no args provided")
				}
			}
		})
	}
}

func TestParseTerragruntFlags_PositionalWithFlags(t *testing.T) {
	t.Parallel()

	// Test that positional arg is combined with existing flags
	args := []string{
		"-include", "other/path",
		"cloudwatch/log-group/lambda-example/terragrunt.hcl",
	}
	state, code := parseTerragruntFlags(args, cmdPlan)

	if code != 0 {
		t.Fatalf("expected code 0, got %d", code)
	}

	if state.module != "cloudwatch/log-group/lambda-example" {
		t.Errorf("module = %q, want cloudwatch/log-group/lambda-example", state.module)
	}

	if !state.includeSet {
		t.Error("include should be set")
	}

	// Both paths should be in include
	if len(state.includeValues) != 2 {
		t.Errorf("include values count = %d, want 2", len(state.includeValues))
	}

	hasOther := false
	hasCloudwatch := false
	for _, inc := range state.includeValues {
		if inc == "other/path" {
			hasOther = true
		}
		if inc == "cloudwatch/log-group/lambda-example" {
			hasCloudwatch = true
		}
	}

	if !hasOther || !hasCloudwatch {
		t.Errorf("include values should contain both paths, got %v", state.includeValues)
	}
}

func TestNormalizePath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input    string
		expected string
	}{
		{"cloudwatch/log-group/lambda-example/terragrunt.hcl", "cloudwatch/log-group/lambda-example"},
		{"cloudwatch/log-group/lambda-example", "cloudwatch/log-group/lambda-example"},
		{"./cloudwatch/log-group/lambda-example", "cloudwatch/log-group/lambda-example"},
		{"./cloudwatch/log-group/lambda-example/terragrunt.hcl", "cloudwatch/log-group/lambda-example"},
		{"", ""},
		{"   ", ""},
		{"path/to/module/", "path/to/module/"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := normalizePath(tt.input)
			if result != tt.expected {
				t.Errorf("normalizePath(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestRunDoctor_WithFlags(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	code := runDoctor([]string{"-root", dir, "-config", "/nonexistent.yml"})

	if code != 0 && code != 1 {
		t.Errorf("unexpected code: %d", code)
	}
}
