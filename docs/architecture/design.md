# Architecture Design

Understand how Cultivator works under the hood.

## Overview

Cultivator is a **CLI-first tool** that orchestrates Terragrunt stack discovery, filtering, and execution. It runs as a command in CI/CD systems (GitHub Actions, GitLab CI) or locally, and delegates state management entirely to Terraform/OpenTofu backends.

Unlike PR-based automation, Cultivator is **job-triggered**: you call it explicitly in your CI workflow, providing flags to control scope and behavior.

## Core Components

### 1. Config Loader
- Loads built-in defaults, optional config file (`--config`), environment variables, and CLI flags
- Merges configuration with CLI flags and environment variables
- Validates configuration against schema

### 2. Stack Discovery
- Walks the root directory recursively
- Finds all `terragrunt.hcl` files
- Builds list of available stacks

### 3. Scope Filter
- Filters stacks by `--env` (environment)
- Filters stacks by `--include` / `--exclude` (path patterns)
- Filters stacks by `--tags` (tag comments in `terragrunt.hcl`)

### 4. Dependency Graph
- Parses `terragrunt.hcl` files
- Builds dependency graph from `dependency` blocks
- Determines correct execution order

### 5. Executor
- Runs Terragrunt commands (`plan`, `apply`, `destroy`)
- Manages parallel execution via worker pool
- Captures stdout and stderr in a single chronologically-ordered stream using `cmd.CombinedOutput()`
- Captures output on per-stack, per-command basis

### 6. Output Formatter
- Formats results as text or JSON
- Redacts sensitive data (passwords, tokens)
- Reports exit codes and errors

## Data Flow

```md
CLI Invocation (plan/apply/destroy)
    ↓
Config Loader → Parse defaults + optional --config file + env vars + flags
    ↓
Stack Discovery → Find all terragrunt.hcl files
    ↓
Scope Filter → Apply --env, --include, --exclude, --tags
    ↓
Dependency Graph → Build execution order
    ↓
Executor (parallel worker pool) → Run Terragrunt commands
    ↓
Output Formatter → Display results (text or JSON)
    ↓
Exit Code (0 = success, 1 = failure, 2 = usage error)
```

## Key Design Principles

### 1. Stateless Operation
- Cultivator does not manage state, backends, or locks beyond stack-level coordination
- All state is owned by Terraform/OpenTofu
- Locks are per-stack filesystem locks to prevent concurrent writes

### 2. Explicit Control
- No automatic triggers or webhooks
- User (or CI job) explicitly calls `cultivator plan/apply/destroy`
- Flags control what runs and how

### 3. Filter-First
- Start with all stacks, then filter by environment/path/tags
- Filters are composable (combine multiple `--include`, `--exclude`, `--tags`)
- Results show exactly which stacks will be affected

### 4. Dependency-Aware
- Respects Terragrunt `dependency` blocks
- Runs stacks in correct order (topological sort)
- Prevents applying stack A before its dependency B

### 5. Parallel by Default
- Configurable worker pool (default: 4 workers)
- Runs independent stacks concurrently
- Respects dependency graph for safe parallelism

### 6. Output Agnostic
- Formats for human (text) or machine (JSON) consumption
- Redacts secrets from logs
- Compatible with GitHub Actions, GitLab CI, and local debugging

## Configuration Reference

Cultivator reads from `cultivator.yml` in the repository root:

```yaml
root: live                    # Root directory to scan for stacks
parallelism: 4               # Worker pool size
output_format: text          # 'text' or 'json'
non_interactive: false       # Equivalent to -input=false
plan:
  destroy: false             # Defaults for 'plan' subcommand
apply:
  auto_approve: true         # Defaults for 'apply' subcommand
destroy:
  auto_approve: true         # Defaults for 'destroy' subcommand
```

CLI flags and environment variables override the config file. Flags take highest precedence.

See [Configuration](../getting-started/configuration.md) for full reference.

## Security Considerations

### Secrets Redaction
- Cultivator does not implement secret redaction in its own output layer
- Use your CI platform's built-in secret masking (GitHub Actions masked secrets, GitLab CI variable masking)
- Mark sensitive Terraform outputs with `sensitive = true` so Terragrunt/OpenTofu suppress them in plan output

### Access Control
- Cultivator respects IAM permissions of the CI runner
- If the CI job lacks permissions to modify a stack, Terraform will error
- No additional RBAC layer in Cultivator itself

### State Backend Protection
- All state is stored in the backend (S3, Terraform Cloud, etc.)
- Cultivator does not access or modify state directly
- Backend authentication is handled by Terragrunt/Terraform
