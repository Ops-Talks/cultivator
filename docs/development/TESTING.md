# Testing Strategy

This document describes the testing approach, structure, and guidelines for Cultivator.

## Overview

Cultivator uses a comprehensive testing strategy combining unit tests, integration tests, fuzz testing, and benchmarks to ensure reliability and maintainability.

### Current Coverage

> **Note**: Coverage numbers below are approximate and may vary between runs. For current values, run `go test -cover ./...` or check the [Codecov dashboard](https://codecov.io/gh/Ops-Talks/cultivator).

| Package    | Coverage | Status      |
|------------|----------|-------------|
| config     | 92.4%    | Excellent   |
| discovery  | 93.1%    | Excellent   |
| hcl        | 100.0%   | Excellent   |
| logging    | 95.3%    | Excellent   |
| runner     | 90.7%    | Excellent   |
| cli        | 85.5%    | Excellent   |
| cmd        | 50.0%    | Good        |

**Total Project Coverage**: ~88.5% (statement coverage)

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

### 2. Fuzz Testing

Fuzz tests exercise code with randomly generated inputs to find edge cases and panics.

**Files**:

- `internal/config/fuzz_test.go`
- `internal/discovery/fuzz_test.go`
- `internal/hcl/fuzz_test.go`

**Run Fuzz Tests**:

```bash
# Run for 60 seconds
go test -fuzz=FuzzParseBool -fuzztime=60s ./internal/config

# Run all fuzz tests in a package
go test -fuzz=. ./internal/config
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
        _ = ParseBool(input)
    })
}
```

## Running Tests

### All Tests

```bash
go test ./...
```

### With Coverage Report

```bash
go test -cover ./...
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

## Coverage Goals

- **New code**: Minimum 80% coverage
- **Critical paths**: 95%+ coverage
- **Public APIs**: 100% coverage on happy path

## CI/CD Integration

Tests run automatically on:

- Pull requests to main
- Commits to main
- Release builds

**GitHub Actions workflow**: `.github/workflows/ci.yml`

### Coverage Tracking

Coverage is tracked via **Codecov**. PR comments show automatic coverage diffs on pull requests.

To generate coverage locally:

```bash
make coverage
```
