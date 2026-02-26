package github

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFormatPlanOutput(t *testing.T) {
	output := FormatPlanOutput("env/dev/vpc", "Plan output here", true)

	assert.Contains(t, output, "Plan Results")
	assert.Contains(t, output, "env/dev/vpc")
	assert.Contains(t, output, "Changes detected")
	assert.Contains(t, output, "Plan output here")
}

func TestFormatApplyOutput(t *testing.T) {
	output := FormatApplyOutput("env/dev/vpc", "Apply output here", true)

	assert.Contains(t, output, "Apply Results")
	assert.Contains(t, output, "env/dev/vpc")
	assert.Contains(t, output, "Applied successfully")
	assert.Contains(t, output, "Apply output here")
}
