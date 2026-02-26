# Cultivator Development Guide

## Project Status

**Core functionality implemented:**
- Dependency graph analysis with topological sorting
- HCL parser for Terragrunt configurations
- GitHub event handling and PR integration
- Command orchestration (plan, apply, plan-all, apply-all)
- Lock management for concurrent operations
- Output formatting and plan summaries
- Configuration validation
- CI/CD integration via GitHub Actions

## Architecture Overview

```
┌─────────────────────────────────────────────────┐
│                   GitHub PR                      │
│            (Comment or Event Trigger)            │
└───────────────────┬─────────────────────────────┘
                    │
                    ▼
┌─────────────────────────────────────────────────┐
│              GitHub Actions Runner               │
│          (Runs cultivator via action.yml)        │
└───────────────────┬─────────────────────────────┘
                    │
                    ▼
┌─────────────────────────────────────────────────┐
│              Orchestrator (main logic)           │
├─────────────────────────────────────────────────┤
│  1. Parse GitHub event                          │
│  2. Load configuration                          │
│  3. Detect changed modules                      │
│  4. Build dependency graph                      │
│  5. Execute in topological order                │
│  6. Post results to PR                          │
└───────────────────┬─────────────────────────────┘
                    │
        ┌───────────┼───────────┐
        ▼           ▼           ▼
    ┌──────┐  ┌──────────┐  ┌────────┐
    │Parser│  │Detector  │  │Executor│
    └──────┘  └──────────┘  └────────┘
        │           │            │
        ▼           ▼            ▼
    ┌─────────────────────────────────┐
    │         Terragrunt CLI          │
    └─────────────────────────────────┘
```

## Key Components

### 1. Graph Package (`pkg/graph/`)
**Purpose:** Dependency analysis and execution ordering

**Features:**
- Build dependency graphs from Terragrunt modules
- Topological sorting (dependencies first)
- Circular dependency detection
- Find affected modules when changes occur
- Group modules for parallel execution

**Example:**
```go
graph := graph.NewGraph()
graph.AddDependency("app", "database")
graph.AddDependency("database", "vpc")

// Returns: [vpc, database, app]
sorted, _ := graph.TopologicalSort()

// Returns modules that depend on vpc
affected := graph.GetAffectedModules([]string{"vpc"})
```

### 2. Parser Package (`pkg/parser/`)
**Purpose:** Parse Terragrunt HCL files

**Features:**
- Extract `dependency` blocks
- Parse `terraform` source references
- Find `include` configurations
- Discover all modules in a directory tree

**Example:**
```go
parser := parser.NewParser()
config, _ := parser.ParseFile("path/to/terragrunt.hcl")

// Get all dependencies
deps, _ := parser.FindDependencies("module/path")
```

### 3. Detector Package (`pkg/detector/`)
**Purpose:** Detect changed files and modules

**Features:**
- Git diff analysis between commits
- Map changed files to Terragrunt modules
- Find modules affected by dependency changes

**Example:**
```go
detector := detector.NewChangeDetector(baseSHA, headSHA, workingDir)
modules, _ := detector.DetectChangedModules()
```

### 4. Executor Package (`pkg/executor/`)
**Purpose:** Execute Terragrunt commands

**Features:**
- Run `plan`, `apply`, `init`, `validate`
- Support for `run-all` operations
- Capture stdout/stderr
- Handle exit codes

**Example:**
```go
executor := executor.NewExecutor(workingDir, stdout, stderr)
result, _ := executor.Plan(ctx, modulePath)
```

### 5. Events Package (`pkg/events/`)
**Purpose:** Parse GitHub webhook events

**Features:**
- Detect event type (PR, comment, push)
- Extract PR number, SHAs, repository info
- Parse commands from comments (`/cultivator plan`)
- Determine if auto-plan should run

### 6. Orchestrator Package (`pkg/orchestrator/`)
**Purpose:** Coordinate the entire workflow

**Workflow:**
1. Parse GitHub event
2. Load configuration
3. Authenticate with GitHub
4. Detect changed modules
5. Build dependency graph
6. Execute commands in order
7. Format and post results

### 7. Formatter Package (`pkg/formatter/`)
**Purpose:** Format outputs for PR comments

**Features:**
- Parse Terraform plan summaries
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
