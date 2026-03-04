# Configuration

Complete reference for Cultivator runtime configuration.

## Is `cultivator.yml` required?

No. Configuration file is optional.

Cultivator runs with defaults plus CLI flags and environment variables.

## Using a config file

When using a file, pass it explicitly via `--config`:

```bash
cultivator plan --config=cultivator.yml
```

Supported filename is up to you (`cultivator.yml`, `.cultivator.yaml`, etc.).

## Minimal config example

```yaml
root: live
parallelism: 4
non_interactive: true

include:
  - envs/prod
exclude:
  - envs/prod/experimental

tags:
  - critical

plan:
  destroy: false
apply:
  auto_approve: false
destroy:
  auto_approve: false
```

## Precedence order

Highest to lowest:

1. CLI flags
2. Environment variables
3. Config file (when passed with `--config`)
4. Built-in defaults

## Supported keys

### `root`

- **Type**: String
- **Default**: `.`
- **Description**: Root directory used to discover `terragrunt.hcl` stacks

### `env`

- **Type**: String
- **Default**: empty
- **Description**: Environment filter derived from the first directory under `root`

### `include` / `exclude`

- **Type**: String list
- **Default**: empty
- **Description**: Relative paths under `root` to include or exclude

### `tags`

- **Type**: String list
- **Default**: empty
- **Description**: Optional tags filter parsed from `terragrunt.hcl` comments (example: `# cultivator:tags=app,db`)

### `parallelism`

- **Type**: Integer
- **Default**: number of CPUs
- **Description**: Maximum number of stacks to execute in parallel

### `non_interactive`

- **Type**: Boolean
- **Default**: `false`
- **Description**: Adds `-input=false` to Terragrunt commands

### `plan.destroy`

- **Type**: Boolean
- **Default**: `false`
- **Description**: Run `terragrunt plan -destroy`

### `apply.auto_approve`

- **Type**: Boolean
- **Default**: `false`
- **Description**: Add `-auto-approve` to `terragrunt apply`

### `destroy.auto_approve`

- **Type**: Boolean
- **Default**: `false`
- **Description**: Add `-auto-approve` to `terragrunt destroy`

## Environment Variables

Environment variables override the config file:

- `CULTIVATOR_ROOT`
- `CULTIVATOR_ENV`
- `CULTIVATOR_INCLUDE`
- `CULTIVATOR_EXCLUDE`
- `CULTIVATOR_TAGS`
- `CULTIVATOR_PARALLELISM`
- `CULTIVATOR_NON_INTERACTIVE`
- `CULTIVATOR_PLAN_DESTROY`
- `CULTIVATOR_APPLY_AUTO_APPROVE`
- `CULTIVATOR_DESTROY_AUTO_APPROVE`

### Logging

- `CULTIVATOR_LOG_LEVEL` — minimum log severity emitted by Cultivator.
  Accepted values: `debug`, `info`, `warning`, `error`. Default: `info`.
  Terragrunt output is always printed regardless of this setting.

Example:

```bash
export CULTIVATOR_ROOT=live
export CULTIVATOR_ENV=prod
export CULTIVATOR_NON_INTERACTIVE=true
export CULTIVATOR_LOG_LEVEL=debug   # enable verbose Cultivator logs
```

## CLI Flags

Flags override environment variables and config file

```bash
cultivator plan \
  --root=live \
  --env=prod \
  --include=envs/prod/app1 \
  --exclude=envs/prod/experimental \
  --tags=app,db \
  --parallelism=4 \
  --non-interactive
```

## OpenTofu or Terraform

Cultivator runs Terragrunt, which in turn calls either OpenTofu or Terraform. Choose the binary in your `terragrunt.hcl`:

```hcl
terraform_binary = "tofu" # Use OpenTofu
# terraform_binary = "terraform" # Use HashiCorp Terraform instead
```

Make sure the selected binary is installed in your CI environment.
