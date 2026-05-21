# Bitbucket Pipelines Integration

This guide shows production-ready patterns for running Cultivator in Bitbucket Pipelines.

Key features of this approach:

- Pre-compiled binaries: No compilation overhead; fast and efficient
- Versioned tools: Pin OpenTofu, Terragrunt, and Cultivator versions
- Structured pipeline: Doctor check before plan/apply for early error detection
- Magic Mode: Auto-detects `BITBUCKET_PR_DESTINATION_BRANCH` as the base ref

Unlike GitHub Actions or GitLab CI, Bitbucket Pipelines uses:

- `bitbucket-pipelines.yml` as the pipeline definition file
- `pipelines.pull-requests` for PR-triggered pipelines
- `pipelines.branches` for branch-triggered pipelines (e.g. after merge to `main`)
- `trigger: manual` for steps requiring human approval
- `deployment:` for environment promotion gates
- PR comments posted via the Bitbucket REST API using an App Password

---

## Prerequisites

### Repository variables

Set the following in **Repository settings > Repository variables**:

| Variable | Example | Description |
|---|---|---|
| `CULTIVATOR_VERSION` | `v0.4.10` | Cultivator release to install |
| `TOFU_VERSION` | `1.11.5` | OpenTofu version |
| `TERRAGRUNT_VERSION` | `0.99.1` | Terragrunt version |
| `CULTIVATOR_ROOT` | `providers` | Root directory for Terragrunt modules |
| `CULTIVATOR_ENV` | _(empty)_ | Optional environment filter |
| `CULTIVATOR_PARALLELISM` | `4` | Max parallel executions |
| `BITBUCKET_USERNAME` | `ci-bot` | Bitbucket username for API auth |
| `BITBUCKET_APP_PASSWORD` | `...` | App Password with **Pull requests: Write** scope |

Cloud credentials (e.g. `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`) should also be stored as repository variables.

### Creating a Bitbucket App Password

1. Go to **Personal settings > App passwords**.
2. Click **Create app password**.
3. Grant the **Pull requests: Write** permission.
4. Store the generated token as `BITBUCKET_APP_PASSWORD`.

---

## Recommended pipeline

```yaml
# bitbucket-pipelines.yml

image: alpine:3.21

pipelines:
  pull-requests:
    '**':
      - step:
          name: Doctor
          script:
            - apk add --no-cache wget unzip curl ca-certificates
            - wget -q -O /usr/local/bin/terragrunt
                https://github.com/gruntwork-io/terragrunt/releases/download/v${TERRAGRUNT_VERSION}/terragrunt_linux_amd64
            - chmod +x /usr/local/bin/terragrunt
            - wget -q -O /usr/local/bin/cultivator
                https://github.com/Ops-Talks/cultivator/releases/download/${CULTIVATOR_VERSION}/cultivator-linux-amd64
            - chmod +x /usr/local/bin/cultivator
            - cultivator doctor --root "$CULTIVATOR_ROOT"

      - step:
          name: Plan
          clone:
            depth: full  # Required for Magic Mode (git diff)
          script:
            - # ... install tools ...
            - |
              ARGS=(
                --root "$CULTIVATOR_ROOT"
                --parallelism "$CULTIVATOR_PARALLELISM"
                --non-interactive=true
                --changed-only  # Magic Mode: auto-detects BITBUCKET_PR_DESTINATION_BRANCH
              )
              if [ -n "$CULTIVATOR_ENV" ]; then ARGS+=(--env "$CULTIVATOR_ENV"); fi

              cultivator plan "${ARGS[@]}" 2>&1 | tee plan_output.txt
              CULTIVATOR_EXIT="${PIPESTATUS[0]}"

              # Post plan output as a PR comment
              if [ -n "${BITBUCKET_APP_PASSWORD:-}" ]; then
                PLAN_OUTPUT=$(cat plan_output.txt)
                COMMENT=$(printf '## Cultivator Plan\n\n```\n%s\n```' "${PLAN_OUTPUT}")
                curl --silent --show-error --fail \
                  --user "${BITBUCKET_USERNAME}:${BITBUCKET_APP_PASSWORD}" \
                  --header "Content-Type: application/json" \
                  --request POST \
                  "https://api.bitbucket.org/2.0/repositories/${BITBUCKET_REPO_FULL_NAME}/pullrequests/${BITBUCKET_PR_ID}/comments" \
                  --data "$(jq -n --arg body "${COMMENT}" '{"content":{"raw":$body}}')"
              fi

              exit "$CULTIVATOR_EXIT"
          artifacts:
            - plan_output.txt

  branches:
    main:
      - step:
          name: Doctor
          # ... same as above ...

      - step:
          name: Apply
          trigger: manual
          deployment: production
          clone:
            depth: full
          script:
            - # ... install tools ...
            - |
              ARGS=(
                --root "$CULTIVATOR_ROOT"
                --parallelism "$CULTIVATOR_PARALLELISM"
                --non-interactive=true
                --auto-approve=true
              )
              if [ -n "$CULTIVATOR_ENV" ]; then ARGS+=(--env "$CULTIVATOR_ENV"); fi

              cultivator apply "${ARGS[@]}" 2>&1 | tee apply_output.txt
          artifacts:
            - apply_output.txt
```

A complete reference pipeline is available in [`examples/bitbucket-pipelines.yml`](../../examples/bitbucket-pipelines.yml).

---

## Magic Mode (changed-only)

When `--changed-only` is active, Cultivator automatically reads the
`BITBUCKET_PR_DESTINATION_BRANCH` environment variable to determine the git
base reference. This means you do **not** need to pass `--base` explicitly in
Bitbucket Pipelines — the right branch is detected automatically.

```bash
# This is sufficient; --base is resolved automatically from the pipeline env.
cultivator plan --changed-only --root providers
```

If you are running outside a PR pipeline (e.g. a manual branch run), set
`BITBUCKET_PR_DESTINATION_BRANCH` explicitly or pass `--base <branch>` on the
command line.

---

## PR comments via the Bitbucket REST API

Bitbucket does not provide a built-in write token equivalent to GitHub's
`GITHUB_TOKEN`. Comments must be posted using the
[Bitbucket REST API](https://developer.atlassian.com/cloud/bitbucket/rest/api-group-pullrequests/)
authenticated with an App Password:

```bash
curl --silent --show-error --fail \
  --user "${BITBUCKET_USERNAME}:${BITBUCKET_APP_PASSWORD}" \
  --header "Content-Type: application/json" \
  --request POST \
  "https://api.bitbucket.org/2.0/repositories/${BITBUCKET_REPO_FULL_NAME}/pullrequests/${BITBUCKET_PR_ID}/comments" \
  --data "$(jq -n --arg body "## Cultivator Plan\n\n\`\`\`\n$(cat plan_output.txt)\n\`\`\`" '{"content":{"raw":$body}}')"
```

Available Bitbucket pipeline environment variables:

| Variable | Description |
|---|---|
| `BITBUCKET_REPO_FULL_NAME` | `workspace/repo-slug` |
| `BITBUCKET_PR_ID` | Pull request ID (only set in PR pipelines) |
| `BITBUCKET_PR_DESTINATION_BRANCH` | Target branch of the PR |
| `BITBUCKET_BRANCH` | Current branch |
| `BITBUCKET_COMMIT` | Current commit SHA |
| `BITBUCKET_BUILD_NUMBER` | Build number |

---

## Apply after merge

Bitbucket Pipelines does not expose a "PR merged" event in the same way as
GitHub Actions. The recommended pattern is to trigger `apply` via a
`pipelines.branches.main` pipeline that runs on every push to `main`
(which in practice means after a PR is merged):

```yaml
branches:
  main:
    - step:
        name: Apply
        trigger: manual       # Requires a human to click "Run" in the UI
        deployment: production
        script:
          - cultivator apply --root providers --auto-approve=true
```

With `trigger: manual`, the pipeline pauses at the Apply step until someone
approves it in the Bitbucket UI, providing a safety gate equivalent to GitHub's
environment approvals or GitLab's `when: manual`.

---

## Optional: use a config file

```bash
cultivator plan \
  --config=cultivator.yml \
  --non-interactive=true
```

```yaml
# cultivator.yml
root: providers
parallelism: 4
non_interactive: true
```

---

## Key differences vs GitHub Actions / GitLab CI

| Aspect | Bitbucket Pipelines | GitHub Actions | GitLab CI |
|---|---|---|---|
| Pipeline file | `bitbucket-pipelines.yml` | `.github/workflows/*.yml` | `.gitlab-ci.yml` |
| PR event | `pipelines.pull-requests` | `on: pull_request` | `rules: merge_request_event` |
| Manual step | `trigger: manual` | `environment:` (approval gates) | `when: manual` |
| PR write token | App Password (repository variable) | `secrets.GITHUB_TOKEN` (built-in) | `GITLAB_TOKEN` (variable) |
| Target branch env var | `BITBUCKET_PR_DESTINATION_BRANCH` | `GITHUB_BASE_REF` | `CI_MERGE_REQUEST_TARGET_BRANCH_NAME` |

---

## Execution flow on Pull Request

When a PR is opened or updated:

1. `Doctor` runs first.
2. `Plan` runs only after `Doctor` succeeds.
3. Plan output is saved as a pipeline artifact.
4. A PR comment is posted with the plan output (requires `BITBUCKET_APP_PASSWORD`).
5. `Apply` is not triggered — it runs only on the `main` branch after merge.

When a PR is merged into `main`:

1. `Doctor` runs.
2. `Apply` is presented as a manual step in the Bitbucket UI.
3. Once approved, `Apply` runs with `--auto-approve=true`.
4. Apply output is saved as a pipeline artifact.

---

## Secrets and credentials

Store cloud credentials in **Repository settings > Repository variables**:

```
AWS_ACCESS_KEY_ID     = <value>
AWS_SECRET_ACCESS_KEY = <value>
AWS_REGION            = us-east-1
```

Cultivator does not manage credentials; Terragrunt/OpenTofu/Terraform reads them
from the environment.

---

## Troubleshooting

### `cultivator: command not found`

Verify the install step ran successfully and the binary was moved to `/usr/local/bin`.

### `terragrunt: command not found`

Cultivator delegates to Terragrunt. Install both binaries in the same step.

### PR comment step fails with `401 Unauthorized`

Ensure `BITBUCKET_USERNAME` and `BITBUCKET_APP_PASSWORD` are set correctly and
the App Password has the **Pull requests: Write** permission.

### PR comment step fails with `404 Not Found`

Verify `BITBUCKET_REPO_FULL_NAME` and `BITBUCKET_PR_ID` are available. These
variables are only set automatically in `pull-requests` pipeline steps.

### Magic Mode finds no changes

Ensure `clone.depth: full` is set on the Plan step. A shallow clone (the
Bitbucket default) may not contain the full history needed for `git diff`.

### No stacks discovered

Check `CULTIVATOR_ROOT` and optional `CULTIVATOR_ENV`.

---

## Further reading

- [Quickstart](../getting-started/quickstart.md)
- [Configuration](../getting-started/configuration.md)
- [CLI Reference](cli-reference.md)
- [Bitbucket Pipelines documentation](https://support.atlassian.com/bitbucket-cloud/docs/get-started-with-bitbucket-pipelines/)
- [GitHub Actions equivalent](github-actions.md)
- [GitLab CI equivalent](gitlab-pipelines.md)
