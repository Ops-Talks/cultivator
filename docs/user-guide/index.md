# User Guide

Learn how to use Cultivator to automate Terragrunt workflows in CI and locally.

## Sections

- **[Features](features.md)** - What Cultivator can do
- **[Workflows](workflows.md)** - How to use Cultivator commands
- **[GitHub Actions](github-actions.md)** - Integrate with GitHub Actions
- **[GitLab Pipelines](gitlab-pipelines.md)** - Integrate with GitLab CI/CD
- **[Configuration](../getting-started/configuration.md)** - Configure your setup

## Quick Overview

Cultivator is a lightweight CLI that discovers Terragrunt modules and orchestrates execution:

```
You call cultivator → Discovers modules → Applies filters → Executes Terragrunt → Shows results
```

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

### Module Discovery
Cultivator recursively walks the root directory, finds all `terragrunt.hcl` files, and maintains a registry of available modules.

### Scope Filters
Use multiple filters to narrow execution:
- `--env=dev` - by environment variable
- `--include=envs/prod/* --include=envs/staging/*` - by path pattern
- `--exclude=experimental` - exclude paths
- `--tags=critical,prod` - by inline tags (e.g., `# cultivator:tags=critical,prod`)

### Dependency Ordering
Cultivator parses `dependency` blocks in Terragrunt configs and runs modules in the correct order automatically.

### Parallelism
Run independent modules concurrently. Default: 4 workers. Override with `--parallelism=N`.

### Output Formats
- `text` (default) - human-readable logs
- `json` - machine-parseable for CI integrations

## Integration Paths

### GitHub Actions
Add a workflow file to run Cultivator on PR and main branch:
- See [GitHub Actions](github-actions.md) for setup and examples
- Features: matrix strategies, secrets integration, caching, artifacts

### GitLab CI/CD
Define jobs in `.gitlab-ci.yml` to orchestrate Cultivator:
- See [GitLab Pipelines](gitlab-pipelines.md) for setup and examples
- Features: parallel jobs, dependency stages, caching, environment-based deployment

### Local Development
Run Cultivator directly in your terminal:
1. Build: `make build` or `go build -o cultivator ./cmd/cultivator`
2. Execute: `./cultivator plan --root=live --env=dev`
3. Review output and commit changes

## Getting Help

- **[Features](features.md)** - Explore what Cultivator can do
- **[Workflows](workflows.md)** - Deep dive into commands, flags, and exit codes
- **[Configuration](../getting-started/configuration.md)** - Full config file reference
- **[FAQ](../faq.md)** - Common questions answered
- **[Architecture](../architecture/design.md)** - Understand how Cultivator works


