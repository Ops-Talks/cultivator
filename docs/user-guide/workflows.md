# Workflows

Learn about the available Cultivator commands and workflows.

## Available Commands

### Plan Command

Run Terraform/Terragrunt plan on affected modules.

```
cultivator plan
```

**Behavior**:
- Detects changed modules
- Builds dependency graph
- Runs plan in dependency order
- Posts results to PR

**Options**:
```
cultivator plan --all          # Plan all modules (not just affected)
cultivator plan --module vpc   # Plan specific module
```

### Apply Command

Apply approved Terraform/Terragrunt changes.

```
cultivator apply
```

**Behavior**:
- Requires approval (if configured)
- Runs apply in dependency order
- Locks modules during apply
- Posts results to PR

**Requirements**:
- Plan must have been run first
- Must have necessary permissions
- Branch protection rules respected

**Options**:
```
cultivator apply --all         # Apply all modules
cultivator apply --force       # Skip approval requirement (if authorized)
```

## Workflow Examples

### Basic Workflow

1. Open a Pull Request with infrastructure changes
2. Cultivator automatically runs plan (if `auto_plan: true`)
3. Review plan results in PR comment
4. Comment: `cultivator apply`
5. Cultivator executes apply
6. Results posted to PR

### Approval Workflow

1. Create PR with changes
2. Comment: `cultivator plan`
3. Review plan output
4. If changes look good: `cultivator apply`
5. If `require_approval: true`, another reviewer must approve
6. Apply executes and results appear in PR

### Multi-Module Workflow

Infrastructure structure:
```
vpc/
  - main.tf
security-groups/
  - main.tf (depends on vpc)
app/
  - main.tf (depends on vpc and security-groups)
```

PR changes `vpc/main.tf`:

1. Cultivator detects: vpc affected
2. Determines: security-groups and app also affected (due to dependencies)
3. Plans: vpc → security-groups → app
4. Shows dependency chain in PR comment

## Status Indicators

Cultivator uses status indicators in PR comments:

| Status | Meaning |
|--------|---------|
| ✓ Plan Complete | Plan succeeded, ready to apply |
| ✗ Plan Failed | Plan had errors, needs fixes |
| ⊗ Apply In Progress | Currently applying changes |
| ✓ Applied | Apply succeeded |
| ✗ Apply Failed | Apply had errors |

## Error Handling

If a command fails, Cultivator will:

1. Post error details to PR comment
2. Suggest troubleshooting steps
3. Provide error context and logs
4. Maintain lock (if apply failed)

Fix the issue and comment again to retry:

```
cultivator plan
```

## Concurrency & Locking

- Only one apply per module at a time
- Lock timeout: configurable (default 30 minutes)
- Automatic cleanup on completion
- Prevents accidental concurrent applies

## Rate Limiting

GitHub API has rate limits. For large infrastructure:

- Stagger multiple runs (don't run all at once)
- Combine PR changes to reduce plan runs
- Plan all modules, not frequently

---

See [Features](features.md) for more about what Cultivator can do.
