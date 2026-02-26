# User Guide

Learn how to use Cultivator to automate Terragrunt workflows in CI.

## Sections

- **[Features](features.md)** - What Cultivator can do
- **[Workflows](workflows.md)** - How to use Cultivator commands
- **[GitLab Pipelines](gitlab-pipelines.md)** - Integrate with GitLab CI/CD
- **[Configuration](../getting-started/configuration.md)** - Configure your setup

## Quick Overview

Cultivator is a CLI that discovers Terragrunt modules and executes commands with consistent output:

```
CI job starts -> Cultivator discovers modules -> Terragrunt runs -> Logs + exit code
```

## Key Concepts

### Module Discovery
Cultivator walks the root directory, finds `terragrunt.hcl` modules, and applies filters.

### Filters
Use `--env`, `--include`, `--exclude`, and `--tags` to scope execution.

### Parallelism
Run modules concurrently with a configurable worker pool.

### Output Formats
Choose `text` or `json` logging for CI visibility and automation.

## Getting Help

- Check [Features](features.md) for what you can do
- Read [Workflows](workflows.md) for command details
- See [Configuration](../getting-started/configuration.md) for setup options
- Browse the [FAQ](../faq.md)
