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
				if !cfg.Plan.Destroy {
					t.Errorf("expected plan.Destroy=true, got %v", cfg.Plan)
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

func TestLoadFile_LegacyDoctorBlockIgnored(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "legacy-doctor.yml")
	content := []byte("root: /tmp\nparallelism: 2\ndoctor:\n  enabled: true\n")
	if err := os.WriteFile(configPath, content, 0o600); err != nil {
		t.Fatalf("write config file: %v", err)
	}

	cfg, extra, found, err := LoadFile(configPath)
	if err != nil {
		t.Fatalf("LoadFile() error = %v", err)
	}
	if !found {
		t.Fatal("expected found=true")
	}
	if cfg.Root != "/tmp" {
		t.Fatalf("expected root=/tmp, got %q", cfg.Root)
	}
	if _, ok := extra["doctor"]; ok {
		t.Fatalf("legacy doctor key should be tolerated, got extras: %v", extra)
	}
}

func TestParseBoolStrict(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		want    bool
		wantErr bool
	}{
		{name: "true literal", input: "true", want: true},
		{name: "yes literal", input: "yes", want: true},
		{name: "on literal", input: "on", want: true},
		{name: "false literal", input: "false", want: false},
		{name: "off literal", input: "off", want: false},
		{name: "invalid literal", input: "maybe", wantErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := ParseBoolStrict(tc.input)
			if (err != nil) != tc.wantErr {
				t.Fatalf("ParseBoolStrict() error = %v, wantErr %v", err, tc.wantErr)
			}
			if tc.wantErr {
				return
			}
			if got != tc.want {
				t.Fatalf("ParseBoolStrict() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestPrecedence_DefaultFileEnvCLI(t *testing.T) {
	t.Parallel()

	defaultCfg := DefaultConfig()
	defaultCfg.Root = "/from-default"
	defaultCfg.Parallelism = 1

	fileCfg := Config{Root: "/from-file", Parallelism: 2}
	envCfg := Config{Root: "/from-env", Parallelism: 4}
	cliParallelism := 8
	cliRoot := "/from-cli"

	merged := MergeConfig(defaultCfg, fileCfg)
	merged = MergeConfig(merged, envCfg)
	merged = ApplyOverrides(merged, Overrides{
		Root:        &cliRoot,
		Parallelism: &cliParallelism,
	})

	if merged.Root != "/from-cli" {
		t.Fatalf("Root = %q, want /from-cli", merged.Root)
	}
	if merged.Parallelism != 8 {
		t.Fatalf("Parallelism = %d, want 8", merged.Parallelism)
	}
}

func TestApplyOverrides_BasicFields(t *testing.T) {
	t.Parallel()

	sp := func(v string) *string { return &v }
	ip := func(v int) *int { return &v }
	bp := func(v bool) *bool { return &v }

	cfg := ApplyOverrides(DefaultConfig(), Overrides{
		Root:           sp("envs"),
		Parallelism:    ip(5),
		NonInteractive: bp(true),
		DryRun:         bp(true),
	})

	if cfg.Root != "envs" {
		t.Errorf("Root = %q, want %q", cfg.Root, "envs")
	}
	if cfg.Parallelism != 5 {
		t.Errorf("Parallelism = %d, want 5", cfg.Parallelism)
	}
	if !cfg.NonInteractive {
		t.Error("NonInteractive = false, want true")
	}
	if !cfg.DryRun {
		t.Error("DryRun = false, want true")
	}
}

func TestApplyOverrides_Filters(t *testing.T) {
	t.Parallel()

	cfg := ApplyOverrides(DefaultConfig(), Overrides{
		Include:    []string{"prod"},
		IncludeSet: true,
		Exclude:    []string{"test"},
		ExcludeSet: true,
		Tags:       []string{"db"},
		TagsSet:    true,
	})

	if !reflect.DeepEqual(cfg.Include, []string{"prod"}) {
		t.Errorf("Include = %v, want [prod]", cfg.Include)
	}
	if !reflect.DeepEqual(cfg.Exclude, []string{"test"}) {
		t.Errorf("Exclude = %v, want [test]", cfg.Exclude)
	}
	if !reflect.DeepEqual(cfg.Tags, []string{"db"}) {
		t.Errorf("Tags = %v, want [db]", cfg.Tags)
	}
}

func TestApplyOverrides_CommandFlags(t *testing.T) {
	t.Parallel()

	bp := func(v bool) *bool { return &v }

	cfg := ApplyOverrides(DefaultConfig(), Overrides{
		PlanDestroy:        bp(true),
		ApplyAutoApprove:   bp(true),
		DestroyAutoApprove: bp(true),
	})

	if !cfg.Plan.Destroy {
		t.Error("Plan.Destroy = false, want true")
	}
	if !cfg.Apply.AutoApprove {
		t.Error("Apply.AutoApprove = false, want true")
	}
	if !cfg.Destroy.AutoApprove {
		t.Error("Destroy.AutoApprove = false, want true")
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
		{
			name: "subcommand env vars",
			envs: map[string]string{
				prefix + "_PLAN_DESTROY":         "true",
				prefix + "_APPLY_AUTO_APPROVE":   "true",
				prefix + "_DESTROY_AUTO_APPROVE": "true",
			},
			want: func(t *testing.T, cfg Config) {
				if !cfg.Plan.Destroy {
					t.Error("Plan.Destroy = false, want true")
				}
				if !cfg.Apply.AutoApprove {
					t.Error("Apply.AutoApprove = false, want true")
				}
				if !cfg.Destroy.AutoApprove {
					t.Error("Destroy.AutoApprove = false, want true")
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
				Plan:  PlanConfig{Destroy: true},
				Apply: ApplyConfig{AutoApprove: true},
			},
			validate: func(t *testing.T, cfg Config) {
				if !cfg.Plan.Destroy {
					t.Error("Plan.Destroy = false, want true")
				}
				if !cfg.Apply.AutoApprove {
					t.Error("Apply.AutoApprove = false, want true")
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
		{
			// Regression: override must apply even when the value equals the base,
			// because the user explicitly set it in an env-level config file.
			name: "parallelism override equals base value",
			base: Config{
				Root:        ".",
				Parallelism: 4,
			},
			override: Config{
				Parallelism: 4,
			},
			validate: func(t *testing.T, cfg Config) {
				if cfg.Parallelism != 4 {
					t.Errorf("Parallelism = %d, want 4", cfg.Parallelism)
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
