# Cultivator

![Cultivator](assets/logo.svg)

**Cultivator** is a Go-based CLI that orchestrates **Terragrunt** pipelines in CI (GitHub Actions and GitLab CI). It standardizes `plan`, `apply`, and `destroy` runs without requiring a separate backend.

## Overview

Cultivator focuses on predictable CI execution for Terragrunt repositories:

- Module discovery from a root layout
- Filters by environment, include/exclude paths, and tags
- Parallel execution with configurable limits
- Text or JSON output for CI logs
- Stateless operation using existing Terragrunt/Terraform backends

## Quick Links

<div class="grid cards" markdown>

- **[Getting Started](getting-started/index.md)** - Installation and first steps
- **[User Guide](user-guide/index.md)** - How to use Cultivator
- **[Architecture](architecture/design.md)** - How it works under the hood
- **[Development](development/contributing.md)** - Contribute to the project

</div>

## How It Works

```
1. CI triggers a pipeline job
2. Cultivator discovers Terragrunt modules
3. Terragrunt runs per module (plan/apply/destroy)
4. Results are logged with a consistent exit code
```

## Example Usage

```bash
# Run plan for dev modules
cultivator plan --root=live --env=dev --non-interactive

# Apply changes with auto-approve
cultivator apply --root=live --env=dev --non-interactive --auto-approve
```

## Getting Help

- **Documentation**: Check the [User Guide](user-guide/index.md)
- **Issues**: [GitHub Issues](https://github.com/Ops-Talks/cultivator/issues)
- **Discussions**: [GitHub Discussions](https://github.com/Ops-Talks/cultivator/discussions)

## License

Cultivator is licensed under the MIT License. See [LICENSE](https://github.com/Ops-Talks/cultivator/blob/main/LICENSE) for details.
