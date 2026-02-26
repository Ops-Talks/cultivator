# Cultivator

**Cultivator** is a Go-based CLI that orchestrates **Terragrunt** pipelines in CI (GitHub Actions and GitLab CI). It standardizes `plan`, `apply`, and `destroy` runs without requiring a separate backend.

## Why Cultivator?

Terragrunt repositories often need a simple, repeatable way to execute `plan` and `apply` across multiple modules. Cultivator focuses on:

- Module discovery from a root layout
- Consistent CLI flags and CI-friendly defaults
- Optional filtering by environment, path, and tags
- Stateless execution that relies on Terragrunt/Terraform backends

## Features

- **Terragrunt-first CLI** - `plan`, `apply`, and `destroy` from one entrypoint
- **Module discovery** - finds `terragrunt.hcl` modules under a root directory
- **Scope filters** - filter by environment, include/exclude paths, or tags
- **Parallel execution** - configurable worker pool for faster runs
- **CI-ready output** - text or JSON logs with consistent exit codes
- **No backend** - uses existing Terragrunt/Terraform state backends

## Quick Start

### Build locally

```bash
go build -o cultivator ./cmd/cultivator
```

### Run a plan

```bash
./cultivator plan --root=live --env=dev --non-interactive
```

### Apply changes

```bash
./cultivator apply --root=live --env=dev --non-interactive --auto-approve
```

### Destroy

```bash
./cultivator destroy --root=live --env=dev --non-interactive --auto-approve
```

## Configuration

Create a `.cultivator.yaml`, `.cultivator.yml`, `cultivator.yaml`, or `cultivator.yml` file in your repository root:

```yaml
root: live
parallelism: 4
output_format: text
non_interactive: true

plan:
  destroy: false
apply:
  auto_approve: true
destroy:
  auto_approve: true
```

Environment variables and CLI flags override the config file. Flags take precedence over environment variables.

## Project Structure

```
cultivator/
├── cmd/
│   └── cultivator/          # CLI entry point
├── internal/
│   ├── cli/                 # CLI parsing and orchestration
│   ├── config/              # Config load/merge
│   ├── discovery/           # Module discovery
│   ├── runner/              # Terragrunt execution
│   └── logging/             # Output formatting
├── docs/                    # Documentation
└── mkdocs.yml               # MkDocs configuration
```

## Requirements

- Terragrunt >= 0.50.0
- OpenTofu >= 1.6.0 or Terraform >= 1.5.0
- Go >= 1.25 (for building from source)

## Contributing

Contributions are welcome. Please read the [Contributing Guide](CONTRIBUTING.md) for details.

## License

MIT License - see [LICENSE](LICENSE) for details.
