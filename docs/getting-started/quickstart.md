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

To use a config file, pass it explicitly with `--config`:

```bash
./cultivator plan --config=cultivator.yml
```

Example `cultivator.yml`:

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

## Step 4: Add a CI job

For production-ready CI integration, see the dedicated guides:

- **[GitHub Actions](../user-guide/github-actions.md)** -- plan on PRs, apply on merge
- **[GitLab Pipelines](../user-guide/gitlab-pipelines.md)** -- plan on MRs, apply on merge

Ready-to-use pipeline files are also available in the [`examples/`](https://github.com/Ops-Talks/cultivator/tree/main/examples) directory.

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
