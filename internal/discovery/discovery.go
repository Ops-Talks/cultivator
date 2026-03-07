// Package discovery finds Terragrunt modules within a root directory,
// applying optional filters for environment, path inclusion/exclusion, and tags.
package discovery

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/Ops-Talks/cultivator/internal/hcl"
	"github.com/Ops-Talks/cultivator/internal/logging"
)

// Module represents a discovered Terragrunt module with its path, environment, and tags.
type Module struct {
	Path         string
	Env          string
	Tags         []string
	Dependencies []string // List of absolute paths to dependent modules
}

// Options controls which modules are returned by Discover.
type Options struct {
	Env     string
	Include []string
	Exclude []string
	Tags    []string
	Logger  *logging.Logger
}

var (
	tagCommentPattern = regexp.MustCompile(`(?i)cultivator:tags\s*=\s*([a-z0-9,_-]+)`)
	tagListPattern    = regexp.MustCompile(`(?i)cultivator_tags\s*=\s*\[(.*?)\]`)
	tagItemPattern    = regexp.MustCompile(`"([a-zA-Z0-9_-]+)"`)
)

// Discover walks root recursively and returns all Terragrunt modules that match options.
func Discover(root string, options Options) ([]Module, error) {
	if root == "" {
		return nil, fmt.Errorf("root directory is required")
	}

	root = filepath.Clean(root)
	scanner := &moduleScanner{
		root:    root,
		include: normalizePaths(root, options.Include),
		exclude: normalizePaths(root, options.Exclude),
		options: options,
	}

	if options.Logger != nil {
		options.Logger.Debug("starting discovery", logging.Fields{
			"root":    root,
			"env":     options.Env,
			"tags":    options.Tags,
			"include": scanner.include,
			"exclude": scanner.exclude,
		})
	}

	if err := filepath.WalkDir(root, scanner.walk); err != nil {
		return nil, fmt.Errorf("discover modules: %w", err)
	}

	return scanner.modules, nil
}

// moduleScanner holds the accumulated state for a single Discover walk.
type moduleScanner struct {
	root    string
	include []string
	exclude []string
	options Options
	modules []Module
}

// debugLog emits a debug message when a logger is configured.
func (s *moduleScanner) debugLog(msg string, fields logging.Fields) {
	if s.options.Logger != nil {
		s.options.Logger.Debug(msg, fields)
	}
}

// walk is the filepath.WalkDir callback.
func (s *moduleScanner) walk(path string, entry os.DirEntry, walkErr error) error {
	if walkErr != nil {
		return walkErr
	}
	if entry.IsDir() {
		if shouldSkipDir(entry.Name()) {
			s.debugLog("skipping directory", logging.Fields{"path": path})
			return filepath.SkipDir
		}
		return nil
	}
	if entry.Name() != "terragrunt.hcl" {
		return nil
	}
	s.debugLog("found terragrunt.hcl", logging.Fields{"path": path})
	return s.visitModule(path)
}

// visitModule processes a single terragrunt.hcl file, applying all filters and
// appending a Module to s.modules if all pass.
func (s *moduleScanner) visitModule(hclPath string) error {
	moduleDir := filepath.Dir(hclPath)

	if !matchesIncludeExclude(moduleDir, s.include, s.exclude) {
		s.debugLog("skipping module: path filter mismatch", logging.Fields{"path": moduleDir})
		return nil
	}

	env := envFromPath(s.root, moduleDir)
	if s.options.Env != "" && s.options.Env != env {
		s.debugLog("skipping module: environment mismatch", logging.Fields{
			"path": moduleDir, "expected": s.options.Env, "actual": env,
		})
		return nil
	}

	module := Module{Path: moduleDir, Env: env, Tags: parseTags(hclPath)}
	if !matchesTags(module.Tags, s.options.Tags) {
		s.debugLog("skipping module: tag mismatch", logging.Fields{
			"path": moduleDir, "required": s.options.Tags, "actual": module.Tags,
		})
		return nil
	}

	deps, err := hcl.ExtractDependencies(hclPath)
	if err != nil {
		return fmt.Errorf("extract dependencies from %s: %w", hclPath, err)
	}
	for _, dep := range deps {
		module.Dependencies = append(module.Dependencies, dep.Path)
	}

	s.debugLog("module discovered", logging.Fields{
		"path": module.Path, "env": module.Env,
		"tags": module.Tags, "dependencies": module.Dependencies,
	})
	s.modules = append(s.modules, module)
	return nil
}

func normalizePaths(root string, paths []string) []string {
	items := make([]string, 0, len(paths))
	for _, item := range paths {
		if item == "" {
			continue
		}

		clean := filepath.Clean(item)
		if !filepath.IsAbs(clean) {
			clean = filepath.Join(root, clean)
		}
		items = append(items, clean)
	}

	return items
}

func matchesIncludeExclude(modulePath string, include []string, exclude []string) bool {
	modulePath = filepath.Clean(modulePath)

	if len(include) > 0 {
		matched := false
		for _, inc := range include {
			if hasPathPrefix(modulePath, inc) {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}

	for _, exc := range exclude {
		if hasPathPrefix(modulePath, exc) {
			return false
		}
	}

	return true
}

func hasPathPrefix(path string, prefix string) bool {
	path = filepath.Clean(path)
	prefix = filepath.Clean(prefix)

	if path == prefix {
		return true
	}

	rel, err := filepath.Rel(prefix, path)
	if err != nil {
		return false
	}

	return rel != "." && !strings.HasPrefix(rel, "..")
}

func envFromPath(root string, modulePath string) string {
	rel, err := filepath.Rel(root, modulePath)
	if err != nil {
		return ""
	}

	parts := strings.Split(filepath.ToSlash(rel), "/")
	if len(parts) == 0 {
		return ""
	}

	return parts[0]
}

func matchesTags(moduleTags []string, required []string) bool {
	if len(required) == 0 {
		return true
	}

	set := map[string]struct{}{}
	for _, tag := range moduleTags {
		set[strings.ToLower(tag)] = struct{}{}
	}

	for _, requiredTag := range required {
		if _, ok := set[strings.ToLower(requiredTag)]; ok {
			return true
		}
	}

	return false
}

func parseTags(path string) []string {
	// #nosec G304 G703 -- module paths are discovered from the validated root directory
	file, err := os.Open(path)
	if err != nil {
		return nil
	}
	// Close error is intentionally ignored on read-only open; the file
	// descriptor will be released by the OS when the function returns.
	defer func() { _ = file.Close() }()

	content, err := io.ReadAll(file)
	if err != nil {
		return nil
	}

	input := string(content)
	matches := tagCommentPattern.FindStringSubmatch(input)
	if len(matches) == 2 {
		return splitTags(matches[1])
	}

	listMatch := tagListPattern.FindStringSubmatch(input)
	if len(listMatch) == 2 {
		items := tagItemPattern.FindAllStringSubmatch(listMatch[1], -1)
		if len(items) == 0 {
			return nil
		}
		result := make([]string, 0, len(items))
		for _, item := range items {
			if len(item) < 2 {
				continue
			}
			result = append(result, item[1])
		}
		return result
	}

	return nil
}

func splitTags(value string) []string {
	parts := strings.FieldsFunc(value, func(r rune) bool {
		return r == ',' || r == ';'
	})

	result := make([]string, 0, len(parts))
	for _, part := range parts {
		clean := strings.TrimSpace(part)
		if clean == "" {
			continue
		}
		result = append(result, clean)
	}

	return result
}

func shouldSkipDir(name string) bool {
	switch name {
	case ".git", ".terragrunt-cache", ".terraform", "node_modules", "vendor", "dist":
		return true
	default:
		return strings.HasPrefix(name, ".")
	}
}
