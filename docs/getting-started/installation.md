# Installation

Install Cultivator as a CLI in your CI pipelines or locally for development.

## Prerequisites

- **Terragrunt** v0.50.0+ (or latest)
- **Go** v1.25+ for building from source
- **OpenTofu** or **Terraform** installed in your CI environment

## Option 1: GitHub Actions

For maximum performance in CI, downloading the pre-compiled binary is recommended. Create `.github/workflows/cultivator.yml`:

```yaml
name: Cultivator Plan

on:
  pull_request:

jobs:
  plan:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Install Cultivator
        run: |
          CULTIVATOR_VERSION="v1.0.0" # Use the desired version
          curl -L https://github.com/Ops-Talks/cultivator/releases/download/${CULTIVATOR_VERSION}/cultivator-linux-amd64 -o cultivator
          chmod +x cultivator
          sudo mv cultivator /usr/local/bin/cultivator

      - name: Install Terragrunt
        run: |
          curl -L https://github.com/gruntwork-io/terragrunt/releases/latest/download/terragrunt_linux_amd64 -o terragrunt
          chmod +x terragrunt
          sudo mv terragrunt /usr/local/bin/terragrunt

      - name: Cultivator plan
        run: cultivator plan --root=live --env=dev --non-interactive
```

## Option 2: GitLab CI

Example `.gitlab-ci.yml` using a Docker image or a direct download:

```yaml
stages:
  - plan
  - apply

plan:
  stage: plan
  image: alpine:latest
  before_script:
    - apk add --no-cache curl
    - CULTIVATOR_VERSION="v1.0.0"
    - curl -L https://github.com/Ops-Talks/cultivator/releases/download/${CULTIVATOR_VERSION}/cultivator-linux-amd64 -o /usr/local/bin/cultivator
    - chmod +x /usr/local/bin/cultivator
    - # Install Terragrunt and OpenTofu/Terraform
  script:
    - cultivator plan --root=live --env=dev --non-interactive
  only:
    - merge_requests

apply:
  stage: apply
  image: alpine:latest
  before_script:
    - # Install dependencies
  script:
    - cultivator apply --root=live --env=dev --non-interactive --auto-approve
  when: manual
  only:
    - main
```

## Option 3: Local Installation

Build from source:

```bash
go build -o cultivator ./cmd/cultivator
./cultivator plan --root=live --env=dev --non-interactive
```

Alternatively, download a binary from the GitHub [Releases page](https://github.com/Ops-Talks/cultivator/releases).

## Configuration

The configuration file is optional. Cultivator uses the following precedence for settings:

1. **CLI Flags** (e.g., `--env`, `--root`)
2. **Environment Variables** (e.g., `CULTIVATOR_ENV`)
3. **Config File** (`cultivator.yml` or passed via `--config`)
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
