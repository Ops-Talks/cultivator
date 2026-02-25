package detector

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/cultivator-dev/cultivator/pkg/module"
	"github.com/cultivator-dev/cultivator/pkg/parser"
)

// Terragrunt file extensions that are relevant for change detection (constant list - DRY principle)
var terragruntRelevantExtensions = [...]string{".hcl", ".tf", ".tfvars"}

const terragruntConfigFile = "terragrunt.hcl"

// Module represents a Terragrunt module that was changed
type Module struct {
	Path             string                 `json:"path"`
	RelativePath     string                 `json:"relative_path"`
	Dependencies     []string               `json:"dependencies,omitempty"`
	ExternalModules  []module.SourceInfo    `json:"external_modules,omitempty"` // External module sources
	Changed          bool                   `json:"changed"`
	Affected         bool                   `json:"affected"` // Affected by dependency changes
}

// ChangeDetector detects which Terragrunt modules have changed
type ChangeDetector struct {
	baseRef    string
	headRef    string
	workingDir string
}

// NewChangeDetector creates a new change detector
func NewChangeDetector(baseRef, headRef, workingDir string) *ChangeDetector {
	return &ChangeDetector{
		baseRef:    baseRef,
		headRef:    headRef,
		workingDir: workingDir,
	}
}

// DetectChangedFiles returns a list of files that changed between base and head
func (d *ChangeDetector) DetectChangedFiles() ([]string, error) {
	cmd := exec.Command("git", "diff", "--name-only", d.baseRef, d.headRef)
	cmd.Dir = d.workingDir

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to detect changed files: %w", err)
	}

	files := strings.Split(strings.TrimSpace(string(output)), "\n")
	return files, nil
}

// DetectChangedModules detects which Terragrunt modules were changed
func (d *ChangeDetector) DetectChangedModules() ([]Module, error) {
	files, err := d.DetectChangedFiles()
	if err != nil {
		return nil, err
	}

	moduleMap := make(map[string]*Module)
	hclParser := parser.NewParser()

	for _, file := range files {
		if file == "" {
			continue
		}

		// Skip non-terragrunt files
		if !d.isRelevantFile(file) {
			continue
		}

		// Find the module directory (directory containing terragrunt.hcl)
		moduleDir := d.findModuleDir(file)
		if moduleDir == "" {
			continue
		}

		if _, exists := moduleMap[moduleDir]; !exists {
			moduleMap[moduleDir] = &Module{
				Path:         filepath.Join(d.workingDir, moduleDir),
				RelativePath: moduleDir,
				Changed:      true,
			}

			// Extract external modules from this module (DRY: single responsibility)
			if externalSources, err := hclParser.GetExternalModuleSources(filepath.Join(d.workingDir, moduleDir)); err == nil {
				moduleMap[moduleDir].ExternalModules = externalSources
			}
		}
	}

	// Convert map to slice
	modules := make([]Module, 0, len(moduleMap))
	for _, mod := range moduleMap {
		modules = append(modules, *mod)
	}

	return modules, nil
}

// isRelevantFile checks if a file is relevant to Terragrunt (DRY principle - consolidated logic)
func (d *ChangeDetector) isRelevantFile(file string) bool {
	for _, ext := range terragruntRelevantExtensions {
		if strings.HasSuffix(file, ext) {
			return true
		}
	}
	return false
}

// findModuleDir finds the directory containing terragrunt.hcl for a given file
func (d *ChangeDetector) findModuleDir(file string) string {
	dir := filepath.Dir(file)

	// Check if this directory or any parent contains terragrunt.hcl
	for dir != "." && dir != "" {
		if d.hasTerragruntConfig(dir) {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached filesystem root
			break
		}
		dir = parent
	}

	return ""
}

// hasTerragruntConfig checks if a directory contains terragrunt.hcl (cross-platform, using os.Stat)
func (d *ChangeDetector) hasTerragruntConfig(dir string) bool {
	configPath := filepath.Join(d.workingDir, dir, terragruntConfigFile)
	_, err := os.Stat(configPath)
	return err == nil
}

// DetectAffectedModules detects modules affected by dependency changes
func (d *ChangeDetector) DetectAffectedModules(changedModules []Module, allModules []Module) ([]Module, error) {
	// Build dependency graph
	dependentMap := d.buildDependencyGraph(allModules)

	// Find all affected modules using DFS
	affected := d.findAffectedModules(changedModules, dependentMap)

	// Build result by filtering and marking modules as affected
	return d.buildAffectedResult(allModules, affected), nil
}

// buildDependencyGraph creates a map of module -> modules that depend on it (DRY principle)
func (d *ChangeDetector) buildDependencyGraph(modules []Module) map[string][]string {
	graph := make(map[string][]string)
	for _, module := range modules {
		for _, dep := range module.Dependencies {
			graph[dep] = append(graph[dep], module.Path)
		}
	}
	return graph
}

// findAffectedModules uses DFS to find all modules affected by changes (extracted method)
func (d *ChangeDetector) findAffectedModules(changedModules []Module, dependentMap map[string][]string) map[string]bool {
	affected := make(map[string]bool)

	var dfs func(string)
	dfs = func(path string) {
		if affected[path] {
			return
		}
		affected[path] = true

		// Add all modules that depend on this one
		for _, dependent := range dependentMap[path] {
			dfs(dependent)
		}
	}

	// Start DFS from all changed modules
	for _, module := range changedModules {
		dfs(module.Path)
	}

	return affected
}

// buildAffectedResult constructs the result list with affected modules marked (DRY helper method)
func (d *ChangeDetector) buildAffectedResult(allModules []Module, affected map[string]bool) []Module {
	result := make([]Module, 0, len(allModules))

	for _, module := range allModules {
		if affected[module.Path] {
			affectedModule := module
			affectedModule.Affected = true
			result = append(result, affectedModule)
		}
	}

	return result
}

// DetectExternalModuleChanges detects changes in external module versions
// Compares external module references between base and head commits
// Returns modules where external modules have been updated
func (d *ChangeDetector) DetectExternalModuleChanges(ctx context.Context, modules []Module) ([]Module, error) {
	changedModules := make([]Module, 0)
	hclParser := parser.NewParser()
	moduleSourceParser := module.NewSourceParser()

	for _, mod := range modules {
		// Get terragrunt.hcl at base commit
		baseTfContent, err := d.getFileAtRef(mod.RelativePath+"/terragrunt.hcl", d.baseRef)
		if err != nil {
			// File might not exist in base - skip
			continue
		}

		// Get terragrunt.hcl at head commit
		headTfContent, err := d.getFileAtRef(mod.RelativePath+"/terragrunt.hcl", d.headRef)
		if err != nil {
			continue
		}

		// Compare external module sources
		if d.hasExternalModuleChanges(baseTfContent, headTfContent, moduleSourceParser) {
			changedModules = append(changedModules, mod)
		}
	}

	// Extract external module sources for changed modules
	for i, mod := range changedModules {
		if externalSources, err := hclParser.GetExternalModuleSources(filepath.Join(d.workingDir, mod.RelativePath)); err == nil {
			changedModules[i].ExternalModules = externalSources
		}
	}

	return changedModules, nil
}

// getFileAtRef retrieves the content of a file at a specific git ref
// Following Single Responsibility: single method for git operations
func (d *ChangeDetector) getFileAtRef(filePath, ref string) (string, error) {
	cmd := exec.Command("git", "show", ref+":"+filePath)
	cmd.Dir = d.workingDir

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get file from git: %w", err)
	}

	return string(output), nil
}

// hasExternalModuleChanges detects if external module sources have changed
// DRY principle: single location for external module comparison
func (d *ChangeDetector) hasExternalModuleChanges(baseTfContent, headTfContent string, parser *module.SourceParser) bool {
	baseSource := d.extractSource(baseTfContent)
	headSource := d.extractSource(headTfContent)

	if baseSource == headSource {
		return false
	}

	// If sources differ, check if they're both external
	if (strings.HasPrefix(baseSource, "git::") || strings.HasPrefix(baseSource, "http")) &&
		(strings.HasPrefix(headSource, "git::") || strings.HasPrefix(headSource, "http")) {

		// Check if the version/ref has changed
		return d.compareExternalSources(baseSource, headSource, parser)
	}

	return true
}

// extractSource extracts the source value from terraform block content
// Helper method to keep hasExternalModuleChanges clean
func (d *ChangeDetector) extractSource(hclContent string) string {
	lines := strings.Split(hclContent, "\n")
	for _, line := range lines {
		if strings.Contains(line, "source") && strings.Contains(line, "=") {
			// Simple extraction: source = "..."
			parts := strings.Split(line, "=")
			if len(parts) == 2 {
				source := strings.TrimSpace(parts[1])
				source = strings.Trim(source, "\"")
				return source
			}
		}
	}
	return ""
}

// compareExternalSources compares two external sources to determine if they've changed
// DRY: single location for external source comparison logic
func (d *ChangeDetector) compareExternalSources(base, head string, parser *module.SourceParser) bool {
	baseInfo, err1 := parser.Parse(base)
	headInfo, err2 := parser.Parse(head)

	if err1 != nil || err2 != nil {
		// If parsing fails, consider it a change
		return true
	}

	// Compare URLs and references
	if baseInfo.URL != headInfo.URL {
		return true // Different repository
	}

	if baseInfo.Ref != headInfo.Ref {
		return true // Different version/tag/branch
	}

	return false
}
