# Features

Cultivator provides practical features for automating Terragrunt workflows in CI.

## Module Discovery

Cultivator discovers modules by walking the root directory and locating `terragrunt.hcl` files.

## Scope Filters

Limit execution to specific areas:

- `--env` for environment filtering
- `--include` / `--exclude` for path filtering
- `--tags` for tag filtering (via `cultivator:tags` comments)

## Parallel Execution

Run modules concurrently with a configurable worker pool:

```yaml
parallelism: 4
```

## Consistent Output

Choose `text` or `json` output to fit CI log requirements:

```yaml
output_format: json
```

## Stateless Operation

Cultivator does not manage state or backends. Terragrunt and Terraform/OpenTofu handle state as usual.

## Cross-Platform Support

- GitHub Actions
- GitLab CI
- Local execution

---

See [User Guide](index.md) for how to use these features.
