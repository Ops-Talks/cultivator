# Cultivator

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.25%2B-blue.svg)](https://golang.org)
[![Release](https://img.shields.io/github/v/release/Ops-Talks/cultivator.svg)](https://github.com/Ops-Talks/cultivator/releases)
[![Build Status](https://img.shields.io/github/actions/workflow/status/Ops-Talks/cultivator/ci.yml?branch=main)](https://github.com/Ops-Talks/cultivator/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/Ops-Talks/cultivator)](https://goreportcard.com/report/github.com/Ops-Talks/cultivator)
[![codecov](https://codecov.io/gh/Ops-Talks/cultivator/branch/main/graph/badge.svg?token=X26B4X86L5)](https://codecov.io/gh/Ops-Talks/cultivator)
[![GitHub Stars](https://img.shields.io/github/stars/Ops-Talks/cultivator)](https://github.com/Ops-Talks/cultivator/stargazers)

**Cultivator** is a lightweight CLI that orchestrates **Terragrunt** stack discovery, filtering, and execution across CI/CD systems and local environments.

## ✨ Why Cultivator?

- 🚀 **CLI-first**: Works in GitHub Actions, GitLab CI, and locally
- 🔍 **Smart Discovery**: Finds all Terragrunt stacks automatically
- 🎯 **Flexible Filtering**: Environment, path patterns, and custom tags
- ⚡ **Parallel Execution**: Configurable worker pool with dependency awareness
- 🔒 **No Server**: Pure CLI; uses existing Terraform/OpenTofu backends
- 📊 **Multiple Formats**: Human-readable or JSON output

## 🚀 Quick Start

### Build
```bash
go build -o cultivator ./cmd/cultivator
```

### Plan
```bash
./cultivator plan --root=live --env=dev --non-interactive
```

### Apply
```bash
./cultivator apply --root=live --env=dev --non-interactive --auto-approve
```

## 📋 Requirements

- **Terragrunt** v0.50.0+ (recommended: v1.0+)
- **OpenTofu** v1.6+ or **Terraform** v1.5+
- **Go** v1.25+ (for building from source)

## 📚 Documentation

Full documentation is available at: **[https://ops-talks.github.io/cultivator](https://ops-talks.github.io/cultivator)**

- [Getting Started](https://ops-talks.github.io/cultivator/getting-started/) - Installation and configuration
- [User Guide](https://ops-talks.github.io/cultivator/user-guide/) - Commands, workflows, CI integration
- [Architecture](https://ops-talks.github.io/cultivator/architecture/design/) - Design and internals
- [FAQ](https://ops-talks.github.io/cultivator/faq/) - Common questions

## 🛠️ Development

```bash
# Clone
git clone https://github.com/Ops-Talks/cultivator.git
cd cultivator

# Build
go build -o cultivator ./cmd/cultivator

# Test
go test ./...

# Lint
golangci-lint run
```

For detailed development guide, see [CONTRIBUTING.md](CONTRIBUTING.md).

## 📄 License

MIT License - see [LICENSE](LICENSE)
