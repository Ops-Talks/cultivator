# Cultivator Development Guide

## Project Status

**Status**: Actively maintained at **v0.2.0**  
**Language**: Go 1.25+  
**Module**: `github.com/Ops-Talks/cultivator`  
**External Dependency**: `gopkg.in/yaml.v3` v3.0.1

**Core Functionality**:
- ✅ Module discovery (recursive walk for `terragrunt.hcl`)
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
Module Discovery (find all terragrunt.hcl files)
    ↓
Scope Filter (apply --env, --include, --exclude, --tags)
    ↓
Dependency Graph (parse dependencies, build execution order)
    ↓
Executor (parallel worker pool)
    ├─→ Run Terragrunt command per module
    ├─→ Capture output on per-module basis
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
├── pkg/
│   ├── cmd/
│   │   └── root.go              # Root command definition
│   ├── config/
│   │   ├── config.go            # Config loading + merging
│   │   └── config_test.go       # Config tests
│   ├── constants/
│   │   └── constants.go         # Version, defaults, etc.
│   ├── detector/
│   │   ├── detector.go          # Module discovery
│   │   └── detector_test.go
│   ├── errors/
│   │   └── errors.go            # Custom error types
│   ├── events/
│   │   ├── events.go            # Flag/config parsing
│   │   └── events_test.go
│   ├── executor/
│   │   └── executor.go          # Terragrunt command execution
│   ├── formatter/
│   │   ├── formatter.go         # Output formatting
│   │   └── formatter_test.go
│   ├── github/
│   │   ├── client.go            # GitHub API (for future use)
│   │   └── client_test.go
│   ├── graph/
│   │   ├── graph.go             # Dependency graph builder
│   │   └── graph_test.go
│   ├── lock/
│   │   ├── lock.go              # File-based locking
│   │   └── lock_test.go
│   ├── module/
│   │   ├── source.go            # Module source parsing
│   │   ├── source_test.go
│   │   ├── git.go               # Git utilities
│   │   └── http.go              # HTTP utilities
│   ├── orchestrator/
│   │   └── orchestrator.go      # Main workflow orchestration
│   ├── parser/
│   │   └── parser.go            # HCL parsing for dependencies
│   └── util/
│       └── strings.go           # String utilities
├── go.mod                       # Module definition
├── go.sum                       # Dependency checksums
├── Makefile                     # Build, test, lint targets
├── Dockerfile                   # Multi-stage Docker image
└── docs/                        # MkDocs documentation
```

## Key Components

### 1. Config Package (`pkg/config/`)
**Purpose**: Load and merge configuration from multiple sources

**Responsibilities**:
- Read `cultivator.yml` / `cultivator.yaml` from filesystem
- Merge with environment variables (CULTIVATOR_* prefix)
- Merge with CLI flags (highest precedence)
- Validate configuration schema
- Provide typed access to settings

**Example**:
```go
cfg, _ := config.Load()  // Loads cultivator.yml
cfg.Root = "live"        // Override with flag
log.Printf("Executing in root: %s", cfg.Root)
```

**Test Coverage**: 73.8%

### 2. Detector Package (`pkg/detector/`)
**Purpose**: Discover Terragrunt modules in the filesystem

**Responsibilities**:
- Recursively walk directory tree from root
- Find all `terragrunt.hcl` files
- Parse module metadata (filters, tags)
- Build list of available modules

**Example**:
```go
detector := detector.NewDetector("live")
modules, _ := detector.Discover(ctx)
// Returns: [{Path: "envs/dev/vpc"}, {Path: "envs/dev/app"}, ...]
```

**Test Coverage**: 73.9%

### 3. Parser Package (`pkg/parser/`)
**Purpose**: Parse HCL to extract Terragrunt dependencies

**Responsibilities**:
- Parse `terragrunt.hcl` files to extract `dependency` blocks
- Build dependency map for graph construction
- Handle `include` configurations

**Example**:
```go
parser := parser.NewParser()
modules, _ := parser.ParseDependencies("live")
// Returns: map[string][]string{"app": ["vpc", "network"]}
```

**Test Coverage**: Not measured in dedicated tests (embedded in graph tests)

### 4. Graph Package (`pkg/graph/`)
**Purpose**: Build dependency graph and determine execution order

**Responsibilities**:
- Accept module list + dependencies
- Perform topological sort
- Detect circular dependencies
- Provide execution plan groups

**Example**:
```go
graph := graph.NewGraph(modules, depMap)
groups, _ := graph.ExecutionGroups()
// Returns: [[vpc], [app, network], [monitoring]]
```

**Test Coverage**: 73.9%

### 5. Executor Package (`pkg/executor/`)
**Purpose**: Execute Terragrunt commands and capture results

**Responsibilities**:
- Run `terragrunt plan`, `apply`, `destroy` per module
- Handle Terragrunt initialization
- Capture stdout/stderr per module
- Report results (success/failure/error messages)

**Example**:
```go
exec := executor.NewExecutor(cfg, logger)
result, _ := exec.Execute(ctx, "plan", "envs/dev/vpc")
// Returns: {Module: "envs/dev/vpc", Status: "success", Output: "..."}
```

**Test Coverage**: 65.4%

### 6. Formatter Package (`pkg/formatter/`)
**Purpose**: Format results for human and machine consumption

**Responsibilities**:
- Parse Terragrunt/Terraform output
- Redact sensitive data (passwords, tokens, keys)
- Format as text or JSON
- Summarize changes (Added, Changed, Destroyed resources)

**Example**:
```go
formatter := formatter.NewFormatter("json")
output, _ := formatter.Format(results)
// Returns: {"modules": [{"name": "vpc", "status": "success", ...}]}
```

**Test Coverage**: 73.9%

### 7. Lock Package (`pkg/lock/`)
**Purpose**: Prevent concurrent operations on the same module

**Responsibilities**:
- Create file-based locks per module
- Implement lock timeout and retry logic
- Provide clear error messages on lock contention

**Example**:
```go
lock := lock.NewLock("live/envs/prod/app", 30*time.Minute)
defer lock.Release()
lock.Acquire(ctx)  // Waits up to 30 minutes
```

**Test Coverage**: 73.9%

### 8. Orchestrator Package (`pkg/orchestrator/`)
**Purpose**: Coordinate the entire Cultivator workflow

**Responsibilities**:
- Load configuration
- Discover modules
- Apply filters
- Build graph
- Execute in parallel (respecting dependencies)
- Format and return results

**Workflow**:
```
1. Load config + parse CLI flags
2. Initialize detector
3. Discover all modules
4. Apply scope filters (env, include, exclude, tags)
5. Parse dependencies
6. Build execution graph
7. Create worker pool (parallelism N)
8. Execute groups sequentially, modules within groups concurrently
9. Collect results
10. Format output
11. Return exit code
```

## Development Workflow

### Local Setup
```bash
# Clone and setup
git clone https://github.com/Ops-Talks/cultivator.git
cd cultivator

# Install dependencies
go mod download

# Run tests
make test

# Run linter
make lint

# Run all checks (tests, lint, vet)
make check

# Build binary
make build

# Run locally
./cultivator plan --root=live --env=dev --non-interactive
```

### Running Tests
```bash
# All tests
make test

# Specific package
go test ./pkg/config/

# Verbose with coverage
go test -v -cover ./...

# Generate coverage report
make coverage

# View coverage in HTML
go tool cover -html=coverage.out
```

### Running Linter
```bash
# Lint everything
make lint

# Format code
go fmt ./...

# Vet for suspicious code
go vet ./...

# Using golangci-lint directly
golangci-lint run
```

### Build and Release
```bash
# Build for current OS
make build

# Build all platforms
make build-all

# Build Docker image
make docker-build

# Create release binaries
goreleaser release --rm-dist  # (when releasing)
```

## Testing Guidelines

### Test Organization
- Tests live in same package (white-box testing preferred)
- Test files: `*_test.go` in same directory as code
- Separate integration tests can go in `internal/integration/`

### Table-Driven Tests
```go
func TestDetector_Discover(t *testing.T) {
    tests := []struct {
        name      string
        root      string
        want      []string
        wantErr   bool
    }{
        {"finds vpc", "live", []string{"vpc"}, false},
        {"empty root", "empty", []string{}, false},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            d := detector.NewDetector(tt.root)
            got, err := d.Discover(context.Background())
            if (err != nil) != tt.wantErr {
                t.Errorf("unexpected error: %v", err)
            }
            if !sliceEqual(got, tt.want) {
                t.Errorf("got %v, want %v", got, tt.want)
            }
        })
    }
}
```

### Coverage Targets
- Global: Aim for 70%+ coverage
- Critical paths (config, graph, executor): 80%+
- Utilities: 60%+
- Don't obsess over 100%; prioritize meaningful tests

## Code Quality Standards

### Formatting
```bash
go fmt ./...            # Required before commit
goimports -w ./...      # Auto-organize imports
golangci-lint run       # Run linter
```

### Error Handling
```go
// Good: Add context
if err != nil {
    return fmt.Errorf("failed to load config: %w", err)
}

// Bad: Ignore errors
_ = executor.Run(ctx, cmd)  // Never do this
```

### Variable Names
- Use clear, descriptive names
- Avoid abbreviations (use `err` not `e`, `cfg` is OK for `config`)
- Avoid stuttering (`detector.Detector` → use `detector.New()`)

### Documentation
- Document all exported functions/types
- Start with function/type name
- Use clear, English prose

```go
// NodeGraph represents a directed acyclic graph of module dependencies.
// It supports topological sorting and circular dependency detection.
type NodeGraph struct {
    // ...
}

// AddNode adds a module node to the graph.
func (g *NodeGraph) AddNode(id string) error {
    // ...
}
```

## Debugging Tips

### Enable Verbose Logging
```bash
export CULTIVATOR_LOG_LEVEL=debug
./cultivator plan --root=live --env=dev
```

### Test Single Module
```bash
./cultivator plan --root=live --include=envs/dev/vpc --non-interactive
```

### Run Terragrunt Manually to Debug
```bash
cd live/envs/dev/vpc
terragrunt init
terragrunt plan
```

### Check Dependencies
```bash
./cultivator doctor  # (if implemented)
# Or parse manually
cat live/envs/dev/vpc/terragrunt.hcl | grep -A5 "dependency"
```

## Continuous Integration

### GitHub Actions (`.github/workflows/ci.yml`)
- Runs on every PR and push
- Tests: `go test ./...`
- Lint: `golangci-lint-action@v6` (Go 1.25 compatible)
- Build: `make build-all`
- Coverage: Upload to Codecov

### Release Process (`.github/workflows/release.yml`)
On push to tag `v*`:
1. Checkout code
2. Build multi-platform binaries (Linux, macOS, Windows)
3. Generate checksums
4. Create GitHub Release with binaries

## Contributing

See [CONTRIBUTING.md](../../CONTRIBUTING.md) for:
- Code style guidelines
- Testing requirements
- PR process
- Commit message standards

## Future Work

Planned enhancements:
- [ ] Metrics/telemetry export
- [ ] Web UI for plan visualization
- [ ] Custom hooks (pre/post execution)
- [ ] Parallel module fetching optimization
- [ ] Better error messages and recovery suggestions

## Resources

- **[Cobra](https://cobra.dev/)** - CLI framework documentation
- **[HCL2](https://github.com/hashicorp/hcl2)** - HCL parsing library
- **[Terragrunt](https://terragrunt.gruntwork.io/)** - Terragrunt documentation
- **[Go Code Review Comments](https://go.dev/wiki/CodeReviewComments)** - Go best practices
- Clean ANSI codes from output
- Truncate long outputs
- Format module lists

## Development Workflow

### Build and Test

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
1. **Parallel execution** - Run independent modules in parallel
2. **Caching** - Cache Terraform/Terragrunt downloads
3. **Incremental plans** - Only plan truly affected modules

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
