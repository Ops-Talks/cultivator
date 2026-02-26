package formatter

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParsePlanOutput(t *testing.T) {
	tests := []struct {
		name     string
		output   string
		expected PlanSummary
	}{
		{
			name:   "with changes",
			output: "Plan: 3 to add, 2 to change, 1 to destroy.",
			expected: PlanSummary{
				ToAdd:      3,
				ToChange:   2,
				ToDestroy:  1,
				HasChanges: true,
			},
		},
		{
			name:   "no changes",
			output: "No changes. Infrastructure is up-to-date.",
			expected: PlanSummary{
				ToAdd:      0,
				ToChange:   0,
				ToDestroy:  0,
				HasChanges: false,
			},
		},
		{
			name:   "only additions",
			output: "Plan: 5 to add, 0 to change, 0 to destroy.",
			expected: PlanSummary{
				ToAdd:      5,
				ToChange:   0,
				ToDestroy:  0,
				HasChanges: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParsePlanOutput(tt.output)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestPlanSummary_String(t *testing.T) {
	tests := []struct {
		name     string
		summary  PlanSummary
		expected string
	}{
		{
			name: "all changes",
			summary: PlanSummary{
				ToAdd:      3,
				ToChange:   2,
				ToDestroy:  1,
				HasChanges: true,
			},
			expected: "**+3** to add, **~2** to change, **-1** to destroy",
		},
		{
			name: "no changes",
			summary: PlanSummary{
				HasChanges: false,
			},
			expected: "No changes",
		},
		{
			name: "only additions",
			summary: PlanSummary{
				ToAdd:      5,
				HasChanges: true,
			},
			expected: "**+5** to add",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.summary.String()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTruncateOutput(t *testing.T) {
	longOutput := ""
	for i := 0; i < 100; i++ {
		longOutput += "Line " + string(rune(i)) + "\n"
	}

	truncated := TruncateOutput(longOutput, 20)
	lines := len(strings.Split(truncated, "\n"))

	// Should have approximately maxLines (plus truncation notice)
	assert.Less(t, lines, 30)
	assert.Contains(t, truncated, "omitted")
}

func TestCleanTerraformOutput(t *testing.T) {
	input := "\x1b[32mGreen text\x1b[0m\n\n\n\nToo many newlines"
	expected := "Green text\n\nToo many newlines"

	result := CleanTerraformOutput(input)
	assert.Equal(t, expected, result)
}

func TestRedactSensitive(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		mustNotContain []string
		mustContain    []string
	}{
		{
			name: "key-value pairs",
			input: strings.Join([]string{
				"token=abc123",
				"password: supersecret",
				"api_key = mykey123",
			}, "\n"),
			mustNotContain: []string{"abc123", "supersecret", "mykey123"},
			mustContain:    []string{"token=[REDACTED]", "password=[REDACTED]", "api_key=[REDACTED]"},
		},
		{
			name:           "JSON secrets",
			input:          `{"api_key":"xyz789","other":"value"}`,
			mustNotContain: []string{"xyz789"},
			mustContain:    []string{`"api_key":"[REDACTED]"`},
		},
		{
			name:           "AWS access keys",
			input:          "AWS key: AKIA1234567890ABCDEF",
			mustNotContain: []string{"AKIA1234567890ABCDEF"},
			mustContain:    []string{"[REDACTED]"},
		},
		{
			name:           "AWS session tokens",
			input:          "Session token: ASIA1234567890ABCDEF",
			mustNotContain: []string{"ASIA1234567890ABCDEF"},
			mustContain:    []string{"[REDACTED]"},
		},
		{
			name:           "GitHub tokens",
			input:          "GitHub: ghp_abcdefghijklmnopqrstuvwxyz1234567890",
			mustNotContain: []string{"ghp_abcdefghijklmnopqrstuvwxyz1234567890"},
			mustContain:    []string{"[REDACTED]"},
		},
		{
			name:           "Base64-like secrets",
			input:          "secret=dGVzdHNlY3JldGRhdGF0aGF0aXNsb25nZW5vdWdodG9iZWJhc2U2NA==",
			mustNotContain: []string{"dGVzdHNlY3JldGRhdGF0aGF0aXNsb25nZW5vdWdodG9iZWJhc2U2NA=="},
			mustContain:    []string{"secret=[REDACTED]"},
		},
		{
			name:           "Private key headers",
			input:          "-----BEGIN RSA PRIVATE KEY-----\nMIIEowIBAAKCAQEA...",
			mustNotContain: []string{"BEGIN RSA PRIVATE KEY"},
			mustContain:    []string{"[REDACTED]"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			redacted := RedactSensitive(tt.input)

			for _, sensitiveValue := range tt.mustNotContain {
				assert.NotContains(t, redacted, sensitiveValue, "Sensitive value should be redacted")
			}

			for _, expectedValue := range tt.mustContain {
				assert.Contains(t, redacted, expectedValue, "Expected redaction marker not found")
			}
		})
	}
}
