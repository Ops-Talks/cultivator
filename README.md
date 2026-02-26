# Cultivator

**Cultivator** is a CI/CD automation tool for Terragrunt that runs plan and apply operations directly from Pull Requests - similar to what Digger/Atlantis do for Terraform, but built specifically for Terragrunt workflows.

## Why Cultivator?

While tools like Atlantis and Digger work great for Terraform, they don't fully support Terragrunt's unique features:
- **Dependencies between modules** (`dependency` blocks)
- **Run-all operations** across multiple modules
- **Hierarchical configuration** with `terragrunt.hcl` inheritance
- **Impact detection** when parent configs change

Cultivator is built from the ground up to handle these Terragrunt-specific scenarios.

## Features

- **PR-based workflows** - Run terragrunt commands via PR comments  
- **Smart change detection** - Detects which modules are affected by changes  
- **Dependency-aware execution** - Respects Terragrunt dependencies and runs in correct order  
- **Run-all support** - Execute plans/applies across multiple modules  
- **No separate server** - Runs in your existing CI/CD (GitHub Actions, GitLab CI, etc)  
- **Locking mechanism** - Prevents concurrent applies to the same module  
- **Rich PR comments** - Beautiful formatted outputs with plan summaries  

## Quick Start

### 1. Add GitHub Action to your repository

Create `.github/workflows/cultivator.yml`:

```yaml
name: Cultivator

on:
  pull_request:
    types: [opened, synchronize, reopened]
  issue_comment:
    types: [created]

jobs:
  cultivator:
    runs-on: ubuntu-latest
    permissions:
      contents: read
      pull-requests: write
      
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0
          
      - name: Setup Terragrunt
        uses: autero1/action-terragrunt@v1
        with:
          terragrunt-version: 0.55.0
      
      - name: Setup Terraform
        uses: hashicorp/setup-terraform@v2
        with:
          terraform_version: 1.7.0
          
      - name: Run Cultivator
        uses: cultivator-dev/cultivator-action@v1
        with:
          github-token: ${{ secrets.GITHUB_TOKEN }}
```

### 2. Use commands in PR comments

- `/cultivator plan` - Run plan on affected modules
- `/cultivator apply` - Apply changes (requires approval)
- `/cultivator plan-all` - Run plan on all modules in directory
- `/cultivator apply-all` - Apply all modules in correct order
- `/cultivator unlock` - Remove locks if needed

## How it Works

1. **Detects Changes**: When a PR is opened/updated, Cultivator analyzes which files changed
2. **Finds Affected Modules**: Determines which Terragrunt modules are impacted (including dependencies)
3. **Runs Operations**: Executes `terragrunt plan` or `apply` in the correct order
4. **Posts Results**: Comments on the PR with formatted output and summaries

## Configuration

Create a `cultivator.yml` in your repository root:

```yaml
version: 1

# Projects to manage
projects:
  - name: infrastructure
    dir: infrastructure/
    terragrunt_version: 0.55.0
    terraform_version: 1.7.0
    
  - name: environments
    dir: environments/
    workflow: custom
    apply_requirements:
      - approved
      - mergeable

# Global settings
settings:
  auto_plan: true
  lock_timeout: 10m
  parallel_plan: true
  max_parallel: 5
  
# Execute custom commands before/after
hooks:
  pre_plan:
    - terragrunt validate-all
  post_apply:
    - echo "Applied successfully"
```

## Architecture

```
┌─────────────────┐
│   Pull Request  │
│    (GitHub)     │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  GitHub Actions │
│   (CI Runner)   │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│   Cultivator    │
│      Core       │
├─────────────────┤
│ • Change Detect │
│ • Dependency    │
│ • Executor      │
│ • Lock Manager  │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│   Terragrunt    │
│   Commands      │
└─────────────────┘
```

## Project Structure

```
cultivator/
├── cmd/
│   └── cultivator/          # CLI entry point
├── pkg/
│   ├── detector/            # Change detection
│   ├── parser/              # Terragrunt HCL parser
│   ├── executor/            # Command execution
│   ├── lock/                # Locking mechanism
│   ├── github/              # GitHub API integration
│   └── formatter/           # Output formatting
├── action/                  # GitHub Action wrapper
└── docs/                    # Documentation
```

## Requirements

- Terragrunt >= 0.50.0
- OpenTofu >= 1.6.0 (open-source Terraform alternative) or Terraform >= 1.5.0
- Go >= 1.25 (for building from source)

## Contributing

Contributions are welcome! Please read our [Contributing Guide](CONTRIBUTING.md) for details.

## License

MIT License - see [LICENSE](LICENSE) for details.

## Roadmap

- [ ] GitHub Actions support
- [ ] GitLab CI support
- [ ] Azure DevOps support
- [ ] Cost estimation integration
- [ ] Drift detection
- [ ] Policy as code (OPA support)
- [ ] Slack/Discord notifications
- [ ] RBAC for different environments

## Credits

Inspired by:
- [Digger](https://github.com/diggerhq/digger)
- [Atlantis](https://www.runatlantis.io/)
- [Terragrunt](https://terragrunt.gruntwork.io/)
