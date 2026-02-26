package discovery

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDiscoverFilters(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	prod := filepath.Join(root, "prod", "app1")
	dev := filepath.Join(root, "dev", "app2")
	gitDir := filepath.Join(root, ".git")

	if err := os.MkdirAll(prod, 0o750); err != nil {
		t.Fatalf("mkdir prod: %v", err)
	}
	if err := os.MkdirAll(dev, 0o750); err != nil {
		t.Fatalf("mkdir dev: %v", err)
	}
	if err := os.MkdirAll(gitDir, 0o750); err != nil {
		t.Fatalf("mkdir git: %v", err)
	}

	prodFile := filepath.Join(prod, "terragrunt.hcl")
	devFile := filepath.Join(dev, "terragrunt.hcl")
	gitFile := filepath.Join(gitDir, "terragrunt.hcl")

	if err := os.WriteFile(prodFile, []byte("# cultivator:tags=app,db\n"), 0o600); err != nil {
		t.Fatalf("write prod file: %v", err)
	}
	if err := os.WriteFile(devFile, []byte("# cultivator:tags=api\n"), 0o600); err != nil {
		t.Fatalf("write dev file: %v", err)
	}
	if err := os.WriteFile(gitFile, []byte(""), 0o600); err != nil {
		t.Fatalf("write git file: %v", err)
	}

	modules, err := Discover(root, Options{Env: "prod"})
	if err != nil {
		t.Fatalf("discover: %v", err)
	}
	if len(modules) != 1 {
		t.Fatalf("expected 1 module, got %d", len(modules))
	}
	if modules[0].Env != "prod" {
		t.Fatalf("expected env prod, got %q", modules[0].Env)
	}

	modules, err = Discover(root, Options{Include: []string{"prod"}, Exclude: []string{"prod/app1"}})
	if err != nil {
		t.Fatalf("discover include/exclude: %v", err)
	}
	if len(modules) != 0 {
		t.Fatalf("expected 0 modules after exclude, got %d", len(modules))
	}

	modules, err = Discover(root, Options{Tags: []string{"db"}})
	if err != nil {
		t.Fatalf("discover tags: %v", err)
	}
	if len(modules) != 1 {
		t.Fatalf("expected 1 tagged module, got %d", len(modules))
	}
}
