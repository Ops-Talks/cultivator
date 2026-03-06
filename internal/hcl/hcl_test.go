package hcl

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestExtractDependencies(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	// Create some dummy dependency directories
	vpcPath := filepath.Join(tmpDir, "vpc")
	dbPath := filepath.Join(tmpDir, "db")
	_ = os.Mkdir(vpcPath, 0o755)
	_ = os.Mkdir(dbPath, 0o755)

	tests := []struct {
		name     string
		content  string
		wantDeps []Dependency
		wantErr  bool
	}{
		{
			name: "single dependency",
			content: `
dependency "vpc" {
  config_path = "../vpc"
}
`,
			wantDeps: []Dependency{
				{Name: "vpc", Path: vpcPath},
			},
		},
		{
			name: "multiple dependencies",
			content: `
dependency "vpc" {
  config_path = "../vpc"
}

dependency "db" {
  config_path = "../db"
}
`,
			wantDeps: []Dependency{
				{Name: "vpc", Path: vpcPath},
				{Name: "db", Path: dbPath},
			},
		},
		{
			name: "dependency with comments and different spacing",
			content: `
# This is a comment
dependency    "vpc"   {
  config_path = "../vpc" # Another comment
}
`,
			wantDeps: []Dependency{
				{Name: "vpc", Path: vpcPath},
			},
		},
		{
			name: "no dependencies",
			content: `
terraform {
  source = "..."
}
`,
			wantDeps: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hclDir := filepath.Join(tmpDir, tt.name)
			_ = os.Mkdir(hclDir, 0o755)
			hclPath := filepath.Join(hclDir, "terragrunt.hcl")

			if err := os.WriteFile(hclPath, []byte(tt.content), 0o644); err != nil {
				t.Fatalf("failed to write test file: %v", err)
			}

			got, err := ExtractDependencies(hclPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("ExtractDependencies() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !reflect.DeepEqual(got, tt.wantDeps) {
				t.Errorf("ExtractDependencies() got = %v, want %v", got, tt.wantDeps)
			}
		})
	}
}

func TestExtractDependencies_FileNotFound(t *testing.T) {
	t.Parallel()
	_, err := ExtractDependencies("non-existent.hcl")
	if err == nil {
		t.Error("expected error for non-existent file, got nil")
	}
}
