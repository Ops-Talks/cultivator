# Testing Strategy

This document describes the testing approach, structure, and guidelines for Cultivator.

## Overview

Cultivator uses a comprehensive testing strategy combining unit tests, integration tests, fuzz testing, and benchmarks to ensure reliability and maintainability.

### Current Coverage

| Package    | Coverage | Status      |
|------------|----------|-------------|
| config     | 97.4%    | Excellent   |
| discovery  | 93.7%    | Excellent   |
| logging    | 94.9%    | Excellent   |
| cli        | 69.4%    | Good        |
| runner     | 70.8%    | Good        |
| cmd        | 50.0%    | Good        |

**Total Project Coverage**: ~79.4% (statement coverage)

## Test Types

### 1. Unit Tests

Located in `*_test.go` files alongside source code.

**Purpose**: Test individual functions and methods in isolation.

**Example**:
```bash
go test -v ./internal/config
```

**Guidelines**:
- One test file per source file
- Use table-driven tests for multiple cases
- Name tests descriptively: `Test_functionName_scenario`
- Use `t.Parallel()` for independent tests
- Avoid external dependencies; use mocks when needed

**Sample Tests**:
- `internal/config/config_test.go` - Configuration loading and merging
- `internal/discovery/discovery_test.go` - Stack discovery logic
- `internal/cli/cli_test.go` - CLI flag parsing and routing
- `internal/runner/runner_test.go` - Terragrunt execution and argument construction
- `cmd/cultivator/main_test.go` - Binary entry point exit codes

### 2. Fuzz Testing

Fuzz tests exercise code with randomly generated inputs to find edge cases and panics.

**Files**:
- `internal/config/fuzz_test.go`
- `internal/discovery/fuzz_test.go`

**Fuzz Functions**:

| Function            | Package    | Purpose                          |
|---------------------|------------|----------------------------------|
| FuzzParseBool       | config     | Boolean parsing robustness       |
| FuzzParseInt        | config     | Integer parsing with edge cases  |
| FuzzMergeConfig     | config     | Config merging with random data  |
| FuzzLoadEnv         | config     | Environment variable loading     |
| FuzzParseTags       | discovery  | Tag parsing with varied input    |
| FuzzSplitTags       | discovery  | Tag splitting edge cases         |
| FuzzEnvFromPath     | discovery  | Path-to-environment parsing      |

**Run Fuzz Tests**:
```bash
# Run for 60 seconds
go test -fuzz=FuzzParseBool -fuzztime=60s ./internal/config

# Run all fuzz tests
go test -fuzz=. ./internal/config ./internal/discovery
```

**Expected Behavior**:
- No panics discovered
- Handles malformed input gracefully
- Edge cases identified and documented

### 3. Integration Tests

Integration tests verify multiple components working together.

**Files**:
- `internal/config/integration_test.go`

**Purpose**: Validate end-to-end workflows.

**Example**:
```bash
go test -v ./internal/config -run TestIntegration
```

### 4. Benchmark Tests

Benchmark tests measure performance characteristics.

**Files**:
- `internal/config/benchmark_test.go`
- `internal/discovery/benchmark_test.go`

**Run Benchmarks**:
```bash
# Run benchmarks with memory stats
go test -bench=. -benchmem ./internal/config

# Run specific benchmark for 10 seconds
go test -bench=BenchmarkLoadFile -benchtime=10s ./internal/config
```

## Test Fixtures

Test fixture files are located in the `testdata/` directory at the project root.

**Structure**:
```
testdata/
├── terragrunt-large/
│   ├── dev/
│   ├── prod/
│   ├── staging/
│   └── test/
└── terragrunt-structure/
    ├── dev/
    └── prod/
```

**Purpose**: Provide realistic Terragrunt configurations for testing.

## How to Write Tests

### Basic Unit Test

```go
func TestConfigLoadFile(t *testing.T) {
    t.Parallel()

    tests := []struct {
        name      string
        filePath  string
        wantErr   bool
    }{
        {
            name:     "valid config file",
            filePath: "valid.yaml",
            wantErr:  false,
        },
        {
            name:     "missing file",
            filePath: "nonexistent.yaml",
            wantErr:  true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            _, err := LoadFile(tt.filePath)
            if (err != nil) != tt.wantErr {
                t.Errorf("LoadFile() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

### Fuzz Test

```go
func FuzzParseBool(f *testing.F) {
    testcases := []string{"true", "false", "yes", "no", "1", "0"}
    for _, tc := range testcases {
        f.Add(tc)
    }
    
    f.Fuzz(func(t *testing.T, input string) {
        _, _ = ParseBool(input)
        // If we reach here without panic, test passes
    })
}
```

## Running Tests

### All Tests
```bash
go test ./...
```

### Specific Package
```bash
go test './internal/config' -v
```

### With Coverage Report
```bash
go test -cover ./...
```

### Generate Coverage HTML
```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### With Race Detector
```bash
go test -race ./...
```

## Test Conventions

1. **Naming**: Test functions start with `Test`, fuzz functions with `Fuzz`, benchmarks with `Benchmark`
2. **Table-Driven**: Use subtests for multiple test cases
3. **Parallel**: Add `t.Parallel()` to independent tests
4. **Helpers**: Mark helper functions with `t.Helper()`
5. **Cleanup**: Use `t.Cleanup()` to clean up resources
6. **Assertions**: Check errors immediately; use custom error messages
7. **Mocking**: Mock external dependencies; avoid real I/O when possible

## Coverage Goals

- **New code**: Minimum 80% coverage
- **Critical paths**: 95%+ coverage
- **Public APIs**: 100% coverage on happy path

Current package coverage reflects these goals. Untested functions are typically:
- Integration code requiring complex setup
- Error paths that are difficult to reproduce
- Signal handling and OS-level interactions

## CI/CD Integration

Tests run automatically on:
- Pull requests to `main` and `develop`
- Commits to `main` and `develop`
- Release builds

**GitHub Actions workflow**: `.github/workflows/ci.yml`

### CI Pipeline

The CI pipeline runs the following checks:

1. **Test Job**:
   - Runs all tests with race detector
   - Generates coverage report (`coverage.out`)
   - Generates JUnit XML test results (`test-results.xml`) via `gotestsum`
   - Uploads coverage to [Codecov](https://codecov.io) for visualization and PR comments
   - Uploads test results to Codecov for failure tracking

2. **Lint Job**:
   - Runs `golangci-lint` with comprehensive rules
   - Checks code formatting and style

3. **Build Job**:
   - Compiles binary for Linux
   - Uploads artifact for validation

### Coverage Tracking

Coverage is tracked via **Codecov**:

- **Dashboard**: [https://app.codecov.io/gh/Ops-Talks/cultivator](https://app.codecov.io/gh/Ops-Talks/cultivator)
- **Badge**: Displayed in README.md
- **PR Comments**: Automatic coverage diffs on pull requests
- **Trends**: Historical coverage trends over time

To generate coverage locally:
```bash
make coverage
```

This will generate `coverage.out` and open an HTML report in your browser.

## Troubleshooting

### Test Timeouts
```bash
go test -timeout=30s ./...
```

### Fuzz Test Not Finding Seed
Check that fuzz test is properly structured and accepts `*testing.F` parameter.

### Coverage Measurement Issues
Clear fuzzing cache:
```bash
go clean -fuzzcache
```

## Resources

- [Go Testing Handbook](https://golang.org/doc/effective_go#testing)
- [Table-Driven Tests](https://github.com/golang/go/wiki/TableDrivenTests)
- [Fuzzing in Go](https://go.dev/doc/fuzz/)

