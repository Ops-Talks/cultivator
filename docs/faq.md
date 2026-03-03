# Frequently Asked Questions

## General Questions

### What is Cultivator?
Cultivator is a **CLI that orchestrates Terragrunt execution** in CI/CD pipelines and local environments. It discovers stacks, applies filters, respects dependencies, and orchestrates parallel execution of `plan`, `apply`, and `destroy` operations.

### How is it different from Atlantis or other GitHub automation tools?
Unlike Atlantis (which is comment-triggered automation in GitHub):
- **Atlantis**: Webhook-based; comments on PRs trigger automation inside GitHub
- **Cultivator**: CLI-based; you explicitly call it from CI jobs (GitHub Actions, GitLab CI, etc.)

**Advantages of Cultivator's approach:**
- Works in any CI system (GitHub Actions, GitLab CI, local development)
- Simpler to debug (just run the CLI command locally)
- Better separation of concerns (CI orchestrates, Cultivator executes)
- No GitHub-specific logic in the tool

### Why not just use shell scripts with Terragrunt directly?
While you can write custom bash scripts to orchestrate Terragrunt stacks, you'll need to:
- Manually list and maintain every stack in your pipeline
- Implement discovery, filtering, and dependency logic yourself
- Handle parallelization with complex semaphore patterns
- Rewrite scripts for different CI systems

Cultivator automates all of this. See [Why Cultivator?](index.md#why-cultivator) for detailed comparison with practical examples.

### Do I need a separate server?
No. Cultivator is a **CLI binary** that runs inside your existing CI/CD system. No additional infrastructure or webhooks required.

## Installation and Setup

### What versions of Terragrunt are supported?
Cultivator supports Terragrunt **v0.50.0+**. For best results, use recent versions (v1.0+).

### What about OpenTofu/Terraform version?
Cultivator works with any OpenTofu or Terraform version supported by your Terragrunt version.
- **Recommended**: OpenTofu v1.6+ or Terraform v1.5+
- Older versions may work but are not tested

### Can I build Cultivator from source?
Yes. Clone the repository and run:
```bash
go build -o cultivator ./cmd/cultivator
```
**Requires Go 1.25+**

### Can I run Cultivator in Docker?
Yes. A Dockerfile is included:
```bash
make docker-build          # Builds docker image
docker run cultivator:latest plan --help
```

## Usage

### How do I run a plan?
```bash
cultivator plan --root=live --env=dev --non-interactive
```

Then review the output, test locally, and proceed to apply.

### How do I approve and apply changes?
In your CI workflow:
```bash
# After PR review
cultivator apply --root=live --env=dev --non-interactive --auto-approve
```

(Approval is enforced at the CI level via branch protection rules, not by Cultivator)

### Can I run all stacks?
Yes. Omit filters to run all discovered stacks:
```bash
cultivator plan --root=live
```

Or filter by specific criteria:
```bash
# By environment
cultivator plan --root=live --env=prod

# By path
cultivator plan --root=live --include=envs/prod/*

# By tag
cultivator plan --root=live --tags=critical

# Combination
cultivator plan --root=live --env=prod --tags=critical --exclude=experimental
```

###What if a stack fails in the middle?
Cultivator stops execution and reports:
- Which stack failed
- The error output
- Exit code `1` (failure)

To retry:
1. Fix the underlying issue (Terraform/Terragrunt/infrastructure)
2. Run Cultivator again with the same flags
3. It will re-attempt all stacks (unchanged ones may be skipped by Terraform caching)

### How do I handle dependencies between stacks?
Cultivator automatically parses `dependency` blocks in `terragrunt.hcl`:

```hcl
dependency "vpc" {
  config_path = "../vpc"
}

inputs = {
  vpc_id = dependency.vpc.outputs.vpc_id
}
```

Cultivator ensures the VPC stack runs before dependent stacks. You don't need to specify order manually.

### Can I run Cultivator locally?
Yes! Cultivator is a local CLI tool. Useful for:
- **Local development**: Test changes before pushing
- **Debugging**: Run the exact same command as CI
- **Manual operations**: Apply changes immediately without CI jobs

```bash
./cultivator plan --root=live --env=dev
./cultivator apply --root=live --env=dev --auto-approve
```

## Security and Permissions

### How do I handle Terraform/cloud credentials in CI?
Use CI secrets management:

**GitHub Actions:**
```yaml
env:
  AWS_ACCESS_KEY_ID: ${{ secrets.AWS_ACCESS_KEY_ID }}
  AWS_SECRET_ACCESS_KEY: ${{ secrets.AWS_SECRET_ACCESS_KEY }}
```

**GitLab CI:**
```yaml
variables:
  AWS_ACCESS_KEY_ID: $AWS_ACCESS_KEY_ID
  AWS_SECRET_ACCESS_KEY: $AWS_SECRET_ACCESS_KEY
```

Cultivator doesn't manage credentials—it passes them through to Terragrunt/Terraform.

### Is my state safe?
Yes. Cultivator does **not** manage state:
- All state is stored in your Terraform/OpenTofu backend (S3, Terraform Cloud, etc.)
- Cultivator only orchestrates `plan` and `apply` commands
- Backend authentication is handled by Terragrunt/Terraform

### What about sensitive outputs?
Mark sensitive outputs in Terraform:
```hcl
output "database_password" {
  value       = aws_db_instance.main.password
  sensitive   = true
  description = "Database password (hidden)"
}
```

Cultivator respects `sensitive = true` and Terragrunt's redaction patterns.

### Who can run Cultivator?
In your CI workflow, you control who can trigger jobs:

**GitHub Actions:**
- PR authors (via `pull_request` event)
- Maintainers (via branch protection + required reviews)

**GitLab CI:**
- Developers (via `only:` rules)
- Protected branches (via branch protection)

Permissions check:
```bash
# If the CI user lacks cloud credentials, Terraform will error
# Example: AWS STS error due to invalid credentials
Error: error configuring Terraform AWS Provider: ...
```

## Output and Debugging

### How do I see which stacks will be affected?

Set `CULTIVATOR_OUTPUT_FORMAT=json` before running — there is no `--output-format` CLI flag:

```bash
CULTIVATOR_OUTPUT_FORMAT=json cultivator plan --root=live --env=prod
```

The JSON output includes:
- List of discovered stacks
- Stacks affected by filters
- Plan summary per stack

### How do I capture output for CI logs?
Cultivator writes to stdout/stderr. CI systems capture automatically:

**GitHub Actions:**
- Logs visible in the job's "Run cultivator plan" step

**GitLab CI:**
- Logs visible in the job's output section

**Local:**
- Logs printed to terminal; save with redirects:
  ```bash
  cultivator plan --root=live > cultivator.log 2>&1
  ```

### How do I debug Cultivator in detail?
Set verbose logging:
```bash
CULTIVATOR_LOG_LEVEL=debug cultivator plan --root=live --env=dev
```

Or switch to JSON output to see per-stack details:
```bash
CULTIVATOR_OUTPUT_FORMAT=json cultivator plan --root=live | jq .
```

## Troubleshooting

### Cultivator doesn't find any stacks
Check:
```bash
# Verify the root directory exists
ls -la live/

# Look for terragrunt.hcl files
find live -name terragrunt.hcl

# Run with explicit root
cultivator plan --root=./live
```

### Stack execution fails with "dependency not found"
Example error:
```
Error: dependency "vpc" not found
```

Solution:
- Verify the referenced stack's `config_path` exists
- Ensure all dependencies are under the same root
- Check spelling in `dependency` blocks

### Environment variables not being used
Verify precedence: **CLI flags > environment variables > config file**

```bash
# Example: This flag overrides CULTIVATOR_ENV
cultivator plan --env=prod --root=live
```

Check what config is loaded:
```bash
# Run without config file (uses defaults + env + flags)
cultivator plan --root=live --env=prod

# Run with config file (explicit)
cultivator plan --config=cultivator.yml --env=prod
```

## Advanced Questions

### Can I use Cultivator with Helm/Kustomize?
No, Cultivator is designed for **Terraform/Terragrunt only**. 

For Helm/Kustomize orchestration, consider:
- ArgoCD
- Flux
- Helm Operator

### Can I run custom scripts before/after Cultivator?
Yes, in your CI workflow:

**GitHub Actions:**
```yaml
- name: Pre-flight checks
  run: ./scripts/validate.sh

- name: Run Cultivator
  run: cultivator plan --root=live --env=prod

- name: Post-execution analysis
  run: ./scripts/analyze-logs.sh
```

**GitLab CI:**
```yaml
script:
  - ./scripts/validate.sh
  - cultivator plan --root=live --env=prod
  - ./scripts/analyze-logs.sh
```

### How do I monitor Cultivator runs?
- **GitHub Actions**: Check "Actions" → workflow run → job logs
- **GitLab CI**: Check "CI/CD" → pipeline → job logs
- **Local**: Check stdout/stderr and saved logs

### Can I integrate Cultivator with Slack/PagerDuty?
Yes, add a post-step in your CI workflow:

**GitHub Actions (Slack):**
```yaml
- name: Notify Slack on failure
  if: failure()
  uses: slackapi/slack-github-action@v1
  with:
    payload: |
      {"text": "Cultivator plan failed"}
```

**GitLab CI (Slack):**
```yaml
on_failure:
  - curl -X POST -H 'Content-type: application/json' \
    https://hooks.slack.com/... \
    --data '{"text":"Cultivator failed"}'
```

## Support and Contribution

### Where do I find more help?
- **Documentation**: [Full documentation](/)
- **Issues**: [GitHub Issues](https://github.com/Ops-Talks/cultivator/issues)
- **Discussions**: [GitHub Discussions](https://github.com/Ops-Talks/cultivator/discussions)
- **Contributing**: [Contributing Guide](../CONTRIBUTING.md)

### How do I report bugs?
1. Search [existing issues](https://github.com/Ops-Talks/cultivator/issues)
2. Create a new issue with:
   - Cultivator version: `cultivator version`
   - Terragrunt version: `terragrunt version`
   - OS/environment details
   - Steps to reproduce
   - Actual vs. expected output

### Can I contribute?
Yes! See [CONTRIBUTING.md](../CONTRIBUTING.md) for guidelines.
