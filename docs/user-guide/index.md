# User Guide

Learn how to use Cultivator to automate Terragrunt workflows in CI and locally.

## Sections

- **[Features](features.md)** - What Cultivator can do
- **[CLI Reference](cli-reference.md)** - How to use Cultivator commands
- **[GitHub Actions](github-actions.md)** - Integrate with GitHub Actions
- **[GitLab Pipelines](gitlab-pipelines.md)** - Integrate with GitLab CI/CD
- **[Configuration](../getting-started/configuration.md)** - Configure your setup

## Quick Overview

Cultivator is a lightweight CLI that discovers Terragrunt stacks and orchestrates execution:

1. Call cultivator
2. Discovers stacks
3. Applies filters
4. Executes Terragrunt
5. Shows results

## Use Cases

### Local Development

Run plans locally to validate changes before committing:

```bash
cultivator plan --root=live --env=dev
```

### Pull Request CI

Test changes safely on PRs—no approval required until review:

```bash
# GitHub Actions or GitLab CI can run this automatically
cultivator plan --root=live --env=dev --non-interactive
```

### Main Branch Deployment

Apply approved changes after merging to main:

```bash
cultivator apply --root=live --env=prod --non-interactive --auto-approve
```

### Cross-Environment Promotion

Test in one environment, apply in others:

```bash
# Test in staging
cultivator plan --root=live --env=staging --tags=critical

# Apply to production after manual approval
cultivator apply --root=live --env=prod --tags=critical --auto-approve
```

## Key Concepts

### Stack Discovery

Cultivator recursively walks the root directory, finds all `terragrunt.hcl` files, and maintains a registry of available stacks.

### Scope Filters

Use multiple filters to narrow execution:

- `--env=dev` - by environment name
- `--include=envs/prod/* --include=envs/staging/*` - by path pattern
- `--exclude=experimental` - exclude paths
- `--tags=critical,prod` - by inline tags (recommended: `# cultivator:tags = critical,prod`)

### Dependency Ordering

Cultivator parses `dependency` blocks in Terragrunt configs and runs stacks in the correct order automatically.

### Parallelism

Run independent stacks concurrently. Default: number of CPUs. Override with `--parallelism=N`.

## Integration Paths

### GitHub Actions

Add a workflow file to run Cultivator on PR and main branch:

- See [GitHub Actions](github-actions.md) for setup and examples
- Features: matrix strategies, secrets integration, caching, artifacts

### GitLab CI/CD

Define jobs in `.gitlab-ci.yml` to orchestrate Cultivator:

- See [GitLab Pipelines](gitlab-pipelines.md) for setup and examples
- Features: parallel jobs, dependency stages, caching, environment-based deployment

### Local CLI Usage

Run Cultivator directly in your terminal:

1. Build from source or download binary
2. Execute: `./cultivator plan --root=live --env=dev`
3. Review output and commit changes

## Getting Help

- **[Features](features.md)** - Explore what Cultivator can do
- **[CLI Reference](cli-reference.md)** - Deep dive into commands, flags, and exit codes
- **[Configuration](../getting-started/configuration.md)** - Full config file reference
- **[FAQ](../faq.md)** - Common questions answered
- **[Architecture](../architecture/design.md)** - Understand how Cultivator works
