package git

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestGetChangedFiles(t *testing.T) {
	t.Parallel()

	t.Run("uses local base ref", func(t *testing.T) {
		t.Parallel()
		repoDir, baseBranch := setupRepoWithFeatureChange(t)
		got, err := GetChangedFiles(context.Background(), repoDir, baseBranch, nil)
		if err != nil {
			t.Fatalf("GetChangedFiles() error = %v", err)
		}
		assertContainsBaseName(t, got, "file2.txt")
	})

	t.Run("falls back to origin base ref when local branch is missing", func(t *testing.T) {
		t.Parallel()
		repoDir, baseBranch := setupRepoWithFeatureChange(t)
		runCmd(t, repoDir, "git", "branch", "-D", baseBranch)

		got, err := GetChangedFiles(context.Background(), repoDir, baseBranch, nil)
		if err != nil {
			t.Fatalf("GetChangedFiles() error = %v", err)
		}
		assertContainsBaseName(t, got, "file2.txt")
	})

	t.Run("returns error for unknown base ref", func(t *testing.T) {
		t.Parallel()
		repoDir, _ := setupRepoWithFeatureChange(t)
		_, err := GetChangedFiles(context.Background(), repoDir, "definitely-not-a-real-ref", nil)
		if err == nil {
			t.Fatal("GetChangedFiles() error = nil, want non-nil")
		}
		if !strings.Contains(err.Error(), "git diff failed for base refs") {
			t.Fatalf("GetChangedFiles() error = %q, want base refs context", err.Error())
		}
	})

	t.Run("resolves changed paths from repository root when working dir is nested", func(t *testing.T) {
		t.Parallel()
		repoDir, baseBranch, changedRelPath := setupRepoWithProviderChange(t)

		got, err := GetChangedFiles(context.Background(), filepath.Join(repoDir, "providers"), baseBranch, nil)
		if err != nil {
			t.Fatalf("GetChangedFiles() error = %v", err)
		}

		want := filepath.Join(repoDir, changedRelPath)
		assertContainsPath(t, got, want)
		for _, changed := range got {
			if strings.Contains(changed, string(filepath.Separator)+"providers"+string(filepath.Separator)+"providers"+string(filepath.Separator)) {
				t.Fatalf("changed file path contains duplicated providers prefix: %q", changed)
			}
		}
	})
}

func runCmd(t *testing.T, dir string, name string, args ...string) {
	t.Helper()
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	if err := cmd.Run(); err != nil {
		t.Fatalf("command %s %v failed: %v", name, args, err)
	}
}

func runCmdOutput(t *testing.T, dir string, name string, args ...string) string {
	t.Helper()
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("command %s %v failed: %v", name, args, err)
	}
	return strings.TrimSpace(string(output))
}

func setupRepoWithFeatureChange(t *testing.T) (string, string) {
	t.Helper()
	tmpDir := t.TempDir()
	remoteDir := filepath.Join(tmpDir, "remote.git")
	runCmd(t, tmpDir, "git", "init", "--bare", remoteDir)

	repoDir := filepath.Join(tmpDir, "work")
	runCmd(t, tmpDir, "git", "clone", remoteDir, repoDir)
	runCmd(t, repoDir, "git", "config", "user.email", "test@example.com")
	runCmd(t, repoDir, "git", "config", "user.name", "test")

	file1 := filepath.Join(repoDir, "file1.txt")
	if err := os.WriteFile(file1, []byte("hello"), 0o644); err != nil {
		t.Fatalf("write file1: %v", err)
	}
	runCmd(t, repoDir, "git", "add", "file1.txt")
	runCmd(t, repoDir, "git", "commit", "-m", "initial commit")

	baseBranch := runCmdOutput(t, repoDir, "git", "branch", "--show-current")
	if baseBranch == "" {
		t.Fatal("base branch is empty")
	}
	runCmd(t, repoDir, "git", "push", "-u", "origin", baseBranch)

	runCmd(t, repoDir, "git", "checkout", "-b", "feature")
	file2 := filepath.Join(repoDir, "file2.txt")
	if err := os.WriteFile(file2, []byte("world"), 0o644); err != nil {
		t.Fatalf("write file2: %v", err)
	}
	runCmd(t, repoDir, "git", "add", "file2.txt")

	return repoDir, baseBranch
}

func setupRepoWithProviderChange(t *testing.T) (string, string, string) {
	t.Helper()
	tmpDir := t.TempDir()
	remoteDir := filepath.Join(tmpDir, "remote.git")
	runCmd(t, tmpDir, "git", "init", "--bare", remoteDir)

	repoDir := filepath.Join(tmpDir, "work")
	runCmd(t, tmpDir, "git", "clone", remoteDir, repoDir)
	runCmd(t, repoDir, "git", "config", "user.email", "test@example.com")
	runCmd(t, repoDir, "git", "config", "user.name", "test")

	changedRelPath := filepath.Join("providers", "aws", "dev", "terragrunt.hcl")
	changedAbsPath := filepath.Join(repoDir, changedRelPath)
	if err := os.MkdirAll(filepath.Dir(changedAbsPath), 0o755); err != nil {
		t.Fatalf("mkdir providers tree: %v", err)
	}
	if err := os.WriteFile(changedAbsPath, []byte("terraform {}\n"), 0o644); err != nil {
		t.Fatalf("write initial terragrunt file: %v", err)
	}
	runCmd(t, repoDir, "git", "add", changedRelPath)
	runCmd(t, repoDir, "git", "commit", "-m", "initial provider stack")

	baseBranch := runCmdOutput(t, repoDir, "git", "branch", "--show-current")
	if baseBranch == "" {
		t.Fatal("base branch is empty")
	}
	runCmd(t, repoDir, "git", "push", "-u", "origin", baseBranch)

	runCmd(t, repoDir, "git", "checkout", "-b", "feature")
	if err := os.WriteFile(changedAbsPath, []byte("terraform {}\n# change\n"), 0o644); err != nil {
		t.Fatalf("write changed terragrunt file: %v", err)
	}
	runCmd(t, repoDir, "git", "add", changedRelPath)

	return repoDir, baseBranch, changedRelPath
}

func assertContainsBaseName(t *testing.T, files []string, want string) {
	t.Helper()
	for _, f := range files {
		if filepath.Base(f) == want {
			return
		}
	}
	t.Fatalf("changed files %v do not contain %q", files, want)
}

func assertContainsPath(t *testing.T, files []string, want string) {
	t.Helper()
	want = filepath.Clean(want)
	for _, f := range files {
		if filepath.Clean(f) == want {
			return
		}
	}
	t.Fatalf("changed files %v do not contain %q", files, want)
}

func Test_buildBaseRefCandidates(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		baseRef string
		want    []string
	}{
		{
			name:    "head only",
			baseRef: "HEAD",
			want:    []string{"HEAD"},
		},
		{
			name:    "branch adds origin fallbacks",
			baseRef: "main",
			want:    []string{"main", "origin/main", "refs/remotes/origin/main"},
		},
		{
			name:    "origin ref unchanged",
			baseRef: "origin/main",
			want:    []string{"origin/main"},
		},
		{
			name:    "refs namespace unchanged",
			baseRef: "refs/heads/main",
			want:    []string{"refs/heads/main"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := buildBaseRefCandidates(tt.baseRef)
			if len(got) != len(tt.want) {
				t.Fatalf("got len %d (%v), want %d (%v)", len(got), got, len(tt.want), tt.want)
			}
			for i := range tt.want {
				if got[i] != tt.want[i] {
					t.Fatalf("got[%d]=%q, want %q", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestIsGitRepo(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()

	if IsGitRepo(context.Background(), tmpDir, nil) {
		t.Errorf("expected %s NOT to be a git repo", tmpDir)
	}

	runCmd(t, tmpDir, "git", "init")
	if !IsGitRepo(context.Background(), tmpDir, nil) {
		t.Errorf("expected %s to be a git repo", tmpDir)
	}
}
