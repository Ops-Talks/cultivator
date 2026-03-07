# GitHub Actions Integration

This guide shows production-ready patterns for running Cultivator in GitHub Actions.

Key features of this approach:

- Pre-compiled binaries: No compilation overhead; fast and efficient
- Versioned tools: Pin OpenTofu, Terragrunt, and Cultivator versions
- Structured workflow: Doctor check before plan/apply for early error detection

Unlike GitLab CI, GitHub Actions uses:

- `on:` events (`pull_request`, `workflow_dispatch`)
- job-level `if:` expressions
- explicit `permissions` for API operations
- PR comments usually posted via `actions/github-script`

---

## Recommended workflow

```yaml
# .github/workflows/cultivator.yml

name: Cultivator

on:
  pull_request:
    branches: [main]
    types: [opened, synchronize, reopened, closed]
  workflow_dispatch:

permissions:
  contents: read
  pull-requests: write

env:
  CULTIVATOR_VERSION: v0.3.10
  TOFU_VERSION: 1.11.5
  TERRAGRUNT_VERSION: 0.99.0
  CULTIVATOR_ROOT: providers
  CULTIVATOR_ENV: ""
  CULTIVATOR_PARALLELISM: "4"

jobs:
  doctor:
    name: Doctor
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@v4

      - name: Install tools
        shell: bash
        run: |
          set -euo pipefail

          sudo apt-get update
          sudo apt-get install -y wget unzip curl

          wget -q https://github.com/opentofu/opentofu/releases/download/v${TOFU_VERSION}/tofu_${TOFU_VERSION}_linux_amd64.zip
          sudo unzip -q tofu_${TOFU_VERSION}_linux_amd64.zip -d /usr/local/bin
          rm tofu_${TOFU_VERSION}_linux_amd64.zip

          wget -q -O terragrunt \
            https://github.com/gruntwork-io/terragrunt/releases/download/v${TERRAGRUNT_VERSION}/terragrunt_linux_amd64
          chmod +x terragrunt
          sudo mv terragrunt /usr/local/bin/terragrunt

          wget -q -O cultivator \
            https://github.com/Ops-Talks/cultivator/releases/download/${CULTIVATOR_VERSION}/cultivator-linux-amd64
          chmod +x cultivator
          sudo mv cultivator /usr/local/bin/cultivator

      - name: Run doctor
        run: cultivator doctor --root "$CULTIVATOR_ROOT"

  plan:
    name: Plan
    runs-on: ubuntu-latest
    needs: doctor
    if: (github.event_name == 'pull_request' && github.event.action != 'closed') || github.event_name == 'workflow_dispatch'
    steps:
      - name: Checkout
        uses: actions/checkout@v4
        with:
          fetch-depth: 0 # Required for Magic Mode (git diff)

      - name: Install tools
        shell: bash
        run: |
          set -euo pipefail

          sudo apt-get update
          sudo apt-get install -y wget unzip curl

          wget -q https://github.com/opentofu/opentofu/releases/download/v${TOFU_VERSION}/tofu_${TOFU_VERSION}_linux_amd64.zip
          sudo unzip -q tofu_${TOFU_VERSION}_linux_amd64.zip -d /usr/local/bin
          rm tofu_${TOFU_VERSION}_linux_amd64.zip

          wget -q -O terragrunt \
            https://github.com/gruntwork-io/terragrunt/releases/download/v${TERRAGRUNT_VERSION}/terragrunt_linux_amd64
          chmod +x terragrunt
          sudo mv terragrunt /usr/local/bin/terragrunt

          wget -q -O cultivator \
            https://github.com/Ops-Talks/cultivator/releases/download/${CULTIVATOR_VERSION}/cultivator-linux-amd64
          chmod +x cultivator
          sudo mv cultivator /usr/local/bin/cultivator

      - name: Run plan
        shell: bash
        run: |
          set -euo pipefail

          args=(
            --root "$CULTIVATOR_ROOT"
            --parallelism "$CULTIVATOR_PARALLELISM"
            --non-interactive=true
          )

          # Enable Magic Mode on Pull Requests to target only changed modules
          if [[ "${{ github.event_name }}" == "pull_request" ]]; then
            args+=(--changed-only --base "${{ github.base_ref }}")
          fi

          if [[ -n "$CULTIVATOR_ENV" ]]; then
            args+=(--env "$CULTIVATOR_ENV")
          fi

          # 2>&1 captures Terragrunt output (written to stderr) alongside stdout.
          cultivator plan "${args[@]}" 2>&1 | tee plan_output.txt

      - name: Upload plan output
        if: always()
        uses: actions/upload-artifact@v4
        with:
          name: plan-output
          path: plan_output.txt

      - name: Comment plan on PR
        if: always() && github.event_name == 'pull_request' && github.event.pull_request.head.repo.fork == false
        uses: actions/github-script@v7
        with:
          github-token: ${{ secrets.GITHUB_TOKEN }}
          script: |
            const fs = require('fs');
            const plan = fs.existsSync('plan_output.txt')
              ? fs.readFileSync('plan_output.txt', 'utf8')
              : 'No plan output file found.';

            const body = [
              '## Cultivator Plan',
              '',
              '```text',
              plan.slice(0, 65000),
              '```'
            ].join('\n');

            await github.rest.issues.createComment({
              owner: context.repo.owner,
              repo: context.repo.repo,
              issue_number: context.issue.number,
              body
            });

  apply:
    name: Apply
    runs-on: ubuntu-latest
    needs: doctor
    if: github.event_name == 'pull_request' && github.event.action == 'closed' && github.event.pull_request.merged == true
    environment: production
    steps:
      - name: Checkout
        uses: actions/checkout@v4
        with:
          ref: ${{ github.event.pull_request.base.ref }}
          fetch-depth: 0 # Required for accurate change mapping and DAG resolution

      - name: Install tools
        shell: bash
        run: |
          set -euo pipefail

          sudo apt-get update
          sudo apt-get install -y wget unzip curl

          wget -q https://github.com/opentofu/opentofu/releases/download/v${TOFU_VERSION}/tofu_${TOFU_VERSION}_linux_amd64.zip
          sudo unzip -q tofu_${TOFU_VERSION}_linux_amd64.zip -d /usr/local/bin
          rm tofu_${TOFU_VERSION}_linux_amd64.zip

          wget -q -O terragrunt \
            https://github.com/gruntwork-io/terragrunt/releases/download/v${TERRAGRUNT_VERSION}/terragrunt_linux_amd64
          chmod +x terragrunt
          sudo mv terragrunt /usr/local/bin/terragrunt

          wget -q -O cultivator \
            https://github.com/Ops-Talks/cultivator/releases/download/${CULTIVATOR_VERSION}/cultivator-linux-amd64
          chmod +x cultivator
          sudo mv cultivator /usr/local/bin/cultivator

      - name: Run apply
        shell: bash
        run: |
          set -euo pipefail

          args=(
            --root "$CULTIVATOR_ROOT"
            --parallelism "$CULTIVATOR_PARALLELISM"
            --non-interactive=true
            --auto-approve=true
          )

          if [[ -n "$CULTIVATOR_ENV" ]]; then
            args+=(--env "$CULTIVATOR_ENV")
          fi

          # 2>&1 captures Terragrunt output (written to stderr) alongside stdout.
          # Execution order is automatically determined by the built-in DAG engine.
          cultivator apply "${args[@]}" 2>&1 | tee apply_output.txt

      - name: Upload apply output
        if: always()
        uses: actions/upload-artifact@v4
        with:
          name: apply-output
          path: apply_output.txt

      - name: Comment apply on PR
        if: always() && github.event_name == 'pull_request'
        uses: actions/github-script@v7
        with:
          github-token: ${{ secrets.GITHUB_TOKEN }}
          script: |
            const fs = require('fs');
            const output = fs.existsSync('apply_output.txt')
              ? fs.readFileSync('apply_output.txt', 'utf8')
              : 'No apply output file found.';

            const body = [
              '## Cultivator Apply',
              '',
              '```text',
              output.slice(0, 65000),
              '```'
            ].join('\n');

            await github.rest.issues.createComment({
              owner: context.repo.owner,
              repo: context.repo.repo,
              issue_number: context.issue.number,
              body
            });
```

---

## Workflow example

A complete reference workflow is available in [`examples/github-actions.yml`](../../examples/github-actions.yml). It demonstrates the full plan → apply lifecycle with `doctor`, PR comments, and artifact uploads.



---

## Optional: use a config file

A config file is optional in GitHub Actions. If you use one, pass it explicitly with `--config`.

```yaml
- name: Plan with config file
  run: |
    cultivator plan \
      --config=cultivator.yml \
      --non-interactive=true
```

---

## Key differences vs GitLab CI

- GitHub uses `on:` events; GitLab uses `rules:` and pipeline sources.
- GitHub PR comments are usually posted with `actions/github-script` and `secrets.GITHUB_TOKEN`.
- GitHub requires explicit `permissions` in workflow for PR write operations.
- GitHub environment approvals are configured in repository environments (`environment: production`).

## Execution flow on Pull Request

When a PR is opened or updated:

1. `doctor` runs first.
2. `plan` runs only after `doctor` succeeds (`needs: doctor`).
3. `apply` is skipped (it runs only when the PR is merged).
4. Plan output is uploaded as artifact.
5. PR comment is posted only for non-fork PRs in this example.

When a PR is merged (event `pull_request` + action `closed` + `merged == true`):

1. `doctor` runs.
2. `apply` runs after `doctor`.
3. Apply output is uploaded as artifact.
4. A new PR comment is created with the `apply` result.

When a PR is closed without merge (`merged == false`):

1. The workflow can still be triggered by the `closed` action.
2. The `apply` job is skipped by the `if` condition.
3. No infrastructure change is executed.

## Approval vs merge

GitHub Actions can reliably gate on `merged == true` in this event. Approval is a repository policy concern.

- If branch protection requires approvals, a merged PR is also approved by policy.
- If branch protection does not require approvals, a PR may be merged without approval.

## Branch protection (recommended)

To enforce the policy "only run apply after approved and merged PR", configure branch protection on `main`:

1. Require at least 1 approval before merge.
2. Require status checks to pass (for example `doctor` and `plan`).
3. Optionally require conversation resolution before merge.

With these settings, merge is blocked until approval and successful checks, and this workflow runs `apply` only after merge (`merged == true`).

---

## Secrets and credentials

Store cloud credentials in **Settings → Secrets and variables → Actions** and expose them in the job:

```yaml
env:
  AWS_ACCESS_KEY_ID: ${{ secrets.AWS_ACCESS_KEY_ID }}
  AWS_SECRET_ACCESS_KEY: ${{ secrets.AWS_SECRET_ACCESS_KEY }}
  AWS_REGION: us-east-1
```

Cultivator does not manage credentials; Terragrunt/OpenTofu/Terraform reads them from the environment.

---

## Troubleshooting

### `cultivator: command not found`

Verify install step ran successfully and binary was moved to `/usr/local/bin`.

### `terragrunt: command not found`

Cultivator delegates to Terragrunt. Install both binaries in the same job.

### PR comment step fails with 403

Ensure workflow includes:

```yaml
permissions:
  pull-requests: write
```

### No stacks discovered

Check `CULTIVATOR_ROOT` and optional `CULTIVATOR_ENV` filter.

---

## Further reading

- [Quickstart](../getting-started/quickstart.md)
- [Configuration](../getting-started/configuration.md)
- [CLI Reference](cli-reference.md)
- [GitHub Actions documentation](https://docs.github.com/en/actions)
- [GitLab CI equivalent](gitlab-pipelines.md)
