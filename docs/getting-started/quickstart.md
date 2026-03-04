# Quick Start

Get Cultivator running in your repository in a few minutes.

A `cultivator.yml` configuration file is not required for basic usage; Cultivator will use default settings if it is missing.

## Step 1: Build or Install the CLI

Build from source:

```bash
go build -o cultivator ./cmd/cultivator
```

Alternatively, you can download a pre-compiled binary from the [Releases page](https://github.com/Ops-Talks/cultivator/releases).

## Step 2: Run a plan locally

Cultivator automatically discovers Terragrunt stacks. Run a plan by specifying the root directory and the environment:

```bash
./cultivator plan --root=live --env=dev --non-interactive
```

If your Terragrunt root is the current directory, you can also run:

```bash
./cultivator plan --non-interactive
```

## Step 3 (optional): Create a config file

Cultivator automatically looks for `cultivator.yml` in the current directory. You can use it to set project-wide defaults:

```yaml
root: live
parallelism: 4
non_interactive: true

plan:
  destroy: false
apply:
  auto_approve: false
destroy:
  auto_approve: false
```

You can also specify a custom configuration file path using the `--config` flag:

```bash
./cultivator plan --config=custom-config.yml
```

## Step 4: Add a CI job

Example GitHub Actions workflow for plan on pull requests. Note that downloading a pre-compiled binary is more efficient than building from source in every job:

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
