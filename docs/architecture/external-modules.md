# External Modules

Learn about Cultivator's support for external Terraform modules.

## Overview

Cultivator can automatically download and manage external Terraform modules, whether they come from:

- **Public Registry**: Terraform Registry, GitHub, HTTP sources
- **Private Registry**: Custom registries, private repositories
- **Local Sources**: Git repositories, S3, etc.

## Configuration

Define external modules in your Terragrunt configuration:

```hcl
terraform {
  source = "github.com/example/module.git//path?ref=v1.0.0"
}
```

Or using the Terraform Registry:

```hcl
terraform {
  source = "terraform-aws-modules/vpc/aws"
  version = "~> 5.0"
}
```

## Module Preparation

Cultivator automatically prepares external modules before execution:

1. **Download**: Fetches module source
2. **Cache**: Stores for reuse
3. **Initialize**: Runs `terraform init`
4. **Validate**: Checks module syntax

## Execution Flow

```
Change Detection
    ↓
Module Parsing (discover external module sources)
    ↓
Dependency Resolution (build execution graph)
    ↓
Module Preparation (download, init, validate)
    ↓
Plan/Apply Execution
```

## Caching

External modules are cached to improve performance:

```
/tmp/cultivator-cache/
├── github.com/
│   └── example/
│       └── module.git/
├── registry.terraform.io/
│   └── terraform-aws-modules/
│       └── vpc/
│           └── aws/
```

Cache location: `${TMPDIR}/cultivator-cache` or `/tmp/cultivator-cache`

## Authentication

For private modules, provide credentials:

### GitHub Modules

```bash
export GITHUB_TOKEN=ghp_xxxxxxx
```

### Private Registry

```bash
export TF_REGISTRY_AUTH_TOKEN_example_com=xxxxxxxxx
```

### AWS (via CodeArtifact)

```bash
# Configure AWS credentials in GitHub Secrets
AWS_ACCESS_KEY_ID
AWS_SECRET_ACCESS_KEY
```

## Troubleshooting

### Module Not Found

```
Error: Invalid source address

The source address is invalid: ...
```

Solutions:
- Check module source URL is correct
- Verify network connectivity
- Check authentication credentials

### Version Constraints

```
Error: Unsupported version constraint
```

Solutions:
- Update Terraform to latest version
- Check version constraint syntax
- Consult [version constraint docs](https://www.terraform.io/language/settings/version-constraints)

### Large Module Downloads

For very large modules, you may need to increase timeouts:

```yaml
settings:
  executor_timeout: 10m
```

---

See [Architecture](design.md) for how modules are integrated.
