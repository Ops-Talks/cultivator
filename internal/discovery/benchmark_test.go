package discovery

import (
	"path/filepath"
	"runtime"
	"testing"
)

func BenchmarkDiscover(b *testing.B) {
	testDataDir := getBenchDataDir(&testing.T{})
	rootDir := filepath.Join(testDataDir, "terragrunt-structure")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Discover(rootDir, Options{})
	}
}

func BenchmarkDiscover_WithFilters(b *testing.B) {
	testDataDir := getBenchDataDir(&testing.T{})
	rootDir := filepath.Join(testDataDir, "terragrunt-structure")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Discover(rootDir, Options{
			Env:  "prod",
			Tags: []string{"app"},
		})
	}
}

func BenchmarkDiscover_LargeStructure(b *testing.B) {
	testDataDir := getBenchDataDir(&testing.T{})
	rootDir := filepath.Join(testDataDir, "terragrunt-large")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Discover(rootDir, Options{})
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

func getBenchDataDir(t *testing.T) string {
	_, file, _, ok := runtime.Caller(1)
	if !ok {
		return "testdata"
	}
	projectRoot := filepath.Join(filepath.Dir(file), "..", "..")
	return filepath.Join(projectRoot, "testdata")
}
