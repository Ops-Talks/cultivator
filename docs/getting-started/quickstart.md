# Quick Start

Get Cultivator running in your repository in a few minutes.

No `cultivator.yml` is required for the first run.

## Step 1: Build the CLI

```bash
go build -o cultivator ./cmd/cultivator
```

## Step 2: Run a plan locally

```bash
./cultivator plan --root=live --env=dev --non-interactive
```

If your Terragrunt root is the current directory, you can also run:

```bash
./cultivator plan --non-interactive
```

## Step 3 (optional): Create a config file

Create `cultivator.yml` in the repository root only if you want defaults:

```yaml
root: live
parallelism: 4
output_format: text
non_interactive: true

plan:
  destroy: false
apply:
  auto_approve: false
destroy:
  auto_approve: false
```

Use it explicitly:

```bash
./cultivator plan --config=cultivator.yml
```

## Step 4: Add a CI job

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
- Review [CLI reference](../user-guide/cli-reference.md)

## Common Commands

```bash
# Plan dev stacks
cultivator plan --root=live --env=dev --non-interactive

# Apply changes
cultivator apply --root=live --env=dev --non-interactive --auto-approve

# Destroy
cultivator destroy --root=live --env=dev --non-interactive --auto-approve
```
