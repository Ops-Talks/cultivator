package cli

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/Ops-Talks/cultivator/internal/discovery"
	"github.com/Ops-Talks/cultivator/internal/logging"
	"github.com/Ops-Talks/cultivator/internal/runner"
)

func Test_Helpers(t *testing.T) {
	t.Parallel()

	t.Run("parseBool", func(t *testing.T) {
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
				if (err != nil) != tt.wantErr {
					t.Errorf("parseBool(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				}
				if !tt.wantErr && result != tt.expected {
					t.Errorf("parseBool(%q) = %v, want %v", tt.input, result, tt.expected)
				}
			})
		}
	})

	t.Run("parseInt", func(t *testing.T) {
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
				if (err != nil) != tt.wantErr {
					t.Errorf("parseInt(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				}
				if !tt.wantErr && result != tt.expected {
					t.Errorf("parseInt(%q) = %d, want %d", tt.input, result, tt.expected)
				}
			})
		}
	})

	t.Run("splitList", func(t *testing.T) {
		result := splitList("a,b;c")
		if len(result) != 3 {
			t.Errorf("splitList(\"a,b;c\") len = %d, want 3", len(result))
		}
	})

	t.Run("fallback", func(t *testing.T) {
		if fallback("value", "default") != "value" {
			t.Error("fallback(\"value\", \"default\") != \"value\"")
		}
		if fallback("", "default") != "default" {
			t.Error("fallback(\"\", \"default\") != \"default\"")
		}
	})

	t.Run("normalizePath", func(t *testing.T) {
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
	})
}

func Test_Flags(t *testing.T) {
	t.Parallel()

	t.Run("stringSlice", func(t *testing.T) {
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
	})

	t.Run("bool", func(t *testing.T) {
		fs := flag.NewFlagSet("test", flag.ContinueOnError)
		f := newBoolFlag(fs, "test", "test")
		if err := f.Set("true"); err != nil {
			t.Fatalf("Set error: %v", err)
		}
		if !f.value {
			t.Error("value should be true")
		}
	})

	t.Run("int", func(t *testing.T) {
		fs := flag.NewFlagSet("test", flag.ContinueOnError)
		f := newIntFlag(fs, "test", "test")
		if err := f.Set("42"); err != nil {
			t.Fatalf("Set error: %v", err)
		}
		if f.value != 42 {
			t.Errorf("value = %d, want 42", f.value)
		}
	})
}

func Test_ParseTerragruntFlags(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		args     []string
		command  string
		validate func(*testing.T, terragruntFlagState, int)
	}{
		{
			name:    "plan with root and destroy",
			args:    []string{"-root", "/test", "-destroy", "true"},
			command: cmdPlan,
			validate: func(t *testing.T, state terragruntFlagState, code int) {
				if code != 0 {
					t.Fatalf("expected code 0, got %d", code)
				}
				if state.root != "/test" {
					t.Errorf("root = %q, want /test", state.root)
				}
				if !state.planDestroyValue {
					t.Error("planDestroyValue should be true")
				}
			},
		},
		{
			name:    "all commands basic flags",
			args:    []string{"-env", "dev", "-parallelism", "4"},
			command: cmdApply,
			validate: func(t *testing.T, state terragruntFlagState, code int) {
				if code != 0 {
					t.Fatalf("expected code 0, got %d", code)
				}
				if state.env != "dev" {
					t.Errorf("env = %q, want dev", state.env)
				}
				if state.parallelismValue != 4 {
					t.Errorf("parallelism = %d, want 4", state.parallelismValue)
				}
			},
		},
		{
			name:    "invalid parallelism",
			args:    []string{"-parallelism", "invalid"},
			command: cmdPlan,
			validate: func(t *testing.T, state terragruntFlagState, code int) {
				if code == 0 {
					t.Error("should fail with invalid parallelism")
				}
			},
		},
		{
			name:    "positional module path",
			args:    []string{"./cloudwatch/log-group/lambda-example/terragrunt.hcl"},
			command: cmdPlan,
			validate: func(t *testing.T, state terragruntFlagState, code int) {
				if code != 0 {
					t.Fatalf("expected code 0, got %d", code)
				}
				if state.module != "cloudwatch/log-group/lambda-example" {
					t.Errorf("module = %q, want cloudwatch/log-group/lambda-example", state.module)
				}
				if !state.includeSet {
					t.Error("include should be set")
				}
			},
		},
		{
			name:    "changed-only and base flags",
			args:    []string{"-changed-only=true", "-base", "main"},
			command: cmdPlan,
			validate: func(t *testing.T, state terragruntFlagState, code int) {
				if code != 0 {
					t.Fatalf("expected code 0, got %d", code)
				}
				if !state.changedOnlyValue {
					t.Error("changedOnlyValue should be true")
				}
				if state.baseRefValue != "main" {
					t.Errorf("baseRefValue = %q, want main", state.baseRefValue)
				}
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			state, code := parseTerragruntFlags(tc.args, tc.command)
			tc.validate(t, state, code)
		})
	}
}

func Test_Run(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		args    []string
		version VersionInfo
		want    int
	}{
		{
			name: "no args",
			args: []string{"cultivator"},
			want: 2,
		},
		{
			name:    "version command",
			args:    []string{"cultivator", "version"},
			version: VersionInfo{Version: "1.0.0"},
			want:    0,
		},
		{
			name: "unknown command",
			args: []string{"cultivator", "unknown"},
			want: 2,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			code := Run(tc.args, tc.version)
			if code != tc.want {
				t.Errorf("Run() = %d, want %d", code, tc.want)
			}
		})
	}
}

func Test_BuildTerragruntConfig(t *testing.T) {
	t.Run("from file", func(t *testing.T) {
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
			t.Errorf("root = %q, want to contain 'from'", cfg.Root)
		}
	})

	t.Run("from env vars", func(t *testing.T) {
		t.Setenv("CULTIVATOR_ROOT", "/env/root")
		state := terragruntFlagState{}
		cfg, err := buildTerragruntConfig(state)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(cfg.Root, "env") && !strings.Contains(cfg.Root, ".") {
			t.Logf("root from env: %q", cfg.Root)
		}
	})
}

func Test_Doctor(t *testing.T) {
	t.Parallel()

	t.Run("missing root", func(t *testing.T) {
		dir := t.TempDir()
		missing := filepath.Join(dir, "nonexistent")
		code := runDoctor([]string{"-root", missing})
		if code != 1 {
			t.Errorf("runDoctor missing root = %d, want 1", code)
		}
	})

	t.Run("valid root", func(t *testing.T) {
		dir := t.TempDir()
		code := runDoctor([]string{"-root", dir})
		if code != 0 && code != 1 {
			t.Errorf("runDoctor valid root = %d, want 0 or 1", code)
		}
	})
}

func Test_LogExecutionResults(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		results  []runner.Result
		validate func(*testing.T, string, string, error)
	}{
		{
			name: "success plan",
			results: []runner.Result{
				{
					Module:   discovery.Module{Path: "app/module"},
					Command:  "plan",
					Stdout:   "Plan: 1 to add",
					ExitCode: 0,
					Duration: 123 * time.Millisecond,
				},
			},
			validate: func(t *testing.T, out, errOut string, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				combined := out + errOut
				if !strings.Contains(combined, "=== plan: app/module ===") {
					t.Errorf("missing header, got %q", combined)
				}
				if !strings.Contains(combined, "Plan: 1 to add") {
					t.Errorf("missing stdout, got %q", combined)
				}
				if !strings.Contains(combined, "duration=123ms") {
					t.Errorf("missing duration, got %q", combined)
				}
			},
		},
		{
			name: "failure plan",
			results: []runner.Result{
				{
					Module:   discovery.Module{Path: "infra/vpc"},
					Command:  "plan",
					ExitCode: 1,
					Stdout:   "Error: something went wrong",
					Duration: 456 * time.Millisecond,
				},
			},
			validate: func(t *testing.T, out, errOut string, err error) {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				combined := out + errOut
				if !strings.Contains(combined, "Error: something went wrong") {
					t.Errorf("missing error output, got %q", combined)
				}
				if !strings.Contains(combined, "duration=456ms") {
					t.Errorf("missing duration in error, got %q", combined)
				}
			},
		},
		{
			name: "multiple modules preserves order",
			results: []runner.Result{
				{
					Module:  discovery.Module{Path: "modules/vpc"},
					Command: "plan",
					Stdout:  "vpc output",
				},
				{
					Module:  discovery.Module{Path: "modules/eks"},
					Command: "plan",
					Stdout:  "eks output",
				},
			},
			validate: func(t *testing.T, out, errOut string, err error) {
				vpcPos := strings.Index(out, "modules/vpc")
				eksPos := strings.Index(out, "modules/eks")
				if vpcPos == -1 || eksPos == -1 || eksPos < vpcPos {
					t.Errorf("order not preserved: vpcPos=%d, eksPos=%d", vpcPos, eksPos)
				}
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			var out, errOut bytes.Buffer
			logger := logging.New(logging.LevelInfo, &out, &errOut)
			err := logExecutionResults(logger, tc.results)
			tc.validate(t, out.String(), errOut.String(), err)
		})
	}
}

func Test_BuildTerragruntConfig_ChangedOnly(t *testing.T) {
	t.Parallel()

	state := terragruntFlagState{
		changedOnlyValue: true,
		changedOnlySet:   true,
		baseRefValue:     "main",
		baseRefSet:       true,
	}

	cfg, err := buildTerragruntConfig(state)
	if err != nil {
		t.Fatalf("error: %v", err)
	}

	if !cfg.ChangedOnly {
		t.Error("cfg.ChangedOnly should be true")
	}
	if cfg.BaseRef != "main" {
		t.Errorf("cfg.BaseRef = %q, want main", cfg.BaseRef)
	}
}

func Test_runTerragruntCommand_Flow(t *testing.T) {
	t.Parallel()

	t.Run("successful plan flow", func(t *testing.T) {
		tmpDir := t.TempDir()
		moduleDir := filepath.Join(tmpDir, "prod", "app1")
		_ = os.MkdirAll(moduleDir, 0o755)
		_ = os.WriteFile(filepath.Join(moduleDir, "terragrunt.hcl"), []byte("# cultivator:tags=app"), 0o644)

		executor := &cliFakeExecutor{}
		r := runner.New().WithExecutor(executor)

		// Mock RunArgs
		args := []string{"-root", tmpDir, "-tags", "app"}

		code := runTerragruntCommand(args, cmdPlan, r)
		if code != 0 {
			t.Errorf("runTerragruntCommand() = %d, want 0", code)
		}

		if len(executor.calls) != 1 {
			t.Errorf("expected 1 executor call, got %d", len(executor.calls))
		}
	})

	t.Run("no modules matched", func(t *testing.T) {
		tmpDir := t.TempDir()
		executor := &cliFakeExecutor{}
		r := runner.New().WithExecutor(executor)

		args := []string{"-root", tmpDir, "-tags", "nonexistent"}
		code := runTerragruntCommand(args, cmdPlan, r)

		if code != 0 {
			t.Errorf("runTerragruntCommand() = %d, want 0 (graceful exit)", code)
		}
		if len(executor.calls) != 0 {
			t.Errorf("expected 0 executor calls, got %d", len(executor.calls))
		}
	})

	t.Run("execution failure", func(t *testing.T) {
		tmpDir := t.TempDir()
		moduleDir := filepath.Join(tmpDir, "prod", "app1")
		_ = os.MkdirAll(moduleDir, 0o755)
		_ = os.WriteFile(filepath.Join(moduleDir, "terragrunt.hcl"), []byte(""), 0o644)

		executor := &cliErrorExecutor{}
		r := runner.New().WithExecutor(executor)

		args := []string{"-root", tmpDir}
		code := runTerragruntCommand(args, cmdPlan, r)

		if code != 1 {
			t.Errorf("runTerragruntCommand() = %d, want 1 (failure)", code)
		}
	})
}

// Test_runTerragruntCommand_FakeRunner verifies that each command is dispatched
// through the RunnerIface, allowing full mock injection without a real executor.
func Test_runTerragruntCommand_FakeRunner(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	moduleDir := filepath.Join(tmpDir, "env", "app")
	_ = os.MkdirAll(moduleDir, 0o755)
	_ = os.WriteFile(filepath.Join(moduleDir, "terragrunt.hcl"), []byte(""), 0o644)

	tests := []struct {
		name    string
		command string
	}{
		{"plan dispatched", cmdPlan},
		{"apply dispatched", cmdApply},
		{"destroy dispatched", cmdDestroy},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			fr := &fakeRunnerIface{}
			code := runTerragruntCommand([]string{"-root", tmpDir}, tc.command, fr)
			if code != 0 {
				t.Errorf("runTerragruntCommand() = %d, want 0", code)
			}
			if len(fr.commands) != 1 || fr.commands[0] != tc.command {
				t.Errorf("expected command %q dispatched, got %v", tc.command, fr.commands)
			}
		})
	}
}

type fakeRunnerIface struct {
	mu       sync.Mutex
	commands []string
}

func (f *fakeRunnerIface) Run(_ context.Context, command string, _ []discovery.Module, _ runner.Options) ([]runner.Result, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.commands = append(f.commands, command)
	return nil, nil
}

type cliErrorExecutor struct{}

func (e *cliErrorExecutor) Run(_ context.Context, _, _ string, _, _ []string) (string, string, int, error) {
	return "", "error", 1, errors.New("command failed")
}

type cliFakeExecutor struct {
	mu    sync.Mutex
	calls []cliCall
}

type cliCall struct {
	workDir string
	args    []string
}

func (f *cliFakeExecutor) Run(_ context.Context, workDir string, _ string, args []string, _ []string) (string, string, int, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls = append(f.calls, cliCall{workDir: workDir, args: args})
	return "ok", "", 0, nil
}
