# Cultivator

![Cultivator](assets/logo.svg)

**Cultivator** is a lightweight CLI that orchestrates **Terragrunt** stack discovery, filtering, and execution across CI/CD systems and local environments.

## Choose Your Path

To get started quickly, select the path that best matches your goals.

### Path A: Infrastructure Automation (End-Users)

If you are looking for a solution to automate your Terragrunt stacks in CI/CD pipelines or local development:

- **[Quick Start](getting-started/quickstart.md)** - Get running in 5 minutes with minimal configuration.
- **[Installation](getting-started/installation.md)** - Learn how to install and setup Cultivator for your environment.
- **[CLI Reference](user-guide/cli-reference.md)** - Full reference for commands, flags, and common workflows.
- **[Configuration](getting-started/configuration.md)** - Customize Cultivator using YAML, environment variables, or flags.
- **[CI/CD Guides](user-guide/index.md)** - Detailed integration guides for **GitHub Actions** and **GitLab CI/CD**.
- **[FAQ](faq.md)** - Frequently asked questions about usage, security, and troubleshooting.

---

### Path B: Project Contribution (Technical Users)

If you are interested in understanding the internal mechanics of Cultivator or want to contribute to its evolution:

- **[Architecture Design](architecture/design.md)** - Detailed technical overview of components, data flow, and design principles.
- **[Development Guide](development/development.md)** - Setup your local environment, project structure, and coding standards.
- **[Testing Strategy](development/TESTING.md)** - Comprehensive guide on unit tests, integration tests, fuzzing, and benchmarks.
- **[External Modules](architecture/external-modules.md)** - Understand how Cultivator and Terragrunt handle remote modules and authentication.
- **[Contributing](development/contributing.md)** - Our workflow for pull requests, issues, and code reviews.

---

## Key Advantages

Cultivator simplifies Terragrunt orchestration by providing:

- **Automatic Stack Discovery** - Recursive search for `terragrunt.hcl` files.
- **Smart Filtering** - Scope execution by environment, path patterns, and custom tags.
- **Dependency Awareness** - Automatically parses and respects HCL dependency blocks.
- **Magic Mode** - Automatically filter modules by Git changes (`--changed-only`).
- **Parallel Execution** - Configurable worker pool for safe, concurrent runs.
- **Stateless Operation** - No server required; leverages your existing Terraform/OpenTofu backends.

## Getting Help

- **[GitHub Issues](https://github.com/Ops-Talks/cultivator/issues)** - Report bugs and request features.
- **[GitHub Discussions](https://github.com/Ops-Talks/cultivator/discussions)** - Ask questions and share ideas with the community.

## Requirements

- **Terragrunt** v0.50.0+ (recommended: v1.0+)
- **OpenTofu** v1.6+ or **Terraform** v1.5+
- **Go** v1.25+ (for building from source)

## License

Cultivator is licensed under the GNU GPL v3 License. See [LICENSE](https://github.com/Ops-Talks/cultivator/blob/main/LICENSE) for details.
