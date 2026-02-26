# Example Terragrunt Project Structure

This directory contains an example Terragrunt project structure to demonstrate how Cultivator works.

## Structure

```
examples/
├── terragrunt/
│   ├── terragrunt.hcl              # Root Terragrunt config
│   ├── environments/
│   │   ├── dev/
│   │   │   ├── vpc/
│   │   │   │   └── terragrunt.hcl
│   │   │   ├── database/
│   │   │   │   └── terragrunt.hcl
│   │   │   └── app/
│   │   │       └── terragrunt.hcl
│   │   ├── staging/
│   │   │   ├── vpc/
│   │   │   │   └── terragrunt.hcl
│   │   │   ├── database/
│   │   │   │   └── terragrunt.hcl
│   │   │   └── app/
│   │   │       └── terragrunt.hcl
│   │   └── prod/
│   │       ├── vpc/
│   │       │   └── terragrunt.hcl
│   │       ├── database/
│   │       │   └── terragrunt.hcl
│   │       └── app/
│   │           └── terragrunt.hcl
│   └── modules/
│       ├── vpc/
│       ├── database/
│       └── app/
```

## How Cultivator Detects Changes

1. **File Changes**: When you modify any `.hcl` or `.tf` file, Cultivator detects it
2. **Module Detection**: It finds which Terragrunt module contains the changed file
3. **Dependency Analysis**: It checks if any other modules depend on the changed module
4. **Execution Order**: It determines the correct order to run operations based on dependencies

## Example Scenarios

### Scenario 1: Change VPC in Dev
- **Changed**: `environments/dev/vpc/terragrunt.hcl`
- **Affected modules**: 
  - `environments/dev/vpc` (direct)
  - `environments/dev/database` (depends on vpc)
  - `environments/dev/app` (depends on vpc and database)
- **Execution order**: vpc → database → app

### Scenario 2: Change App in Staging
- **Changed**: `environments/staging/app/terragrunt.hcl`
- **Affected modules**: 
  - `environments/staging/app` (direct)
- **Execution order**: app

### Scenario 3: Change Root Config
- **Changed**: `terragrunt.hcl` (root)
- **Affected modules**: ALL (because all inherit from root)
- **Execution order**: Topologically sorted based on dependencies

## Pipeline Examples

This directory also contains example pipeline configurations for different CI/CD platforms:

### GitHub Actions
See the main [Quick Start](../docs/getting-started/quickstart.md) for GitHub Actions examples.

### GitLab CI/CD
See [`.gitlab-ci.yml`](.gitlab-ci.yml) for a complete GitLab pipeline example.

To use in your project:
```bash
cp examples/.gitlab-ci.yml .gitlab-ci.yml
```

Then customize the `TERRAFORM_VERSION` and `TERRAGRUNT_VERSION` variables as needed.

For more details, see [GitLab Pipelines Integration](../docs/user-guide/gitlab-pipelines.md).
