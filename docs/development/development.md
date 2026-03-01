# Cultivator Development Guide

## Project Status

**Status**: Actively maintained at **v0.2.0**  
**Language**: Go 1.25+  
**Module**: `github.com/Ops-Talks/cultivator`  
**External Dependency**: `gopkg.in/yaml.v3` v3.0.1

**Core Functionality**:
- ✅ Stack discovery (recursive walk for `terragrunt.hcl`)
- ✅ Scope filtering (environment, paths, tags)
- ✅ Dependency graph (topological sorting)
- ✅ Parallel execution (configurable worker pool)
- ✅ Output formatting (text, JSON)
- ✅ Configuration management (YAML + CLI flags + env vars)
- ✅ Secret redaction (automatic masking of sensitive data)
- ✅ CI/CD integration (GitHub Actions, GitLab CI)

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

**Test Coverage**: 97.5%

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

**Test Coverage**: 92.4%

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

**Test Coverage**: 67.9%

### 4. Logging Package (`internal/logging/`)

**Purpose**: Structured logging throughout the application.

**Responsibilities**:
- Create and manage loggers
- Format log messages
- Support multiple output levels
- Redact sensitive information

**Test Files**:
- `logger_test.go` - Unit tests

**Test Coverage**: 83.9%

### 5. Runner Package (`internal/runner/`)

**Purpose**: Execute Terragrunt commands and capture results.

**Responsibilities**:
- Execute `terragrunt` commands per stack
- Handle command initialization and lifecycle
- Capture stdout and stderr
- Report execution results
- Manage concurrent execution with worker pool

**Test Files**:
- `runner_test.go` - Unit tests

**Test Coverage**: 65.4%

## Testing

For comprehensive information on testing strategy, test types, and guidelines, see the [Testing Guide](TESTING.md).

### Coverage Summary

Current test coverage reflects a focus on critical functionality:

- **Core libraries** (config, discovery): 92-97% coverage
- **CLI interface**: 67.9% coverage
- **Utilities**: 65-84% coverage

Lower coverage in `cmd/main.go` is expected, as the main function contains OS exit calls that cannot be tested directly.

## Development Workflow

### Building

```bash
# Build binary
go build -o cultivator ./cmd/cultivator

# Build with version information
go build -ldflags="-X main.version=v0.2.0 -X main.commit=$(git rev-parse HEAD)" \
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
    Root string
    // ...
}

// LoadFile reads and parses a YAML configuration file.
func LoadFile(path string) (Config, error) {
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
_ = runner.Execute(ctx, cmd)
```

### Naming Conventions

- Use clear, descriptive names
- Avoid single-letter variables except for loop indices and receivers
- Package names should be concise and lowercase
- Avoid repetition in names (e.g., `discovery.Discover` instead nof `discovery.DiscoveryDiscover`)

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

### Local Development

```bash
# Build
go build -o cultivator ./cmd/cultivator

# Run
./cultivator --help
./cultivator validate -c cultivator.yml
./cultivator version
```

### Testing with Docker

```bash
# Build image
docker build -t cultivator:dev .

# Run
docker run -v $(pwd):/workspace cultivator:dev validate
```

## Testing Strategy

### Unit Tests
- Each package has `*_test.go` files
- Test core logic in isolation
- Mock external dependencies

### Integration Tests (TODO)
- Test full workflow end-to-end
- Use test repositories
- Verify GitHub integration

### Manual Testing
1. Create a test PR in a Terragrunt repo
2. Trigger Cultivator via comment or event
3. Verify plan/apply results
4. Check PR comments

## Configuration

### Example `cultivator.yml`
```yaml
version: 1

projects:
  - name: production
    dir: envs/prod
    terragrunt_version: 0.55.0
    terraform_version: 1.7.0
    apply_requirements:
      - approved
    auto_plan: true

settings:
  auto_plan: true
  parallel_plan: true
  max_parallel: 5
  lock_timeout: 10m
```

## GitHub Action Integration

### Workflow File (`.github/workflows/cultivator.yml`)
```yaml
name: Cultivator

on:
  pull_request:
    types: [opened, synchronize, reopened]
  issue_comment:
    types: [created]

jobs:
  cultivator:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: cultivator-dev/cultivator-action@v1
        with:
          github-token: ${{ secrets.GITHUB_TOKEN }}
```

## Next Steps for Production

### Must-Have Features
1. **Proper approval checking** - Verify PR is approved before apply
2. **State locking backend** - Use DynamoDB/S3 for distributed locks
3. **Improved error handling** - Better error messages and recovery
4. **Plan file storage** - Save plans and apply exactly what was planned

### Nice-to-Have Features
1. **Cost estimation** - Integrate with Infracost
2. **Drift detection** - Detect infrastructure drift
3. **Policy as code** - OPA/Sentinel integration
4. **Multi-CI support** - GitLab CI, Azure DevOps, CircleCI
5. **Slack/Discord notifications**
6. **RBAC** - Role-based access control for environments

### Performance Improvements
1. **Parallel execution** - Run independent stacks in parallel
2. **Caching** - Cache Terraform/Terragrunt downloads
3. **Incremental plans** - Only plan truly affected stacks

## Common Development Tasks

### Adding a New Command
1. Create command function in `pkg/cmd/root.go`
2. Implement logic (possibly in new package)
3. Add tests
4. Update documentation

### Adding GitHub Event Support
1. Update `pkg/events/events.go`
2. Add parser for new event type
3. Handle in orchestrator
4. Test with sample event payload

### Improving Output Formatting
1. Update `pkg/formatter/formatter.go`
2. Add new formatting functions
3. Use in `pkg/github/client.go`
4. Test with real Terragrunt output

## Troubleshooting

### Tests Failing
```bash
# Run specific test
go test -v ./pkg/graph -run TestGraph_TopologicalSort

# Check for race conditions
go test -race ./...
```

### Build Issues
```bash
# Clean and rebuild
make clean
go mod tidy
make build
```

### GitHub Action Not Triggering
- Check workflow file syntax
- Verify permissions are set correctly
- Check event filters (PR types, comment patterns)

## Contributing

See [Contributing Guide](/CONTRIBUTING.md) for:
- Code style guidelines
- PR process
- Commit message format
- Testing requirements

## Resources

- [Terragrunt Documentation](https://terragrunt.gruntwork.io/)
- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [Go Testing](https://go.dev/doc/tutorial/add-a-test)
- [Cobra CLI Framework](https://github.com/spf13/cobra)
