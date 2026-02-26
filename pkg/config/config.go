// Package config handles configuration loading and validation.
package config

import (
	"os"
	"time"

	custErrors "github.com/cultivator-dev/cultivator/pkg/errors"
	"github.com/cultivator-dev/cultivator/pkg/lock"
	yaml "gopkg.in/yaml.v3"
)

// Default configuration values (constants for DRY principle)
const (
	defaultVersion     = 1
	defaultMaxParallel = 5
)

// Config represents the cultivator configuration
type Config struct {
	Version  int                 `yaml:"version"`
	Projects []ProjectConfig     `yaml:"projects"`
	Settings GlobalSettings      `yaml:"settings"`
	Hooks    map[string][]string `yaml:"hooks,omitempty"`
}

// ProjectConfig represents a single project/directory to manage
type ProjectConfig struct {
	Name              string   `yaml:"name"`
	Dir               string   `yaml:"dir"`
	TerragruntVersion string   `yaml:"terragrunt_version,omitempty"`
	TerraformVersion  string   `yaml:"terraform_version,omitempty"`
	Workflow          string   `yaml:"workflow,omitempty"`
	ApplyRequirements []string `yaml:"apply_requirements,omitempty"`
	AutoPlan          *bool    `yaml:"auto_plan,omitempty"`
}

// GlobalSettings represents global configuration settings
type GlobalSettings struct {
	AutoPlan        bool          `yaml:"auto_plan"`
	LockTimeout     time.Duration `yaml:"-"`
	LockTimeoutStr  string        `yaml:"lock_timeout,omitempty"`
	ParallelPlan    bool          `yaml:"parallel_plan"`
	MaxParallel     int           `yaml:"max_parallel,omitempty"`
	RequireApproval bool          `yaml:"require_approval"`
}

// LoadConfig loads the cultivator configuration from a file
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, custErrors.NewFileError("read", path, err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, custErrors.NewParseError("YAML configuration", err).
			WithContext("file_path", path)
	}

	// Apply defaults and validate
	config.applyDefaults()

	return &config, nil
}

// applyDefaults sets default values for configuration (DRY principle - centralized defaults)
func (c *Config) applyDefaults() {
	if c.Version == 0 {
		c.Version = defaultVersion
	}

	// Parse lock timeout
	c.Settings.LockTimeout = lock.ParseDuration(c.Settings.LockTimeoutStr)
	if c.Settings.MaxParallel == 0 {
		c.Settings.MaxParallel = defaultMaxParallel
	}
}

// GetProject returns the project config for a given directory
func (c *Config) GetProject(dir string) *ProjectConfig {
	for i := range c.Projects {
		if c.Projects[i].Dir == dir {
			return &c.Projects[i]
		}
	}
	return nil
}
