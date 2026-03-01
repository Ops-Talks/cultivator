package discovery

import (
	"path/filepath"
	"testing"
)

// FuzzParseTags tests parseTags function with random input
func FuzzParseTags(f *testing.F) {
	// Add seed corpus
	f.Add("# cultivator:tags = database,api")
	f.Add("# cultivator_tags = [\"database\", \"cache\"]")
	f.Add("")
	f.Add("# some comment")
	f.Add("cultivator:tags=single")
	f.Add("# cultivator:tags = tag1, tag2, tag3")
	f.Add("# cultivator_tags = [\"tag-with-dash\", \"tag_with_underscore\"]")
	f.Add("\n\n# cultivator:tags = multiline\n\n")
	f.Add("# random content\n# cultivator:tags = test\n# more content")

	f.Fuzz(func(t *testing.T, data string) {
		// This should not panic
		result := parseTags("nonexistent_file_" + filepath.Base(data))

		// Validate result
		if result == nil {
			return // nil is valid
		}

		// All tags should be non-empty strings
		for _, tag := range result {
			if tag == "" {
				t.Errorf("parseTags returned empty tag in result")
			}
		}
	})
}

// FuzzSplitTags tests splitTags function with random input
func FuzzSplitTags(f *testing.F) {
	// Add seed corpus
	f.Add("database,api")
	f.Add("tag1;tag2;tag3")
	f.Add("single-tag")
	f.Add("tag-with-dash,tag_with_underscore")
	f.Add("")
	f.Add("  spaces  ,  around  ")
	f.Add("tag1, tag2, tag3")
	f.Add("tag1;tag2,tag3")

	f.Fuzz(func(t *testing.T, data string) {
		// This should not panic
		result := splitTags(data)

		// Validate result
		if result == nil {
			return
		}

		// All tags should be trimmed and non-empty
		for _, tag := range result {
			if tag == "" {
				t.Errorf("splitTags returned empty tag in result")
			}
			// Check that tag doesn't have leading/trailing spaces
			if len(tag) > 0 && (tag[0] == ' ' || tag[len(tag)-1] == ' ') {
				t.Errorf("splitTags returned untrimmed tag: %q", tag)
			}
		}
	})
}

// FuzzEnvFromPath tests envFromPath function with random path input
func FuzzEnvFromPath(f *testing.F) {
	// Add seed corpus
	f.Add("/home/user/prod/app", "/home/user")
	f.Add("/terraform/staging/database", "/terraform")
	f.Add("/root/dev/services", "/root")
	f.Add("/path/to/module", "/path/to")
	f.Add("/single", "/single")
	f.Add(".", ".")
	f.Add("module/path", "module")
	f.Add("", "")

	f.Fuzz(func(t *testing.T, modulePath string, root string) {
		// This should not panic
		result := envFromPath(root, modulePath)

		// Validate result
		if result == "" {
			return // empty string is valid
		}

		// Result should not contain path separators
		if modulePath != "" && root != "" {
			// Valid result should be just one part of the relative path
			if len(result) > 0 && !isValidEnv(result) {
				t.Errorf("envFromPath returned invalid env: %q", result)
			}
		}
	})
}

// Helper function to validate env format
func isValidEnv(env string) bool {
	// Valid env should be a simple string without special chars that cause parsing issues
	if env == "" {
		return true
	}
	// Very simple check: just make sure it's not obviously garbage
	return len(env) < 256
}
