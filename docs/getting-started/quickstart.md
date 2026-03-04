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

## Next: CI/CD Integration

Once you're comfortable running Cultivator locally, set up CI/CD pipelines:

- [GitHub Actions Integration](../user-guide/github-actions.md) - Production-ready workflows
- [GitLab CI Integration](../user-guide/gitlab-pipelines.md) - Complete pipeline examples

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
