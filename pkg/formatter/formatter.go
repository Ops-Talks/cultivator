package formatter

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// Compiled regex patterns (cached for performance and DRY principle)
var (
	planSummaryRegex   = regexp.MustCompile(`Plan:\s+(\d+)\s+to\s+add,\s+(\d+)\s+to\s+change,\s+(\d+)\s+to\s+destroy`)
	ansiCodeRegex      = regexp.MustCompile(`\x1b\[[0-9;]*m`)
	multipleNewlinesRx = regexp.MustCompile(`\n{3,}`)

	// Sensitive data patterns (compiled once for performance)
	secretKeyValueRegex   = regexp.MustCompile(`(?i)\b(token|password|secret|access_key|secret_key|api_key|apikey)\b\s*[:=]\s*([^\s"']+)`)
	secretJSONRegex       = regexp.MustCompile(`(?i)("(?:token|password|secret|access_key|secret_key|api_key|apikey)"\s*:\s*)"[^"]+"`)
	awsAccessKeyRegex     = regexp.MustCompile(`AKIA[0-9A-Z]{16}`)
	awsSessionTokenRegex  = regexp.MustCompile(`ASIA[0-9A-Z]{16}`)
	githubTokenRegex      = regexp.MustCompile(`gh[pousr]_[A-Za-z0-9_]{10,}`)
	base64LikeRegex       = regexp.MustCompile(`(?i)\b(token|password|secret|key)\b\s*[:=]\s*([A-Za-z0-9+/]{40,}={0,2})`)
	privateKeyHeaderRegex = regexp.MustCompile(`-----BEGIN (?:RSA |EC )?PRIVATE KEY-----`)
)

// Constants for formatting (DRY principle)
const (
	maxModulesDisplay   = 5
	changeSymbolAdd     = "+"
	changeSymbolChange  = "~"
	changeSymbolDestroy = "-"
	addFormatStr        = "**+%d** to add"
	changeFormatStr     = "**~%d** to change"
	destroyFormatStr    = "**-%d** to destroy"
	noChangesMsg        = "No changes"
	changesDetectedMsg  = "Changes detected"
	redactedPlaceholder = "[REDACTED]"
)

// redactionRule defines a pattern and its replacement strategy
type redactionRule struct {
	name    string
	pattern *regexp.Regexp
	replace string
}

// getSensitiveDataRules returns the list of redaction rules (DRY principle)
func getSensitiveDataRules() []redactionRule {
	return []redactionRule{
		{
			name:    "key-value pairs",
			pattern: secretKeyValueRegex,
			replace: `$1=` + redactedPlaceholder,
		},
		{
			name:    "JSON secrets",
			pattern: secretJSONRegex,
			replace: `$1"` + redactedPlaceholder + `"`,
		},
		{
			name:    "AWS access keys",
			pattern: awsAccessKeyRegex,
			replace: redactedPlaceholder,
		},
		{
			name:    "AWS session tokens",
			pattern: awsSessionTokenRegex,
			replace: redactedPlaceholder,
		},
		{
			name:    "GitHub tokens",
			pattern: githubTokenRegex,
			replace: redactedPlaceholder,
		},
		{
			name:    "Base64-like secrets",
			pattern: base64LikeRegex,
			replace: `$1=` + redactedPlaceholder,
		},
		{
			name:    "Private key headers",
			pattern: privateKeyHeaderRegex,
			replace: redactedPlaceholder,
		},
	}
}

// PlanSummary extracts a summary from Terraform plan output
type PlanSummary struct {
	ToAdd      int
	ToChange   int
	ToDestroy  int
	HasChanges bool
}

// ParsePlanOutput parses Terraform/Terragrunt plan output to extract summary
func ParsePlanOutput(output string) PlanSummary {
	summary := PlanSummary{}

	matches := planSummaryRegex.FindStringSubmatch(output)
	if len(matches) == 4 {
		// Use strconv for better performance than fmt.Sscanf
		summary.ToAdd, _ = strconv.Atoi(matches[1])
		summary.ToChange, _ = strconv.Atoi(matches[2])
		summary.ToDestroy, _ = strconv.Atoi(matches[3])
		summary.HasChanges = summary.ToAdd > 0 || summary.ToChange > 0 || summary.ToDestroy > 0
		return summary
	}

	// Check for "No changes" message
	if strings.Contains(output, noChangesMsg) || strings.Contains(output, "no changes") {
		summary.HasChanges = false
		return summary
	}

	// If we can't parse, assume there might be changes
	summary.HasChanges = true
	return summary
}

// String formats the plan summary as a Markdown string (implements Stringer interface)
func (s PlanSummary) String() string {
	if !s.HasChanges {
		return noChangesMsg
	}

	return s.buildChangeSummary()
}

// buildChangeSummary constructs the formatted change summary (extracted method - DRY principle)
func (s PlanSummary) buildChangeSummary() string {
	var sb strings.Builder

	if s.ToAdd > 0 {
		if sb.Len() > 0 {
			sb.WriteString(", ")
		}
		fmt.Fprintf(&sb, addFormatStr, s.ToAdd)
	}

	if s.ToChange > 0 {
		if sb.Len() > 0 {
			sb.WriteString(", ")
		}
		fmt.Fprintf(&sb, changeFormatStr, s.ToChange)
	}

	if s.ToDestroy > 0 {
		if sb.Len() > 0 {
			sb.WriteString(", ")
		}
		fmt.Fprintf(&sb, destroyFormatStr, s.ToDestroy)
	}

	if sb.Len() == 0 {
		return changesDetectedMsg
	}

	return sb.String()
}

// TruncateOutput truncates long output to a reasonable size for PR comments
func TruncateOutput(output string, maxLines int) string {
	lines := strings.Split(output, "\n")

	if len(lines) <= maxLines {
		return output
	}

	half := maxLines / 2
	return buildTruncatedOutput(lines, half, maxLines)
}

// buildTruncatedOutput constructs the truncated output (extracted method - DRY principle)
func buildTruncatedOutput(lines []string, half, maxLines int) string {
	var sb strings.Builder
	omittedCount := len(lines) - maxLines

	for i := 0; i < half; i++ {
		sb.WriteString(lines[i])
		sb.WriteString("\n")
	}

	fmt.Fprintf(&sb, "... (%d lines omitted) ...\n", omittedCount)

	for i := len(lines) - half; i < len(lines); i++ {
		sb.WriteString(lines[i])
		sb.WriteString("\n")
	}

	return strings.TrimSuffix(sb.String(), "\n")
}

// FormatModuleList formats a list of module paths as Markdown bullet list
func FormatModuleList(modules []string) string {
	if len(modules) == 0 {
		return "None"
	}

	if len(modules) <= maxModulesDisplay {
		return formatModuleItems(modules)
	}

	return formatModuleItemsWithMore(modules)
}

// formatModuleItems builds formatted list when count is small (extracted method - DRY principle)
func formatModuleItems(modules []string) string {
	var sb strings.Builder
	for _, m := range modules {
		fmt.Fprintf(&sb, "- `%s`\n", m)
	}
	return strings.TrimSuffix(sb.String(), "\n")
}

// formatModuleItemsWithMore builds truncated list showing first N items (extracted method - DRY principle)
func formatModuleItemsWithMore(modules []string) string {
	var sb strings.Builder
	for i := 0; i < maxModulesDisplay; i++ {
		fmt.Fprintf(&sb, "- `%s`\n", modules[i])
	}
	fmt.Fprintf(&sb, "- ... and %d more", len(modules)-maxModulesDisplay)
	return sb.String()
}

// CleanTerraformOutput removes ANSI color codes, redacts sensitive data, and normalizes whitespace
// Uses cached regex patterns for performance (follows established DRY principle)
func CleanTerraformOutput(output string) string {
	// Remove ANSI color codes using cached regex
	cleaned := ansiCodeRegex.ReplaceAllString(output, "")

	// Remove excessive blank lines using cached regex
	cleaned = multipleNewlinesRx.ReplaceAllString(cleaned, "\n\n")

	// Redact sensitive data before returning output
	cleaned = RedactSensitive(cleaned)

	return strings.TrimSpace(cleaned)
}

// RedactSensitive masks common secret patterns from output
// Applies all redaction rules using pre-compiled regex patterns (DRY + Performance)
func RedactSensitive(output string) string {
	redacted := output
	for _, rule := range getSensitiveDataRules() {
		redacted = rule.pattern.ReplaceAllString(redacted, rule.replace)
	}
	return redacted
}

// HighlightChanges preserves resource change markers without modification
// Note: This is a no-op method; consider enhancing or removing in future iterations
func HighlightChanges(output string) string {
	return output
}
