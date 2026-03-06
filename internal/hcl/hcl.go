// Package hcl provides a lightweight parser for Terragrunt HCL files
// to extract module dependencies.
package hcl

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
)

// Dependency represents a Terragrunt dependency link.
type Dependency struct {
	Name string
	Path string // Absolute path to the dependent module
}

var (
	// dependencyRegex matches 'dependency "name" {'
	dependencyRegex = regexp.MustCompile(`(?m)^dependency\s+"([^"]+)"\s+\{`)
	// configPathRegex matches 'config_path = "path"'
	configPathRegex = regexp.MustCompile(`config_path\s+=\s+"([^"]+)"`)
)

// ExtractDependencies reads a terragrunt.hcl file and extracts all dependency paths.
// It returns a list of absolute paths to the dependency directories.
func ExtractDependencies(hclPath string) ([]Dependency, error) {
	content, err := os.ReadFile(hclPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read HCL file: %w", err)
	}

	hclDir := filepath.Dir(hclPath)
	matches := dependencyRegex.FindAllStringSubmatchIndex(string(content), -1)
	
	var deps []Dependency
	for _, match := range matches {
		name := string(content[match[2]:match[3]])
		
		// Look for config_path within the block (until the next closing brace)
		// This is a simplified approach for the lightweight parser.
		blockContent := string(content[match[1]:])
		pathMatch := configPathRegex.FindStringSubmatch(blockContent)
		
		if len(pathMatch) > 1 {
			relPath := pathMatch[1]
			absPath := filepath.Clean(filepath.Join(hclDir, relPath))
			
			deps = append(deps, Dependency{
				Name: name,
				Path: absPath,
			})
		}
	}

	return deps, nil
}
