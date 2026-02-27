# Architecture Design

Understand how Cultivator works under the hood.

## Overview

Cultivator is a **CLI-first tool** that orchestrates Terragrunt module discovery, filtering, and execution. It runs as a command in CI/CD systems (GitHub Actions, GitLab CI) or locally, and delegates state management entirely to Terraform/OpenTofu backends.

Unlike PR-based automation, Cultivator is **job-triggered**: you call it explicitly in your CI workflow, providing flags to control scope and behavior.

## Core Components

### 1. Config Loader
- Reads `cultivator.yml` / `.cultivator.yaml` from repository root
- Merges configuration with CLI flags and environment variables
- Validates configuration against schema

### 2. Module Discovery
- Walks the root directory recursively
- Finds all `terragrunt.hcl` files
- Builds list of available modules

### 3. Scope Filter
- Filters modules by `--env` (environment)
- Filters modules by `--include` / `--exclude` (path patterns)
- Filters modules by `--tags` (tag comments in `terragrunt.hcl`)

### 4. Dependency Graph
- Parses `terragrunt.hcl` files
- Builds dependency graph from `dependency` blocks
- Determines correct execution order

### 5. Executor
- Runs Terragrunt commands (`plan`, `apply`, `destroy`)
- Manages parallel execution via worker pool
- Handles locks to prevent concurrent applies
- Captures output on per-module, per-command basis

### 6. Output Formatter
- Formats results as text or JSON
- Redacts sensitive data (passwords, tokens)
- Reports exit codes and errors

## Data Flow

```
CLI Invocation (plan/apply/destroy)
    ↓
Config Loader → Parse flags + env vars + cultivator.yml
    ↓
Module Discovery → Find all terragrunt.hcl files
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
- Cultivator does not manage state, backends, or locks beyond module-level coordination
- All state is owned by Terraform/OpenTofu
- Locks are per-module filesystem locks to prevent concurrent writes

### 2. Explicit Control
- No automatic triggers or webhooks
- User (or CI job) explicitly calls `cultivator plan/apply/destroy`
- Flags control what runs and how

### 3. Filter-First
- Start with all modules, then filter by environment/path/tags
- Filters are composable (combine multiple `--include`, `--exclude`, `--tags`)
- Results show exactly which modules will be affected

### 4. Dependency-Aware
- Respects Terragrunt `dependency` blocks
- Runs modules in correct order (topological sort)
- Prevents applying module A before its dependency B

### 5. Parallel by Default
- Configurable worker pool (default: 4 workers)
- Runs independent modules concurrently
- Respects dependency graph for safe parallelism

### 6. Output Agnostic
- Formats for human (text) or machine (JSON) consumption
- Redacts secrets from logs
- Compatible with GitHub Actions, GitLab CI, and local debugging

## Configuration Reference

Cultivator reads from `cultivator.yml` in the repository root:

```yaml
root: live                    # Root directory to scan for modules
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

## Locking Mechanism

Cultivator uses **file-based locks** to prevent concurrent applies on the same module:

- Lock file: `.terraform/cultivator.lock` (per module)
- Timeout: 30 minutes (configurable)
- Retry: Exponential backoff
- Failure: Prints clear error message with lock holder details

This is a safety mechanism; actual state locking is managed by Terraform/OpenTofu.

## Security Considerations

### Secrets Redaction
- Environment variables and output are scanned for common secret patterns
- Passwords, API keys, and tokens are masked in logs
- Compatible with GitHub Actions and GitLab CI secret masking

### Access Control
- Cultivator respects IAM permissions of the CI runner
- If the CI job lacks permissions to modify a module, Terraform will error
- No additional RBAC layer in Cultivator itself

### State Backend Protection
- All state is stored in the backend (S3, Terraform Cloud, etc.)
- Cultivator does not access or modify state directly
- Backend authentication is handled by Terragrunt/Terraform
