# User Guide

Learn how to use Cultivator to automate your Terragrunt workflows.

## Sections

- **[Features](features.md)** - What Cultivator can do
- **[Configuration](../getting-started/configuration.md)** - How to configure your setup
- **[Workflows](workflows.md)** - How to use Cultivator commands

## Quick Overview

Cultivator runs Terragrunt operations from PR comments:

```
User comments → Cultivator detects → Executes with dependencies → Posts results
```

## Key Concepts

### Change Detection
Cultivator automatically detects which modules are affected by changes in your PR.

### Dependency Awareness
Operations respect Terragrunt dependencies and run modules in the correct order.

### Locking
Prevents concurrent applies to the same module, ensuring safe deployments.

### Rich Output
Results are formatted nicely in PR comments with plan summaries and status indicators.

## Getting Help

- Check [Features](features.md) for what you can do
- Read [Workflows](workflows.md) for command details
- See [Configuration](../getting-started/configuration.md) for setup options
- Browse the [FAQ](../faq.md)
