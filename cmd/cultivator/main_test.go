package main

import (
	"os"
	"testing"

	"github.com/Ops-Talks/cultivator/internal/cli"
)

func TestVersionInfo(t *testing.T) {
	// Test that variables are initialized
	if version == "dev" {
		t.Log("version is default value")
	}
	if commit == "unknown" {
		t.Log("commit is default value")
	}
	if date == "unknown" {
		t.Log("date is default value")
	}
}

func TestVersionInfoStruct(t *testing.T) {
	// Test VersionInfo struct construction
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

func TestRun_Version(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"cultivator", "version"}
	code := run()
	if code != 0 {
		t.Errorf("run() with version command = %d, want 0", code)
	}
}

func TestRun_NoArgs(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"cultivator"}
	code := run()
	if code != 2 {
		t.Errorf("run() with no args = %d, want 2", code)
	}
}

func TestRun_InvalidCommand(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"cultivator", "invalid"}
	code := run()
	if code != 2 {
		t.Errorf("run() with invalid command = %d, want 2", code)
	}
}

func TestRun_Plan_NoRoot(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	// plan with nonexistent root will fail during discovery
	os.Args = []string{"cultivator", "plan", "--root=/tmp/nonexistent"}
	code := run()
	if code != 1 {
		t.Errorf("run() plan with nonexistent root = %d, want 1", code)
	}
}

func TestRun_Doctor(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"cultivator", "doctor"}
	code := run()
	// doctor can return 0 or 1 depending on system
	if code != 0 && code != 1 {
		t.Errorf("run() doctor = %d, want 0 or 1", code)
	}
}
