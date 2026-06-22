// Package git provides a wrapper around git commands to identify changed files.
package git

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/Ops-Talks/cultivator/internal/logging"
)

// ErrBaseRefNotFound is returned by GetChangedFiles when none of the candidate
// base refs can be resolved locally (typical of CI environments that do not
// fetch the destination branch by default, such as Bitbucket Pipelines).
// Callers can use errors.Is to detect this case and offer a recovery path,
// for example by fetching the missing branch and retrying.
var ErrBaseRefNotFound = errors.New("base ref not found locally")

// GetChangedFiles returns a list of files that have changed since the baseRef.
// It uses 'git diff --name-only <baseRef>' to identify changes.
// BaseRef can be a branch name (e.g., 'main') or a commit hash.
//
// When every candidate ref fails because git cannot resolve it locally, the
// returned error wraps ErrBaseRefNotFound.
func GetChangedFiles(ctx context.Context, workingDir string, baseRef string, logger *logging.Logger) ([]string, error) {
	baseRef = strings.TrimSpace(baseRef)
	if baseRef == "" {
		baseRef = "HEAD"
	}
	repoRoot, err := gitRepoRoot(ctx, workingDir)
	if err != nil {
		return nil, fmt.Errorf("resolve git repository root: %w", err)
	}

	baseRefs := buildBaseRefCandidates(baseRef)
	debugLog(logger, "collecting changed files", logging.Fields{
		"working_dir": workingDir,
		"repo_root":   repoRoot,
		"base_ref":    baseRef,
		"candidates":  strings.Join(baseRefs, ","),
	})

	var (
		output             []byte
		lastErr            error
		usedRef            string
		allUnknownRevision = true
	)
	for _, ref := range baseRefs {
		out, err := gitDiffNameOnly(ctx, workingDir, ref)
		if err != nil {
			lastErr = err
			if !isUnknownRevisionErr(err) {
				allUnknownRevision = false
			}
			debugLog(logger, "git diff failed for candidate base ref", logging.Fields{
				"base_ref": ref,
				"error":    err.Error(),
			})
			continue
		}
		output = out
		usedRef = ref
		allUnknownRevision = false
		break
	}
	if output == nil {
		if allUnknownRevision {
			return nil, fmt.Errorf("%w: tried %s: %w", ErrBaseRefNotFound, strings.Join(baseRefs, ","), lastErr)
		}
		return nil, fmt.Errorf("git diff failed for base refs %s: %w", strings.Join(baseRefs, ","), lastErr)
	}

	lines := strings.Split(string(output), "\n")
	var changedFiles []string
	for _, line := range lines {
		clean := strings.TrimSpace(line)
		if clean == "" {
			continue
		}
		// Git returns paths relative to the repository root.
		absPath := filepath.Join(repoRoot, clean)
		changedFiles = append(changedFiles, absPath)
	}
	debugLog(logger, "changed files collected", logging.Fields{
		"count":         len(changedFiles),
		"resolved_base": usedRef,
	})

	return changedFiles, nil
}

// FetchRemoteBranch runs 'git fetch --no-tags <remote> <branch>' in workingDir.
// It is intended as a recovery hook for CI environments (notably Bitbucket
// Pipelines) where the PR destination branch is not present locally.
// The command output is captured and returned as part of the error on failure
// so callers can surface a useful diagnostic.
func FetchRemoteBranch(ctx context.Context, workingDir, remote, branch string, logger *logging.Logger) error {
	remote = strings.TrimSpace(remote)
	branch = strings.TrimSpace(branch)
	if remote == "" {
		return fmt.Errorf("remote is required")
	}
	if branch == "" {
		return fmt.Errorf("branch is required")
	}

	debugLog(logger, "fetching remote branch", logging.Fields{
		"working_dir": workingDir,
		"remote":      remote,
		"branch":      branch,
	})

	// #nosec G204 -- remote and branch are validated above and supplied by trusted callers.
	cmd := exec.CommandContext(ctx, "git", "fetch", "--no-tags", remote, branch)
	cmd.Dir = workingDir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git fetch %s %s: %w: %s", remote, branch, err, strings.TrimSpace(string(out)))
	}
	return nil
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

// isUnknownRevisionErr reports whether err corresponds to git failing to
// resolve the ref (as opposed to a permission/IO/transport failure). It
// inspects the captured stderr from *exec.ExitError because git reports
// missing refs with messages such as "fatal: ambiguous argument 'origin/X':
// unknown revision or path not in the working tree." and exits with code 128.
func isUnknownRevisionErr(err error) bool {
	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) {
		return false
	}
	stderr := strings.ToLower(string(exitErr.Stderr))
	if stderr == "" {
		return false
	}
	switch {
	case strings.Contains(stderr, "unknown revision"):
		return true
	case strings.Contains(stderr, "ambiguous argument"):
		return true
	case strings.Contains(stderr, "bad revision"):
		return true
	case strings.Contains(stderr, "not a valid object name"):
		return true
	}
	return false
}

func gitRepoRoot(ctx context.Context, workingDir string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", "rev-parse", "--show-toplevel")
	cmd.Dir = workingDir
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git rev-parse --show-toplevel failed: %w", err)
	}

	root := strings.TrimSpace(string(output))
	if root == "" {
		return "", fmt.Errorf("git repository root is empty")
	}

	return filepath.Clean(root), nil
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

