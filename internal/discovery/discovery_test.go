package discovery

import (
	"bytes"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/Ops-Talks/cultivator/internal/logging"
)

func TestDiscover(t *testing.T) {
	t.Parallel()

	root := t.TempDir()

	// Setup a complex directory structure
	// root/
	//   prod/
	//     app1/terragrunt.hcl (tags: app, db)
	//     app2/terragrunt.hcl (tags: app, api)
	//   dev/
	//     app3/terragrunt.hcl (tags: api)
	//   .git/
	//     terragrunt.hcl (should be ignored)

	dirs := []string{
		filepath.Join(root, "prod", "app1"),
		filepath.Join(root, "prod", "app2"),
		filepath.Join(root, "dev", "app3"),
		filepath.Join(root, ".git"),
	}

	for _, d := range dirs {
		if err := os.MkdirAll(d, 0o750); err != nil {
			t.Fatalf("setup: mkdir %s: %v", d, err)
		}
	}

	files := []struct {
		path    string
		content string
	}{
		{filepath.Join(root, "prod", "app1", "terragrunt.hcl"), "# cultivator:tags=app,db\n"},
		{filepath.Join(root, "prod", "app2", "terragrunt.hcl"), "cultivator_tags = [\"app\", \"api\"]\ndependency \"vpc\" { config_path = \"../app1\" }\n"},
		{filepath.Join(root, "dev", "app3", "terragrunt.hcl"), "# cultivator:tags=api\n"},
		{filepath.Join(root, ".git", "terragrunt.hcl"), ""},
	}

	for _, f := range files {
		if err := os.WriteFile(f.path, []byte(f.content), 0o600); err != nil {
			t.Fatalf("setup: write file %s: %v", f.path, err)
		}
	}

	tests := []struct {
		name      string
		options   Options
		wantCount int
		validate  func(*testing.T, []Module)
	}{
		{
			name:      "no filters finds all modules except hidden",
			options:   Options{},
			wantCount: 3,
		},
		{
			name:      "filter by env prod",
			options:   Options{Env: "prod"},
			wantCount: 2,
			validate: func(t *testing.T, modules []Module) {
				for _, m := range modules {
					if m.Env != "prod" {
						t.Errorf("expected env prod, got %q for path %s", m.Env, m.Path)
					}
				}
			},
		},
		{
			name:      "filter by env dev",
			options:   Options{Env: "dev"},
			wantCount: 1,
			validate: func(t *testing.T, modules []Module) {
				if modules[0].Env != "dev" {
					t.Errorf("expected env dev, got %q", modules[0].Env)
				}
			},
		},
		{
			name:      "filter by tags (api)",
			options:   Options{Tags: []string{"api"}},
			wantCount: 2, // prod/app2 and dev/app3
		},
		{
			name:      "filter by tags (db)",
			options:   Options{Tags: []string{"db"}},
			wantCount: 1, // prod/app1
		},
		{
			name: "include specific path",
			options: Options{
				Include: []string{"prod/app1"},
			},
			wantCount: 1,
		},
		{
			name: "exclude specific path",
			options: Options{
				Exclude: []string{"prod/app2"},
			},
			wantCount: 2, // prod/app1 and dev/app3
		},
		{
			name: "include and exclude",
			options: Options{
				Include: []string{"prod"},
				Exclude: []string{"prod/app1"},
			},
			wantCount: 1, // prod/app2
		},
		{
			name:      "extract dependencies",
			options:   Options{Include: []string{"prod/app2"}},
			wantCount: 1,
			validate: func(t *testing.T, modules []Module) {
				m := modules[0]
				if len(m.Dependencies) != 1 {
					t.Fatalf("expected 1 dependency, got %d", len(m.Dependencies))
				}
				if !strings.Contains(m.Dependencies[0], "prod/app1") {
					t.Errorf("expected dependency on prod/app1, got %q", m.Dependencies[0])
				}
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			modules, err := Discover(root, tc.options)
			if err != nil {
				t.Fatalf("Discover() error: %v", err)
			}
			if len(modules) != tc.wantCount {
				t.Fatalf("Discover() got %d modules, want %d", len(modules), tc.wantCount)
			}
			if tc.validate != nil {
				tc.validate(t, modules)
			}
		})
	}
}

func TestDiscover_Verbose(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	moduleDir := filepath.Join(root, "prod", "app1")
	_ = os.MkdirAll(moduleDir, 0o755)
	_ = os.WriteFile(filepath.Join(moduleDir, "terragrunt.hcl"), []byte("# cultivator:tags=app"), 0o644)

	var buf bytes.Buffer
	logger := logging.New(logging.LevelDebug, &buf, &buf)

	_, err := Discover(root, Options{
		Logger: logger,
	})
	if err != nil {
		t.Fatalf("Discover() error: %v", err)
	}

	output := buf.String()
	expected := []string{
		"starting discovery",
		"found terragrunt.hcl",
		"module discovered",
	}

	for _, s := range expected {
		if !strings.Contains(output, s) {
			t.Errorf("expected output to contain %q, but it didn't. Output:\n%s", s, output)
		}
	}
}

func TestParseTags_CompatibilityFormats(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		content string
		want    []string
	}{
		{
			name:    "comment style with mixed separators and case",
			content: "# cultivator:tags = Prod,critical; API\n",
			want:    []string{"prod", "critical", "api"},
		},
		{
			name: "list style with duplicates and whitespace",
			content: `locals {
  cultivator_tags = ["Prod", "api", "api"]
}`,
			want: []string{"prod", "api"},
		},
		{
			name: "both styles merged and deduplicated",
			content: `# cultivator:tags = app
locals {
  cultivator_tags = ["APP", "db"]
}`,
			want: []string{"app", "db"},
		},
		{
			name: "malformed tags are ignored",
			content: `// cultivator:tags = valid, bad tag, ???,other
locals {
  cultivator_tags = ["still_good", "bad tag"]
}`,
			want: []string{"valid", "other", "still_good"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tmpDir := t.TempDir()
			filePath := filepath.Join(tmpDir, "terragrunt.hcl")
			if err := os.WriteFile(filePath, []byte(tc.content), 0o600); err != nil {
				t.Fatalf("write file: %v", err)
			}

			got := parseTags(filePath)
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("parseTags() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestMatchesIncludeExclude_PrefixCollision(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	target := filepath.Join(root, "prod", "app")
	sibling := filepath.Join(root, "prod", "app2")

	include := normalizeFilterPaths(root, []string{"prod/app"})
	if !matchesIncludeExclude(target, include, nil) {
		t.Fatal("expected include path to match exact module subtree")
	}
	if matchesIncludeExclude(sibling, include, nil) {
		t.Fatal("prefix collision should not include sibling module")
	}

	exclude := normalizeFilterPaths(root, []string{"prod/app"})
	if matchesIncludeExclude(target, nil, exclude) {
		t.Fatal("expected exclude path to filter exact module subtree")
	}
	if !matchesIncludeExclude(sibling, nil, exclude) {
		t.Fatal("prefix collision should not exclude sibling module")
	}
}
