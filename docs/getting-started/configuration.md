# Configuration

Complete reference for Cultivator configuration.

## Configuration File

Cultivator looks for a config file in this order:

- `.cultivator.yaml`
- `.cultivator.yml`
- `cultivator.yaml`
- `cultivator.yml`

Place the config file in the root of your repository.

## Simple Configuration

For basic local CLI usage:

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

## Complete Schema (GitHub/GitLab)

For CI/CD pipelines, use the full schema:

```yaml
version: 1

projects:
  - name: production
    dir: envs/prod
    terragrunt_version: 0.55.0
    terraform_version: 1.7.0
    auto_plan: true
    apply_requirements:
      - approved
      - mergeable

  - name: staging
    dir: envs/staging
    terragrunt_version: 0.55.0
    terraform_version: 1.7.0
    auto_plan: true

settings:
  auto_plan: true
  lock_timeout: 10m
  parallel_plan: true
  max_parallel: 5
  require_approval: true
  delete_plans: false

hooks:
  pre_plan:
    - echo "Starting plan..."
    - terragrunt validate-all
    
  post_plan:
    - echo "Plan completed"
    - ./scripts/notify-slack.sh
    
  pre_apply:
    - echo "Starting apply..."
    - ./scripts/backup-state.sh
    
  post_apply:
    - echo "Apply completed"
    - ./scripts/update-docs.sh
    
  on_error:
    - echo "Error occurred"
    - ./scripts/notify-error.sh
```

## Configuration Reference

### Root Level (Simple Mode)

#### `root`
- **Type**: String
- **Default**: `.`
- **Description**: Root directory used to discover `terragrunt.hcl` stacks

#### `env`
- **Type**: String
- **Default**: empty
- **Description**: Environment filter derived from the first directory under `root`

#### `include` / `exclude`
- **Type**: String list
- **Default**: empty
- **Description**: Relative paths under `root` to include or exclude

#### `tags`
- **Type**: String list
- **Default**: empty
- **Description**: Optional tags filter parsed from `terragrunt.hcl` comments (example: `# cultivator:tags=app,db`)

#### `parallelism`
- **Type**: Integer
- **Default**: number of CPUs
- **Description**: Maximum number of stacks to execute in parallel

#### `output_format`
- **Type**: String
- **Default**: `text`
- **Valid values**: `text`, `json`
- **Description**: Log output format

#### `non_interactive`
- **Type**: Boolean
- **Default**: `false`
- **Description**: Adds `-input=false` to Terragrunt commands

#### `plan.destroy`
- **Type**: Boolean
- **Default**: `false`
- **Description**: Run `terragrunt plan -destroy`

#### `apply.auto_approve`
- **Type**: Boolean
- **Default**: `false`
- **Description**: Add `-auto-approve` to `terragrunt apply`

#### `destroy.auto_approve`
- **Type**: Boolean
- **Default**: `false`
- **Description**: Add `-auto-approve` to `terragrunt destroy`

### Root Level (CI/CD Mode)

#### `version` (required)
Version of the config schema. Currently `1`.

#### `projects` (required)
List of Terragrunt project definitions.

##### Project Fields

- **`name`** (required): Unique identifier for the project
- **`dir`** (required): Path to directory containing Terragrunt files
- **`terragrunt_version`** (optional): Specific Terragrunt version
- **`terraform_version`** (optional): Specific Terraform version
- **`workflow`** (optional): Workflow type (`default` or `custom`)
- **`auto_plan`** (optional): Auto-run plan on PR
- **`apply_requirements`** (optional): Requirements for apply
  - `approved`: PR must have at least one approval
  - `mergeable`: PR must be in mergeable state
  - `status_checks`: All status checks must pass

#### `settings` (optional)
Global settings for all projects.

- **`auto_plan`** (default: `true`): Auto-run plan on PR open/update
- **`lock_timeout`** (default: `10m`): Lock timeout duration
- **`parallel_plan`** (default: `true`): Run plans in parallel
- **`max_parallel`** (default: `5`): Maximum parallel operations
- **`require_approval`** (default: `true`): Require PR approval before apply
- **`delete_plans`** (default: `false`): Delete plan files after apply

#### `hooks` (optional)
Custom commands at different stages.

Hook types:
- **`pre_plan`**: Before running plan
- **`post_plan`**: After successful plan
- **`pre_apply`**: Before running apply
- **`post_apply`**: After successful apply
- **`on_error`**: When any operation fails

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

CI/CD-specific:
- `CULTIVATOR_CONFIG`: Path to config file
- `GITHUB_TOKEN`: GitHub API token
- `CULTIVATOR_LOG_LEVEL`: Log level (`debug`, `info`, `warn`, `error`)

Example:

```bash
export CULTIVATOR_ROOT=live
export CULTIVATOR_ENV=prod
export CULTIVATOR_NON_INTERACTIVE=true
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
