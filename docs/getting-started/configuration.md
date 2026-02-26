# Configuration

Configure Cultivator to match your infrastructure and workflow needs.

## Configuration File

Create a `cultivator.yml` file in your repository root:

```yaml
version: 1

settings:
  # Auto plan on PR events
  auto_plan: true
  
  # Require approval for apply operations
  require_approval: true
  
  # Lock timeout to prevent deadlocks
  lock_timeout: 30m
  
  # Run modules in parallel (experimental)
  parallel_plan: false
  max_parallel: 1
```

## Settings Reference

### `auto_plan`
- **Type**: Boolean
- **Default**: `true`
- **Description**: Automatically run plan when PR is opened or updated

### `require_approval`
- **Type**: Boolean
- **Default**: `true`
- **Description**: Require explicit approval comment before applying changes

### `lock_timeout`
- **Type**: Duration
- **Default**: `30m`
- **Valid values**: `1m`, `5m`, `10m`, `30m`, `1h`, etc.
- **Description**: Maximum time to wait for lock before timing out

### `parallel_plan`
- **Type**: Boolean
- **Default**: `false`
- **Description**: Run modules in parallel (experimental, not recommended for interdependent modules)

### `max_parallel`
- **Type**: Integer
- **Default**: `1`
- **Description**: Maximum number of modules to execute in parallel

## GitHub Secrets

Configure these secrets in your GitHub repository settings:

### Required

| Secret | Description |
|--------|-------------|
| `GITHUB_TOKEN` | GitHub token (auto-provided by GitHub Actions) |

### Optional (based on your infrastructure)

| Secret | Description |
|--------|-------------|
| `AWS_ACCESS_KEY_ID` | AWS credentials |
| `AWS_SECRET_ACCESS_KEY` | AWS credentials |
| `AZURE_CLIENT_ID` | Azure credentials |
| `AZURE_CLIENT_SECRET` | Azure credentials |
| `GCP_PROJECT_ID` | GCP project ID |
| `GCP_SA_KEY` | GCP service account key (JSON) |

## Branch Protection Rules

Recommended settings for your main branch:

1. Go to Settings → Branches → Add rule
2. Branch name pattern: `main`
3. Enable:
   - Require pull request reviews before merging
   - Dismiss stale pull request approvals
   - Require status checks to pass (include Cultivator if desired)

## Permissions

Cultivator requires these GitHub permissions:

```yaml
permissions:
  contents: read
  pull-requests: write
  statuses: write
  checks: write
```

## Filtering Runs

Run Cultivator only on specific path changes:

```yaml
on:
  pull_request:
    paths:
      - 'infrastructure/**'
      - 'terragrunt/**'
      - 'cultivator.yml'
      - '.github/workflows/cultivator.yml'
```

## Environment Variables

Set environment variables in your workflow:

```yaml
env:
  TF_VAR_environment: production
  TF_VAR_region: us-east-1
  TERRAGRUNT_DOWNLOAD: /tmp/terragrunt-cache
```

## OpenTofu or Terraform

Cultivator runs Terragrunt, which in turn calls either OpenTofu or Terraform.
Choose the binary in your `terragrunt.hcl`:

```hcl
terraform_binary = "tofu" # Use OpenTofu (default in our examples)
# terraform_binary = "terraform" # Use HashiCorp Terraform instead
```

Make sure the selected binary is installed in your CI environment.

## Advanced Configuration

### Custom modules path

```yaml
settings:
  modules_path: ./infrastructure
```

### Ignore specific patterns

```yaml
settings:
  ignore_patterns:
    - '**/.terraform'
    - '**/cache/**'
```

### Redaction patterns

```yaml
settings:
  redaction:
    enabled: true
    patterns:
      - 'password'
      - 'secret'
      - 'token'
      - 'key'
```

## See Also

- [Quick Start](quickstart.md) - Get started in 5 minutes
- [User Guide](../user-guide/workflows.md) - Available commands
- [GitHub Actions Docs](https://docs.github.com/en/actions)
