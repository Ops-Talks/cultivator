# Configuration Guide

Complete reference for `cultivator.yml` configuration file.

## File Location

Place `cultivator.yml` in the root of your repository.

## Schema

### Top-level Structure

```yaml
version: 1                # Required: Config version
projects: []              # Required: List of projects
settings: {}              # Optional: Global settings
hooks: {}                 # Optional: Custom hooks
```

## Projects

Define Terragrunt projects/directories to manage.

```yaml
projects:
  - name: production              # Required: Unique project name
    dir: envs/prod                # Required: Directory path
    terragrunt_version: 0.55.0    # Optional: Terragrunt version
    terraform_version: 1.7.0      # Optional: Terraform version
    workflow: custom               # Optional: Workflow type (default, custom)
    auto_plan: true                # Optional: Auto-run plan on PR
    apply_requirements:            # Optional: Requirements for apply
      - approved                   # PR must be approved
      - mergeable                  # PR must be mergeable
```

### Project Fields

#### `name` (required)
Unique identifier for the project.

#### `dir` (required)
Path to the directory containing Terragrunt files. Relative to repository root.

#### `terragrunt_version` (optional)
Specific Terragrunt version to use for this project.

#### `terraform_version` (optional)
Specific Terraform version to use for this project.

#### `workflow` (optional)
- `default`: Standard plan/apply workflow
- `custom`: Custom workflow defined in hooks

#### `auto_plan` (optional)
Override global `auto_plan` setting for this project.

#### `apply_requirements` (optional)
List of requirements that must be met before apply:
- `approved`: PR must have at least one approval
- `mergeable`: PR must be in mergeable state
- `status_checks`: All status checks must pass

## Settings

Global settings that apply to all projects (unless overridden).

```yaml
settings:
  auto_plan: true           # Automatically run plan on PR open/update
  lock_timeout: 10m         # Lock timeout duration
  parallel_plan: true       # Run plans in parallel when possible
  max_parallel: 5           # Maximum parallel operations
  require_approval: true    # Require approval before apply
  delete_plans: false       # Delete plan files after apply
```

### Settings Fields

#### `auto_plan` (default: `true`)
Automatically run `terragrunt plan` when PR is opened or updated.

#### `lock_timeout` (default: `10m`)
How long to wait for a lock before timing out. Format: `5m`, `1h`, etc.

#### `parallel_plan` (default: `true`)
Run plan operations in parallel for independent modules.

#### `max_parallel` (default: `5`)
Maximum number of parallel operations.

#### `require_approval` (default: `true`)
Require PR approval before allowing apply.

#### `delete_plans` (default: `false`)
Automatically delete plan files after successful apply.

## Hooks

Execute custom commands at different stages.

```yaml
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

### Hook Types

- `pre_plan`: Before running plan
- `post_plan`: After successful plan
- `pre_apply`: Before running apply
- `post_apply`: After successful apply
- `on_error`: When any operation fails

## Complete Example

```yaml
version: 1

projects:
  # Production environment
  - name: production
    dir: environments/prod
    terragrunt_version: 0.55.0
    terraform_version: 1.7.0
    apply_requirements:
      - approved
      - mergeable
      - status_checks
    auto_plan: true

  # Staging environment
  - name: staging
    dir: environments/staging
    terragrunt_version: 0.55.0
    terraform_version: 1.7.0
    apply_requirements:
      - approved
    auto_plan: true

  # Development environment
  - name: development
    dir: environments/dev
    terragrunt_version: 0.55.0
    terraform_version: 1.7.0
    auto_plan: true
    # No apply requirements for dev

settings:
  auto_plan: true
  lock_timeout: 15m
  parallel_plan: true
  max_parallel: 10
  require_approval: true
  delete_plans: false

hooks:
  pre_plan:
    - echo "=== Starting Terragrunt Plan ==="
    - terragrunt validate-all
    
  post_plan:
    - echo "=== Plan Completed ==="
    - ./scripts/analyze-costs.sh
    
  pre_apply:
    - echo "=== Starting Terragrunt Apply ==="
    - ./scripts/create-backup.sh
    - ./scripts/notify-team.sh "Apply starting"
    
  post_apply:
    - echo "=== Apply Completed ==="
    - ./scripts/update-documentation.sh
    - ./scripts/notify-team.sh "Apply completed successfully"
    
  on_error:
    - echo "=== Error Occurred ==="
    - ./scripts/notify-team.sh "Error in Cultivator"
    - ./scripts/create-incident.sh
```

## Environment Variables

Cultivator supports these environment variables:

- `CULTIVATOR_CONFIG`: Path to config file (default: `cultivator.yml`)
- `GITHUB_TOKEN`: GitHub API token
- `CULTIVATOR_LOG_LEVEL`: Log level (`debug`, `info`, `warn`, `error`)
- `CULTIVATOR_AUTO_PLAN`: Override auto_plan setting (`true`/`false`)

## Validation

Validate your configuration:

```bash
cultivator validate --config cultivator.yml
```

## Migration from Other Tools

### From Atlantis

```yaml
# Atlantis atlantis.yaml
version: 3
projects:
  - dir: environments/prod
    terraform_version: v1.7.0
```

Converts to:

```yaml
# Cultivator cultivator.yml
version: 1
projects:
  - name: prod
    dir: environments/prod
    terraform_version: 1.7.0
```

### From Digger

Cultivator config is similar to Digger but adds Terragrunt-specific features.

## Best Practices

1. **Start Simple**: Begin with minimal config and add features as needed
2. **Use Projects**: Separate different environments into projects
3. **Set Requirements**: Use `apply_requirements` to enforce safety
4. **Add Hooks**: Automate common tasks with hooks
5. **Version Control**: Keep `cultivator.yml` in version control
6. **Test Changes**: Test config changes in a dev environment first

## Next Steps

- Review advanced features in the [User Guide](user-guide/features.md)
- Check [Examples](../examples/)
- Read [Troubleshooting](faq.md)
