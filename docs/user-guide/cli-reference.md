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

### Apply

Run `terragrunt apply` on matching stacks.

```bash
cultivator apply --root=live --env=dev --non-interactive --auto-approve
```

**Options**:

- `--auto-approve` - add `-auto-approve`
- `--non-interactive` - add `-input=false`

### Destroy

Run `terragrunt destroy` on matching stacks.

```bash
cultivator destroy --root=live --env=dev --non-interactive --auto-approve
```

**Options**:

- `--auto-approve` - add `-auto-approve`
- `--non-interactive` - add `-input=false`

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

## Environment Variables

| Variable               | Values                              | Default | Description                                                  |
|------------------------|-------------------------------------|---------|--------------------------------------------------------------|
| `CULTIVATOR_LOG_LEVEL` | `debug`, `info`, `warning`, `error` | `info`  | Minimum log level emitted by Cultivator. Terragrunt output is always printed regardless of this setting. |

**Example — enable debug logging:**

```bash
CULTIVATOR_LOG_LEVEL=debug cultivator plan --root=live --env=dev
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
