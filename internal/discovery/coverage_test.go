package discovery

import (
	"os"
	"path/filepath"
	"testing"
)

// Test parseTags with simple comment format
func TestParseTags_Simple(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "simple.hcl")
	content := []byte("# cultivator:tags = prod,database,api")

	if err := os.WriteFile(filePath, content, 0o600); err != nil {
		t.Fatalf("write file: %v", err)
	}

	tags := parseTags(filePath)
	if len(tags) != 3 {
		t.Errorf("expected 3 tags, got %d: %v", len(tags), tags)
	}
}

// Test parseTags with list format
func TestParseTags_ListFormat(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "list.hcl")
	content := []byte(`locals {
  cultivator_tags = ["prod", "critical"]
}`)

	if err := os.WriteFile(filePath, content, 0o600); err != nil {
		t.Fatalf("write file: %v", err)
	}

	tags := parseTags(filePath)
	if len(tags) != 2 {
		t.Errorf("expected 2 tags, got %d: %v", len(tags), tags)
	}
}

// Test parseTags with no tags
func TestParseTags_None(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "notags.hcl")
	content := []byte("terraform { source = \"../module\" }\n")

	if err := os.WriteFile(filePath, content, 0o600); err != nil {
		t.Fatalf("write file: %v", err)
	}

	tags := parseTags(filePath)
	if tags != nil {
		t.Errorf("expected nil, got %v", tags)
	}
}

// Test parseTags with empty file
func TestParseTags_Empty(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "empty.hcl")

	if err := os.WriteFile(filePath, []byte(""), 0o600); err != nil {
		t.Fatalf("write file: %v", err)
	}

	tags := parseTags(filePath)
	if tags != nil {
		t.Errorf("expected nil, got %v", tags)
	}
}

// Test splitTags with commas
func TestSplitTags_Comma(t *testing.T) {
	t.Parallel()

	const firstTag = "tag1"
	tags := splitTags("tag1,tag2,tag3")
	if len(tags) != 3 || tags[0] != firstTag {
		t.Errorf("expected 3 tags starting with %s, got %v", firstTag, tags)
	}
}

// Test splitTags with semicolons
func TestSplitTags_Semicolon(t *testing.T) {
	t.Parallel()

	tags := splitTags("tag1;tag2;tag3")
	if len(tags) != 3 || tags[0] != "tag1" {
		t.Errorf("expected 3 tags, got %v", tags)
	}
}

// Test splitTags with mixed separators
func TestSplitTags_Mixed(t *testing.T) {
	t.Parallel()

	tags := splitTags("tag1,tag2;tag3")
	if len(tags) != 3 {
		t.Errorf("expected 3 tags with mixed separators, got %d: %v", len(tags), tags)
	}
}

// Test splitTags with spaces
func TestSplitTags_Spaces(t *testing.T) {
	t.Parallel()

	tags := splitTags("tag1 , tag2 ; tag3")
	if len(tags) != 3 {
		t.Errorf("expected 3 trimmed tags, got %v", tags)
	}
	for _, tag := range tags {
		if tag != "tag1" && tag != "tag2" && tag != "tag3" {
			t.Errorf("unexpected tag: %q", tag)
		}
	}
}

// Test envFromPath with standard path
func TestEnvFromPath_Standard(t *testing.T) {
	t.Parallel()

	env := envFromPath("/root", "/root/prod/app")
	if env != "prod" {
		t.Errorf("expected env 'prod', got %q", env)
	}
}

// Test envFromPath with nested path
func TestEnvFromPath_Nested(t *testing.T) {
	t.Parallel()

	env := envFromPath("/root", "/root/staging/service/db")
	if env != "staging" {
		t.Errorf("expected env 'staging', got %q", env)
	}
}

// Test envFromPath with same paths
func TestEnvFromPath_Same(t *testing.T) {
	t.Parallel()

	env := envFromPath("/root", "/root")
	if env != "." {
		t.Errorf("expected env '.', got %q", env)
	}
}

// Test envFromPath with empty module path
func TestEnvFromPath_EmptyModule(t *testing.T) {
	t.Parallel()

	env := envFromPath("/root", "")
	if env != "" {
		t.Errorf("expected empty env, got %q", env)
	}
}

// Test shouldSkipDir with hidden dirs
func TestShouldSkipDir_Hidden(t *testing.T) {
	t.Parallel()

	if !shouldSkipDir(".git") {
		t.Error(".git should be skipped")
	}
	if !shouldSkipDir(".terraform") {
		t.Error(".terraform should be skipped")
	}
	if !shouldSkipDir(".cache") {
		t.Error(".cache should be skipped")
	}
}

// Test shouldSkipDir with normal dirs
func TestShouldSkipDir_Normal(t *testing.T) {
	t.Parallel()

	if shouldSkipDir("prod") {
		t.Error("prod should not be skipped")
	}
	if shouldSkipDir("staging") {
		t.Error("staging should not be skipped")
	}
	if shouldSkipDir("modules") {
		t.Error("modules should not be skipped")
	}
}

// Test hasPathPrefix with exact match
func TestHashPathPrefix_Exact(t *testing.T) {
	t.Parallel()

	if !hasPathPrefix("/root/prod", "/root/prod") {
		t.Error("exact path should match")
	}
}

// Test hasPathPrefix with prefix
func TestHashPathPrefix_Prefix(t *testing.T) {
	t.Parallel()

	if !hasPathPrefix("/root/prod/app", "/root/prod") {
		t.Error("nested path should have prefix")
	}
}

// Test hasPathPrefix with non-matching path
func TestHashPathPrefix_NoMatch(t *testing.T) {
	t.Parallel()

	if hasPathPrefix("/root/prod", "/root/staging") {
		t.Error("different paths should not match")
	}
}

// Test normalizePaths with relative paths
func TestNormalizePaths_Relative(t *testing.T) {
	t.Parallel()

	paths := normalizePaths("/root", []string{"prod", "staging"})
	if len(paths) != 2 {
		t.Errorf("expected 2 paths, got %d", len(paths))
	}
	for _, path := range paths {
		if !filepath.IsAbs(path) {
			t.Errorf("path should be absolute: %q", path)
		}
	}
}

// Test normalizePaths with absolute paths
func TestNormalizePaths_Absolute(t *testing.T) {
	t.Parallel()

	paths := normalizePaths("/root", []string{"/prod", "/staging"})
	if len(paths) != 2 {
		t.Errorf("expected 2 paths, got %d", len(paths))
	}
}

// Test normalizePaths with empty strings
func TestNormalizePaths_Empty(t *testing.T) {
	t.Parallel()

	paths := normalizePaths("/root", []string{"", "prod", ""})
	if len(paths) != 1 {
		t.Errorf("empty strings should be filtered, got %d paths", len(paths))
	}
}

// Test matchesTags with matching tags
func TestMatchesTags_Match(t *testing.T) {
	t.Parallel()

	if !matchesTags([]string{"prod", "database"}, []string{"database"}) {
		t.Error("should match when tag is present")
	}
}

// Test matchesTags with no required tags
func TestMatchesTags_NoRequired(t *testing.T) {
	t.Parallel()

	if !matchesTags([]string{"prod"}, []string{}) {
		t.Error("should match when no tags required")
	}
}

// Test matchesTags with no matching tags
func TestMatchesTags_NoMatch(t *testing.T) {
	t.Parallel()

	if matchesTags([]string{"prod"}, []string{"staging"}) {
		t.Error("should not match when tag is absent")
	}
}

// Test matchesTags case-insensitive
func TestMatchesTags_CaseInsensitive(t *testing.T) {
	t.Parallel()

	if !matchesTags([]string{"PROD"}, []string{"prod"}) {
		t.Error("matching should be case-insensitive")
	}
}

func TestDiscover_RootValidationAndWalkError(t *testing.T) {
	t.Parallel()

	t.Run("empty root", func(t *testing.T) {
		t.Parallel()

		modules, err := Discover("", Options{})
		if err == nil {
			t.Fatal("expected error for empty root")
		}
		if modules != nil {
			t.Fatalf("expected nil modules, got %v", modules)
		}
	})

	t.Run("missing root", func(t *testing.T) {
		t.Parallel()

		missing := filepath.Join(t.TempDir(), "not-found")
		modules, err := Discover(missing, Options{})
		if err == nil {
			t.Fatal("expected walk error for missing root")
		}
		if modules != nil {
			t.Fatalf("expected nil modules, got %v", modules)
		}
	})
}

func TestDiscover_SkipByTagsAndEnv(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	prodDir := filepath.Join(root, "prod", "service")
	devDir := filepath.Join(root, "dev", "service")

	if err := os.MkdirAll(prodDir, 0o750); err != nil {
		t.Fatalf("mkdir prod: %v", err)
	}
	if err := os.MkdirAll(devDir, 0o750); err != nil {
		t.Fatalf("mkdir dev: %v", err)
	}

	prodFile := filepath.Join(prodDir, "terragrunt.hcl")
	devFile := filepath.Join(devDir, "terragrunt.hcl")

	if err := os.WriteFile(prodFile, []byte("# cultivator:tags = backend"), 0o600); err != nil {
		t.Fatalf("write prod file: %v", err)
	}
	if err := os.WriteFile(devFile, []byte("# cultivator:tags = web"), 0o600); err != nil {
		t.Fatalf("write dev file: %v", err)
	}

	modules, err := Discover(root, Options{Env: "prod", Tags: []string{"web"}})
	if err != nil {
		t.Fatalf("discover: %v", err)
	}
	if len(modules) != 0 {
		t.Fatalf("expected 0 modules for unmatched tag in selected env, got %d", len(modules))
	}
}

func TestParseTags_ListWithoutQuotedItems(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "list-no-items.hcl")
	content := []byte("locals {\n  cultivator_tags = []\n}\n")

	if err := os.WriteFile(filePath, content, 0o600); err != nil {
		t.Fatalf("write file: %v", err)
	}

	tags := parseTags(filePath)
	if tags != nil {
		t.Fatalf("expected nil tags, got %v", tags)
	}
}
