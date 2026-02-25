package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadConfig(t *testing.T) {
	// Create a temporary config file
	configContent := `version: 1

projects:
  - name: test-project
    dir: test/dir
    terragrunt_version: 0.55.0
    terraform_version: 1.7.0

settings:
  auto_plan: true
  parallel_plan: true
  max_parallel: 5
  lock_timeout: 10m
`

	tmpFile, err := os.CreateTemp("", "cultivator-test-*.yml")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(configContent)
	require.NoError(t, err)
	tmpFile.Close()

	// Load config
	cfg, err := LoadConfig(tmpFile.Name())
	require.NoError(t, err)

	assert.Equal(t, 1, cfg.Version)
	assert.Len(t, cfg.Projects, 1)
	assert.Equal(t, "test-project", cfg.Projects[0].Name)
	assert.Equal(t, "test/dir", cfg.Projects[0].Dir)
	assert.True(t, cfg.Settings.AutoPlan)
	assert.Equal(t, 5, cfg.Settings.MaxParallel)
}

func TestGetProject(t *testing.T) {
	cfg := &Config{
		Projects: []ProjectConfig{
			{Name: "proj1", Dir: "dir1"},
			{Name: "proj2", Dir: "dir2"},
		},
	}

	proj := cfg.GetProject("dir1")
	require.NotNil(t, proj)
	assert.Equal(t, "proj1", proj.Name)

	proj = cfg.GetProject("nonexistent")
	assert.Nil(t, proj)
}
