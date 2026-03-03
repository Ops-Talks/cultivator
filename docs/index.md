# Cultivator

![Cultivator](assets/logo.svg)

**Cultivator** is a lightweight CLI that orchestrates **Terragrunt** stack discovery, filtering, and execution across CI/CD systems and local environments.

## Key Features

- **Automatic Stack Discovery** — Find all `terragrunt.hcl` files recursively
- **Smart Filtering** — Scope execution by environment, paths, and custom tags
- **Dependency-Aware** — Respects Terragrunt dependencies, runs stacks in correct order
- **Parallel Execution** — Configurable worker pool for fast, safe concurrent runs
- **No Server Required** — Pure CLI; works in any CI system (GitHub Actions, GitLab CI, etc.)
- **Multi-Format Output** — Human-readable text or machine-parseable JSON

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

## Why Cultivator?

### The Problem: Managing Multiple Stacks Without Orchestration

Managing Terragrunt stacks across environments typically requires custom pipeline scripts:

```yaml
# Without Cultivator: Manual stack management
plan:
  script:
    # Manual discovery
    - STACKS=$(find live/prod -name "terragrunt.hcl" -type f)
    
    # Manual filtering
    - FILTERED=()
    - for stack in $STACKS; do
        if [[ "$stack" == *"prod"* ]] && [[ "$stack" != *"experimental"* ]]; then
          FILTERED+=("$stack")
        fi
      done
    
    # Sequential execution (slow!)
    - FAILED=()
    - for stack in "${FILTERED[@]}"; do
        DIR=$(dirname "$stack")
        cd "$DIR"
        if ! terragrunt plan; then
          FAILED+=("$DIR")
        fi
        cd -
      done
    
    # Manual error reporting
    - if [ ${#FAILED[@]} -gt 0 ]; then
        echo "Failed stacks: ${FAILED[*]}"
        exit 1
      fi
```

**Problems:**
- **Manual maintenance** — Update pipeline when adding/removing stacks
- **No filtering** — Complex bash logic for environment/tag filtering
- **Sequential execution** — Slow; implementing parallelization is complex
- **No dependency handling** — Must manually order stacks
- **Inconsistent errors** — Hard to parse failures across stacks
- **Not portable** — Rewrite for each CI system

### The Solution: Automated Orchestration

Cultivator automates discovery, filtering, and execution:

```yaml
# With Cultivator: One command
plan:
  script:
    - cultivator plan --root=live --env=prod --exclude=experimental --parallelism=8 --non-interactive
```

**Benefits:**
- **Automatic discovery** — Finds all stacks with `terragrunt.hcl`
- **Smart filtering** — By environment, path patterns, and tags
- **Parallel execution** — Safe concurrent runs with configurable workers
- **Dependency awareness** — Parses and respects stack dependencies
- **Structured output** — JSON or text format with clear exit codes
- **Same everywhere** — Identical command works in CI and locally

### Real-World Example

```bash
# Scenario: Run plan on production stacks tagged "critical", exclude experiments

# Without Cultivator:
# - Write ~50 lines of bash
# - Parse terragrunt.hcl for tags manually
# - Implement dependency graph
# - Handle parallel execution with semaphores
# - Aggregate errors and format output

# With Cultivator:
cultivator plan --root=live --env=prod --tags=critical --exclude=experimental --parallelism=4
```

### Key Advantages Over Shell Scripts

| Requirement | Shell Scripts | Cultivator |
|-------------|---------------|------------|
| **Add new stack** | Update pipeline code | Automatic discovery |
| **Filter by env** | Manual bash logic | `--env=prod` |
| **Filter by tags** | Parse HCL manually | `--tags=critical` |
| **Parallel execution** | Implement semaphores | `--parallelism=8` |
| **Handle dependencies** | Manual ordering | Automatic graph parsing |
| **Local testing** | Rewrite for local use | Same command everywhere |
| **Error reporting** | Custom aggregation | Structured JSON/text output |

## What's Different?

| Feature | Cultivator | Raw Terragrunt | Atlantis |
|---------|-----------|-----------------|----------|
| **CLI-based** | Yes | Yes | No (webhook) |
| **Works locally** | Yes | Yes | No |
| **Stack discovery** | Yes | No | Yes |
| **Dependency graph** | Yes | No | Yes |
| **CI/CD agnostic** | Yes | Yes | No (GitHub only) |
| **Parallel execution** | Yes | Partial | Yes |
| **No server required** | Yes | Yes | No |

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
