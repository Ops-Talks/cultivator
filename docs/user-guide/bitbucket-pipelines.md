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
- `pipelines.branches` for branch-triggered pipelines (for example after merge to `main`)
- `trigger: manual` for steps requiring human approval
- `deployment:` for environment promotion gates
- PR comments posted via the Bitbucket REST API using a Repository Access Token
  (App Passwords are deprecated by Atlassian)

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
| `REPO_ACCESS_TOKEN` | `ATCTT3xFfGN0...` | Repository Access Token with the `pullrequest` write scope. Used as a Bearer token to post PR comments. Mark the variable as **Secured**. |

Cloud credentials (for example `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`)
should also be stored as repository variables. Prefer OIDC where the cloud
provider supports it; see [Secrets and credentials](#secrets-and-credentials).

### Creating a Repository Access Token

Bitbucket Cloud is deprecating App Passwords. The replacement for CI-style
automation is a **Repository Access Token** scoped to a single repository:

1. Open **Repository settings > Access tokens** for the repository.
2. Click **Create Repository Access Token**.
3. Grant the `pullrequest` write scope (this implicitly includes
   `pullrequest:read` and `repository:read`).
4. Optionally set an expiry date and enable token rotation.
5. Copy the token (it is shown only once) and store it as a **Secured**
   repository variable named `REPO_ACCESS_TOKEN`.

Authenticate API calls with `Authorization: Bearer $REPO_ACCESS_TOKEN`. No
username is required; the token is tied to the repository, not to a user.

For user-level scripting use cases outside Pipelines, prefer an API token over
an App Password (see Atlassian's
[API tokens documentation](https://support.atlassian.com/bitbucket-cloud/docs/api-tokens/)).

---

## Recommended pipeline

The structure below is the minimal skeleton. The full reference, including the
install script that downloads OpenTofu, Terragrunt, and Cultivator, lives in
[`examples/bitbucket-pipelines.yml`](https://github.com/Ops-Talks/cultivator/blob/main/examples/bitbucket-pipelines.yml).
Each step references an anchored definition under `definitions.steps` so the
install logic is declared once and reused.

```yaml
# bitbucket-pipelines.yml (skeleton)

image: alpine:3.21

definitions:
  steps:
    - step: &doctor-step
        name: Doctor
        script:
          - # install terragrunt + cultivator (see the reference example)
          - cultivator doctor --root "$CULTIVATOR_ROOT"

    - step: &plan-step
        name: Plan
        clone:
          depth: full          # required for Magic Mode (git diff)
        script:
          - # install opentofu + terragrunt + cultivator
          - |
            cultivator plan \
              --root "$CULTIVATOR_ROOT" \
              --parallelism "$CULTIVATOR_PARALLELISM" \
              --non-interactive=true \
              --changed-only   # auto-detects BITBUCKET_PR_DESTINATION_BRANCH
        artifacts:
          - plan_output.txt

    - step: &apply-step
        name: Apply
        trigger: manual         # human approval gate
        deployment: production
        clone:
          depth: full
        script:
          - # install opentofu + terragrunt + cultivator
          - |
            cultivator apply \
              --root "$CULTIVATOR_ROOT" \
              --parallelism "$CULTIVATOR_PARALLELISM" \
              --non-interactive=true \
              --auto-approve=true

pipelines:
  pull-requests:
    "**":
      - step: *doctor-step
      - step: *plan-step

  branches:
    main:
      - step: *doctor-step
      - step: *apply-step
```

The `branches.main` pipeline is what runs after a PR is merged. With
`trigger: manual` on `*apply-step`, the Apply step pauses until someone clicks
**Run** in the Bitbucket UI - the equivalent of a GitHub environment approval
gate or GitLab `when: manual`.

---

## Magic Mode (changed-only)

When `--changed-only` is active, Cultivator automatically reads the
`BITBUCKET_PR_DESTINATION_BRANCH` environment variable to determine the git
base reference. This means you do **not** need to pass `--base` explicitly in
Bitbucket Pipelines - the right branch is detected automatically.

```bash
# This is sufficient; --base is resolved automatically from the pipeline env.
cultivator plan --changed-only --root providers
```

If you are running outside a PR pipeline (for example a manual branch run),
set `BITBUCKET_PR_DESTINATION_BRANCH` explicitly or pass `--base <branch>` on
the command line.

### Pull-request merge semantics

Bitbucket clones the source branch and **merges the destination branch into it
before running the script**. As a consequence:

- `git diff origin/<destination>` shows only the changes introduced by the
  pull request, not the full source-branch history.
- Pipelines must clone enough history to reach the merge base. The default
  `clone.depth` is **50 commits**; set `clone.depth: full` on every step that
  uses Magic Mode.

### Source-branch globs

Glob patterns under `pull-requests:` match the **source branch** of the pull
request, never the destination. For example, `pull-requests: { 'feature/*': ... }`
fires when a PR is opened from a `feature/*` branch toward any target.

### Forked repositories

Pull requests opened from a forked repository do **not** trigger the
`pull-requests` pipeline. This is an Atlassian platform limitation and cannot
be worked around from Cultivator itself.

### Parallel runs with `branches:` and `default:`

`pull-requests:` runs **in addition to** any matching `branches:` or `default:`
entry for the same event. To avoid two pipelines starting in parallel for one
push, scope `branches:` to the trunk (for example `main`) and rely on
`pull-requests:` for everything else.

### Auto-fetch of the destination branch

When `--changed-only` cannot resolve the base ref locally (a common situation
in Bitbucket because the destination branch is not cloned as a tracking ref),
Cultivator now fetches it automatically:

```bash
git fetch --no-tags origin "$BITBUCKET_PR_DESTINATION_BRANCH"
```

The retry runs at most once per invocation and is enabled only for Bitbucket
Pipelines pull-request builds. Pass `--no-auto-fetch` to opt out (for example
when you want to enforce a manual fetch step earlier in the pipeline).
Cultivator still exits non-zero when the destination ref cannot be resolved
after the fetch attempt; it never silently falls back to running every stack.

### Modern alternative: `triggers.pullrequest-fulfilled`

Custom pipelines can subscribe to specific pull-request events via the
`triggers` property:

```yaml
pipelines:
  custom:
    apply-after-merge:
      - step:
          name: Apply on merge
          trigger: automatic
          script:
            - cultivator apply --root providers --auto-approve=true
  triggers:
    pullrequest-fulfilled:
      - pipeline: apply-after-merge
        target:
          branches:
            include:
              - main
```

`triggers` events fire only for the custom pipeline referenced and require
`destination` matching. Known limits documented by Atlassian: up to 20
conditions per repository and 100 referenced pipelines per event type.

---

## PR comments via the Bitbucket REST API

Bitbucket does not provide a built-in write token equivalent to GitHub's
`GITHUB_TOKEN`. Comments must be posted using the
[Bitbucket REST API](https://developer.atlassian.com/cloud/bitbucket/rest/api-group-pullrequests/)
authenticated with a Repository Access Token (or, for user-level scripts, an
API token). Both are sent as a Bearer token; App Passwords with HTTP Basic
auth are deprecated.

```bash
curl --silent --show-error --fail \
  --header "Authorization: Bearer ${REPO_ACCESS_TOKEN}" \
  --header "Content-Type: application/json" \
  --request POST \
  "https://api.bitbucket.org/2.0/repositories/${BITBUCKET_WORKSPACE}/${BITBUCKET_REPO_SLUG}/pullrequests/${BITBUCKET_PR_ID}/comments" \
  --data "$(jq -n --arg body "## Cultivator Plan\n\n\`\`\`\n$(cat plan_output.txt)\n\`\`\`" '{"content":{"raw":$body}}')"
```

Available Bitbucket pipeline environment variables:

| Variable | Description |
|---|---|
| `BITBUCKET_WORKSPACE` | Workspace slug containing the repository. |
| `BITBUCKET_REPO_SLUG` | Repository slug. |
| `BITBUCKET_REPO_FULL_NAME` | Convenience form: `workspace/repo-slug`. |
| `BITBUCKET_CLONE_DIR` | Absolute path of the directory where the repository is cloned. |
| `BITBUCKET_BRANCH` | In PR builds this is the **source branch**, not the destination. |
| `BITBUCKET_COMMIT` | Commit SHA being built. |
| `BITBUCKET_BUILD_NUMBER` | Monotonic build counter. |
| `BITBUCKET_PR_ID` | Pull request ID (set only in PR builds). |
| `BITBUCKET_PR_DESTINATION_BRANCH` | Destination branch of the PR. |
| `BITBUCKET_PR_DESTINATION_COMMIT` | Tip commit of the destination branch at the time of the build. |
| `BITBUCKET_STEP_OIDC_TOKEN` | Short-lived OIDC token minted for the current step (when OIDC is configured). |

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
| PR write token | Repository Access Token (Bearer; App Passwords deprecated) | `secrets.GITHUB_TOKEN` (built-in) | `GITLAB_TOKEN` (variable) |
| Target branch env var | `BITBUCKET_PR_DESTINATION_BRANCH` | `GITHUB_BASE_REF` | `CI_MERGE_REQUEST_TARGET_BRANCH_NAME` |

---

## Execution flow on Pull Request

When a PR is opened or updated:

1. `Doctor` runs first.
2. `Plan` runs only after `Doctor` succeeds.
3. Plan output is saved as a pipeline artifact.
4. A PR comment is posted with the plan output (requires `REPO_ACCESS_TOKEN`).
5. `Apply` is not triggered - it runs only on the `main` branch after merge.

When a PR is merged into `main`:

1. `Doctor` runs.
2. `Apply` is presented as a manual step in the Bitbucket UI.
3. Once approved, `Apply` runs with `--auto-approve=true`.
4. Apply output is saved as a pipeline artifact.

---

## Secrets and credentials

### Prefer OIDC for cloud credentials

Bitbucket Pipelines can issue a short-lived OIDC token to each step
(`BITBUCKET_STEP_OIDC_TOKEN`). Configure your cloud provider to trust the
Bitbucket OIDC issuer and exchange the token for temporary credentials at the
beginning of every step. This avoids storing long-lived secrets in repository
variables and is the recommended option for AWS, GCP, and Azure.

See Atlassian's
[OIDC integration documentation](https://support.atlassian.com/bitbucket-cloud/docs/integrate-pipelines-with-resource-servers-using-oidc/)
for end-to-end setup steps.

### Fallback: long-lived credentials

When OIDC is not available, store cloud credentials in **Repository settings >
Repository variables**:

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

Ensure `REPO_ACCESS_TOKEN` is set, marked as **Secured**, and was created with
the `pullrequest` write scope. The token is sent as a Bearer token; do **not**
use HTTP Basic auth with an App Password (Atlassian has deprecated App
Passwords).

### PR comment step fails with `404 Not Found`

Verify `BITBUCKET_WORKSPACE`, `BITBUCKET_REPO_SLUG`, and `BITBUCKET_PR_ID` are
available. `BITBUCKET_PR_ID` is only set automatically in `pull-requests`
pipeline steps.

### Magic Mode finds no changes

Ensure `clone.depth: full` is set on every step that uses Magic Mode. The
default `clone.depth` in Bitbucket Pipelines is **50 commits**, which is often
enough for small PRs but breaks down for long-lived branches.

If the destination branch still cannot be resolved, Cultivator now performs a
single `git fetch --no-tags origin "$BITBUCKET_PR_DESTINATION_BRANCH"` and
retries the diff. The CLI exits non-zero when the ref is still unreachable;
it never silently runs every stack. Pass `--no-auto-fetch` to disable the
retry.

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
