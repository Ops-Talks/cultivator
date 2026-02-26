# Installation

Install Cultivator as a CLI in your CI pipelines or locally for development.

## Prerequisites

- Terragrunt v0.50.0+ (or latest)
- Go v1.25+ for building from source
- OpenTofu or Terraform installed in your CI environment

## Option 1: GitHub Actions

Create `.github/workflows/cultivator.yml`:

```yaml
name: Cultivator Plan

on:
  pull_request:

jobs:
  plan:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Setup Go
        uses: actions/setup-go@v5
        with:
          go-version: 'stable'

      - name: Build Cultivator
        run: go build -o bin/cultivator ./cmd/cultivator

      - name: Install Terragrunt
        run: |
          curl -L https://github.com/gruntwork-io/terragrunt/releases/latest/download/terragrunt_linux_amd64 -o terragrunt
          chmod +x terragrunt
          sudo mv terragrunt /usr/local/bin/terragrunt

      - name: Cultivator plan
        run: ./bin/cultivator plan --root=live --env=dev --non-interactive
```

## Option 2: GitLab CI

Create `.gitlab-ci.yml`:

```yaml
stages:
  - plan
  - apply

plan:
  stage: plan
  image: golang:1.25
  script:
    - go build -o bin/cultivator ./cmd/cultivator
    - ./bin/cultivator plan --root=live --env=dev --non-interactive
  only:
    - merge_requests

apply:
  stage: apply
  image: golang:1.25
  script:
    - go build -o bin/cultivator ./cmd/cultivator
    - ./bin/cultivator apply --root=live --env=dev --non-interactive --auto-approve
  when: manual
  only:
    - main
```

## Option 3: Local Installation

```bash
go build -o cultivator ./cmd/cultivator
./cultivator plan --root=live --env=dev --non-interactive
```

## Configuration

Create a `.cultivator.yaml` (or `cultivator.yml`) in your repository root. See [Configuration](configuration.md) for details.

## Next Steps

- [Quick Start](quickstart.md) - Run your first commands
- [Configuration](configuration.md) - Customize your setup
- [User Guide](../user-guide/index.md) - Learn available commands
