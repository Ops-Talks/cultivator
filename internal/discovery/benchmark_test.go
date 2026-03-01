package discovery

import (
	"path/filepath"
	"runtime"
	"testing"
)

func BenchmarkDiscover(b *testing.B) {
	testDataDir := getBenchDataDir()
	rootDir := filepath.Join(testDataDir, "terragrunt-structure")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := Discover(rootDir, Options{}); err != nil {
			b.Errorf("Discover error: %v", err)
		}
	}
}

func BenchmarkDiscover_WithFilters(b *testing.B) {
	testDataDir := getBenchDataDir()
	rootDir := filepath.Join(testDataDir, "terragrunt-structure")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := Discover(rootDir, Options{
			Env:  "prod",
			Tags: []string{"app"},
		}); err != nil {
			b.Errorf("Discover error: %v", err)
		}
	}
}

func BenchmarkDiscover_LargeStructure(b *testing.B) {
	testDataDir := getBenchDataDir()
	rootDir := filepath.Join(testDataDir, "terragrunt-large")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := Discover(rootDir, Options{}); err != nil {
			b.Errorf("Discover error: %v", err)
		}
	}
}

func BenchmarkMatchesTags(b *testing.B) {
	moduleTags := []string{"app", "db", "prod-critical"}
	filterTags := []string{"app"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		matchesTags(moduleTags, filterTags)
	}
}

func getBenchDataDir() string {
	_, file, _, ok := runtime.Caller(1)
	if !ok {
		return "testdata"
	}
	projectRoot := filepath.Join(filepath.Dir(file), "..", "..")
	return filepath.Join(projectRoot, "testdata")
}
