# Configuration

Configure Cultivator to match your Terragrunt layout and CI needs.

## Configuration File

Cultivator looks for a config file in this order:

- `.cultivator.yaml`
- `.cultivator.yml`
- `cultivator.yaml`
- `cultivator.yml`

Example configuration:

```yaml
root: live
parallelism: 4
output_format: text
non_interactive: true

include:
  - envs/prod
exclude:
  - envs/prod/experimental

tags:
  - app

plan:
  destroy: false
apply:
  auto_approve: false
destroy:
  auto_approve: true
```

## Settings Reference

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

### `output_format`
- **Type**: String
- **Default**: `text`
- **Valid values**: `text`, `json`
- **Description**: Log output format

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
- `CULTIVATOR_OUTPUT_FORMAT`
- `CULTIVATOR_NON_INTERACTIVE`
- `CULTIVATOR_PLAN_DESTROY`
- `CULTIVATOR_APPLY_AUTO_APPROVE`
- `CULTIVATOR_DESTROY_AUTO_APPROVE`

Example:

```yaml
env:
  CULTIVATOR_ROOT: live
  CULTIVATOR_ENV: prod
  CULTIVATOR_NON_INTERACTIVE: "true"
```

## CLI Flags

Flags override environment variables:

```bash
cultivator plan \
  --root=live \
  --env=prod \
  --include=envs/prod/app1 \
  --exclude=envs/prod/experimental \
  --tags=app,db \
  --parallelism=4 \
  --output-format=json \
  --non-interactive
```

## OpenTofu or Terraform

Cultivator runs Terragrunt, which in turn calls either OpenTofu or Terraform. Choose the binary in your `terragrunt.hcl`:

```hcl
terraform_binary = "tofu" # Use OpenTofu
# terraform_binary = "terraform" # Use HashiCorp Terraform instead
```

Make sure the selected binary is installed in your CI environment.
