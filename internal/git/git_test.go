package git

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestGetChangedFiles(t *testing.T) {
	// This test depends on git being installed and the project being a git repo.
	// In a real CI environment, we might want to mock the git command or create a dummy repo.

	tmpDir := t.TempDir()

	// Create a dummy git repo
	runCmd(t, tmpDir, "git", "init")
	runCmd(t, tmpDir, "git", "config", "user.email", "test@example.com")
	runCmd(t, tmpDir, "git", "config", "user.name", "test")

	file1 := filepath.Join(tmpDir, "file1.txt")
	_ = os.WriteFile(file1, []byte("hello"), 0o644)
	runCmd(t, tmpDir, "git", "add", "file1.txt")
	runCmd(t, tmpDir, "git", "commit", "-m", "initial commit")

	// Create a branch and change another file
	runCmd(t, tmpDir, "git", "checkout", "-b", "feature")
	file2 := filepath.Join(tmpDir, "file2.txt")
	_ = os.WriteFile(file2, []byte("world"), 0o644)
	runCmd(t, tmpDir, "git", "add", "file2.txt")
	// file2 is staged but not committed, 'git diff' will see it if we compare against master

	tests := []struct {
		name     string
		baseRef  string
		wantFile string
	}{
		{
			name:     "compare feature against master",
			baseRef:  "master",
			wantFile: "file2.txt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetChangedFiles(context.Background(), tmpDir, tt.baseRef)
			if err != nil {
				// git might use 'main' instead of 'master'
				got, err = GetChangedFiles(context.Background(), tmpDir, "main")
				if err != nil {
					t.Fatalf("GetChangedFiles() error = %v", err)
				}
			}

			found := false
			for _, f := range got {
				if filepath.Base(f) == tt.wantFile {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("GetChangedFiles() did not find %q in %v", tt.wantFile, got)
			}
		})
	}
}

func runCmd(t *testing.T, dir string, name string, args ...string) {
	t.Helper()
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	if err := cmd.Run(); err != nil {
		t.Fatalf("command %s %v failed: %v", name, args, err)
	}
}

func TestIsGitRepo(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()

	if IsGitRepo(tmpDir) {
		t.Errorf("expected %s NOT to be a git repo", tmpDir)
	}

	runCmd(t, tmpDir, "git", "init")
	if !IsGitRepo(tmpDir) {
		t.Errorf("expected %s to be a git repo", tmpDir)
	}
}
