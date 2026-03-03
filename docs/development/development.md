# Cultivator Development Guide

## Project Status

**Status**: Actively maintained at **v0.3.10**  
**Language**: Go 1.25+  
**Module**: `github.com/Ops-Talks/cultivator`  
**External Dependency**: `gopkg.in/yaml.v3` v3.0.1

**Core Functionality**:

- Stack discovery (recursive walk for `terragrunt.hcl`)
- Scope filtering (environment, paths, tags)
- Dependency graph (topological sorting)
- Parallel execution (configurable worker pool)
- Output formatting (text, JSON)
- Configuration management (YAML + CLI flags + env vars)
- CI/CD integration (GitHub Actions, GitLab CI)

## Architecture Overview

```
User invokes Cultivator
    ↓
Config Loader (merge cultivator.yml + env vars + CLI flags)
    ↓
Stack Discovery (find all terragrunt.hcl files)
    ↓
Scope Filter (apply --env, --include, --exclude, --tags)
    ↓
Dependency Graph (parse dependencies, build execution order)
    ↓
Executor (parallel worker pool)
    ├─→ Run Terragrunt command per stack
    ├─→ Capture output on per-stack basis
    └─→ Handle locks for concurrent operations
    ↓
Output Formatter (redact secrets, format text/JSON)
    ↓
Exit with status code (0=success, 1=failure, 2=usage error)
```

## Directory Structure

```
cultivator/
├── cmd/
│   └── cultivator/
│       ├── main.go              # Entry point with ldflags injection
│       └── main_test.go         # CLI tests
├── internal/
│   ├── cli/
│   │   ├── cli.go               # CLI command routing and flag parsing
│   │   └── cli_test.go          # CLI tests
│   ├── config/
│   │   ├── config.go            # Configuration loading and merging
│   │   ├── config_test.go       # Unit tests
│   │   ├── benchmark_test.go    # Performance benchmarks
│   │   ├── coverage_test.go     # Additional coverage tests
│   │   ├── fuzz_test.go         # Fuzz testing for robustness
│   │   └── integration_test.go  # End-to-end tests
│   ├── discovery/
│   │   ├── discovery.go         # Stack discovery and filtering
│   │   ├── discovery_test.go    # Unit tests
│   │   ├── benchmark_test.go    # Performance benchmarks
│   │   ├── coverage_test.go     # Additional coverage tests
│   │   └── fuzz_test.go         # Fuzz testing for robustness
│   ├── logging/
│   │   ├── logger.go            # Structured logging
│   │   └── logger_test.go       # Logger tests
│   └── runner/
│       ├── runner.go            # Terragrunt command execution
│       └── runner_test.go       # Runner tests
├── testdata/
│   ├── terragrunt-large/        # Large-scale test fixtures
│   │   ├── dev/
│   │   ├── prod/
│   │   ├── staging/
│   │   └── test/
│   └── terragrunt-structure/    # Simple test fixtures
│       ├── dev/
│       └── prod/
├── docs/                        # MkDocs documentation
├── go.mod                       # Module definition
├── go.sum                       # Dependency checksums
├── Makefile                     # Build, test, lint targets
└── Dockerfile                   # Multi-stage Docker image
```

## Key Packages

### 1. Config Package (`internal/config/`)

**Purpose**: Load and merge configuration from multiple sources.

**Responsibilities**:
- Read `cultivator.yml` / `cultivator.yaml` from filesystem
- Merge with environment variables (CULTIVATOR_* prefix)
- Merge with CLI flags (highest precedence)
- Validate configuration schema
- Provide typed access to settings

**Test Files**:
- `config_test.go` - Unit tests
- `benchmark_test.go` - Performance measurements
- `coverage_test.go` - Additional coverage gaps
- `fuzz_test.go` - Fuzzing for edge cases
- `integration_test.go` - End-to-end workflows

**Test Coverage**: 97.4%

### 2. Discovery Package (`internal/discovery/`)

**Purpose**: Discover and filter Terragrunt stacks in the filesystem.

**Responsibilities**:
- Recursively walk directory tree from root
- Find all `terragrunt.hcl` files
- Filter stacks by environment, path patterns, and tags
- Parse stack metadata
- Build list of available stacks

**Test Files**:
- `discovery_test.go` - Unit tests
- `benchmark_test.go` - Performance measurements
- `coverage_test.go` - Additional coverage gaps
- `fuzz_test.go` - Fuzzing for edge cases

**Test Coverage**: 93.7%

### 3. CLI Package (`internal/cli/`)

**Purpose**: Handle command-line interface and flag parsing.

**Responsibilities**:
- Parse CLI arguments and flags
- Route commands (plan, apply, destroy, version, doctor)
- Build configuration from CLI flags
- Format and display output
- Handle user interaction

**Test Files**:
- `cli_test.go` - Comprehensive unit tests

**Test Coverage**: 69.2%

### 4. Logging Package (`internal/logging/`)

**Purpose**: Structured logging throughout the application.

**Responsibilities**:
- Create and manage loggers
- Format log messages
- Support multiple output levels
- Redact sensitive information

**Test Files**:
- `logger_test.go` - Unit tests

**Test Coverage**: 94.9%

### 5. Runner Package (`internal/runner/`)

**Purpose**: Execute Terragrunt commands and capture results.

**Responsibilities**:
- Execute `terragrunt` commands per stack
- Handle command initialization and lifecycle
- Capture stdout and stderr in a single chronologically-ordered stream using `cmd.CombinedOutput()`
- Report execution results
- Manage concurrent execution with worker pool

**Test Files**:
- `runner_test.go` - Unit tests

**Test Coverage**: 95.5%

## Testing

For comprehensive information on testing strategy, test types, and guidelines, see the [Testing Guide](TESTING.md).

### Coverage Summary

Current test coverage reflects a focus on critical functionality:

- **Core libraries** (config, discovery, logging, runner): 93-97% coverage
- **CLI interface**: 69.2% coverage
- **cmd/main**: 50.0% (expected; OS exit calls cannot be tested directly)

## Development Workflow

### Building

```bash
# Build binary
go build -o cultivator ./cmd/cultivator

# Build with version information
go build -ldflags="-X main.version=v0.3.10 -X main.commit=$(git rev-parse HEAD)" \
  -o cultivator ./cmd/cultivator
```

### Testing

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run specific package tests
go test -v ./internal/cli
```

### Linting

```bash
# Using golangci-lint
golangci-lint run

# Format code
go fmt ./...
goimports -w ./...
```

## Contributing Guidelines

See [Contributing](contributing.md) for detailed contribution guidelines.

Key points:
- Write tests for new functionality
- Maintain or improve test coverage
- Follow Go conventions from the [Go Code Review Comments](https://go.dev/wiki/CodeReviewComments)
- Update documentation when making changes


## Code Quality Standards

### Documentation

All exported functions, types, and packages must be documented.

```go
// Config represents application configuration loaded from multiple sources.
// Configuration is merged in precedence order: defaults, YAML file, env vars, CLI flags.
type Config struct {
    Root        string
    Env         string
    Parallelism int
    // ...
}

// LoadFile reads and parses a YAML configuration file from the given path.
// It returns the parsed config, the resolved path, whether a file was found, and any error.
func LoadFile(path string) (Config, string, bool, error) {
    // ...
}
```

### Error Handling

Always wrap errors with context:

```go
// Good: Add context using %w
if err != nil {
    return fmt.Errorf("failed to load config from %s: %w", path, err)
}

// Avoid: Ignoring errors
_ = someOperation()
```

### Naming Conventions

- Use clear, descriptive names
- Avoid single-letter variables except for loop indices and receivers
- Package names should be concise and lowercase
- Avoid repetition in names (e.g., `discovery.Discover` instead of `discovery.DiscoveryDiscover`)

## Local Development Setup

```bash
# Clone repository
git clone https://github.com/Ops-Talks/cultivator.git
cd cultivator

# Download dependencies
go mod download

# Build
go build -o cultivator ./cmd/cultivator

# Run tests
go test ./...
```

## Continuous Integration

All pull requests must pass:
- Unit tests: `go test ./...`
- Linting: `golangci-lint run`
- Code formatting: `go fmt ./...`
- Coverage: Maintained at project baseline

Tests run automatically on all pull requests to `main` and commits to `main`.

```bash
# Install dependencies
make deps

# Build binary
make build

# Run tests
make test

# Run tests with coverage
make coverage

# Lint code
make lint

# Format code
make fmt

# Run all checks
make check
```

### Testing with Docker

```bash
# Build image
docker build -t cultivator:dev .

# Run help
docker run cultivator:dev plan --help
```

## Common Development Tasks

### Adding a New Subcommand

1. Add a constant in `internal/cli/cli.go` alongside `cmdPlan`, `cmdApply`, `cmdDestroy`
2. Add a case in the `switch command` block in `Run`
3. Implement the handler function
4. Write tests in `internal/cli/cli_test.go`
5. Update CLI reference documentation

### Extending Stack Discovery

Discovery logic lives in `internal/discovery/discovery.go`. Adding a new filter type:

1. Add the field to `Options` struct
2. Apply the filter inside `Discover`
3. Add unit tests in `discovery_test.go` covering the new filter
4. Add fuzz test seeds in `fuzz_test.go` if the field is string-typed

### Modifying Configuration

Configuration loading lives in `internal/config/config.go`:

1. Add the field to the `Config` struct
2. Set the default in `DefaultConfig`
3. Map the environment variable in `LoadEnv`
4. Add merge logic in `MergeConfig` and `ApplyOverrides`
5. Validate in `Validate` if the field has constraints
6. Update `docs/getting-started/configuration.md`

### Changing Runner Behavior

The runner lives in `internal/runner/runner.go`. It uses `cmd.CombinedOutput()` to merge stdout and stderr in write order, preserving chronological output from Terragrunt. When changing how output is captured, update the corresponding tests in `runner_test.go`.

## Troubleshooting

### Tests failing

```bash
# Run a specific test with verbose output
go test -v ./internal/cli -run TestBuildTerragruntConfig

# Run all tests with the race detector
go test -race ./...

# Run Make full check (fmt, vet, lint, test, coverage)
make check
```

### Build issues

```bash
# Clean build artifacts and rebuild
make clean
go mod tidy
make build
```

### Linting errors

```bash
# See all lint issues
golangci-lint run

# Auto-format code
make fmt
```

## Contributing

See the [Contributing Guide](contributing.md) for:
- Contribution workflow
- Pull request checklist
- Code style requirements
- Testing requirements

## Resources

- [Terragrunt Documentation](https://terragrunt.gruntwork.io/)
- [OpenTofu Documentation](https://opentofu.org/docs/)
- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [Go Testing](https://go.dev/doc/tutorial/add-a-test)
- [Effective Go](https://go.dev/doc/effective_go)
