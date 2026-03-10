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
	if baseRef == "" {
		baseRef = "HEAD"
	}
	debugLog(logger, "collecting changed files", logging.Fields{"working_dir": workingDir, "base_ref": baseRef})

	// #nosec G204 -- baseRef is expected to be a valid git reference
	cmd := exec.CommandContext(ctx, "git", "diff", "--name-only", baseRef)
	cmd.Dir = workingDir

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git diff failed: %w", err)
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
	debugLog(logger, "changed files collected", logging.Fields{"count": len(changedFiles)})

	return changedFiles, nil
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
