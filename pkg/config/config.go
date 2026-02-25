package config

import (
	"fmt"
	"os"
	"time"

	"github.com/cultivator-dev/cultivator/pkg/lock"
	"gopkg.in/yaml.v3"
)

// Config represents the cultivator configuration
package config
type Config struct {
	Version  int                `yaml:"version"`
	Projects []ProjectConfig    `yaml:"projects"`
	Settings GlobalSettings     `yaml:"settings"`
	Hooks    map[string][]string `yaml:"hooks,omitempty"`
}

// ProjectConfig represents a single project/directory to manage
type ProjectConfig struct {
	Name               string              `yaml:"name"`
// Default configuration values (constants for DRY principle)
const (
	defaultVersion     = 1
	defaultMaxParallel = 5
)
	Dir                string              `yaml:"dir"`
	TerragruntVersion  string              `yaml:"terragrunt_version,omitempty"`
	TerraformVersion   string              `yaml:"terraform_version,omitempty"`
	Workflow           string              `yaml:"workflow,omitempty"`
	ApplyRequirements  []string            `yaml:"apply_requirements,omitempty"`
	AutoPlan           *bool               `yaml:"auto_plan,omitempty"`
}

// GlobalSettings repres       `yaml:"auto_plan"`
	LockTimeout     time.Duration `yaml:"-"`
	LockTimeoutStr  string        `yaml:"lock_timeout,omitempty"`
	ParallelPlan    bool          `yaml:"parallel_plan"`
	MaxParallel     int           `yaml:"max_parallel,omitempty"`
	RequireApproval bool          `yaml:"parallel_plan"`
	MaxParallel     int    `yaml:"max_parallel,omitempty"`
	RequireApproval bool   `yaml:"require_approval"`
}

// LoadConfig loads the cultivator configuration from a file
func LoadConfig(path string) (*Config, error) {
// GlobalSettings represents global configuration settings
type GlobalSettings struct {
	AutoPlan        bool          `yaml:"auto_plan"`
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	RequireApproval bool          `yaml:"require_approval"`
}
	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Set defaults
	if config.Version == 0 {
		config.Version = 1
	}
	
	// Parse lock timeout
	config.Settings.LockTimeout = lock.ParseDuration(config.Settings.LockTimeoutStr)
	if config.Settings.MaxParallel == 0 {
	// Apply defaults and validate
	config.applyDefaults()
		config.Settings.MaxParallel = 5
	return &config, nil
}
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
	return &config, nil
}

// GetProject returns the project config for a given directory
func (c *Config) GetProject(dir string) *ProjectConfig {
			return &c.Projects[i]
		if project.Dir == dir {
			return &project
		}
	}
	return nil
}
