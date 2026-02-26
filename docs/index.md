# Cultivator

![Cultivator](assets/logo.svg)

**Cultivator** is a CI/CD automation tool for Terragrunt that runs plan and apply operations directly from Pull Requests - similar to what Digger/Atlantis do for Terraform, but built specifically for Terragrunt workflows.

## Overview

While tools like Atlantis and Digger work great for Terraform, they don't fully support Terragrunt's unique features:

- **Dependencies between modules** (`dependency` blocks)
- **Run-all operations** across multiple modules
- **Hierarchical configuration** with `terragrunt.hcl` inheritance
- **Impact detection** when parent configs change

Cultivator is built from the ground up to handle these Terragrunt-specific scenarios.

## Key Features

- **PR-based workflows** - Run terragrunt commands via PR comments
- **Smart change detection** - Detects which modules are affected by changes
- **Dependency-aware execution** - Respects Terragrunt dependencies and runs in correct order
- **Run-all support** - Execute plans/applies across multiple modules
- **No separate server** - Runs in your existing CI/CD (GitHub Actions, GitLab CI, etc)
- **Locking mechanism** - Prevents concurrent applies to the same module
- **Rich PR comments** - Beautiful formatted outputs with plan summaries

## Quick Links

<div class="grid cards" markdown>

- **[Getting Started](getting-started/index.md)** - Installation and first steps
- **[User Guide](user-guide/index.md)** - How to use Cultivator
- **[Architecture](architecture/design.md)** - How it works under the hood
- **[Development](development/contributing.md)** - Contribute to the project

</div>

## How It Works

```
1. User comments on PR:
   "cultivator plan" or "cultivator apply"
   
2. Cultivator detects changed modules
   
3. Builds dependency graph
   
4. Executes operations respecting dependencies
   
5. Posts formatted results as PR comment
```

## Example Usage

```bash
# Comment on PR to run plan for affected modules
cultivator plan

# Run plan for all modules
cultivator plan --all

# Apply changes (requires approval in some configurations)
cultivator apply
```

## Getting Help

- **Documentation**: Check the [User Guide](user-guide/index.md)
- **Issues**: [GitHub Issues](https://github.com/Ops-Talks/cultivator/issues)
- **Discussions**: [GitHub Discussions](https://github.com/Ops-Talks/cultivator/discussions)

## License

Cultivator is licensed under the MIT License. See [LICENSE](https://github.com/Ops-Talks/cultivator/blob/main/LICENSE) for details.

---

**Latest Version**: 1.0.0 | **Last Updated**: February 2026
