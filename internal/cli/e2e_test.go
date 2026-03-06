package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Ops-Talks/cultivator/internal/runner"
)

func TestE2E_TerragruntStructure(t *testing.T) {
	// Use absolute path to testdata
	wd, _ := os.Getwd()
	testdataDir := filepath.Join(wd, "..", "..", "testdata", "terragrunt-structure")

	tests := []struct {
		name           string
		args           []string
		command        string
		expectedCalls  int
		expectedStatus int
		validate       func(*testing.T, *cliFakeExecutor)
	}{
		{
			name:           "full discovery and execution",
			args:           []string{"-root", testdataDir},
			command:        cmdPlan,
			expectedCalls:  3, // dev/app3, prod/app1, prod/app2
			expectedStatus: 0,
		},
		{
			name:           "filter by environment dev",
			args:           []string{"-root", testdataDir, "-env", "dev"},
			command:        cmdPlan,
			expectedCalls:  1,
			expectedStatus: 0,
			validate: func(t *testing.T, executor *cliFakeExecutor) {
				if !strings.Contains(executor.calls[0].workDir, "dev/app3") {
					t.Errorf("expected dev/app3, got %s", executor.calls[0].workDir)
				}
			},
		},
		{
			name:           "filter by tag app",
			args:           []string{"-root", testdataDir, "-tags", "app"},
			command:        cmdPlan,
			expectedCalls:  2, // dev/app3 and prod/app1
			expectedStatus: 0,
		},
		{
			name:           "dry-run mode",
			args:           []string{"-root", testdataDir, "-dry-run=true"},
			command:        cmdPlan,
			expectedCalls:  0, // Executor should not be called in dry-run
			expectedStatus: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			executor := &cliFakeExecutor{}
			r := runner.New().WithExecutor(executor)

			code := runTerragruntCommand(tt.args, tt.command, r)

			if code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, code)
			}

			if len(executor.calls) != tt.expectedCalls {
				t.Errorf("expected %d calls, got %d", tt.expectedCalls, len(executor.calls))
			}

			if tt.validate != nil {
				tt.validate(t, executor)
			}
		})
	}
}

func TestE2E_TerragruntLarge(t *testing.T) {
	wd, _ := os.Getwd()
	testdataDir := filepath.Join(wd, "..", "..", "testdata", "terragrunt-large")

	t.Run("parallel execution on large structure", func(t *testing.T) {
		executor := &cliFakeExecutor{}
		r := runner.New().WithExecutor(executor)

		// Run plan on prod environment which has many modules
		args := []string{"-root", testdataDir, "-env", "prod", "-parallelism", "10"}
		code := runTerragruntCommand(args, cmdPlan, r)

		if code != 0 {
			t.Errorf("expected status 0, got %d", code)
		}

		// Each env (dev, prod, staging, test) has ~25 modules
		if len(executor.calls) < 20 {
			t.Errorf("expected at least 20 calls, got %d", len(executor.calls))
		}
	})
}
