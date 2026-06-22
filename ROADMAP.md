# Roadmap

This document tracks planned work for the Cultivator project. Items are
organized as **Epic > History > Task > Sub-task**.

Conventions:

- Each Epic represents a high-level outcome (a bug fix, a feature area, or a
  documentation effort).
- Each History (user story) describes the value delivered from a user or
  operator perspective.
- Each Task is an actionable unit of work scoped to a single concern.
- Each Sub-task is a concrete step that can be reviewed and tested in isolation.

Status legend: `[ ]` not started, `[~]` in progress, `[x]` done.

---

## Epic 1: Fix `--changed-only` (Magic Mode) over-matching modules

**Context.** When `--changed-only` is enabled, Cultivator detects the set of
files changed against the configured base ref and selects only the affected
Terragrunt stacks. A regression in the ancestor heuristic of
`moduleHasChanges` (see `internal/cli/cli.go`) causes any change to a file at
the repository root (for example `bitbucket-pipelines.yml`) to mark every
discovered module as changed. The result is that all stacks are executed,
defeating the purpose of Magic Mode and producing dangerous full-stack runs in
CI pipelines.

**Goal.** Restore correct behavior so that:

1. Changes outside the configured `--root` never trigger module selection.
2. The shared-config heuristic (a change to `root.hcl` or `environment.hcl`
   propagating to descendants) keeps working for files inside the root.
3. False positives caused by naive string-prefix checks are eliminated.

### History 1.1: As a developer, only the stacks I actually changed should run when I use `--changed-only`

#### Task 1.1.1: Refactor `moduleHasChanges` to scope the ancestor heuristic to the configured root

- [x] Sub-task 1.1.1.1: Extend the `moduleHasChanges` signature to accept the
  absolute, cleaned root path (`rootPath string`) in addition to the module
  path and the changed file list.
- [x] Sub-task 1.1.1.2: Replace the unconditional ancestor check with a guard
  that skips files whose cleaned path does not live inside `rootPath`. Use a
  separator-aware check (equality with `rootPath`, or prefix
  `rootPath + string(os.PathSeparator)`).
- [x] Sub-task 1.1.1.3: Replace the first `strings.HasPrefix(changedPath, modPath)`
  check with a separator-aware comparison so that paths like
  `/repo/providers/aws-extra/...` do not match `/repo/providers/aws`.
- [x] Sub-task 1.1.1.4: Preserve the existing behavior for shared parent
  configs (`root.hcl`, `environment.hcl`) located inside the configured root.

#### Task 1.1.2: Wire the configured root into the filter pipeline

- [x] Sub-task 1.1.2.1: In `filterChangedModules`, compute the absolute and
  cleaned form of `cfg.Root` once and pass it to `moduleHasChanges`.
- [x] Sub-task 1.1.2.2: Verify that `cfg.Root` is already normalized upstream;
  if not, perform the normalization at a single, well-documented call site.
- [x] Sub-task 1.1.2.3: Add a debug log line emitting the resolved root used
  for change filtering to ease future troubleshooting.

#### Task 1.1.3: Strengthen unit tests for `moduleHasChanges`

- [x] Sub-task 1.1.3.1: Update the existing `Test_ModuleHasChanges` table to
  pass the new `rootPath` parameter.
- [x] Sub-task 1.1.3.2: Add a regression case: change at the repository root
  (`/repo/bitbucket-pipelines.yml`) with `rootPath = /repo/providers` must
  return `false`.
- [x] Sub-task 1.1.3.3: Add a case for a sibling tree inside the root that
  does not contain the module (must return `false`).
- [x] Sub-task 1.1.3.4: Add a case for a path-prefix collision
  (`/repo/providers/aws-extra/file` vs `/repo/providers/aws`) verifying it
  returns `false`.
- [x] Sub-task 1.1.3.5: Keep coverage for the positive cases: file inside the
  module, parent `root.hcl`, and parent `environment.hcl`.

#### Task 1.1.4: Validate end-to-end behavior

- [x] Sub-task 1.1.4.1: Run `go test ./internal/cli/... ./internal/git/...`
  and confirm all suites pass.
- [x] Sub-task 1.1.4.2: Run `go vet ./...` and `golangci-lint run` (if
  configured locally) without new findings.
- [ ] Sub-task 1.1.4.3: Build the binary (`go build ./cmd/cultivator`) and
  exercise `cultivator plan --changed-only` on a repository where only a
  subset of modules is changed; confirm that unrelated stacks are skipped.

---

## Epic 2: Bitbucket Pipelines auto-fetch for base branch resolution

**Context.** Bitbucket Pipelines does not create remote-tracking refs for the
PR destination branch. When Cultivator runs `git diff <base>` it iterates the
candidates `base`, `origin/base`, and `refs/remotes/origin/base`. If none of
these resolve, `GetChangedFiles` fails and the CLI exits with code `1`. The
current failure mode is non-obvious and forces users to add a manual
`git fetch origin "$BITBUCKET_PR_DESTINATION_BRANCH"` step before invoking
Cultivator.

**Goal.** When Cultivator detects it is running inside a Bitbucket Pipelines
pull-request build and the base ref is not locally resolvable, Cultivator
should fetch the destination branch automatically and retry. The fetch must
be opt-out-safe, side-effect-aware, and produce clear logs.

### History 2.1: As a Bitbucket Pipelines user, I want Cultivator to recover from missing remote refs without manual fetch steps

#### Task 2.1.1: Detect when a base ref cannot be resolved

- [x] Sub-task 2.1.1.1: In `internal/git/git.go`, surface a distinct error
  (for example `ErrBaseRefNotFound`) when all candidate refs returned by
  `buildBaseRefCandidates` fail to resolve, so callers can react
  programmatically.
- [x] Sub-task 2.1.1.2: Keep the existing aggregate error message for
  unrelated git failures (network, permissions) so they are not silently
  retried.

#### Task 2.1.2: Implement scoped auto-fetch for Bitbucket PR builds

- [x] Sub-task 2.1.2.1: Add a helper `tryFetchBaseRef(ctx, workingDir, remote,
  branch, logger)` in the `git` package that wraps `git fetch --no-tags
  --prune <remote> <branch>` and returns a typed error on failure.
- [x] Sub-task 2.1.2.2: In the CLI flow (`filterChangedModules` or a small
  wrapper), when `ci.Detect()` reports `ProviderBitbucket` with `IsPR=true`
  and `GetChangedFiles` returns `ErrBaseRefNotFound`, invoke
  `tryFetchBaseRef` for `origin` and `BITBUCKET_PR_DESTINATION_BRANCH`,
  then retry `GetChangedFiles` exactly once.
- [x] Sub-task 2.1.2.3: Emit informative logs around the auto-fetch attempt
  (`Info` on success, `Warn` on fetch failure including the suggested manual
  command). Never swallow the original error if the retry still fails.
- [x] Sub-task 2.1.2.4: Guard the behavior so it only activates for the
  Bitbucket CI provider; do not auto-fetch on local runs, GitHub Actions, or
  GitLab CI where the ref is already available.

#### Task 2.1.3: Make the behavior controllable

- [x] Sub-task 2.1.3.1: Add a boolean configuration flag
  `--no-auto-fetch` (and matching `cfg.NoAutoFetch`) that disables the
  auto-fetch path even when the conditions are met.
- [x] Sub-task 2.1.3.2: Document the default (`auto-fetch enabled in
  Bitbucket PR builds`) and the opt-out flag.

#### Task 2.1.4: Tests for the auto-fetch behavior

- [x] Sub-task 2.1.4.1: Add unit tests for `tryFetchBaseRef` using a
  temporary local git repository with a fake remote (or by stubbing
  `exec.CommandContext` via an injection seam).
- [x] Sub-task 2.1.4.2: Add a CLI-level test that simulates the Bitbucket
  environment (`BITBUCKET_BUILD_NUMBER`, `BITBUCKET_PR_ID`,
  `BITBUCKET_PR_DESTINATION_BRANCH`) and verifies the auto-fetch path is
  taken when the base ref is missing.
- [x] Sub-task 2.1.4.3: Add a negative test that ensures the auto-fetch is
  skipped when `--no-auto-fetch` is set or when the provider is not
  Bitbucket.

#### Task 2.1.5: Improve error message when auto-fetch is impossible

- [x] Sub-task 2.1.5.1: When auto-fetch fails or is disabled and the base
  ref is still unresolvable, return an error that names the missing ref and
  prints the exact manual command the user should run.
- [x] Sub-task 2.1.5.2: Ensure the CLI exits with a non-zero code in that
  scenario and never falls back to running all stacks.

---

## Epic 3: Documentation alignment

**Context.** The behavioral changes from Epics 1 and 2 must be reflected in
the user-facing documentation so operators understand the new guarantees and
configuration knobs.

### History 3.1: As an operator, I want the docs to describe the corrected Magic Mode semantics and the Bitbucket auto-fetch behavior

#### Task 3.1.1: Update Magic Mode documentation

- [x] Sub-task 3.1.1.1: In `docs/user-guide/features.md`, clarify that only
  files inside `--root` participate in module selection. Mention that shared
  parent configs (`root.hcl`, `environment.hcl`) propagate to descendants
  inside the same root.
- [x] Sub-task 3.1.1.2: In `docs/index.md`, refresh the Magic Mode bullet so
  it reflects the corrected scoping (changes outside the root do not trigger
  full-stack execution).
- [x] Sub-task 3.1.1.3: In `docs/getting-started/quickstart.md`, audit any
  example using `--changed-only` and confirm the wording matches the new
  guarantees.

#### Task 3.1.2: Update Bitbucket Pipelines guide

- [x] Sub-task 3.1.2.1: In `docs/user-guide/bitbucket-pipelines.md`, document
  the auto-fetch behavior for PR builds (when it triggers, what command it
  runs, and how to opt out with `--no-auto-fetch`).
- [x] Sub-task 3.1.2.2: Update the "Magic Mode finds no changes"
  troubleshooting entry to reflect that `clone.depth: full` is still
  recommended, but Cultivator will now attempt to fetch the destination
  branch on its own when needed.
- [x] Sub-task 3.1.2.3: Replace any wording that suggests a silent fallback
  to running all stacks with an explicit statement that Cultivator exits
  non-zero when the base ref cannot be resolved.

#### Task 3.1.3: Update CLI reference

- [x] Sub-task 3.1.3.1: In `docs/user-guide/cli-reference.md`, list the new
  `--no-auto-fetch` flag with its default, scope, and behavior.
- [x] Sub-task 3.1.3.2: Ensure the `--changed-only` and `--base` entries
  describe how root scoping interacts with file selection.

#### Task 3.1.4: Update architecture and FAQ entries

- [x] Sub-task 3.1.4.1: In `docs/architecture/design.md`, add or update the
  section that describes change-set computation so it documents the
  root-scoped ancestor heuristic and the Bitbucket auto-fetch hook.
- [x] Sub-task 3.1.4.2: In `docs/faq.md`, add an entry: "Why did
  `--changed-only` run every stack in a previous version?" with a concise
  explanation and the fix reference.

#### Task 3.1.5: Update README and examples

- [x] Sub-task 3.1.5.1: Confirm the `README.md` Magic Mode bullet is still
  accurate; adjust wording if any prior phrasing implied that root-level
  changes affected discovery.
- [x] Sub-task 3.1.5.2: Review `examples/bitbucket-pipelines.yml` and remove
  the explicit `git fetch` workaround (or keep it as a commented hint with a
  reference to the auto-fetch behavior).

---

## Epic 4: Align Bitbucket Pipelines documentation with the official reference

**Context.** A review of the current Bitbucket Pipelines guide
(`docs/user-guide/bitbucket-pipelines.md`) and the reference example
(`examples/bitbucket-pipelines.yml`) against the official Atlassian
documentation (Bitbucket Pipelines configuration reference, Git clone
behavior, Pipeline start conditions, and Variables and secrets) surfaced
several discrepancies. The most impactful ones are: a YAML anchor declared
under an unsupported `definitions.steps` key and never referenced (each step
duplicates the install commands instead); missing documentation on the
implicit destination-branch merge performed by `pull-requests` pipelines;
missing notes on fork PR behavior, default clone depth, parallel execution
with `branches`/`default`, and the modern `triggers.pullrequest-fulfilled`
hook; and an incomplete environment-variable table.

**Goal.** Make the Bitbucket guide and example match the official reference
exactly, eliminate duplicated install scripts via valid YAML anchors, and
explain the PR pipeline behaviors that affect Cultivator users in production.

### History 4.1: As a Bitbucket Pipelines user, the Cultivator docs and example reflect the official Bitbucket reference accurately

#### Task 4.1.1: Fix the YAML anchor usage in `examples/bitbucket-pipelines.yml`

- [x] Sub-task 4.1.1.1: Remove the unsupported `definitions.steps` block.
  `definitions` only accepts `caches`, `services`, `pipelines` (imports),
  and `exports` according to the official reference; arbitrary `steps`
  children are not part of the schema.
- [x] Sub-task 4.1.1.2: Declare the install logic as a top-level YAML anchor
  (for example a `install-tools: &install-tools` mapping with a `script`
  key) per the Bitbucket "YAML anchors" page, and reuse it from every step
  that needs the tools (Doctor, Plan, Apply) via `<<: *install-tools` or by
  expanding the anchored `script` list.
- [x] Sub-task 4.1.1.3: Ensure the anchored script remains compatible with
  the `alpine:3.21` image and with the `clone.depth: full` requirement on
  steps that consume Magic Mode.

#### Task 4.1.2: Document Bitbucket `pull-requests` pipeline semantics

- [x] Sub-task 4.1.2.1: In `docs/user-guide/bitbucket-pipelines.md`, add a
  short section explaining that `pull-requests` pipelines merge the
  destination branch into the source branch before running the script, and
  document the impact on `git diff` and on Cultivator's Magic Mode.
- [x] Sub-task 4.1.2.2: Document that glob patterns under `pull-requests:`
  match the **source branch** of the pull request, not the destination.
- [x] Sub-task 4.1.2.3: Document that pull requests from forked repositories
  do not trigger the pipeline (Atlassian limitation).
- [x] Sub-task 4.1.2.4: Document that `pull-requests` pipelines run **in
  addition to** any matching `branches` or `default` pipelines, which can
  cause two pipelines to start in parallel for the same event; recommend
  scoping `branches:` to `main` (or equivalent) to avoid double runs.

#### Task 4.1.3: Cover the `triggers` property as a modern alternative

- [x] Sub-task 4.1.3.1: Add a sub-section describing the `triggers` property
  with `pullrequest-push`, `pullrequest-fulfilled`, and
  `pullrequest-rejected` events as a modern alternative to the legacy
  `pull-requests` / `branches` selectors.
- [x] Sub-task 4.1.3.2: Provide an example that uses
  `triggers.pullrequest-fulfilled` to run `cultivator apply` automatically
  after a PR is merged, paired with a deployment gate.
- [x] Sub-task 4.1.3.3: Call out the known limitations of triggers:
  custom-pipeline-only references, the 20-conditions-per-type and
  100-pipelines-per-event caps.

#### Task 4.1.4: Correct the `clone.depth` reference

- [x] Sub-task 4.1.4.1: In the troubleshooting "Magic Mode finds no changes"
  entry, update the wording to reflect that the default `clone.depth` is
  `50` commits (not "shallow"), and document `full` as the recommended
  value for change-set computation.
- [x] Sub-task 4.1.4.2: Cross-reference the auto-fetch behavior introduced
  in Epic 2 so users understand that `depth: full` plus auto-fetch is the
  combination that yields reliable Magic Mode results.

#### Task 4.1.5: Expand the environment-variable table

- [x] Sub-task 4.1.5.1: Add the following entries with descriptions matching
  the official reference: `BITBUCKET_WORKSPACE`, `BITBUCKET_REPO_SLUG`,
  `BITBUCKET_CLONE_DIR`, `BITBUCKET_PR_DESTINATION_COMMIT`,
  `BITBUCKET_STEP_OIDC_TOKEN`.
- [x] Sub-task 4.1.5.2: Add a note clarifying that `BITBUCKET_BRANCH` in a
  PR build is the source branch, not the destination, and that the
  destination is exposed via `BITBUCKET_PR_DESTINATION_BRANCH` /
  `BITBUCKET_PR_DESTINATION_COMMIT`.

#### Task 4.1.6: Recommend OIDC for cloud credentials

- [x] Sub-task 4.1.6.1: In the "Secrets and credentials" section, add a
  callout pointing readers to `BITBUCKET_STEP_OIDC_TOKEN` and the official
  Bitbucket OIDC integration page as the preferred alternative to storing
  long-lived `AWS_ACCESS_KEY_ID` / `AWS_SECRET_ACCESS_KEY` as repository
  variables.
- [x] Sub-task 4.1.6.2: Keep the static-credential path as a fallback, but
  mark it explicitly as the less-secure option.

#### Task 4.1.7: Validate the example against the official linter

- [ ] Sub-task 4.1.7.1: After the example refactor, validate
  `examples/bitbucket-pipelines.yml` using the
  [Bitbucket Pipelines validator](https://bitbucket-pipelines.prod.public.atl-paas.net/validator)
  (or the inline Atlassian YAML schema) and confirm it parses cleanly.
- [ ] Sub-task 4.1.7.2: Run `yamllint` against the example using the
  repository's `.yamllint.yaml` to make sure the file matches project
  conventions.

---

## Epic 5: Release readiness

### History 5.1: As a maintainer, I want a clean validation pass before tagging the fix

#### Task 5.1.1: Local validation

- [x] Sub-task 5.1.1.1: Run `make test` (or the equivalent `go test ./...`)
  and confirm the full suite is green.
- [x] Sub-task 5.1.1.2: Run `go vet ./...` and `golangci-lint run`.
- [x] Sub-task 5.1.1.3: Build the binary and execute `cultivator doctor` to
  confirm baseline checks still pass.

#### Task 5.1.2: Manual smoke test in a real Terragrunt layout

- [ ] Sub-task 5.1.2.1: Using a representative repository (for example the
  `testdata/terragrunt-large` layout), simulate a PR that touches three
  stacks plus one file at the repo root. Confirm only the three stacks are
  selected.
- [ ] Sub-task 5.1.2.2: Simulate a Bitbucket PR environment by exporting
  `BITBUCKET_BUILD_NUMBER`, `BITBUCKET_PR_ID`, and
  `BITBUCKET_PR_DESTINATION_BRANCH`, and confirm the auto-fetch path runs
  when the destination branch is not already fetched.

#### Task 5.1.3: Issue closure

- [ ] Sub-task 5.1.3.1: Reference the fixed behavior in the PR description
  and link the GitHub issue.
- [ ] Sub-task 5.1.3.2: After merge, verify the documentation site builds
  cleanly (`mkdocs build`) and the published version reflects the updates.
