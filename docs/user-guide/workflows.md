# Workflows

Learn the key commands and workflows for Cultivator.

## Available Commands

### Plan

Run `terragrunt plan` on matching modules.

```bash
cultivator plan --root=live --env=dev --non-interactive
```

**Options**:
- `--destroy` - run `terragrunt plan -destroy`
- `--non-interactive` - add `-input=false`

### Apply

Run `terragrunt apply` on matching modules.

```bash
cultivator apply --root=live --env=dev --non-interactive --auto-approve
```

**Options**:
- `--auto-approve` - add `-auto-approve`
- `--non-interactive` - add `-input=false`

### Destroy

Run `terragrunt destroy` on matching modules.

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

### Plan specific modules

```bash
./cultivator plan --root=live --include=envs/prod/app1,envs/prod/app2 --non-interactive
```

### Plan excluding experimental modules

```bash
./cultivator plan --root=live --exclude=experimental --non-interactive
```

### Plan only modules with specific tags

```bash
./cultivator plan --root=live --tags=critical --non-interactive
```

### Plan with custom parallelism

```bash
./cultivator plan --root=live --parallelism=8 --non-interactive
```

### Output JSON for CI/CD systems

```bash
./cultivator plan --root=live --output-format=json --non-interactive
```

## Exit Codes

- `0` - Success (all modules executed successfully)
- `1` - Failure (one or more modules failed)
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
