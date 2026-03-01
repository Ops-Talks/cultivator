package main

import (
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
