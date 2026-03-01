package config

import (
	"path/filepath"
	"runtime"
	"testing"
)

func BenchmarkLoadFile(b *testing.B) {
	testDataDir := getBenchTestDataDir(&testing.T{})
	cfgFile := filepath.Join(testDataDir, "terragrunt-structure", ".cultivator.yaml")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		LoadFile(cfgFile)
	}
}

func BenchmarkMergeConfig(b *testing.B) {
	base := DefaultConfig()
	override := DefaultConfig()
	override.Parallelism = 8

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		MergeConfig(base, override)
	}
}

func BenchmarkApplyOverrides(b *testing.B) {
	cfg := DefaultConfig()
	root := "prod"
	parallelism := 4

	overrides := Overrides{
		Root:        &root,
		Parallelism: &parallelism,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ApplyOverrides(cfg, overrides)
	}
}

func BenchmarkValidate(b *testing.B) {
	cfg := DefaultConfig()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Validate(cfg)
	}
}

func getBenchTestDataDir(t *testing.T) string {
	_, file, _, ok := runtime.Caller(1)
	if !ok {
		return "testdata"
	}
	projectRoot := filepath.Join(filepath.Dir(file), "..", "..")
	return filepath.Join(projectRoot, "testdata")
}
