# Quick Start

Get Cultivator running in your repository in a few minutes.

## Step 1: Build the CLI

```bash
go build -o cultivator ./cmd/cultivator
```

## Step 2: Run a plan locally

```bash
./cultivator plan --root=live --env=dev --non-interactive
```

## Step 3: Add a CI job

Example GitHub Actions workflow for `plan` on pull requests:

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

## Next Steps

- Learn about [configuration](configuration.md)
- Review [workflows](../user-guide/workflows.md)

## Common Commands

```bash
# Plan dev stacks
cultivator plan --root=live --env=dev --non-interactive

# Apply changes
cultivator apply --root=live --env=dev --non-interactive --auto-approve

# Destroy
cultivator destroy --root=live --env=dev --non-interactive --auto-approve
```
