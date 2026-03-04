# Features

Cultivator provides practical features for automating Terragrunt workflows in CI.

## Stack Discovery

Cultivator discovers stacks by walking the root directory and locating `terragrunt.hcl` files.

## Scope Filters

Limit execution to specific areas:

- `--env` for environment filtering
- `--include` / `--exclude` for path filtering
- `--tags` for tag filtering (via `cultivator:tags` comments)

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
