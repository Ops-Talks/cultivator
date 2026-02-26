// Package parser reads and parses Terragrunt HCL configuration files.
package parser

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/cultivator-dev/cultivator/pkg/constants"
	custErrors "github.com/cultivator-dev/cultivator/pkg/errors"
	"github.com/cultivator-dev/cultivator/pkg/module"
	hcl "github.com/hashicorp/hcl/v2"
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
	Source       string             // Local or external module source
	ExternalInfo *module.SourceInfo // Parsed info if source is external (git:: or http://)
	IsExternal   bool               // Whether this source is external
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

// ParseFile parses a terragrunt.hcl file with proper error handling
func (p *Parser) ParseFile(path string) (*TerragruntConfig, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, custErrors.NewFileError("get absolute path for", path, err)
	}

	content, err := os.ReadFile(absPath)
	if err != nil {
		return nil, custErrors.NewFileError("read", absPath, err)
	}

	file, diags := p.parser.ParseHCL(content, absPath)
	if diags.HasErrors() {
		return nil, custErrors.NewParseError("HCL file", fmt.Errorf("%v", diags)). //nolint:errorlint
												WithContext("file_path", absPath)
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
			{Type: constants.BlockTypeDependency},
			{Type: constants.BlockTypeTerraform},
			{Type: constants.BlockTypeInclude},
			{Type: constants.BlockTypeInputs},
		},
	})

	if diags.HasErrors() {
		return custErrors.NewParseError("HCL body", fmt.Errorf("%v", diags)) //nolint:errorlint
	}

	// Parse dependency blocks
	for _, block := range content.Blocks {
		switch block.Type {
		case constants.BlockTypeDependency:
			dep := p.parseDependencyBlock(block)
			if dep != nil {
				config.Dependencies = append(config.Dependencies, *dep)
			}
		case constants.BlockTypeTerraform:
			config.Terraform = p.parseTerraformBlock(block)
		case constants.BlockTypeInclude:
			config.Include = p.parseIncludeBlock(block)
		}
	}

	return nil
}

// parseDependencyBlock parses a dependency block with proper error handling
func (p *Parser) parseDependencyBlock(block *hcl.Block) *Dependency {
	if len(block.Labels) == 0 {
		return nil
	}

	dep := &Dependency{
		Name: block.Labels[0],
	}

	attrs, diaG := block.Body.JustAttributes()
	if diaG.HasErrors() {
		// Log error but continue - non-critical failure
		return dep
	}

	if configPathAttr, exists := attrs[constants.AttrConfigPath]; exists {
		val, diaG := configPathAttr.Expr.Value(nil)
		if !diaG.HasErrors() {
			dep.ConfigPath = val.AsString()
		}
	}

	return dep
}

// parseTerraformBlock parses a terraform block with proper null safety
func (p *Parser) parseTerraformBlock(block *hcl.Block) *TerraformBlock {
	if block == nil {
		return nil
	}

	tf := &TerraformBlock{}

	attrs, diag := block.Body.JustAttributes()
	if diag.HasErrors() {
		return tf
	}

	if sourceAttr, exists := attrs[constants.AttrSource]; exists {
		val, diag := sourceAttr.Expr.Value(nil)
		if !diag.HasErrors() {
			sourceStr := val.AsString()
			tf.Source = sourceStr

			// Check if source is external (git:: or http(s)://)
			if p.isExternalSource(sourceStr) {
				tf.IsExternal = true
				// Parse external source information - non-critical failure
				if info, err := p.sourceParser.Parse(sourceStr); err == nil {
					tf.ExternalInfo = info
				}
			}
		}
	}

	return tf
}

// isExternalSource checks if a Terraform source is pointing to an external module
// Following DRY principle: single location for external source detection
func (p *Parser) isExternalSource(source string) bool {
	return strings.HasPrefix(source, constants.GitPrefix) ||
		strings.HasPrefix(source, constants.HTTPPrefix) ||
		strings.HasPrefix(source, constants.HTTPSPrefix)
}

// parseIncludeBlock parses an include block
func (p *Parser) parseIncludeBlock(block *hcl.Block) *IncludeBlock {
	if block == nil {
		return nil
	}

	inc := &IncludeBlock{}

	attrs, diag := block.Body.JustAttributes()
	if diag.HasErrors() {
		return inc
	}

	if pathAttr, exists := attrs[constants.AttrPath]; exists {
		val, diag := pathAttr.Expr.Value(nil)
		if !diag.HasErrors() {
			inc.Path = val.AsString()
		}
	}

	return inc
}

// FindDependencies finds all dependencies for a module
func (p *Parser) FindDependencies(modulePath string) ([]string, error) {
	configPath := filepath.Join(modulePath, constants.TerragruntConfigFile)
	config, err := p.ParseFile(configPath)
	if err != nil {
		return nil, custErrors.NewParseError(
			fmt.Sprintf("dependencies for module %s", modulePath), err,
		).WithContext("module_path", modulePath)
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
			return custErrors.NewFileError("walk", path, err)
		}

		// Check if file is a Terragrunt config file
		if !info.IsDir() && info.Name() == constants.TerragruntConfigFile {
			moduleDir := filepath.Dir(path)
			modules = append(modules, moduleDir)
		}

		return nil
	})

	if err != nil {
		return nil, custErrors.NewFileError("walk", rootDir, err)
	}

	return modules, nil
}

// GetExternalModuleSources extracts external module sources from a terragrunt.hcl file
// Returns both the source string and parsed information
// Following Single Responsibility: extract external sources only
func (p *Parser) GetExternalModuleSources(modulePath string) ([]module.SourceInfo, error) {
	configPath := filepath.Join(modulePath, constants.TerragruntConfigFile)
	config, err := p.ParseFile(configPath)
	if err != nil {
		return nil, custErrors.NewParseError(
			fmt.Sprintf("external sources for module %s", modulePath), err,
		).WithContext("module_path", modulePath)
	}

	var externalSources []module.SourceInfo

	if config.Terraform != nil && config.Terraform.IsExternal && config.Terraform.ExternalInfo != nil {
		externalSources = append(externalSources, *config.Terraform.ExternalInfo)
	}

	return externalSources, nil
}
