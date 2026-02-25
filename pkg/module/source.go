package module

import (
	"context"
	"fmt"
	"net/url"
	"strings"
)

// ModuleSource defines the interface for external Terraform module sources
// Following the Interface Segregation Principle (SOLID)
type ModuleSource interface {
	// Parse extracts repository information from a source string
	Parse(source string) (*SourceInfo, error)

	// FetchVersion retrieves the version/tag/commit in a source
	FetchVersion(ctx context.Context, source string) (string, error)

	// Checkout clones/updates the module to a working directory
	Checkout(ctx context.Context, source string, workdir string) error

	// Type returns the source type (git, http, etc)
	Type() string
}

// SourceInfo contains parsed information about a module source
type SourceInfo struct {
	Type       string // "git" or "http"
	URL        string // Repository URL
	SubPath    string // Path within repo (e.g., //vpc)
	Ref        string // Branch, tag, or commit
	RawSource  string // Original source string
}

// SourceParser creates appropriate ModuleSource implementations
// Following the Dependency Inversion Principle (SOLID)
type SourceParser struct {
	sources map[string]ModuleSource
}

// NewSourceParser creates a new source parser with default implementations
func NewSourceParser() *SourceParser {
	return &SourceParser{
		sources: map[string]ModuleSource{
			"git":  NewGitModuleSource(),
			"http": NewHTTPModuleSource(),
		},
	}
}

// Parse determines the source type and delegates parsing
// This avoids code duplication (DRY principle)
func (sp *SourceParser) Parse(sourceStr string) (*SourceInfo, error) {
	if sourceStr == "" {
		return nil, fmt.Errorf("empty source string")
	}

	// Detect source type from prefix
	sourceType := sp.detectSourceType(sourceStr)
	if sourceType == "" {
		return nil, fmt.Errorf("unknown source type: %s", sourceStr)
	}

	parser, exists := sp.sources[sourceType]
	if !exists {
		return nil, fmt.Errorf("no parser for source type: %s", sourceType)
	}

	info, err := parser.Parse(sourceStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse %s source: %w", sourceType, err)
	}

	return info, nil
}

// detectSourceType identifies the source type from the source string
// Following DRY: single location for type detection logic
func (sp *SourceParser) detectSourceType(source string) string {
	if strings.HasPrefix(source, "git::") {
		return "git"
	}
	if strings.HasPrefix(source, "https://") || strings.HasPrefix(source, "http://") {
		return "http"
	}
	return ""
}

// isValidURL validates if a string is a valid HTTP URL
func isValidURL(urlStr string) error {
	_, err := url.Parse(urlStr)
	return err
}

// extractSubPath extracts the subpath from a Terraform source
// e.g., "git::https://github.com/org/repo//vpc" -> returns "/vpc"
func extractSubPath(source string) string {
	if idx := strings.LastIndex(source, "//"); idx != -1 && idx < len(source)-2 {
		subPath := source[idx+1:] // Remove the "//"
		// Remove query parameters if present
		if qIdx := strings.Index(subPath, "?"); qIdx != -1 {
			subPath = subPath[:qIdx]
		}
		return "/" + subPath
	}
	return ""
}

// extractQueryParam extracts a specific query parameter from a source string
func extractQueryParam(source string, paramName string) string {
	if idx := strings.Index(source, "?"); idx != -1 {
		query := source[idx+1:]
		params := strings.Split(query, "&")
		for _, param := range params {
			parts := strings.Split(param, "=")
			if len(parts) == 2 && parts[0] == paramName {
				return parts[1]
			}
		}
	}
	return ""
}
