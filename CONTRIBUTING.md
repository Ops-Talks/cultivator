# Contributing to Cultivator

Thank you for considering contributing to Cultivator! This document provides guidelines and instructions for contributing.

## Getting Started

1. **Fork the repository**
2. **Clone your fork**
   ```bash
   git clone https://github.com/YOUR_USERNAME/cultivator.git
   cd cultivator
   ```
3. **Install dependencies**
   ```bash
   go mod download
   ```
4. **Build the project**
   ```bash
   go build -o cultivator ./cmd/cultivator
   ```

## Development Setup

### Prerequisites
- Go 1.25 or higher
- Terragrunt 0.55.0 or higher
- Terraform 1.7.0 or higher
- Git

### Running Tests
```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests with verbose output
go test -v ./...
```

### Running Locally
```bash
# Build
go build -o cultivator ./cmd/cultivator

# Run
./cultivator --help
```

### Pre-commit Hooks

We use pre-commit hooks to ensure code quality and consistency before commits are made. This helps catch issues early and maintains a high-quality codebase.

#### Installation

1. **Install pre-commit framework** (one-time setup):
   ```bash
   # Using pip
   pip install pre-commit
   
   # Or using homebrew (macOS)
   brew install pre-commit
   
   # Or using apt (Ubuntu/Debian)
   sudo apt install pre-commit
   ```

2. **Install the hooks** in the repository:
   ```bash
   make pre-commit-install
   
   # Or manually
   pre-commit install --install-hooks
   ```

3. **Verify installation**:
   ```bash
   pre-commit run --all-files
   ```

#### What the Hooks Do

The pre-commit hooks automatically run the following checks before every commit:

**Code Quality:**
- `gofmt` - Formats Go code with simplification
- `goimports` - Organizes and formats import statements
- `go vet` - Checks for common Go mistakes
- `go test` - Runs tests for changed packages
- `go mod tidy` - Ensures go.mod and go.sum are up to date
- `go build` - Verifies the code compiles

**Static Analysis:**
- `golangci-lint` - Comprehensive linting with 30+ linters (see [.golangci.yml](.golangci.yml))
- `go-critic` - Additional Go code checks

**Security:**
- `trufflehog` - Scans for secrets and credentials
- `detect-private-key` - Prevents committing private keys

**General:**
- Trailing whitespace removal
- End-of-file fixer
- YAML validation
- Large file detection
- Merge conflict markers
- Mixed line endings

#### Usage

**Automatic (Recommended):**
Hooks run automatically on `git commit`. If any check fails, the commit is blocked and you'll see what needs to be fixed.

**Manual:**
```bash
# Run all hooks on all files
make pre-commit-run

# Run all hooks on staged files only
pre-commit run

# Run specific hook
pre-commit run golangci-lint --all-files

# Skip hooks (not recommended)
git commit --no-verify -m "message"
```

#### Maintenance

```bash
# Update hooks to latest versions
make pre-commit-update

# Uninstall hooks
make pre-commit-uninstall
```

#### Bypassing Hooks (Emergency Use Only)

If you absolutely need to bypass hooks (e.g., work-in-progress commit), use:
```bash
git commit --no-verify -m "WIP: message"
```

**Note:** CI will still run all checks, so this should only be used for temporary commits.

#### Troubleshooting

**Problem:** `pre-commit: command not found`
```bash
# Ensure pre-commit is installed
pip install --user pre-commit
# Add to PATH if needed
export PATH="$HOME/.local/bin:$PATH"
```

**Problem:** Hooks fail on first run
```bash
# Clean cache and reinstall
pre-commit clean
pre-commit install --install-hooks
pre-commit run --all-files
```

**Problem:** `golangci-lint` takes too long
```bash
# Run only on changed files (default behavior)
pre-commit run golangci-lint

# Or adjust timeout in .pre-commit-config.yaml
```

## Making Changes

### Code Style
- Follow standard Go conventions
- Use `gofmt` to format your code
- Run `go vet` to check for common issues
- Add comments for exported functions and types

### Commit Messages
- Use clear and descriptive commit messages
- Start with a verb in present tense (e.g., "Add", "Fix", "Update")
- Reference issue numbers when applicable

Example:
```
Add support for GitLab CI integration

- Implement GitLab API client
- Add GitLab webhook handler
- Update documentation

Fixes #123
```

### Pull Request Process

1. **Create a branch** for your changes
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. **Make your changes** following the code style guidelines

3. **Add tests** for new functionality

4. **Update documentation** if needed

5. **Run tests** to ensure everything works
   ```bash
   go test ./...
   ```

6. **Commit your changes**
   ```bash
   git add .
   git commit -m "Your descriptive commit message"
   ```

7. **Push to your fork**
   ```bash
   git push origin feature/your-feature-name
   ```

8. **Create a Pull Request** on GitHub

### PR Guidelines
- Provide a clear description of the changes
- Reference related issues
- Ensure all tests pass
- Update documentation as needed
- Keep PRs focused on a single feature/fix

## Project Structure

```
cultivator/
├── cmd/cultivator/          # CLI entry point
├── pkg/
│   ├── cmd/                 # CLI commands
│   ├── config/              # Configuration handling
│   ├── detector/            # Change detection
│   ├── executor/            # Terragrunt execution
│   ├── github/              # GitHub integration
│   ├── lock/                # Lock management
│   └── parser/              # HCL parsing
├── examples/                # Example configurations
└── docs/                    # Documentation
```

## Areas for Contribution

We welcome contributions in these areas:

- **CI/CD Integrations**: GitLab CI, Azure DevOps, CircleCI, etc.
- **Features**: Cost estimation, drift detection, policy enforcement
- **Documentation**: Tutorials, guides, examples
- **Testing**: Unit tests, integration tests, e2e tests
- **Bug Fixes**: Check our issues for bugs to fix

## Questions?

If you have questions or need help:
- Open an issue for discussion
- Check existing documentation
- Reach out to maintainers

## Code of Conduct

Be respectful and inclusive. We're all here to build something great together.

## License

By contributing, you agree that your contributions will be licensed under the MIT License.
