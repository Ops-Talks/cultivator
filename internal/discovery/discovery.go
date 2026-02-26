package discovery

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type Module struct {
	Path string
	Env  string
	Tags []string
}

type Options struct {
	Env     string
	Include []string
	Exclude []string
	Tags    []string
}

var tagCommentPattern = regexp.MustCompile(`(?i)cultivator:tags\s*=\s*([a-z0-9,_-]+)`)
var tagListPattern = regexp.MustCompile(`(?i)cultivator_tags\s*=\s*\[(.*?)\]`)
var tagItemPattern = regexp.MustCompile(`"([a-zA-Z0-9_-]+)"`)

func Discover(root string, options Options) ([]Module, error) {
	if root == "" {
		return nil, fmt.Errorf("root directory is required")
	}

	root = filepath.Clean(root)
	modules := []Module{}

	include := normalizePaths(root, options.Include)
	exclude := normalizePaths(root, options.Exclude)

	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		if entry.IsDir() {
			if shouldSkipDir(entry.Name()) {
				return filepath.SkipDir
			}
			return nil
		}

		if entry.Name() != "terragrunt.hcl" {
			return nil
		}

		moduleDir := filepath.Dir(path)
		if !matchesIncludeExclude(moduleDir, include, exclude) {
			return nil
		}

		env := envFromPath(root, moduleDir)
		if options.Env != "" && options.Env != env {
			return nil
		}

		module := Module{
			Path: moduleDir,
			Env:  env,
			Tags: parseTags(path),
		}

		if !matchesTags(module.Tags, options.Tags) {
			return nil
		}

		modules = append(modules, module)
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("discover modules: %w", err)
	}

	return modules, nil
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
	// #nosec G304 -- module paths are discovered from the configured root
	file, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer func() {
		_ = file.Close()
	}()

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

	scanner := bufio.NewScanner(strings.NewReader(input))
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.Contains(line, "cultivator") {
			continue
		}

		matches = tagCommentPattern.FindStringSubmatch(line)
		if len(matches) == 2 {
			return splitTags(matches[1])
		}
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
