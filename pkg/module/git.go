// Package module provides module source implementations for git and HTTP repositories.
package module

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	custErrors "github.com/cultivator-dev/cultivator/pkg/errors"
)

// GitModuleSource handles Terraform sources from Git repositories
// Single Responsibility: handles only git:// sources
type GitModuleSource struct{}

// NewGitModuleSource creates a new Git module source handler
func NewGitModuleSource() *GitModuleSource {
	return &GitModuleSource{}
}

// Type returns the source type identifier
func (g *GitModuleSource) Type() string {
	return "git"
}

// Parse extracts repository information from a git:: source
// Example: git::https://github.com/org/repo//vpc?ref=v1.0.0
// Example: git::ssh://git@github.com/org/repo.git//database?ref=main
func (g *GitModuleSource) Parse(source string) (*SourceInfo, error) {
	// Remove "git::" prefix
	if !strings.HasPrefix(source, "git::") {
		return nil, custErrors.NewValidationError("git source must start with 'git::'").
			WithContext("source", source)
	}

	sourceWithoutPrefix := strings.TrimPrefix(source, "git::")

	// Extract subpath (everything after //)
	subPath := extractSubPath(sourceWithoutPrefix)

	// Extract reference (branch, tag, or commit)
	ref := extractQueryParam(sourceWithoutPrefix, "ref")
	if ref == "" {
		ref = "HEAD" // Default to HEAD if not specified
	}

	// Remove subpath and query params to get the repository URL
	repoURL := g.extractRepositoryURL(sourceWithoutPrefix)
	if repoURL == "" {
		return nil, custErrors.NewValidationError(fmt.Sprintf("invalid git source: %s", source)).
			WithContext("source", source)
	}

	return &SourceInfo{
		Type:      "git",
		URL:       repoURL,
		SubPath:   subPath,
		Ref:       ref,
		RawSource: source,
	}, nil
}

// FetchVersion retrieves the current version (commit SHA) from the remote repository
// Without cloning the entire repo (efficient)
func (g *GitModuleSource) FetchVersion(ctx context.Context, source string) (string, error) {
	info, err := g.Parse(source)
	if err != nil {
		return "", err
	}

	// Use git ls-remote to get the commit SHA for the given ref
	cmd := exec.CommandContext(ctx, "git", "ls-remote", "--resolve", info.URL, info.Ref)
	output, err := cmd.Output()
	if err != nil {
		return "", custErrors.NewExternalError("git ls-remote", err).
			WithContext("url", info.URL).
			WithContext("ref", info.Ref)
	}

	// Output format: "<SHA>\t<ref>"
	parts := strings.Fields(string(output))
	if len(parts) == 0 {
		return "", custErrors.NewValidationError("no output from git ls-remote").
			WithContext("url", info.URL).
			WithContext("ref", info.Ref)
	}

	return parts[0], nil // Return the commit SHA
}

// Checkout clones or updates the git repository to the specified working directory
func (g *GitModuleSource) Checkout(ctx context.Context, source string, workdir string) error {
	info, err := g.Parse(source)
	if err != nil {
		return err
	}

	// Clone with the specified depth to save bandwidth (DRY: logic in one place)
	return g.gitClone(ctx, info.URL, info.Ref, workdir)
}

// gitClone performs the actual git clone operation
// Extracted method to follow DRY principle and improve testability
func (g *GitModuleSource) gitClone(ctx context.Context, repoURL, ref, workdir string) error {
	// Clone with depth=1 for efficiency, then checkout the specific ref
	cmd := exec.CommandContext(ctx, "git", "clone", "--depth", "1", "--branch", ref, repoURL, workdir)
	if err := cmd.Run(); err != nil {
		// Fallback: try without depth for refs that might not be recognized as branches
		cmd = exec.CommandContext(ctx, "git", "clone", repoURL, workdir)
		if cloneErr := cmd.Run(); cloneErr != nil {
			return fmt.Errorf("failed to clone repository: %w", cloneErr)
		}

		// Checkout the specific ref
		cmd = exec.CommandContext(ctx, "git", "-C", workdir, "checkout", ref)
		if checkoutErr := cmd.Run(); checkoutErr != nil {
			return fmt.Errorf("failed to checkout ref %s: %w", ref, checkoutErr)
		}
	}

	return nil
}

// extractRepositoryURL removes subpath and query parameters from a git URL
// This is a helper method to keep Parse() clean (DRY + Clean Code)
func (g *GitModuleSource) extractRepositoryURL(sourceWithoutPrefix string) string {
	// Find where subpath or query starts
	subPathIdx := strings.LastIndex(sourceWithoutPrefix, "//")
	queryIdx := strings.Index(sourceWithoutPrefix, "?")

	endIdx := len(sourceWithoutPrefix)

	// Use the earliest boundary (subpath or query)
	if subPathIdx != -1 {
		endIdx = subPathIdx
	}
	if queryIdx != -1 && queryIdx < endIdx {
		endIdx = queryIdx
	}

	url := sourceWithoutPrefix[:endIdx]

	// Ensure .git extension for SSH URLs if not present
	if strings.HasPrefix(url, "ssh://") && !strings.HasSuffix(url, ".git") {
		url = url + ".git"
	}

	return url
}
