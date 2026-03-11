# Examples

This directory contains ready-to-use CI/CD pipeline configurations for Cultivator.

## Pipeline Examples

| File | Description |
|------|-------------|
| [`github-actions.yml`](github-actions.yml) | GitHub Actions pipeline: plan on PRs, apply on merge to main |
| [`gitlab-ci.yml`](gitlab-ci.yml) | GitLab CI/CD pipeline: plan on MRs, apply on merge to main |

Copy the relevant file into your infrastructure repository and adjust paths, versions, and environment variables as needed.

## How Cultivator Works in CI

1. **Discovery**: Cultivator walks the root directory and finds all `terragrunt.hcl` files
2. **Filtering**: It applies scope filters (environment, paths, tags, Git changes)
3. **Dependency Analysis**: It parses `dependency` blocks and builds a DAG for execution order
4. **Execution**: It runs Terragrunt commands in parallel, respecting the dependency graph

## Example Scenarios

### Scenario 1: Change VPC in Dev

- **Changed**: `environments/dev/vpc/terragrunt.hcl`
- **Affected modules**:
  - `environments/dev/vpc` (direct)
  - `environments/dev/database` (depends on vpc)
  - `environments/dev/app` (depends on vpc and database)
- **Execution order**: vpc -> database -> app

### Scenario 2: Change App in Staging

- **Changed**: `environments/staging/app/terragrunt.hcl`
- **Affected modules**:
  - `environments/staging/app` (direct)
- **Execution order**: app

### Scenario 3: Change Root Config

- **Changed**: `terragrunt.hcl` (root)
- **Affected modules**: ALL (because all inherit from root)
- **Execution order**: Topologically sorted based on dependencies

## Further Reading

- [Installation Guide](../docs/getting-started/installation.md) for full setup instructions
- [GitHub Actions Guide](../docs/user-guide/github-actions.md) for detailed GitHub Actions patterns
- [GitLab Pipelines Guide](../docs/user-guide/gitlab-pipelines.md) for detailed GitLab CI patterns
