# Installation

Install Cultivator as a CLI in your CI pipelines or locally for development.

## Prerequisites

- **Terragrunt** v0.50.0+ (or latest)
- **Go** v1.25+ for building from source
- **OpenTofu** or **Terraform** installed in your CI environment

## Option 1: Download a Pre-compiled Binary

Download the latest release from the [Releases page](https://github.com/Ops-Talks/cultivator/releases):

```bash
# Replace vX.Y.Z with the desired version from the Releases page
CULTIVATOR_VERSION="vX.Y.Z"
curl -L "https://github.com/Ops-Talks/cultivator/releases/download/${CULTIVATOR_VERSION}/cultivator-linux-amd64" -o cultivator
chmod +x cultivator
sudo mv cultivator /usr/local/bin/cultivator
```

## Option 2: Build from Source

```bash
go build -o cultivator ./cmd/cultivator
./cultivator plan --root=live --env=dev --non-interactive
```

## Option 3: CI/CD Integration

For production-ready CI workflows, see the dedicated guides:

- **[GitHub Actions](../user-guide/github-actions.md)** -- full plan/apply lifecycle with doctor, PR comments, and artifacts
- **[GitLab Pipelines](../user-guide/gitlab-pipelines.md)** -- full plan/apply lifecycle with MR comments and manual approval

Ready-to-use pipeline files are also available in the [`examples/`](https://github.com/Ops-Talks/cultivator/tree/main/examples) directory.

## Configuration

The configuration file is optional. Cultivator uses the following precedence for settings:

1. **CLI Flags** (e.g., `--env`, `--root`)
2. **Environment Variables** (e.g., `CULTIVATOR_ENV`)
3. **Config File** (passed via `--config`)
4. **Defaults**

Example of using a custom config file:

```bash
cultivator plan --config=my-config.yml
```

See [Configuration](configuration.md) for more details.

## Next Steps

- [Quick Start](quickstart.md) - Run your first commands
- [Configuration](configuration.md) - Customize your setup
- [User Guide](../user-guide/index.md) - Learn available commands
