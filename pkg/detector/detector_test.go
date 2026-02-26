package detector

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestChangeDetector_isRelevantFile(t *testing.T) {
	detector := &ChangeDetector{}

	tests := []struct {
		name     string
		filename string
		expected bool
	}{
		{"terragrunt.hcl", "path/to/terragrunt.hcl", true},
		{"main.tf", "path/to/main.tf", true},
		{"variables.tfvars", "path/to/variables.tfvars", true},
		{"README.md", "path/to/README.md", false},
		{"config.yaml", "path/to/config.yaml", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detector.isRelevantFile(tt.filename)
			assert.Equal(t, tt.expected, result)
		})
	}
}
