# Cultivator

![Cultivator](assets/logo.svg)

**Cultivator** is a lightweight CLI that orchestrates **Terragrunt** stack discovery, filtering, and execution across CI/CD systems and local environments.

## Key Features

✅ **Automatic Stack Discovery** — Find all `terragrunt.hcl` files recursively  
✅ **Smart Filtering** — Scope execution by environment, paths, and custom tags  
✅ **Dependency-Aware** — Respects Terragrunt dependencies, runs stacks in correct order  
✅ **Parallel Execution** — Configurable worker pool for fast, safe concurrent runs  
✅ **No Server Required** — Pure CLI; works in any CI system (GitHub Actions, GitLab CI, etc.)  
✅ **Secret Redaction** — Automatically masks sensitive data in logs  
✅ **Multi-Format Output** — Human-readable text or machine-parseable JSON  

## Overview

Cultivator solves the complexity of orchestrating Terragrunt:

- **Discovers** all Terragrunt stacks under a root directory
- **Filters** stacks by environment, path patterns, and tags
- **Respects** stack dependencies and executes in correct order
- **Executes** Terragrunt commands (`plan`, `apply`, `destroy`) in parallel when safe
- **Reports** results with clear exit codes and formatted output

Unlike webhook-based automation, Cultivator is a **CLI you invoke explicitly** from CI jobs or locally.

## Quick Links

<div class="grid cards" markdown>

- **[Getting Started](getting-started/index.md)** - Installation, configuration, first steps
- **[User Guide](user-guide/index.md)** - Commands, workflows, CI integrations
- **[Architecture](architecture/design.md)** - Technical design and internals
- **[FAQ](faq.md)** - Common questions answered

</div>

## Typical Workflow

### 1. Local Development
```bash
./cultivator plan --root=live --env=dev
# Review plan, iterate on infrastructure code
```

### 2. Create Pull Request
```bash
git push origin feature-branch
# CI automatically runs: cultivator plan --root=live --env=dev
```

### 3. Review & Merge
```bash
# Team reviews plan output, merges to main after approval
```

### 4. Production Deployment  
```bash
# CI automatically runs: cultivator apply --root=live --env=prod --auto-approve
```

## Use Case Examples

### Multi-Environment Deployment
```bash
cultivator apply --root=live --env=prod --tags=critical --auto-approve
```

### Target Specific Stacks
```bash
cultivator plan --root=live \
  --include=envs/prod/database \
  --include=envs/prod/app
```

### Exclude Experimental Infrastructure
```bash
cultivator plan --root=live --exclude=experimental --non-interactive
```

## What's Different?

| Feature | Cultivator | Raw Terragrunt | Atlantis |
|---------|-----------|-----------------|----------|
| **CLI-based** | ✅ | ✅ | ❌ (webhook) |
| **Works locally** | ✅ | ✅ | ❌ |
| **Stack discovery** | ✅ | ❌ | ✅ |
| **Dependency graph** | ✅ | ❌ | ✅ |
| **CI/CD agnostic** | ✅ | ✅ | ❌ (GitHub only) |
| **Parallel execution** | ✅ | Partial | ✅ |
| **No server required** | ✅ | ✅ | ❌ |

## Getting Started

### Build Locally
```bash
go build -o cultivator ./cmd/cultivator
./cultivator version
```

### Run in Docker
```bash
make docker-build
docker run cultivator:latest plan --help
```

### Integrate with CI/CD
- **[GitHub Actions](user-guide/github-actions.md)** — Full workflow examples
- **[GitLab CI/CD](user-guide/gitlab-pipelines.md)** — Full pipeline examples

## Getting Help

- **[User Guide](user-guide/index.md)** — Learn all commands and features
- **[FAQ](faq.md)** — Common questions answered
- **[GitHub Issues](https://github.com/Ops-Talks/cultivator/issues)** — Report bugs
- **[GitHub Discussions](https://github.com/Ops-Talks/cultivator/discussions)** — Ask questions
- **[Contributing Guide](../CONTRIBUTING.md)** — Contribute to the project

## Requirements

- **Terragrunt** v0.50.0+ (recommended: v1.0+)
- **OpenTofu** v1.6+ or **Terraform** v1.5+
- **Go** v1.25+ (only to build from source)

## License

Cultivator is licensed under the MIT License. See [LICENSE](../LICENSE) for details.
