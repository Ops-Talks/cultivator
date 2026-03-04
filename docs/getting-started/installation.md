# Installation

Install Cultivator as a CLI in your CI pipelines or locally for development.

## Prerequisites

- Terragrunt v0.50.0+ (or latest)
- Go v1.25+ for building from source
- OpenTofu or Terraform installed in your CI environment

## Local Installation

```bash
go build -o cultivator ./cmd/cultivator
./cultivator plan --root=live --env=dev --non-interactive
```

## Configuration

Configuration file is optional.

- Without file: pass everything via flags and/or environment variables.
- With file: create one (for example `cultivator.yml`) and pass it explicitly with `--config`.

Example:

```bash
./bin/cultivator plan --config=cultivator.yml
```

See [Configuration](configuration.md) for supported keys and precedence.

## CI/CD Integration

For detailed CI/CD setup instructions and complete workflow examples:

- **[GitHub Actions Integration](../user-guide/github-actions.md)** — Full workflow examples with best practices
- **[GitLab CI Integration](../user-guide/gitlab-pipelines.md)** — Complete pipeline configurations

## Next Steps

- [Quick Start](quickstart.md) - Run your first commands
- [Configuration](configuration.md) - Customize your setup
- [User Guide](../user-guide/index.md) - Learn available commands
