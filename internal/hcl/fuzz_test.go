package hcl

import (
	"os"
	"path/filepath"
	"testing"
)

// FuzzExtractDependencies tests ExtractDependencies with random HCL content
func FuzzExtractDependencies(f *testing.F) {
	// Add seed corpus
	f.Add(`dependency "vpc" { config_path = "../vpc" }`)
	f.Add(`dependency "db" { config_path = "../db" }
dependency "vpc" { config_path = "../vpc" }`)
	f.Add("# some comment\ndependency \"app\" { config_path = \"./app\" }")
	f.Add("")
	f.Add("terraform { source = \"...\" }")

	f.Fuzz(func(t *testing.T, content string) {
		tmpDir := t.TempDir()
		hclPath := filepath.Join(tmpDir, "terragrunt.hcl")

		if err := os.WriteFile(hclPath, []byte(content), 0o644); err != nil {
			t.Skip("could not write temp file")
		}

		// This should not panic
		deps, err := ExtractDependencies(hclPath)

		if err == nil {
			for _, dep := range deps {
				if dep.Name == "" {
					t.Errorf("ExtractDependencies returned empty dependency name")
				}
				if dep.Path == "" {
					t.Errorf("ExtractDependencies returned empty dependency path")
				}
			}
		}
	})
}
