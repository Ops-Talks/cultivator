# External Modules

Understand how Cultivator handles external Terraform modules sourced from private and public repositories.

## Overview

Cultivator does **not** download or manage external Terraform modules directly. Instead, **Terragrunt handles all module sourcing**, and Cultivator orchestrates Terragrunt's execution.

This design keeps Cultivator simple and focused on orchestration while leveraging Terragrunt's robust module handling.

## How Module Sourcing Works

1. **Cultivator discovers** `terragrunt.hcl` files under the root directory (Terragrunt stacks)
2. **Terragrunt parses** module sources from `terragrunt.hcl`
3. **Terragrunt downloads** external modules (using `terraform get`)
4. **Terragrunt initializes** the working directory
5. **Cultivator executes** `terragrunt plan/apply/destroy`

## Module Source Types

Terragrunt supports module sources from:

### Public Registry

```hcl
terraform {
  source = "terraform-aws-modules/vpc/aws"
  version = "~> 5.0"
}
```

### GitHub (Public)

```hcl
terraform {
  source = "github.com/example/module.git//path?ref=v1.0.0"
}
```

### GitHub (Private)

```hcl
terraform {
  source = "git::https://github.com/your-org/module.git//path?ref=v1.0.0"
}
```

### Custom Git Repository

```hcl
terraform {
  source = "git::ssh://git@github.example.com/module.git//path?ref=main"
}
```

### HTTP Source

```hcl
terraform {
  source = "https://example.com/modules/vpc.tar.gz"
}
```

### Local Path

```hcl
terraform {
  source = "../shared-modules/vpc"
}
```

## Authentication for Private Modules

Terragrunt uses Git credentials to fetch private modules. Configure in your CI environment:

### GitHub (HTTPS + Personal Access Token)

```bash
# In GitHub Actions secrets
git config --global url."https://ghp_xxxxx:x-oauth-basic@github.com/".insteadOf "https://github.com/"
```

Or use SSH:

```bash
# In GitHub Actions via actions/checkout with SSH key
- uses: actions/checkout@v4
  with:
    ssh-key: ${{ secrets.SSH_KEY }}
```

### GitLab (Terraform Registry)

```bash
# Set CI_JOB_TOKEN in .gitlab-ci.yml
git config --global url."https://gitlab-ci-token:${CI_JOB_TOKEN}@gitlab.com/".insteadOf "https://gitlab.com/"
```

### AWS CodeArtifact

```bash
# Retrieve credentials and configure git
aws codeartifact login --tool git --domain your-domain --repository your-repo --region us-east-1
```

## Cultivator's Role

Cultivator **coordinates execution** but doesn't manage module versioning or caching:

```text
Cultivator (orchestration)
    |
Terragrunt (module sourcing, state management)
    |
Terraform/OpenTofu (module execution, state backend)
    |
External Module (downloaded and executed)
```

When you run:

```bash
cultivator plan --root=live --env=prod
```

Cultivator:

1. Discovers `terragrunt.hcl` files (stacks)
2. Respects dependency graph
3. **For each stack**, calls `terragrunt plan`
4. Terragrunt handles module downloads and initialization

## Caching and Performance

Terragrunt caches downloaded modules in `.terragrunt-cache/` by default.

Terragrunt manages cache invalidation automatically. To clear the cache manually, remove the `.terragrunt-cache/` directories:

```bash
find . -type d -name '.terragrunt-cache' -exec rm -rf {} +
```

## Dependency Resolution

If stack A depends on stack B (via `dependency` blocks), Cultivator ensures B runs and completes before A:

```hcl
# Module A
dependency "base" {
  config_path = "../base"
}

inputs = {
  vpc_id = dependency.base.outputs.vpc_id
}
```

Cultivator parses this graph and executes stacks in the correct order.

## Troubleshooting Module Issues

### Module Not Found

**Solutions:**

- Verify the repository URL is correct
- Check network connectivity from CI runner
- Verify Git credentials are configured
- Ensure the ref/tag exists

### Version Constraint Errors

**Solutions:**

- Update Terraform to latest
- Check version constraint syntax
- Verify the module provides the requested version

### Permission Denied

**Solutions (GitHub Actions):**

```yaml
- uses: actions/checkout@v4
  with:
    ssh-key: ${{ secrets.SSH_PRIVATE_KEY }}

- name: Trust GitHub's SSH key
  run: |
    mkdir -p ~/.ssh
    ssh-keyscan github.com >> ~/.ssh/known_hosts
```

**Solutions (GitLab CI):**

```yaml
before_script:
  - eval $(ssh-agent -s)
  - echo "$SSH_PRIVATE_KEY" | base64 -d | ssh-add -
  - ssh-keyscan gitlab.com >> ~/.ssh/known_hosts
```

## Best Practices

### 1. Version All External Modules

Always use exact versions (e.g., `?ref=v1.2.3`).

### 2. Separate Module Repository from Infrastructure Repo

- Keep modules in a dedicated Git repository
- Tag releases explicitly
- Update module versions deliberately

### 3. Use Dependency Blocks for Cross-Module Data

Use `dependency` blocks instead of hardcoding outputs or using `data` blocks across modules.

### 4. Cache `.terragrunt-cache/` in CI

Use CI caching mechanisms to speed up module downloads.

### 5. Document Module Requirements

List external dependencies, minimum versions, and required credentials in your module README.

---

See [Design](design.md) for Cultivator's overall architecture and [Configuration](../getting-started/configuration.md) for runtime options.
