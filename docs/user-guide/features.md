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

Use tags to logically group stacks and selectively execute them. Tags are defined as comments in `terragrunt.hcl` files.

**Syntax in HCL:**

```hcl
# cultivator:tags=critical,prod
terraform {
  source = "./modules/database"
}
```

**Comma-separated multiple tags:**

```hcl
# cultivator:tags=app,prod,critical
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
