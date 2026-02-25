# Development Guide

Contribute to Cultivator and help improve the project.

## Repository Structure

```
cultivator/
├── cmd/
│   └── cultivator/        # CLI entrypoint
├── pkg/
│   ├── cmd/               # CLI command handlers
│   ├── config/            # Configuration management
│   ├── detector/          # Change detection
│   ├── events/            # GitHub events handling
│   ├── executor/          # Terragrunt execution
│   ├── formatter/         # Output formatting
│   ├── github/            # GitHub API client
│   ├── graph/             # Dependency graph
│   ├── lock/              # Locking mechanism
│   ├── orchestrator/      # Main orchestration
│   └── parser/            # Terragrunt config parsing
├── docs/                  # Documentation
├── examples/              # Example configurations
└── tests/                 # Integration tests
```

## Development Setup

### Prerequisites

- **Go 1.21+**
- **Terragrunt 1.0+**
- **Docker** (for testing)
- **Git**

### Local Setup

```bash
# Clone repository
git clone https://github.com/Ops-Talks/cultivator.git
cd cultivator

# Install dependencies
go mod download

# Build
make build

# Test
make test

# Run linters
make lint
```

## Making Changes

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/your-feature`
3. Make your changes
4. Run tests: `make test`
5. Run linters: `make lint`
6. Commit: `git commit -am 'Add feature'`
7. Push: `git push origin feature/your-feature`
8. Open a Pull Request

## Testing

We use Go's built-in testing framework with high code coverage targets:

```bash
# Run all tests
make test

# Run with coverage
make test-coverage

# Run specific test
go test ./pkg/detector -v
```

## Code Standards

- **Go**: Follow [Effective Go](https://golang.org/doc/effective_go)
- **Names**: Descriptive and clear
- **Comments**: Document public functions and complex logic
- **Tests**: Write tests for new features
- **Linting**: Must pass golangci-lint

## Commit Messages

Follow conventional commits:

```
feat: add support for dependent modules
fix: resolve lock timeout issue
docs: update installation guide
test: add tests for change detector
chore: update dependencies
```

## Documentation

Documentation lives in `docs/` and is built with MkDocs:

```bash
# Build docs locally
mkdocs serve

# Check links
lychee docs/**/*.md
```

## Common Tasks

### Adding a new command

1. Create handler in `pkg/cmd/`
2. Register in orchestrator
3. Add documentation in `docs/`
4. Write tests in corresponding `*_test.go`

### Adding a new configuration option

1. Update `pkg/config/config.go`
2. Add validation logic
3. Update `cultivator.yml` schema
4. Document in `docs/getting-started/configuration.md`

### Bug Reports

Include:
- Minimal reproduction steps
- Expected vs actual behavior
- Cultivator version
- Terragrunt version
- Environment (GitHub Actions, local, etc.)

### Feature Requests

Describe:
- Current behavior
- Desired behavior
- Use case
- Suggested implementation (if any)

## Performance Considerations

- Change detection should be O(n) with file count
- Dependency graph building should be O(n + e) with modules and edges
- Large repositories (1000+ modules) should still execute in reasonable time

## Security

- All inputs must be validated
- No secrets in logs or comments
- Validate GitHub webhook signatures
- Use constant-time comparisons for secrets

---

Questions? Open an issue or discussion on [GitHub](https://github.com/Ops-Talks/cultivator).
