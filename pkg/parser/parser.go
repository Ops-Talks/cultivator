package parser

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/cultivator-dev/cultivator/pkg/module"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
)

// TerragruntConfig represents a parsed terragrunt.hcl file
type TerragruntConfig struct {
	Path         string
	Dependencies []Dependency
	Terraform    *TerraformBlock
	Inputs       map[string]interface{}
	Include      *IncludeBlock
}

// Dependency represents a dependency block
type Dependency struct {
	Name       string
	ConfigPath string
}

// TerraformBlock represents the terraform block
type TerraformBlock struct {
	Source       string                // Local or external module source
	ExternalInfo *module.SourceInfo    // Parsed info if source is external (git:: or http://)
	IsExternal   bool                  // Whether this source is external
}

// IncludeBlock represents the include block
type IncludeBlock struct {
	Path string
}

// Parser parses Terragrunt HCL files
type Parser struct {
	parser       *hclparse.Parser
	sourceParser *module.SourceParser // For parsing external module sources
}

// NewParser creates a new Terragrunt parser
func NewParser() *Parser {
	return &Parser{
		parser:       hclparse.NewParser(),
		sourceParser: module.NewSourceParser(),
	}
}

// ParseFile parses a terragrunt.hcl file
func (p *Parser) ParseFile(path string) (*TerragruntConfig, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %w", err)
	}

	content, err := os.ReadFile(absPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	file, diags := p.parser.ParseHCL(content, absPath)
	if diags.HasErrors() {
		return nil, fmt.Errorf("failed to parse HCL: %w", diags)
	}

	config := &TerragruntConfig{
		Path:         absPath,
		Dependencies: []Dependency{},
		Inputs:       make(map[string]interface{}),
	}

	// Parse the file body
	if file != nil && file.Body != nil {
		if err := p.parseBody(file.Body, config); err != nil {
			return nil, err
		}
	}

	return config, nil
}

// parseBody parses the HCL body
func (p *Parser) parseBody(body hcl.Body, config *TerragruntConfig) error {
	content, _, diags := body.PartialContent(&hcl.BodySchema{
		Blocks: []hcl.BlockHeaderSchema{
			{Type: "dependency"},
			{Type: "terraform"},
			{Type: "include"},
			{Type: "inputs"},
		},
	})

	if diags.HasErrors() {
		return fmt.Errorf("failed to parse body: %w", diags)
	}

	// Parse dependency blocks
	for _, block := range content.Blocks {
		switch block.Type {
		case "dependency":
			dep := p.parseDependencyBlock(block)
			if dep != nil {
				config.Dependencies = append(config.Dependencies, *dep)
			}
		case "terraform":
			config.Terraform = p.parseTerraformBlock(block)
		case "include":
			config.Include = p.parseIncludeBlock(block)
		}
	}

	return nil
}

// parseDependencyBlock parses a dependency block
func (p *Parser) parseDependencyBlock(block *hcl.Block) *Dependency {
	if len(block.Labels) == 0 {
		return nil
	}

	dep := &Dependency{
		Name: block.Labels[0],
	}

	attrs, _ := block.Body.JustAttributes()
	if configPathAttr, exists := attrs["config_path"]; exists {
		val, _ := configPathAttr.Expr.Value(nil)
		dep.ConfigPath = val.AsString()
	}

	return dep
}

// parseTerraformBlock parses a terraform block
func (p *Parser) parseTerraformBlock(block *hcl.Block) *TerraformBlock {
	tf := &TerraformBlock{}

	attrs, _ := block.Body.JustAttributes()
	if sourceAttr, exists := attrs["source"]; exists {
		val, _ := sourceAttr.Expr.Value(nil)
		sourceStr := val.AsString()
		tf.Source = sourceStr

		// Check if source is external (git:: or http(s)://)
		if p.isExternalSource(sourceStr) {
			tf.IsExternal = true
			// Parse external source information
			if info, err := p.sourceParser.Parse(sourceStr); err == nil {
				tf.ExternalInfo = info
			}
		}
	}

	return tf
}

// isExternalSource checks if a Terraform source is pointing to an external module
// Following DRY principle: single location for external source detection
func (p *Parser) isExternalSource(source string) bool {
	return strings.HasPrefix(source, "git::") ||
		strings.HasPrefix(source, "http://") ||
		strings.HasPrefix(source, "https://")
}

// parseIncludeBlock parses an include block
func (p *Parser) parseIncludeBlock(block *hcl.Block) *IncludeBlock {
	inc := &IncludeBlock{}

	attrs, _ := block.Body.JustAttributes()
	if pathAttr, exists := attrs["path"]; exists {
		val, _ := pathAttr.Expr.Value(nil)
		inc.Path = val.AsString()
	}

	return inc
}

// FindDependencies finds all dependencies for a module
func (p *Parser) FindDependencies(modulePath string) ([]string, error) {
	config, err := p.ParseFile(filepath.Join(modulePath, "terragrunt.hcl"))
	if err != nil {
		return nil, err
	}

	deps := make([]string, len(config.Dependencies))
	for i, dep := range config.Dependencies {
		deps[i] = dep.ConfigPath
	}

	return deps, nil
}

// GetAllModules recursively finds all Terragrunt modules in a directory
func (p *Parser) GetAllModules(rootDir string) ([]string, error) {
	var modules []string

	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && info.Name() == "terragrunt.hcl" {
			moduleDir := filepath.Dir(path)
			modules = append(modules, moduleDir)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk directory: %w", err)
	}

	return modules, nil
}

// GetExternalModuleSources extracts external module sources from a terragrunt.hcl file
// Returns both the source string and parsed information
// Following Single Responsibility: extract external sources only
func (p *Parser) GetExternalModuleSources(modulePath string) ([]module.SourceInfo, error) {
	config, err := p.ParseFile(filepath.Join(modulePath, "terragrunt.hcl"))
	if err != nil {
		return nil, err
	}

	var externalSources []module.SourceInfo

	if config.Terraform != nil && config.Terraform.IsExternal && config.Terraform.ExternalInfo != nil {
		externalSources = append(externalSources, *config.Terraform.ExternalInfo)
	}

	return externalSources, nil
}
