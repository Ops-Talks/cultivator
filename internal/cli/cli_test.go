package cli

import (
	"bytes"
	"errors"
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Ops-Talks/cultivator/internal/discovery"
	"github.com/Ops-Talks/cultivator/internal/logging"
	"github.com/Ops-Talks/cultivator/internal/runner"
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
	// Cannot use t.Parallel() with t.Setenv().
	t.Setenv("CULTIVATOR_ROOT", "/env/root")

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

func TestLogExecutionResults_Success(t *testing.T) {
	t.Parallel()

	var out, errOut bytes.Buffer
	logger := logging.New(logging.LevelInfo, &out, &errOut)

	results := []runner.Result{
		{
			Module:   discovery.Module{Path: "app/module"},
			Command:  "plan",
			Stdout:   "Plan: 1 to add",
			ExitCode: 0,
		},
	}

	err := logExecutionResults(logger, results)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	outStr := out.String()
	if !strings.Contains(outStr, "=== plan: app/module ===") {
		t.Errorf("expected module header in output, got %q", outStr)
	}
	if !strings.Contains(outStr, "Plan: 1 to add") {
		t.Errorf("expected stdout in output, got %q", outStr)
	}
	if !strings.Contains(outStr, "completed") {
		t.Errorf("expected 'completed' in output, got %q", outStr)
	}
}

func TestLogExecutionResults_ExitCodeFailureNoError(t *testing.T) {
	t.Parallel()

	var out, errOut bytes.Buffer
	logger := logging.New(logging.LevelInfo, &out, &errOut)

	// exit code != 0 but Error == nil (normal terragrunt failure)
	results := []runner.Result{
		{
			Module:   discovery.Module{Path: "infra/vpc"},
			Command:  "plan",
			ExitCode: 1,
			Stderr:   "Error: something went wrong",
		},
	}

	err := logExecutionResults(logger, results)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	combined := out.String() + errOut.String()
	if !strings.Contains(out.String(), "=== plan: infra/vpc ===") {
		t.Errorf("expected module header in output, got %q", out.String())
	}
	// Should NOT contain "error=<nil>" — the error field must be omitted
	if strings.Contains(combined, "error=<nil>") {
		t.Errorf("log output should not contain 'error=<nil>', got: %q", combined)
	}
	if !strings.Contains(combined, "exit_code=1") {
		t.Errorf("expected exit_code in output, got: %q", combined)
	}
	if !strings.Contains(out.String(), "something went wrong") {
		t.Errorf("expected stderr in output, got: %q", out.String())
	}
}

func TestLogExecutionResults_ErrorSet(t *testing.T) {
	t.Parallel()

	var out, errOut bytes.Buffer
	logger := logging.New(logging.LevelInfo, &out, &errOut)

	execErr := errors.New("process killed")
	results := []runner.Result{
		{
			Module:   discovery.Module{Path: "infra/db"},
			Command:  "apply",
			ExitCode: 1,
			Error:    execErr,
		},
	}

	err := logExecutionResults(logger, results)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	combined := out.String() + errOut.String()
	if !strings.Contains(out.String(), "=== apply: infra/db ===") {
		t.Errorf("expected module header in output, got %q", out.String())
	}
	if !strings.Contains(combined, "process killed") {
		t.Errorf("expected error message in output, got: %q", combined)
	}
	if !strings.Contains(combined, "error=process killed") {
		t.Errorf("expected 'error=process killed' in output, got: %q", combined)
	}
}

func TestLogExecutionResults_MultipleModules(t *testing.T) {
	t.Parallel()

	var out, errOut bytes.Buffer
	logger := logging.New(logging.LevelInfo, &out, &errOut)

	results := []runner.Result{
		{
			Module:   discovery.Module{Path: "modules/vpc"},
			Command:  "plan",
			Stdout:   "Plan: 2 to add",
			ExitCode: 0,
		},
		{
			Module:   discovery.Module{Path: "modules/rds"},
			Command:  "plan",
			Stdout:   "",
			Stderr:   "Error: insufficient permissions",
			ExitCode: 1,
		},
		{
			Module:   discovery.Module{Path: "modules/eks"},
			Command:  "plan",
			Stdout:   "Plan: 1 to add",
			ExitCode: 0,
		},
	}

	err := logExecutionResults(logger, results)
	if err == nil {
		t.Fatal("expected error due to rds failure, got nil")
	}

	outStr := out.String()

	// Each module must have its own header, in order
	vpcIdx := strings.Index(outStr, "=== plan: modules/vpc ===")
	rdsIdx := strings.Index(outStr, "=== plan: modules/rds ===")
	eksIdx := strings.Index(outStr, "=== plan: modules/eks ===")

	if vpcIdx < 0 {
		t.Errorf("missing header for modules/vpc, output: %q", outStr)
	}
	if rdsIdx < 0 {
		t.Errorf("missing header for modules/rds, output: %q", outStr)
	}
	if eksIdx < 0 {
		t.Errorf("missing header for modules/eks, output: %q", outStr)
	}

	// vpc stdout must appear after vpc header and before rds header
	if vpcIdx > 0 && rdsIdx > 0 && !strings.Contains(outStr[vpcIdx:rdsIdx], "Plan: 2 to add") {
		t.Errorf("vpc stdout not between vpc and rds headers, output: %q", outStr)
	}

	// rds stderr must appear in the output
	if !strings.Contains(outStr, "insufficient permissions") {
		t.Errorf("expected rds stderr in output, got: %q", outStr)
	}

	// eks stdout must appear after eks header
	if eksIdx > 0 && !strings.Contains(outStr[eksIdx:], "Plan: 1 to add") {
		t.Errorf("eks stdout not found after eks header, output: %q", outStr)
	}
}
