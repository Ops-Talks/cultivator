// Package config provides YAML config loading, environment variable overrides,
// and CLI flag merging for cultivator.
package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// PlanConfig holds plan-specific Terragrunt options.
type PlanConfig struct {
	Destroy bool `yaml:"destroy"`
}

// ApplyConfig holds apply-specific Terragrunt options.
type ApplyConfig struct {
	AutoApprove bool `yaml:"auto_approve"`
}

// DestroyConfig holds destroy-specific Terragrunt options.
type DestroyConfig struct {
	AutoApprove bool `yaml:"auto_approve"`
}

type Config struct {
	Root           string        `yaml:"root"`
	Env            string        `yaml:"env"`
	Include        []string      `yaml:"include"`
	Exclude        []string      `yaml:"exclude"`
	Tags           []string      `yaml:"tags"`
	Parallelism    int           `yaml:"parallelism"`
	NonInteractive bool          `yaml:"non_interactive"`
	DryRun         bool          `yaml:"dry_run"`
	ShowGraph      bool          `yaml:"show_graph"`
	ChangedOnly    bool          `yaml:"changed_only"`
	BaseRef        string        `yaml:"base_ref"`
	Plan           PlanConfig    `yaml:"plan"`
	Apply          ApplyConfig   `yaml:"apply"`
	Destroy        DestroyConfig `yaml:"destroy"`
}

type Overrides struct {
	Root               *string
	Env                *string
	Include            []string
	IncludeSet         bool
	Exclude            []string
	ExcludeSet         bool
	Tags               []string
	TagsSet            bool
	Parallelism        *int
	NonInteractive     *bool
	DryRun             *bool
	ShowGraph          *bool
	ChangedOnly        *bool
	BaseRef            *string
	PlanDestroy        *bool
	ApplyAutoApprove   *bool
	DestroyAutoApprove *bool
}

func DefaultConfig() Config {
	parallelism := runtime.NumCPU()
	if parallelism < 1 {
		parallelism = 1
	}

	return Config{
		Root:        ".",
		Parallelism: parallelism,
		BaseRef:     "HEAD",
	}
}

func LoadFile(path string) (Config, map[string]interface{}, bool, error) {
	cfg := DefaultConfig()
	extra := map[string]interface{}{}

	if strings.TrimSpace(path) == "" {
		return cfg, extra, false, nil
	}

	// #nosec G304 -- config path is user-supplied
	fileContent, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, extra, false, nil
		}
		return cfg, extra, false, fmt.Errorf("read config: %w", err)
	}

	var raw map[string]interface{}
	if err := yaml.Unmarshal(fileContent, &raw); err != nil {
		return cfg, extra, false, fmt.Errorf("parse config: %w", err)
	}

	if err := yaml.Unmarshal(fileContent, &cfg); err != nil {
		return cfg, extra, false, fmt.Errorf("decode config: %w", err)
	}

	knownKeys := map[string]struct{}{
		"root": {}, "env": {}, "include": {}, "exclude": {}, "tags": {},
		"parallelism": {}, "non_interactive": {}, "dry_run": {}, "show_graph": {},
		"changed_only": {}, "base_ref": {}, "plan": {}, "apply": {}, "destroy": {}, "doctor": {},
	}
	for key := range raw {
		if _, ok := knownKeys[key]; ok {
			continue
		}
		extra[key] = raw[key]
	}

	return cfg, extra, true, nil
}

func LoadEnv(prefix string) Config {
	cfg := DefaultConfig()

	// envLoader pairs an environment variable suffix with the mutation it applies
	// to Config. All setter closures are branch-free so complexity stays in the
	// named helpers (splitEnvList, applyEnvParallelism, parseBool).
	type envLoader struct {
		suffix string
		apply  func(c *Config, val string)
	}

	loaders := []envLoader{
		{"_ROOT", func(c *Config, v string) { c.Root = v }},
		{"_ENV", func(c *Config, v string) { c.Env = v }},
		{"_BASE_REF", func(c *Config, v string) { c.BaseRef = v }},
		{"_INCLUDE", func(c *Config, v string) { c.Include = splitEnvList(v) }},
		{"_EXCLUDE", func(c *Config, v string) { c.Exclude = splitEnvList(v) }},
		{"_TAGS", func(c *Config, v string) { c.Tags = splitEnvList(v) }},
		{"_PARALLELISM", func(c *Config, v string) { applyEnvParallelism(c, v) }},
		{"_NON_INTERACTIVE", func(c *Config, v string) { c.NonInteractive = parseBoolLenient(v) }},
		{"_DRY_RUN", func(c *Config, v string) { c.DryRun = parseBoolLenient(v) }},
		{"_SHOW_GRAPH", func(c *Config, v string) { c.ShowGraph = parseBoolLenient(v) }},
		{"_CHANGED_ONLY", func(c *Config, v string) { c.ChangedOnly = parseBoolLenient(v) }},
		{"_PLAN_DESTROY", func(c *Config, v string) { c.Plan.Destroy = parseBoolLenient(v) }},
		{"_APPLY_AUTO_APPROVE", func(c *Config, v string) { c.Apply.AutoApprove = parseBoolLenient(v) }},
		{"_DESTROY_AUTO_APPROVE", func(c *Config, v string) { c.Destroy.AutoApprove = parseBoolLenient(v) }},
	}

	for _, l := range loaders {
		if val := os.Getenv(prefix + l.suffix); val != "" {
			l.apply(&cfg, val)
		}
	}

	return cfg
}

// splitEnvList splits a comma- or semicolon-separated environment variable
// value into individual non-empty items.
func splitEnvList(v string) []string {
	return strings.FieldsFunc(v, func(r rune) bool {
		return r == ',' || r == ';'
	})
}

// applyEnvParallelism parses val as a positive integer and writes it to
// cfg.Parallelism; invalid or non-positive values are silently ignored.
func applyEnvParallelism(cfg *Config, val string) {
	n, err := ParseInt(val)
	if err != nil || n < 1 {
		return
	}
	cfg.Parallelism = n
}

func MergeConfig(base, override Config) Config {
	result := base

	// Root uses a special guard: the zero value "." is treated as unset.
	if override.Root != "" && override.Root != "." {
		result.Root = override.Root
	}
	mergeString(&result.Env, override.Env)
	mergeString(&result.BaseRef, override.BaseRef)
	mergeSlice(&result.Include, override.Include)
	mergeSlice(&result.Exclude, override.Exclude)
	mergeSlice(&result.Tags, override.Tags)
	mergePositiveInt(&result.Parallelism, override.Parallelism)
	mergeBoolTrue(&result.NonInteractive, override.NonInteractive)
	mergeBoolTrue(&result.DryRun, override.DryRun)
	mergeBoolTrue(&result.ShowGraph, override.ShowGraph)
	mergeBoolTrue(&result.ChangedOnly, override.ChangedOnly)
	mergeBoolTrue(&result.Plan.Destroy, override.Plan.Destroy)
	mergeBoolTrue(&result.Apply.AutoApprove, override.Apply.AutoApprove)
	mergeBoolTrue(&result.Destroy.AutoApprove, override.Destroy.AutoApprove)

	return result
}

func ApplyOverrides(cfg Config, ovr Overrides) Config {
	applyPtr(&cfg.Root, ovr.Root)
	applyPtr(&cfg.Env, ovr.Env)
	applyPtr(&cfg.BaseRef, ovr.BaseRef)
	applyPositiveInt(&cfg.Parallelism, ovr.Parallelism)
	applyPtr(&cfg.NonInteractive, ovr.NonInteractive)
	applyPtr(&cfg.DryRun, ovr.DryRun)
	applyPtr(&cfg.ShowGraph, ovr.ShowGraph)
	applyPtr(&cfg.ChangedOnly, ovr.ChangedOnly)
	if ovr.IncludeSet {
		cfg.Include = ovr.Include
	}
	if ovr.ExcludeSet {
		cfg.Exclude = ovr.Exclude
	}
	if ovr.TagsSet {
		cfg.Tags = ovr.Tags
	}
	applyBoolFlag(&cfg.Plan.Destroy, ovr.PlanDestroy)
	applyBoolFlag(&cfg.Apply.AutoApprove, ovr.ApplyAutoApprove)
	applyBoolFlag(&cfg.Destroy.AutoApprove, ovr.DestroyAutoApprove)

	return cfg
}

func Validate(cfg Config) error {
	if cfg.Root == "" {
		return fmt.Errorf("root directory is required")
	}

	if !filepath.IsAbs(cfg.Root) {
		if _, err := os.Stat(cfg.Root); err != nil && os.IsNotExist(err) {
			return fmt.Errorf("root directory does not exist: %s", cfg.Root)
		}
	}

	if cfg.Parallelism < 1 {
		return fmt.Errorf("parallelism must be >= 1, got %d", cfg.Parallelism)
	}

	return nil
}

func ParseBool(value string) bool {
	return parseBoolLenient(value)
}

func parseBoolLenient(value string) bool {
	lower := strings.ToLower(strings.TrimSpace(value))
	return lower == "true" || lower == "1" || lower == "on" || lower == "yes"
}

var errInvalidBooleanValue = errors.New("invalid boolean value")

// ParseBoolStrict parses boolean values for strict channels (such as CLI flags).
func ParseBoolStrict(value string) (bool, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "true", "1", "t", "yes", "y", "on":
		return true, nil
	case "false", "0", "f", "no", "n", "off":
		return false, nil
	default:
		return false, fmt.Errorf("%w: %s", errInvalidBooleanValue, strings.TrimSpace(value))
	}
}

func ParseInt(value string) (int, error) {
	clean := strings.TrimSpace(value)
	if clean == "" {
		return 0, fmt.Errorf("invalid integer value: %s", clean)
	}

	val, err := strconv.Atoi(clean)
	if err != nil {
		return 0, fmt.Errorf("invalid integer value: %s", clean)
	}
	return val, nil
}

// applyPtr copies *src into *dst when src is non-nil.
func applyPtr[T any](dst *T, src *T) {
	if src != nil {
		*dst = *src
	}
}

// applyBoolFlag sets *dst to true only when src is non-nil and true.
// Used for one-way command flags (plan --destroy, apply --auto-approve) that
// should only ever be enabled, never cleared, by an override.
func applyBoolFlag(dst *bool, src *bool) {
	if src != nil && *src {
		*dst = true
	}
}

// applyPositiveInt writes *src to *dst only when src is non-nil and positive.
func applyPositiveInt(dst *int, src *int) {
	if src != nil && *src > 0 {
		*dst = *src
	}
}

// mergeString writes src to *dst when src is non-empty.
func mergeString(dst *string, src string) {
	if src != "" {
		*dst = src
	}
}

// mergeSlice writes src to *dst when src is non-empty.
func mergeSlice(dst *[]string, src []string) {
	if len(src) > 0 {
		*dst = src
	}
}

// mergePositiveInt writes src to *dst when src is positive.
func mergePositiveInt(dst *int, src int) {
	if src > 0 {
		*dst = src
	}
}

// mergeBoolTrue sets *dst to true when src is true; it never clears a true value.
func mergeBoolTrue(dst *bool, src bool) {
	if src {
		*dst = true
	}
}
