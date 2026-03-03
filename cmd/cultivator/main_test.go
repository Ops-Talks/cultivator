package main

import (
	"os"
	"testing"

	"github.com/Ops-Talks/cultivator/internal/cli"
)

func Test_variables_defaultValues(t *testing.T) {
	t.Parallel()

	if version == "" {
		t.Error("version must not be empty")
	}
	if commit == "" {
		t.Error("commit must not be empty")
	}
	if date == "" {
		t.Error("date must not be empty")
	}
}

func Test_VersionInfo_fields(t *testing.T) {
	t.Parallel()

	vi := cli.VersionInfo{
		Version: "1.0.0",
		Commit:  "abc123",
		Date:    "2024-01-01",
	}

	if vi.Version != "1.0.0" {
		t.Errorf("Version = %q, want 1.0.0", vi.Version)
	}
	if vi.Commit != "abc123" {
		t.Errorf("Commit = %q, want abc123", vi.Commit)
	}
	if vi.Date != "2024-01-01" {
		t.Errorf("Date = %q, want 2024-01-01", vi.Date)
	}
}

func Test_run_exitCodes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		args     []string
		wantCode int
	}{
		{
			name:     "version command",
			args:     []string{"cultivator", "version"},
			wantCode: 0,
		},
		{
			name:     "no arguments",
			args:     []string{"cultivator"},
			wantCode: 2,
		},
		{
			name:     "unknown command",
			args:     []string{"cultivator", "invalid"},
			wantCode: 2,
		},
		{
			name:     "plan with nonexistent root",
			args:     []string{"cultivator", "plan", "--root=/tmp/nonexistent"},
			wantCode: 1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Tests that mutate os.Args cannot run in parallel.
			oldArgs := os.Args
			t.Cleanup(func() { os.Args = oldArgs })

			os.Args = tc.args
			got := run()
			if got != tc.wantCode {
				t.Errorf("run(%v) = %d, want %d", tc.args, got, tc.wantCode)
			}
		})
	}
}

func Test_run_doctor(t *testing.T) {
	// doctor returns 0 or 1 depending on the system environment;
	// any other code indicates an unexpected failure in command dispatch.
	oldArgs := os.Args
	t.Cleanup(func() { os.Args = oldArgs })

	os.Args = []string{"cultivator", "doctor"}
	got := run()
	if got != 0 && got != 1 {
		t.Errorf("run(doctor) = %d, want 0 or 1", got)
	}
}
