# Getting Started

Welcome to Cultivator! This section will guide you through installation, configuration, and your first Terragrunt plan.

You can start using Cultivator without any configuration file. A `cultivator.yml` is optional.

## Prerequisites

- **Go 1.25+** (for building from source)
- **Git** (2.0+)
- **Docker** (optional, for containerized deployment)
- **Git hosting** (GitHub, GitLab, etc.) if using CI integration
- **Terragrunt** (for your infrastructure)

## Choose Your Path

- **Quickest Start**: [Run locally in minutes](quickstart.md)
- **Full Installation**: [Installation Guide](installation.md)
- **Configuration Details**: [Configuration Reference](configuration.md)

## Minimal Path (No config file)

```bash
go build -o cultivator ./cmd/cultivator
./cultivator plan --root=live --env=dev --non-interactive
```

If needed, add a config file later and pass it with `--config`.

## Next Steps

1. **[Quick Start](quickstart.md)** - Get running in 5 minutes
2. **[Installation](installation.md)** - Install Cultivator
3. **[Configuration](configuration.md)** - Configure your repository

---

Need help? Check the [FAQ](../faq.md) or open an [issue on GitHub](https://github.com/Ops-Talks/cultivator/issues).
