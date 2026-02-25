# Features

Cultivator provides powerful features for automating Terragrunt workflows in GitHub.

## Smart Change Detection

Automatically detect which modules are affected by changes in your PR:

```
Your PR changes:
  - infrastructure/vpc/main.tf
  - infrastructure/vpc/variables.tf

Cultivator detects:
  - vpc module is affected
  - Any modules depending on vpc
  
Runs plan for: vpc + dependent modules
```

## Dependency-Aware Execution

Respects Terragrunt module dependencies and executes in the correct order:

```hcl
# infrastructure/app/terragrunt.hcl
dependency "vpc" {
  config_path = "../vpc"
}

dependency "database" {
  config_path = "../database"
}
```

Cultivator will run:
1. VPC (no dependencies)
2. Database (depends on VPC)
3. App (depends on VPC and Database)

## Locking Mechanism

Prevents concurrent apply operations to the same module:

- File-based locking
- Configurable timeout
- Automatic cleanup on completion
- Prevents mistakes during simultaneous deployments

## Rich PR Comments

Formatted output with:

- Plan summary (add/change/destroy counts)
- Affected modules list
- Dependency graph visualization
- Status indicators
- Error messages with troubleshooting tips

Example PR comment:

```
Plan Results for VPC Module

Status: Changes Detected

Resources:
  Add:       3
  Change:    1
  Destroy:   0

Modules: vpc, vpc-security-groups

Applied: ✓ Run with: cultivator apply
```

## Security

### Redaction of Sensitive Data

Automatically redacts:
- AWS access keys
- Password fields
- API tokens
- Database credentials
- Custom patterns

### GitHub Webhook Validation

- Validates webhook signatures
- Prevents unauthorized requests
- Respects GitHub permissions

### IAM Integration

- Requires proper cloud credentials in GitHub Secrets
- Respects IAM policies
- Logs all operations for audit

## Parallel Execution (Experimental)

Run independent modules in parallel for faster execution:

```yaml
settings:
  parallel_plan: true
  max_parallel: 4
```

Warning: Only use for modules without dependencies.

## Cross-Platform Support

Works with:
- GitHub Actions (recommended)
- GitLab CI
- Jenkins
- Other CI/CD platforms

Docker container for flexibility.

## Git Integration

- Detects merge-base automatically
- Works with any branch strategy
- Respects `.gitignore` patterns
- Handles large repositories efficiently

## Module Support

- **Terragrunt**: Full support for all features
- **Terraform**: Works with Terragrunt-wrapped Terraform
- **LocalStack**: Local testing support
- **Both**: Works in hybrid environments

## Approval Workflows

- Require approval for apply operations
- Comment-based approval
- Integration with GitHub branch protection
- Customizable approval policies

## Monitoring & Logging

- GitHub Actions logs
- PR comment history
- Detailed error messages
- Audit trail of all operations

---

See [User Guide](index.md) for how to use these features.
