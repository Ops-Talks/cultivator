package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Root           string                 `yaml:"root"`
	Env            string                 `yaml:"env"`
	Include        []string               `yaml:"include"`
	Exclude        []string               `yaml:"exclude"`
	Tags           []string               `yaml:"tags"`
	Parallelism    int                    `yaml:"parallelism"`
	OutputFormat   string                 `yaml:"output_format"`
	NonInteractive bool                   `yaml:"non_interactive"`
	Plan           map[string]interface{} `yaml:"plan"`
	Apply          map[string]interface{} `yaml:"apply"`
	Destroy        map[string]interface{} `yaml:"destroy"`
	Doctor         map[string]interface{} `yaml:"doctor"`
}

type Overrides struct {
	Root           *string
	Env            *string
	Include        []string
	IncludeSet     bool
	Exclude        []string
	ExcludeSet     bool
	Tags           []string
	TagsSet        bool
	Parallelism    *int
	OutputFormat   *string
	NonInteractive *bool
	PlanDestroy    *bool
	ApplyAutoAppr  *bool
	DestroyAutoApr *bool
}

func DefaultConfig() Config {
	parallelism := runtime.NumCPU()
	if parallelism < 1 {
		parallelism = 1
	}

	return Config{
		Root:         ".",
		Parallelism:  parallelism,
		OutputFormat: "text",
		Plan:         map[string]interface{}{},
		Apply:        map[string]interface{}{},
		Destroy:      map[string]interface{}{},
		Doctor:       map[string]interface{}{},
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

	for key := range raw {
		if key == "root" || key == "env" || key == "include" || key == "exclude" || key == "tags" || key == "parallelism" || key == "output_format" || key == "non_interactive" || key == "plan" || key == "apply" || key == "destroy" || key == "doctor" {
			continue
		}
		extra[key] = raw[key]
	}

	return cfg, extra, true, nil
}

func LoadEnv(prefix string) Config {
	cfg := DefaultConfig()

	if root := os.Getenv(prefix + "_ROOT"); root != "" {
		cfg.Root = root
	}
	if env := os.Getenv(prefix + "_ENV"); env != "" {
		cfg.Env = env
	}
	if include := os.Getenv(prefix + "_INCLUDE"); include != "" {
		cfg.Include = strings.FieldsFunc(include, func(r rune) bool {
			return r == ',' || r == ';'
		})
	}
	if exclude := os.Getenv(prefix + "_EXCLUDE"); exclude != "" {
		cfg.Exclude = strings.FieldsFunc(exclude, func(r rune) bool {
			return r == ',' || r == ';'
		})
	}
	if tags := os.Getenv(prefix + "_TAGS"); tags != "" {
		cfg.Tags = strings.FieldsFunc(tags, func(r rune) bool {
			return r == ',' || r == ';'
		})
	}
	if par := os.Getenv(prefix + "_PARALLELISM"); par != "" {
		val, err := strconv.Atoi(par)
		if err == nil && val > 0 {
			cfg.Parallelism = val
		}
	}
	if format := os.Getenv(prefix + "_OUTPUT_FORMAT"); format != "" {
		cfg.OutputFormat = format
	}
	if nonInt := os.Getenv(prefix + "_NON_INTERACTIVE"); nonInt != "" {
		cfg.NonInteractive = parseBool(nonInt)
	}

	return cfg
}

func MergeConfig(base, override Config) Config {
	result := base

	if override.Root != "" && override.Root != "." {
		result.Root = override.Root
	}
	if override.Env != "" {
		result.Env = override.Env
	}
	if len(override.Include) > 0 {
		result.Include = override.Include
	}
	if len(override.Exclude) > 0 {
		result.Exclude = override.Exclude
	}
	if len(override.Tags) > 0 {
		result.Tags = override.Tags
	}
	if override.Parallelism > 0 && override.Parallelism != base.Parallelism {
		result.Parallelism = override.Parallelism
	}
	if override.OutputFormat != "" && override.OutputFormat != base.OutputFormat {
		result.OutputFormat = override.OutputFormat
	}
	if override.NonInteractive {
		result.NonInteractive = override.NonInteractive
	}
	if len(override.Plan) > 0 {
		for key, value := range override.Plan {
			result.Plan[key] = value
		}
	}
	if len(override.Apply) > 0 {
		for key, value := range override.Apply {
			result.Apply[key] = value
		}
	}
	if len(override.Destroy) > 0 {
		for key, value := range override.Destroy {
			result.Destroy[key] = value
		}
	}
	if len(override.Doctor) > 0 {
		for key, value := range override.Doctor {
			result.Doctor[key] = value
		}
	}

	return result
}

func ApplyOverrides(cfg Config, ovr Overrides) Config {
	if ovr.Root != nil {
		cfg.Root = *ovr.Root
	}
	if ovr.Env != nil {
		cfg.Env = *ovr.Env
	}
	if ovr.IncludeSet {
		cfg.Include = ovr.Include
	}
	if ovr.ExcludeSet {
		cfg.Exclude = ovr.Exclude
	}
	if ovr.TagsSet {
		cfg.Tags = ovr.Tags
	}
	if ovr.Parallelism != nil && *ovr.Parallelism > 0 {
		cfg.Parallelism = *ovr.Parallelism
	}
	if ovr.OutputFormat != nil && *ovr.OutputFormat != "" {
		cfg.OutputFormat = *ovr.OutputFormat
	}
	if ovr.NonInteractive != nil {
		cfg.NonInteractive = *ovr.NonInteractive
	}
	if ovr.PlanDestroy != nil && *ovr.PlanDestroy {
		if cfg.Plan == nil {
			cfg.Plan = make(map[string]interface{})
		}
		cfg.Plan["destroy"] = true
	}
	if ovr.ApplyAutoAppr != nil && *ovr.ApplyAutoAppr {
		if cfg.Apply == nil {
			cfg.Apply = make(map[string]interface{})
		}
		cfg.Apply["auto_approve"] = true
	}
	if ovr.DestroyAutoApr != nil && *ovr.DestroyAutoApr {
		if cfg.Destroy == nil {
			cfg.Destroy = make(map[string]interface{})
		}
		cfg.Destroy["auto_approve"] = true
	}

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

	const (
		textFormat = "text"
		jsonFormat = "json"
	)
	switch cfg.OutputFormat {
	case textFormat, jsonFormat:
		// valid
	default:
		return fmt.Errorf("invalid output format: %s (must be %s or %s)", cfg.OutputFormat, textFormat, jsonFormat)
	}

	return nil
}

func ParseBool(value string) bool {
	return parseBool(value)
}

func parseBool(value string) bool {
	lower := strings.ToLower(strings.TrimSpace(value))
	return lower == "true" || lower == "1" || lower == "on" || lower == "yes"
}

func ParseInt(value string) (int, error) {
	val, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("invalid integer: %s", value)
	}
	return val, nil
}
