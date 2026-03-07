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

// findBlockEnd returns the byte index of the closing '}' that balances the
// implicit opening '{' at the start position (depth begins at 1). Returns -1
// if no matching closing brace is found before the end of content.
func findBlockEnd(content []byte, start int) int {
	depth := 1
	for i := start; i < len(content); i++ {
		switch content[i] {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return i
			}
		}
	}
	return -1
}

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

		// Restrict config_path search to the current block only, preventing
		// a block without config_path from inheriting a path from a later block.
		blockEnd := findBlockEnd(content, match[1])
		if blockEnd < 0 {
			continue
		}
		pathMatch := configPathRegex.FindStringSubmatch(string(content[match[1]:blockEnd]))

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
