package config

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestLoadFile(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()

	tests := []struct {
		name       string
		configPath string
		content    string
		wantFound  bool
		wantErr    bool
		validate   func(*testing.T, Config)
	}{
		{
			name:       "valid config file",
			configPath: filepath.Join(tempDir, "valid.yaml"),
			content:    "root: live\nparallelism: 3\nnon_interactive: true\nplan:\n  destroy: true\napply:\n  auto_approve: true\n",
			wantFound:  true,
			validate: func(t *testing.T, cfg Config) {
				if cfg.Root != "live" {
					t.Errorf("expected root live, got %q", cfg.Root)
				}
				if cfg.Parallelism != 3 {
					t.Errorf("expected parallelism 3, got %d", cfg.Parallelism)
				}
				if !cfg.NonInteractive {
					t.Error("expected non-interactive true")
				}
				if v, ok := cfg.Plan["destroy"]; !ok || v != true {
					t.Errorf("expected plan.destroy=true, got %v", cfg.Plan)
				}
			},
		},
		{
			name:       "missing file",
			configPath: filepath.Join(tempDir, "nonexistent.yaml"),
			wantFound:  false,
		},
		{
			name:       "empty path",
			configPath: "",
			wantFound:  false,
		},
		{
			name:       "invalid yaml",
			configPath: filepath.Join(tempDir, "invalid.yaml"),
			content:    "root: : :",
			wantErr:    true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.content != "" {
				if err := os.WriteFile(tc.configPath, []byte(tc.content), 0o600); err != nil {
					t.Fatalf("write config file: %v", err)
				}
			}

			cfg, _, found, err := LoadFile(tc.configPath)
			if (err != nil) != tc.wantErr {
				t.Fatalf("LoadFile() error = %v, wantErr %v", err, tc.wantErr)
			}
			if found != tc.wantFound {
				t.Fatalf("LoadFile() found = %v, want %v", found, tc.wantFound)
			}
			if tc.validate != nil {
				tc.validate(t, cfg)
			}
		})
	}
}

func TestApplyOverrides(t *testing.T) {
	t.Parallel()

	s := func(v string) *string { return &v }
	i := func(v int) *int { return &v }
	b := func(v bool) *bool { return &v }

	tests := []struct {
		name      string
		overrides Overrides
		want      Config
	}{
		{
			name: "override all basic fields",
			overrides: Overrides{
				Root:           s("envs"),
				Parallelism:    i(5),
				NonInteractive: b(true),
				DryRun:         b(true),
			},
			want: Config{
				Root:           "envs",
				Parallelism:    5,
				NonInteractive: true,
				DryRun:         true,
			},
		},
		{
			name: "override filters",
			overrides: Overrides{
				Include:    []string{"prod"},
				IncludeSet: true,
				Exclude:    []string{"test"},
				ExcludeSet: true,
				Tags:       []string{"db"},
				TagsSet:    true,
			},
			want: Config{
				Include: []string{"prod"},
				Exclude: []string{"test"},
				Tags:    []string{"db"},
			},
		},
		{
			name: "override command flags",
			overrides: Overrides{
				PlanDestroy:    b(true),
				ApplyAutoAppr:  b(true),
				DestroyAutoApr: b(true),
			},
			want: Config{
				Plan:    map[string]any{"destroy": true},
				Apply:   map[string]any{"auto_approve": true},
				Destroy: map[string]any{"auto_approve": true},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cfg := ApplyOverrides(DefaultConfig(), tc.overrides)
			if cfg.Root != "" && tc.want.Root != "" && cfg.Root != tc.want.Root {
				t.Errorf("Root = %q, want %q", cfg.Root, tc.want.Root)
			}
			if tc.want.Parallelism != 0 && cfg.Parallelism != tc.want.Parallelism {
				t.Errorf("Parallelism = %d, want %d", cfg.Parallelism, tc.want.Parallelism)
			}
			if tc.overrides.NonInteractive != nil && cfg.NonInteractive != *tc.overrides.NonInteractive {
				t.Errorf("NonInteractive = %v, want %v", cfg.NonInteractive, *tc.overrides.NonInteractive)
			}
			if tc.overrides.DryRun != nil && cfg.DryRun != *tc.overrides.DryRun {
				t.Errorf("DryRun = %v, want %v", cfg.DryRun, *tc.overrides.DryRun)
			}
			if tc.overrides.IncludeSet && !reflect.DeepEqual(cfg.Include, tc.want.Include) {
				t.Errorf("Include = %v, want %v", cfg.Include, tc.want.Include)
			}
			if tc.overrides.ExcludeSet && !reflect.DeepEqual(cfg.Exclude, tc.want.Exclude) {
				t.Errorf("Exclude = %v, want %v", cfg.Exclude, tc.want.Exclude)
			}
			if tc.overrides.TagsSet && !reflect.DeepEqual(cfg.Tags, tc.want.Tags) {
				t.Errorf("Tags = %v, want %v", cfg.Tags, tc.want.Tags)
			}
			if tc.overrides.PlanDestroy != nil {
				if v, ok := cfg.Plan["destroy"]; !ok || v != true {
					t.Errorf("Plan[destroy] = %v, want true", v)
				}
			}
		})
	}
}

func TestValidate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		cfg     Config
		wantErr bool
	}{
		{
			name:    "default config is valid",
			cfg:     DefaultConfig(),
			wantErr: false,
		},
		{
			name: "missing root",
			cfg: Config{
				Root:        "",
				Parallelism: 1,
			},
			wantErr: true,
		},
		{
			name: "invalid parallelism",
			cfg: Config{
				Root:        ".",
				Parallelism: 0,
			},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := Validate(tc.cfg)
			if (err != nil) != tc.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}

func TestLoadEnv(t *testing.T) {
	const prefix = "TEST_CULTIVATOR"

	tests := []struct {
		name string
		envs map[string]string
		want func(*testing.T, Config)
	}{
		{
			name: "load all fields",
			envs: map[string]string{
				prefix + "_ROOT":            "environments",
				prefix + "_ENV":             "staging",
				prefix + "_INCLUDE":         "app,db",
				prefix + "_EXCLUDE":         "legacy;tmp",
				prefix + "_TAGS":            "frontend,backend",
				prefix + "_PARALLELISM":     "8",
				prefix + "_NON_INTERACTIVE": "true",
			},
			want: func(t *testing.T, cfg Config) {
				if cfg.Root != "environments" {
					t.Errorf("Root = %q, want environments", cfg.Root)
				}
				if cfg.Env != "staging" {
					t.Errorf("Env = %q, want staging", cfg.Env)
				}
				if !reflect.DeepEqual(cfg.Include, []string{"app", "db"}) {
					t.Errorf("Include = %v, want [app db]", cfg.Include)
				}
				if !reflect.DeepEqual(cfg.Exclude, []string{"legacy", "tmp"}) {
					t.Errorf("Exclude = %v, want [legacy tmp]", cfg.Exclude)
				}
				if !reflect.DeepEqual(cfg.Tags, []string{"frontend", "backend"}) {
					t.Errorf("Tags = %v, want [frontend backend]", cfg.Tags)
				}
				if cfg.Parallelism != 8 {
					t.Errorf("Parallelism = %d, want 8", cfg.Parallelism)
				}
				if !cfg.NonInteractive {
					t.Error("NonInteractive = false, want true")
				}
			},
		},
		{
			name: "invalid parallelism fallback",
			envs: map[string]string{
				prefix + "_PARALLELISM": "not-a-number",
			},
			want: func(t *testing.T, cfg Config) {
				if cfg.Parallelism < 1 {
					t.Errorf("Parallelism = %d, want >= 1", cfg.Parallelism)
				}
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			for k, v := range tc.envs {
				t.Setenv(k, v)
			}

			cfg := LoadEnv(prefix)
			tc.want(t, cfg)
		})
	}
}

func TestMergeConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		base     Config
		override Config
		validate func(*testing.T, Config)
	}{
		{
			name: "merge command flags",
			base: DefaultConfig(),
			override: Config{
				Plan:  map[string]any{"destroy": true},
				Apply: map[string]any{"auto_approve": true},
			},
			validate: func(t *testing.T, cfg Config) {
				if v, ok := cfg.Plan["destroy"]; !ok || v != true {
					t.Errorf("Plan[destroy] = %v, want true", v)
				}
				if v, ok := cfg.Apply["auto_approve"]; !ok || v != true {
					t.Errorf("Apply[auto_approve] = %v, want true", v)
				}
			},
		},
		{
			name: "override root and env",
			base: DefaultConfig(),
			override: Config{
				Root: "new-root",
				Env:  "prod",
			},
			validate: func(t *testing.T, cfg Config) {
				if cfg.Root != "new-root" {
					t.Errorf("Root = %q, want new-root", cfg.Root)
				}
				if cfg.Env != "prod" {
					t.Errorf("Env = %q, want prod", cfg.Env)
				}
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			merged := MergeConfig(tc.base, tc.override)
			tc.validate(t, merged)
		})
	}
}
