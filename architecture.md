# Cultivator - System Design

## 1. Overview

**Cultivator** is a Go tool to orchestrate **Terragrunt** pipelines in CI (GitHub Actions and GitLab CI), offering standardized `plan`, `apply`, and `destroy` operations without requiring its own backend (unlike Atlantis / OpenTaco).
It is cloud-provider agnostic, respects the repository Terragrunt layout and `root.hcl`, and delegates authentication to the pipeline environment (IAM roles, service accounts, etc.).

Key features:

- Full focus on **Terragrunt** (does not run raw Terraform).
- Execution via a single CLI (`cultivator`) that can be plugged into any CI executor.
- Support for `plan`, `apply`, and `destroy` with layout conventions and filters (by folder, environment, tag, etc.).
- Automatic triggers on PRs (via CI config) and manual execution options (manual jobs / workflow_dispatch).
- Code in **Go (latest stable)** with linting using **golangci-lint**.
- Markdown documentation published via **MkDocs** on GitHub Pages.

---

## 2. Goals and requirements

### 2.1 Functional requirements

- FR01: Execute `terragrunt plan` for one or more modules, following repository conventions (standard "infrastructure-live" layout).
- FR02: Execute `terragrunt apply` with explicit confirmation (configure "non-interactive" for CI).
- FR03: Execute `terragrunt plan -destroy` and `terragrunt destroy` (or equivalent with flags).
- FR04: Automatically discover Terragrunt modules from the root path (e.g., `live/`, `envs/`, etc.), respecting `root.hcl`.
- FR05: Allow scope filters: by directory, environment, tag, or list of paths.
- FR06: Integrate with CI (GitHub Actions, GitLab CI) through:
  - parameters via environment variables and CLI flags;
  - consistent exit codes for success/failure;
  - structured/readable logs.
- FR07: Support execution modes:
  - "PR Mode": run automatically on PRs (default to `plan` only);
  - "Manual Mode": run `plan`, `apply`, `destroy` on demand (manual workflow or manual job in the pipeline).

### 2.2 Non-functional requirements

- NFR01: Be stateless - no database or dedicated backend; state is handled by Terragrunt/Terraform (S3, GCS, etc.).
- NFR02: Go (latest stable), with lint configuration via **golangci-lint** and optional use of the official GitHub Action.
- NFR03: Portable across executors (GitHub Actions, GitLab CI, local).
- NFR04: Minimal configuration; prefer convention over configuration.
- NFR05: Full documentation in Markdown, published with **MkDocs** on GitHub Pages.

---

## 3. High-level architecture

### 3.1 Main components

- **Cultivator CLI (Go binary)**
  - Main entry points (e.g., `cultivator plan`, `cultivator apply`, `cultivator destroy`).
  - Implements flag/env parsing, module discovery, and orchestration of `terragrunt` calls.

- **Module Discovery**
  - Recursively walks directories from a root (e.g., `live/`).
  - Identifies folders with `terragrunt.hcl` and applies filters (env, tags, etc.).

- **Execution Module (Runner)**
  - Invokes `terragrunt` via `os/exec`, building arguments per operation (`plan`, `apply`, `destroy`).
  - Implements configurable parallelism and concurrency limits.

- **Configuration Module (Config Loader)**
  - Reads an optional config file (e.g., `.cultivator.yaml`) with defaults for root path, concurrency, filters, etc.
  - Merges with environment variables and CLI flags (precedence: flags > env > file).

- **Logging/Reporting Module**
  - Outputs structured logs in text format with consistent formatting.
  - Generates a final summary with module list, status, and log links (when running in CI).

- **Docs**
  - `docs/` directory with Markdown pages.
  - Build and deploy via MkDocs + GitHub Actions to GitHub Pages.

### 3.2 CI flow (high level)

1. A developer opens a PR with changes in a Terragrunt repository.
2. GitHub Actions / GitLab CI triggers the PR pipeline.
3. The job runs:

- repo checkout;
- Go setup (if building Cultivator in the pipeline) or binary download;
- cloud authentication (IAM, service accounts, etc.);
- `cultivator plan` with scope parameters (e.g., environment, folder).

4. Cultivator:

- discovers modules;
- runs `terragrunt plan` for each eligible module;
- returns consolidated output + exit code.

5. CI marks the job as success/failure and can post logs on the PR (platform-native).

For `apply`/`destroy`, the recommendation is to run separate pipelines (e.g., triggered on merge or manual approval in the pipeline), following the pattern suggested by Gruntwork for Terragrunt CI/CD.

---

## 4. CLI design

### 4.1 Commands and subcommands

Suggested interface:

```text
cultivator plan    [MODULE_PATH] [flags]
cultivator apply   [MODULE_PATH] [flags]
cultivator destroy [MODULE_PATH] [flags]
cultivator version
cultivator doctor
```

**Positional Arguments (optional)**:

- `[MODULE_PATH]` (string, optional): specific module path to execute. Can be:
  - A relative path: `cloudwatch/log-group/example`
  - A path with trailing filename: `cloudwatch/log-group/example/terragrunt.hcl`
  - A path with leading `./`: `./cloudwatch/log-group/example`
  - Internally normalized and treated as a filter (equivalent to `--include` flag)

Examples:

```bash
# Run plan on a specific module
cultivator plan cloudwatch/log-group/lambda-example

# With filename
cultivator plan cloudwatch/log-group/lambda-example/terragrunt.hcl

# Combined with other flags
cultivator plan --parallelism=5 cloudwatch/log-group/lambda-example
cultivator plan cloudwatch/log-group/lambda-example --destroy
```

**Backward Compatibility**: All existing flag-based invocations continue to work:

```bash
cultivator plan --include cloudwatch/log-group/lambda-example --parallelism=5
```

Main flags (common):

- `--root` (string): root directory of Terragrunt environments (default: `.` or `live/`).
- `--env` (string): logical environment filter (e.g., `dev`, `stage`, `prod`).
- `--include` (stringSlice): list of relative paths to run (e.g., `--include envs/prod/app1`).
- `--exclude` (stringSlice): paths to ignore.
- `--tags` (stringSlice): logical tags, interpreted by convention (e.g., labels in `terragrunt.hcl`).
- `--parallelism` (int): maximum number of concurrent executions.
- `--non-interactive` (bool): forces non-interactive modes in `terragrunt` (required in CI).

Command-specific flags:

- `plan`:
  - `--destroy` (bool): generate a destroy plan (shortcut for `terragrunt plan -destroy`).

- `apply`:
  - `--auto-approve` (bool): passes `-auto-approve` to `terragrunt` (or equivalent).

- `destroy`:
  - same set as apply, but always in destroy mode.

### 4.2 Layout conventions

Follow the Gruntwork "infrastructure live repo pattern", for example:

```text
live/
  prod/
    eu-west-1/
      app1/terragrunt.hcl
      app2/terragrunt.hcl
  stage/
  dev/
```

- `--root=live/`
- `--env=prod` could map to `live/prod/**/terragrunt.hcl`.
- Evaluate support for multiple layouts via configuration (e.g., `layout: gruntwork` / `layout: flat`).

---

## 5. Module discovery and selection

### 5.1 Discovery algorithm

1. Receive `rootDir` (default or via flag/config).
2. Recursively walk (`filepath.WalkDir`) looking for `terragrunt.hcl` files.
3. For each `terragrunt.hcl`, create a `Module` object:

- `Path` (directory where the file lives);
- `Env` (derived from the path or an attribute in `terragrunt.hcl`);
- `Tags` (optionally derived from comments or extra blocks).

4. Apply filters (`Env`, `Include`, `Exclude`, `Tags`).
5. Return the list of modules to execute.

### 5.2 Respecting root.hcl

- If there is a `root.hcl` in the root directory, use Terragrunt with `--config` pointing to it when appropriate.
- Ensure that the `terragrunt` invocation preserves the expected hierarchy (run from the module directory so the parent `root.hcl` is loaded).

---

## 6. Terragrunt execution

### 6.1 Execution model

- Each module runs in an independent process (`os/exec.Command`).
- Example commands:

```bash
cd live/prod/eu-west-1/app1
terragrunt plan -no-color -input=false
```

- For `apply`:

```bash
cd live/prod/eu-west-1/app1
terragrunt apply -no-color -input=false -auto-approve
```

- For `destroy`:

```bash
cd live/prod/eu-west-1/app1
terragrunt destroy -no-color -input=false -auto-approve
# ou
terragrunt apply -destroy -no-color -input=false -auto-approve
```

- Pass environment variables from the parent process (CI), including `AWS_*`, `GOOGLE_*`, etc. (provider-agnostic).

### 6.2 Parallelism

- Use a worker pool with a `parallelism` limit.
- Simple strategy:
  - Enqueue modules;
  - Run up to `N` workers in parallel;
  - Aggregate results (stdout/stderr, exit code) into a result structure.

---

## 7. CI integration

### 7.1 GitHub Actions

Example workflow for `plan` in PRs:

```yaml
name: Cultivator Plan

on:
  pull_request:

jobs:
  plan:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Setup Go
        uses: actions/setup-go@v5
        with:
          go-version: 'stable'

      - name: Build Cultivator
        run: go build -o bin/cultivator ./cmd/cultivator

      - name: Install Terragrunt
        run: |
          curl -L https://github.com/gruntwork-io/terragrunt/releases/latest/download/terragrunt_linux_amd64 -o terragrunt
          chmod +x terragrunt
          sudo mv terragrunt /usr/local/bin/terragrunt

      - name: Cloud auth
        run: |
          # Ex.: configurar AWS/GCP/Azure via OIDC ou secrets

      - name: Cultivator plan
        run: |
          ./bin/cultivator plan --root=live --env=dev --non-interactive
```

Lint com **golangci-lint** usando a GitHub Action oficial:

```yaml
  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: 'stable'
      - name: golangci-lint
        uses: golangci/golangci-lint-action@v8
        with:
          version: latest
```

Deploying MkDocs docs to GitHub Pages can follow a standard workflow (e.g., the `mkdocs-deploy-gh-pages` action).

### 7.2 GitLab CI

Simplified pipeline example:

```yaml
stages:
  - lint
  - plan
  - apply

lint:
  stage: lint
  image: golang:latest
  script:
    - go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
    - golangci-lint run ./...

plan:
  stage: plan
  image: hashicorp/terraform:latest # ou imagem custom com terragrunt + go
  script:
    - go build -o bin/cultivator ./cmd/cultivator
    - ./bin/cultivator plan --root=live --env=dev --non-interactive
  only:
    - merge_requests

apply:
  stage: apply
  image: hashicorp/terraform:latest
  script:
    - go build -o bin/cultivator ./cmd/cultivator
    - ./bin/cultivator apply --root=live --env=dev --non-interactive --auto-approve
  when: manual
  only:
    - main
```

---

## 8. Go code organization

Proposed repository structure:

```text
.
├── cmd/
│   └── cultivator/
│       └── main.go
├── internal/
│   ├── cli/          # parsing de flags, comandos
│   ├── config/       # load/merge de config (arquivo/env/flags)
│   ├── discovery/    # descoberta de módulos terragrunt
│   ├── runner/       # execução de terragrunt, worker pool
│   ├── logging/      # logger e formatação de saída
│   └── ci/           # helpers específicos para CI (opcional)
├── pkg/              # APIs públicas (se planejado uso como lib)
├── docs/             # Markdown p/ MkDocs
├── mkdocs.yml        # configuração MkDocs
├── .golangci.yml     # config golangci-lint
└── .github/workflows/
    ├── ci.yaml       # build + lint + testes
    └── docs.yaml     # deploy mkdocs p/ gh-pages
```

- Use `golangci-lint` with configuration in `.golangci.yml`, enabling essential linters (e.g., `govet`, `staticcheck`, `gocyclo`, `gosec` as needed).

---

## 9. Documentation (MkDocs + GitHub Pages)

- Docs structure (example):

```text
docs/
  index.md
  getting-started.md
  configuration.md
  ci-github-actions.md
  ci-gitlab-ci.md
  architecture.md   # este system design
  faq.md
```

- Minimal `mkdocs.yml`:

```yaml
site_name: Cultivator
repo_url: https://github.com/<org>/cultivator

nav:
  - Home: index.md
  - Getting started: getting-started.md
  - Configuration: configuration.md
  - CI:
      - GitHub Actions: ci-github-actions.md
      - GitLab CI: ci-gitlab-ci.md
  - Architecture: architecture.md
```

- Deploy to GitHub Pages via GitHub Actions using the MkDocs deploy action.

---

## 10. Future extensions

- Support generating `plan` comments directly on PRs via the platform API (if you want to get closer to the Atlantis UX).
- Support periodic "drift scan" (mode `cultivator drift-check`) running `plan` in read-only mode.
- Possible "graph" mode that generates a dependency visualization between modules based on directory structure.
