# Features

Cultivator provides practical features for automating Terragrunt workflows in CI.

## Stack Discovery

Cultivator discovers stacks by walking the root directory and locating `terragrunt.hcl` files.

## Scope Filters

Limit execution to specific areas:

- `--env` for environment filtering
- `--include` / `--exclude` for path filtering
- `--tags` for tag filtering (via `cultivator:tags` comments)

### Tag Filtering

Use tags to logically group stacks and selectively execute them. Tags can be defined as comments or as an HCL variable in `terragrunt.hcl` files.

**Option 1: Syntax as Comment**

```hcl
# cultivator:tags=critical,prod
terraform {
  source = "./modules/database"
}
```

**Option 2: Syntax as HCL Variable**

```hcl
cultivator_tags = ["app", "api"]
terraform {
  source = "./modules/api"
}
```

**Usage with CLI:**

```bash
# Execute only stacks with 'critical' tag
cultivator plan --root=live --tags=critical --non-interactive

# Execute stacks with either 'prod' or 'staging' tag
cultivator apply --root=live --tags=prod,staging --non-interactive --auto-approve
```

**Usage in `cultivator.yml`:**

```yaml
root: live
tags:
  - critical
  - prod
parallelism: 4
```

**Real-world example:**

Your infrastructure might have:

```
live/
  prod/
    database/
      terragrunt.hcl          # cultivator:tags=critical,prod
    api/
      terragrunt.hcl          # cultivator:tags=prod
    cache/
      terragrunt.hcl          # cultivator:tags=optional,prod
  dev/
    database/
      terragrunt.hcl          # cultivator:tags=dev
```

Execute only critical production systems:

```bash
cultivator plan --root=live --env=prod --tags=critical --non-interactive
```

This will only execute `live/prod/database/terragrunt.hcl`.

## Dependency-Aware Execution (DAG)

Cultivator automatically detects and respects Terragrunt `dependency` blocks. It builds a Directed Acyclic Graph (DAG) to determine the safest and most efficient execution order.

- **Safe Ordering**: If Module B depends on Module A, Cultivator ensures Module A completes successfully before starting Module B.
- **Maximized Parallelism**: Modules that do not depend on each other are executed concurrently, up to the defined `--parallelism` limit.
- **Cycle Detection**: Cultivator validates your project structure and will error out if it detects circular dependencies, preventing infinite loops.

## Magic Mode (Git Integration)

"Magic Mode" allows you to execute commands only on modules affected by code changes in a Pull Request or a specific commit range.

- **Automatic Filtering**: Use `--changed-only` to let Cultivator query Git for modified files and automatically target only the relevant Terragrunt modules.
- **Smart Mapping**: Changes to a file inside a module's directory (or its subdirectories) will trigger that module.
- **Custom Base**: Use `--base=main` (or any branch/commit) to specify the reference point for change detection.

```bash
# In a PR branch, plan only what changed compared to main
cultivator plan --changed-only --base=main
```

## Dry-Run Mode

Preview exactly what Cultivator would do without actually running any Terragrunt commands.

- **Safety First**: Use `--dry-run` to see the list of discovered modules, the execution order determined by the DAG, and the exact `terragrunt` commands that would be invoked.
- **Verification**: Ideal for verifying complex tag or path filters before applying changes to production.

```bash
cultivator apply --tags=critical --dry-run
```

## Parallel Execution

Run stacks concurrently with a configurable worker pool:

```yaml
parallelism: 4
```

## Configurable Log Levels

Control Cultivator's own log verbosity via the `CULTIVATOR_LOG_LEVEL` environment variable:

```bash
CULTIVATOR_LOG_LEVEL=debug cultivator plan --root=live --env=dev
```

Accepted values: `debug`, `info`, `warning`, `error`. Default: `info`.
Terragrunt output is always printed in full, regardless of this setting.

## Stateless Operation

Cultivator does not manage state or backends. Terragrunt and Terraform/OpenTofu handle state as usual.

## Cross-Platform Support

- GitHub Actions
- GitLab CI
- Local execution

---

See [User Guide](index.md) for how to use these features.
