# CLI Reference

Learn the key commands and workflows for Cultivator.

## Available Commands

### Plan

Run `terragrunt plan` on matching stacks.

```bash
cultivator plan --root=live --env=dev --non-interactive
```

**Options**:

- `--destroy` - run `terragrunt plan -destroy`
- `--non-interactive` - add `-input=false`
- `--dry-run` - don't execute terragrunt commands
- `--changed-only` - only execute modules with changed files
- `--base` - git base reference for `--changed-only` (default: `HEAD`)

### Apply

Run `terragrunt apply` on matching stacks.

```bash
cultivator apply --root=live --env=dev --non-interactive --auto-approve
```

**Options**:

- `--auto-approve` - add `-auto-approve`
- `--non-interactive` - add `-input=false`
- `--dry-run` - don't execute terragrunt commands
- `--changed-only` - only execute modules with changed files
- `--base` - git base reference for `--changed-only` (default: `HEAD`)

### Destroy

Run `terragrunt destroy` on matching stacks.

```bash
cultivator destroy --root=live --env=dev --non-interactive --auto-approve
```

**Options**:

- `--auto-approve` - add `-auto-approve`
- `--non-interactive` - add `-input=false`
- `--dry-run` - don't execute terragrunt commands
- `--changed-only` - only execute modules with changed files
- `--base` - git base reference for `--changed-only` (default: `HEAD`)

## CLI Examples

### Plan specific environment

```bash
./cultivator plan --root=live --env=prod --non-interactive
```

### Plan specific stacks

```bash
./cultivator plan --root=live --include=envs/prod/app1,envs/prod/app2 --non-interactive
```

### Plan excluding experimental stacks

```bash
./cultivator plan --root=live --exclude=experimental --non-interactive
```

### Plan only stacks with specific tags

```bash
./cultivator plan --root=live --tags=critical --non-interactive
```

### Plan with custom parallelism

```bash
./cultivator plan --root=live --parallelism=8 --non-interactive
```

### Magic Mode: Plan only changed modules

Compare current branch against `main` and only run `plan` on modules that have file changes.

```bash
./cultivator plan --changed-only --base=main
```

### Doctor

Verify your environment and configuration.

```bash
cultivator doctor --root=live --config=cultivator.yml
```

### Version

Print version information.

```bash
cultivator version
```

## CLI Arguments & Flags

### Positional Arguments

Cultivator supports a positional argument for the module path. This is a shorthand for filtering to a specific module.

```bash
cultivator plan cloudwatch/log-group/example
```
*Note: The path is automatically normalized (e.g., removing leading `./` or trailing `/terragrunt.hcl`).*

## Environment Variables

| Variable | Values | Default | Description |
|----------|--------|---------|-------------|
| `CULTIVATOR_LOG_LEVEL` | `debug`, `info`, `warning`, `error` | `info` | Minimum log level. |
| `CULTIVATOR_ROOT` | path | `.` | Root directory for discovery. |
| `CULTIVATOR_ENV` | string | | Environment filter. |
| `CULTIVATOR_INCLUDE` | comma-separated paths | | Paths to include. |
| `CULTIVATOR_EXCLUDE` | comma-separated paths | | Paths to exclude. |
| `CULTIVATOR_TAGS` | comma-separated tags | | Tag filters. |
| `CULTIVATOR_PARALLELISM` | integer | CPU count | Max parallel executions. |
| `CULTIVATOR_NON_INTERACTIVE` | `true`, `false` | `false` | Force non-interactive mode. |
| `CULTIVATOR_DRY_RUN` | `true`, `false` | `false` | Enable dry-run mode. |
| `CULTIVATOR_CHANGED_ONLY` | `true`, `false` | `false` | Enable Magic Mode (Git changes). |
| `CULTIVATOR_BASE_REF` | string | `HEAD` | Git base reference for Magic Mode. |

**Example — enable Magic Mode in CI:**

```bash
CULTIVATOR_CHANGED_ONLY=true CULTIVATOR_BASE_REF=main cultivator plan
```

## Output Format

Cultivator produces human-readable text output by default, organized by module with clear section headers:

```
=== plan: live/prod/vpc ===
Running terragrunt plan in /path/to/live/prod/vpc...
[output from terragrunt]

=== plan: live/prod/app ===
Running terragrunt plan in /path/to/live/prod/app...
[output from terragrunt]
```

Each stack shows:
- Module path
- Terragrunt output and any errors
- Clear demarcation for easy scanning

## Binary Invocation

Local development (after building):
```bash
go build -o cultivator ./cmd/cultivator
./cultivator plan --root=live --env=dev
```

CI/CD environments (pre-compiled/downloaded binaries in PATH):
```bash
cultivator plan --root=live --env=dev
```

Both forms are equivalent. The difference:
- `./cultivator` - Runs binary in current directory (local builds)
- `cultivator` - Runs binary found in system PATH (CI/production installs)

## Exit Codes

- `0` - Success (all stacks executed successfully)
- `1` - Failure (one or more stacks failed)
- `2` - Usage error (invalid flags or arguments)

## Execution Metrics

Cultivator automatically tracks and reports execution time:
- **Per-module duration**: Displayed in the logs for each module upon completion.
- **Total duration**: Displayed at the end of the command execution.

These metrics help identify infrastructure bottlenecks and optimize CI/CD runtime.

## Standard Workflow

For pull requests + main branch merges:

1. **PR triggers plan:**

   ```bash
   cultivator plan --root=live --env=dev --non-interactive
   ```

2. **Review plan output**

3. **Merge PR**

4. **Main branch triggers apply:**

   ```bash
   cultivator apply --root=live --env=dev --non-interactive --auto-approve
   ```

---

See [Features](features.md) for more about what Cultivator can do.
