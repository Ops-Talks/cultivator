// Package git provides a wrapper around git commands to identify changed files.
package git

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/Ops-Talks/cultivator/internal/logging"
)

// GetChangedFiles returns a list of files that have changed since the baseRef.
// It uses 'git diff --name-only <baseRef>' to identify changes.
// BaseRef can be a branch name (e.g., 'main') or a commit hash.
func GetChangedFiles(ctx context.Context, workingDir string, baseRef string, logger *logging.Logger) ([]string, error) {
	baseRef = strings.TrimSpace(baseRef)
	if baseRef == "" {
		baseRef = "HEAD"
	}
	baseRefs := buildBaseRefCandidates(baseRef)
	debugLog(logger, "collecting changed files", logging.Fields{
		"working_dir": workingDir,
		"base_ref":    baseRef,
		"candidates":  strings.Join(baseRefs, ","),
	})

	var (
		output  []byte
		lastErr error
		usedRef string
	)
	for _, ref := range baseRefs {
		out, err := gitDiffNameOnly(ctx, workingDir, ref)
		if err != nil {
			lastErr = err
			debugLog(logger, "git diff failed for candidate base ref", logging.Fields{
				"base_ref": ref,
				"error":    err.Error(),
			})
			continue
		}
		output = out
		usedRef = ref
		break
	}
	if output == nil {
		return nil, fmt.Errorf("git diff failed for base refs %s: %w", strings.Join(baseRefs, ","), lastErr)
	}

	lines := strings.Split(string(output), "\n")
	var changedFiles []string
	for _, line := range lines {
		clean := strings.TrimSpace(line)
		if clean == "" {
			continue
		}
		// Git returns relative paths from the repo root.
		// We resolve them to absolute paths for consistency.
		absPath := filepath.Join(workingDir, clean)
		changedFiles = append(changedFiles, absPath)
	}
	debugLog(logger, "changed files collected", logging.Fields{
		"count":         len(changedFiles),
		"resolved_base": usedRef,
	})

	return changedFiles, nil
}

func buildBaseRefCandidates(baseRef string) []string {
	candidates := []string{baseRef}
	if shouldTryOriginRef(baseRef) {
		candidates = append(candidates, "origin/"+baseRef)
		candidates = append(candidates, "refs/remotes/origin/"+baseRef)
	}
	return dedupeCandidates(candidates)
}

func shouldTryOriginRef(baseRef string) bool {
	if baseRef == "HEAD" {
		return false
	}
	if strings.HasPrefix(baseRef, "origin/") || strings.HasPrefix(baseRef, "refs/") {
		return false
	}
	return !strings.Contains(baseRef, "/")
}

func dedupeCandidates(candidates []string) []string {
	seen := make(map[string]struct{}, len(candidates))
	result := make([]string, 0, len(candidates))
	for _, candidate := range candidates {
		if candidate == "" {
			continue
		}
		if _, ok := seen[candidate]; ok {
			continue
		}
		seen[candidate] = struct{}{}
		result = append(result, candidate)
	}
	return result
}

func gitDiffNameOnly(ctx context.Context, workingDir, baseRef string) ([]byte, error) {
	// #nosec G204 -- baseRef is expected to be a valid git reference
	cmd := exec.CommandContext(ctx, "git", "diff", "--name-only", baseRef)
	cmd.Dir = workingDir
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git diff failed: %w", err)
	}
	return output, nil
}

// IsGitRepo checks if the given directory is part of a git repository.
func IsGitRepo(ctx context.Context, workingDir string, logger *logging.Logger) bool {
	cmd := exec.CommandContext(ctx, "git", "rev-parse", "--is-inside-work-tree")
	cmd.Dir = workingDir
	err := cmd.Run()
	if err != nil {
		debugLog(logger, "directory is not a git repo", logging.Fields{"working_dir": workingDir, "error": err.Error()})
	}
	return err == nil
}

func debugLog(logger *logging.Logger, msg string, fields logging.Fields) {
	if logger == nil {
		return
	}
	logger.Debug(msg, fields)
}
