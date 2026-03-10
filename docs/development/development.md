# Cultivator Development Guide

## Project Status

**Status**: Actively maintained
**Language**: Go 1.25+
**Module**: `github.com/Ops-Talks/cultivator`
**External Dependencies**: `gopkg.in/yaml.v3` v3.0.1, `github.com/olekukonko/tablewriter` v0.0.5

**Core Functionality**:

- Stack discovery (recursive walk for terragrunt.hcl)
- Scope filtering (environment, paths, tags)
- Dependency graph (topological sorting)
- Parallel execution (configurable worker pool)
- Output formatting (text-based)
- Configuration management (YAML + CLI flags + env vars)
- CI/CD integration (GitHub Actions, GitLab CI)

## Architecture Overview

```text
User invokes Cultivator
    |
Config Loader (merge cultivator.yml + env vars + CLI flags)
    |
Stack Discovery (find all terragrunt.hcl files)
    |
Scope Filter (apply --env, --include, --exclude, --tags)
    |
Dependency Graph (parse dependencies, build execution order)
    |
Executor (parallel worker pool)
    |-- Run Terragrunt command per stack
    |-- Capture output on per-stack basis
    |-- Respect dependency order via DAG signals
    |
Output Formatter (format text)
    |
Exit with status code (0=success, 1=failure, 2=usage error)
```

## Directory Structure

```text
cultivator/
├── cmd/
│   └── cultivator/
│       ├── main.go              # Entry point with ldflags injection
│       └── main_test.go         # CLI tests
├── internal/
│   ├── cli/
│   │   ├── cli.go               # CLI command routing and flag parsing
│   │   ├── cli_test.go          # CLI unit tests
│   │   └── e2e_test.go          # End-to-end integration tests
│   ├── config/
│   │   ├── config.go            # Configuration loading and merging
│   │   ├── config_test.go       # Unit tests
│   │   ├── benchmark_test.go    # Performance benchmarks
│   │   ├── coverage_test.go     # Additional coverage tests
│   │   ├── fuzz_test.go         # Fuzz testing for robustness
│   │   ├── integration_test.go  # End-to-end tests
│   │   └── testconst_test.go    # Shared test constants
│   ├── dag/
│   │   ├── dag.go               # Directed acyclic graph, topological sort, cycle detection
│   │   └── dag_test.go          # DAG unit tests
│   ├── discovery/
│   │   ├── discovery.go         # Stack discovery and filtering
│   │   ├── discovery_test.go    # Unit tests
│   │   ├── benchmark_test.go    # Performance benchmarks
│   │   ├── coverage_test.go     # Additional coverage tests
│   │   ├── fuzz_test.go         # Fuzz testing for robustness
│   │   └── testconst_test.go    # Shared test constants
│   ├── git/
│   │   ├── git.go               # Git diff integration for changed-file detection
│   │   └── git_test.go          # Git integration tests
│   ├── hcl/
│   │   ├── hcl.go               # Lightweight HCL parser for dependency config_path extraction
│   │   ├── hcl_test.go          # HCL parser tests
│   │   └── fuzz_test.go         # Fuzz testing
│   ├── logging/
│   │   ├── logger.go            # Structured logging
│   │   └── logger_test.go       # Logger tests
│   └── runner/
│       ├── runner.go            # Terragrunt command execution
│       ├── runner_test.go       # Runner tests
│       ├── dag_integration_test.go # DAG-aware execution order tests
│       └── dryrun_test.go       # Dry-run mode tests
├── testdata/
│   ├── terragrunt-large/        # Large-scale test fixtures
│   └── terragrunt-structure/    # Simple test fixtures
├── docs/                        # MkDocs documentation
├── go.mod                       # Module definition
├── go.sum                       # Dependency checksums
├── Makefile                     # Build, test, lint targets
└── Dockerfile                   # Multi-stage Docker image
```

## Key Packages

### 1. Config Package (internal/config/)

**Purpose**: Load and merge configuration from multiple sources.

**Responsibilities**:

- Read configuration from filesystem
- Merge with environment variables (CULTIVATOR_* prefix)
- Merge with CLI flags (highest precedence)
- Validates required configuration fields (root path exists, parallelism is positive)
- Provide typed access to settings

### 2. Discovery Package (internal/discovery/)

**Purpose**: Discover and filter Terragrunt stacks in the filesystem.

**Responsibilities**:

- Recursively walk directory tree from root
- Find all terragrunt.hcl files
- Filter stacks by environment, path patterns, and tags
- Parse stack metadata
- Build list of available stacks

### 3. CLI Package (internal/cli/)

**Purpose**: Handle command-line interface and flag parsing.

**Responsibilities**:

- Parse CLI arguments and flags
- Route commands (plan, apply, destroy, version, doctor)
- Build configuration from CLI flags
- Format and display output
- Handle user interaction

### 4. Logging Package (internal/logging/)

**Purpose**: Structured logging throughout the application.

**Responsibilities**:

- Create and manage loggers
- Format log messages
- Support multiple output levels
- Write summary tables at end of execution
- Cultivator relies on CI platform secret masking (e.g., GitHub Actions masked secrets) for sensitive data redaction

#### Logging Ownership Boundary

- CLI owns user-facing lifecycle and result logs (start/failure/success and summary table).
- Engine packages (`discovery`, `runner`, `git`) only emit optional debug logs through injected logger dependencies.
- This keeps normal output stable while still enabling deep diagnostics with `CULTIVATOR_LOG_LEVEL=debug`.

### 5. Runner Package (internal/runner/)

**Purpose**: Execute Terragrunt commands and capture results.

**Responsibilities**:

- Execute terragrunt commands per stack
- Handle command initialization and lifecycle
- Capture stdout and stderr in a single stream
- Report execution results
- Manage concurrent execution with worker pool

### 6. DAG Package (internal/dag/)

**Purpose**: Model and resolve inter-module dependencies.

**Responsibilities**:

- Build a Directed Acyclic Graph (DAG) of module relationships
- Perform topological sorting to determine correct execution order
- Detect and report circular dependencies (DFS-based cycle detection)
- Optionally render the dependency graph in Mermaid format

### 7. HCL Package (internal/hcl/)

**Purpose**: Extract dependency information from `terragrunt.hcl` files.

**Responsibilities**:

- Parse `dependency` blocks using regex-based extraction (no full HCL parser required)
- Resolve `config_path` values to absolute paths
- Provide dependency edges to the DAG builder

### 8. Git Package (internal/git/)

**Purpose**: Detect changed files for Magic Mode (`--changed-only`).

**Responsibilities**:

- Run `git diff --name-only` against a configurable base ref
- Return the set of changed file paths
- Map changed files to affected Terragrunt modules

## Testing

For comprehensive information on testing strategy, test types, and guidelines, see the [Testing Guide](TESTING.md).

## Development Workflow

### Building

```bash
# Build binary
go build -o cultivator ./cmd/cultivator

# Build with version information
go build -ldflags="-X main.version=v1.0.0 -X main.commit=$(git rev-parse HEAD)" \
  -o cultivator ./cmd/cultivator
```

### Running Tests

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
```

## Contributing Guidelines

See [Contributing](contributing.md) for detailed contribution guidelines.

## Code Quality Standards

### Documentation

All exported functions, types, and packages must be documented.

### Error Handling

Always wrap errors with context using fmt.Errorf with %w verb.

### Naming Conventions

- Use clear, descriptive names
- Avoid single-letter variables except for loop indices and receivers
- Package names should be concise and lowercase

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

- Unit tests
- Linting
- Code formatting
- Coverage baseline

## Common Development Tasks

### Adding a New Subcommand

1. Add a constant in internal/cli/cli.go
2. Add a case in the switch command block in Run
3. Implement the handler function
4. Write tests in internal/cli/cli_test.go
5. Update CLI reference documentation

### Extending Stack Discovery

Discovery logic lives in internal/discovery/discovery.go.

### Modifying Configuration

Configuration loading lives in internal/config/config.go.

### Changing Runner Behavior

The runner lives in internal/runner/runner.go. It uses cmd.CombinedOutput() to merge stdout and stderr.

## Troubleshooting

### Test failing

```bash
# Run a specific test with verbose output
go test -v ./internal/cli -run TestSomeFunction
```

### Build issues

```bash
# Clean build artifacts and rebuild
go mod tidy
make build
```

### Linting errors

```bash
# See all lint issues
golangci-lint run
```

## Resources

- [Terragrunt Documentation](https://terragrunt.gruntwork.io/)
- [OpenTofu Documentation](https://opentofu.org/docs/)
- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [Go Testing](https://go.dev/doc/tutorial/add-a-test)
