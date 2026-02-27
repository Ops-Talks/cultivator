# GitLab CI/CD Integration

This guide shows how to run Cultivator inside GitLab CI/CD pipelines.

Cultivator is a CLI binary that orchestrates Terragrunt. To use it in a pipeline you need three things in the same job environment:

1. The `cultivator` binary
2. The `terragrunt` binary
3. `terraform` (or OpenTofu)

---

## Minimal example

```yaml
# .gitlab-ci.yml

stages:
  - validate
  - plan
  - apply

variables:
  CULTIVATOR_VERSION: "v0.2.0"
  TERRAGRUNT_VERSION: "0.67.0"
  TERRAFORM_VERSION: "1.10.0"
  # Cultivator settings
  CULTIVATOR_ROOT: "infrastructure"
  CULTIVATOR_PARALLELISM: "4"
  CULTIVATOR_OUTPUT_FORMAT: "json"

.install_tools: &install_tools
  before_script:
    # Install Terraform
    - wget -q https://releases.hashicorp.com/terraform/${TERRAFORM_VERSION}/terraform_${TERRAFORM_VERSION}_linux_amd64.zip
    - unzip -q terraform_${TERRAFORM_VERSION}_linux_amd64.zip -d /usr/local/bin
    - rm terraform_${TERRAFORM_VERSION}_linux_amd64.zip
    # Install Terragrunt
    - wget -q -O /usr/local/bin/terragrunt
        https://github.com/gruntwork-io/terragrunt/releases/download/v${TERRAGRUNT_VERSION}/terragrunt_linux_amd64
    - chmod +x /usr/local/bin/terragrunt
    # Install Cultivator
    - wget -q -O /usr/local/bin/cultivator
        https://github.com/Ops-Talks/cultivator/releases/download/${CULTIVATOR_VERSION}/cultivator-linux-amd64
    - chmod +x /usr/local/bin/cultivator
    # Verify
    - cultivator doctor

doctor:
  stage: validate
  image: alpine:3.21
  <<: *install_tools
  script:
    - cultivator doctor
  rules:
    - if: '$CI_PIPELINE_SOURCE == "merge_request_event"'
    - if: '$CI_COMMIT_BRANCH == "main"'

plan:
  stage: plan
  image: alpine:3.21
  <<: *install_tools
  script:
    - cultivator plan
        --root "$CULTIVATOR_ROOT"
        --env "$CI_ENVIRONMENT_NAME"
        --parallelism "$CULTIVATOR_PARALLELISM"
        --output-format "$CULTIVATOR_OUTPUT_FORMAT"
        --non-interactive
  rules:
    - if: '$CI_PIPELINE_SOURCE == "merge_request_event"'

apply:
  stage: apply
  image: alpine:3.21
  <<: *install_tools
  script:
    - cultivator apply
        --root "$CULTIVATOR_ROOT"
        --env "$CI_ENVIRONMENT_NAME"
        --parallelism "$CULTIVATOR_PARALLELISM"
        --output-format "$CULTIVATOR_OUTPUT_FORMAT"
        --non-interactive
        --auto-approve
  rules:
    - if: '$CI_COMMIT_BRANCH == "main"'
      when: manual
  environment:
    name: production
```

---

## Using a config file

Instead of passing all flags on the command line, you can commit a `.cultivator.yaml` at the repo root:

```yaml
# .cultivator.yaml
root: infrastructure
parallelism: 4
output_format: json
non_interactive: true
```

The pipeline script then becomes simply:

```yaml
script:
  - cultivator plan --env "$CI_ENVIRONMENT_NAME"
```

See [Configuration](../getting-started/configuration.md) for all available options.

---

## Environment-based deployments

Use GitLab environments to drive which Terragrunt stack is targeted:

```yaml
stages:
  - plan
  - apply

.cultivator_base:
  image: alpine:3.21
  before_script:
    - apk add --no-cache wget unzip
    - *install_tools  # see minimal example above

plan:dev:
  extends: .cultivator_base
  stage: plan
  script:
    - cultivator plan --root infrastructure --env dev --non-interactive
  environment:
    name: dev
  rules:
    - if: '$CI_PIPELINE_SOURCE == "merge_request_event"'

plan:prod:
  extends: .cultivator_base
  stage: plan
  script:
    - cultivator plan --root infrastructure --env prod --non-interactive
  environment:
    name: prod
  rules:
    - if: '$CI_COMMIT_BRANCH == "main"'

apply:prod:
  extends: .cultivator_base
  stage: apply
  script:
    - cultivator apply --root infrastructure --env prod --non-interactive --auto-approve
  environment:
    name: prod
  rules:
    - if: '$CI_COMMIT_BRANCH == "main"'
      when: manual
  needs:
    - plan:prod
```

---

## Tag-based filtering

Run only modules that carry a specific tag (defined via `# cultivator:tags=` comments in `terragrunt.hcl`):

```yaml
plan:networking:
  stage: plan
  script:
    - cultivator plan --root infrastructure --env prod --tags networking --non-interactive
```

---

## Caching

Cache the Terragrunt plugin directory to speed up runs:

```yaml
cache:
  key: terragrunt-$CI_COMMIT_REF_SLUG
  paths:
    - .terragrunt-cache/
```

---

## Cloud credentials

Add credentials as CI/CD variables (**Settings → CI/CD → Variables**) and expose them as environment variables in the job:

```yaml
plan:
  variables:
    AWS_ACCESS_KEY_ID: $AWS_ACCESS_KEY_ID
    AWS_SECRET_ACCESS_KEY: $AWS_SECRET_ACCESS_KEY
    AWS_DEFAULT_REGION: us-east-1
```

Terragrunt (called by Cultivator) will pick them up automatically.

---

## Troubleshooting

### `cultivator: command not found`
The binary was not installed or `/usr/local/bin` is not in `PATH`. Run `cultivator doctor` as part of `before_script` to catch this early.

### `terragrunt: command not found`
Cultivator delegates execution to Terragrunt. Both binaries must be present in the same job. `cultivator doctor` will report if Terragrunt is missing.

### No modules discovered
Check that `--root` points to the directory containing your `terragrunt.hcl` files and that `--env` matches the subdirectory structure.

---

## Further reading

- [Quickstart](../getting-started/quickstart.md) — local usage
- [Configuration](../getting-started/configuration.md) — all options and precedence
- [Workflows](workflows.md) — available commands and flags
- [GitLab CI/CD documentation](https://docs.gitlab.com/ee/ci/)


The easiest way to use Cultivator in GitLab CI is via Docker image.

#### Step 1: Create `.gitlab-ci.yml`

Create `.gitlab-ci.yml` in your repository root:

```yaml
stages:
  - plan
  - apply

variables:
  TERRAFORM_VERSION: "1.7.0"
  TERRAGRUNT_VERSION: "0.55.0"

workflow:
  rules:
    - if: '$CI_PIPELINE_SOURCE == "merge_request_event"'
    - if: '$CI_COMMIT_BRANCH && $CI_OPEN_MRS'
      when: never
    - if: '$CI_COMMIT_BRANCH'

cultivator_plan:
  stage: plan
  image: ubuntu:latest
  rules:
    - if: '$CI_PIPELINE_SOURCE == "merge_request_event"'
  before_script:
    - apt-get update && apt-get install -y wget unzip git curl
    - wget -q https://releases.hashicorp.com/terraform/${TERRAFORM_VERSION}/terraform_${TERRAFORM_VERSION}_linux_amd64.zip
    - unzip -q terraform_${TERRAFORM_VERSION}_linux_amd64.zip -d /usr/local/bin && rm terraform_${TERRAFORM_VERSION}_linux_amd64.zip
    - wget -q https://github.com/gruntwork-io/terragrunt/releases/download/v${TERRAGRUNT_VERSION}/terragrunt_linux_amd64
    - chmod +x terragrunt_linux_amd64 && mv terragrunt_linux_amd64 /usr/local/bin/terragrunt
  script:
    - terragrunt run-all plan --terragrunt-non-interactive
  artifacts:
    paths:
      - .terraform/
      - "**/*.tfplan"
    expire_in: 1 day
  only:
    - merge_requests

cultivator_apply:
  stage: apply
  image: ubuntu:latest
  rules:
    - if: '$CI_PIPELINE_SOURCE == "merge_request_event" && $CI_MERGE_REQUEST_APPROVED == "true"'
      when: manual
  before_script:
    - apt-get update && apt-get install -y wget unzip git curl
    - wget -q https://releases.hashicorp.com/terraform/${TERRAFORM_VERSION}/terraform_${TERRAFORM_VERSION}_linux_amd64.zip
    - unzip -q terraform_${TERRAFORM_VERSION}_linux_amd64.zip -d /usr/local/bin && rm terraform_${TERRAFORM_VERSION}_linux_amd64.zip
    - wget -q https://github.com/gruntwork-io/terragrunt/releases/download/v${TERRAGRUNT_VERSION}/terragrunt_linux_amd64
    - chmod +x terragrunt_linux_amd64 && mv terragrunt_linux_amd64 /usr/local/bin/terragrunt
  script:
    - terragrunt run-all apply --terragrunt-non-interactive
  only:
    - merge_requests
  dependencies:
    - cultivator_plan
```

#### Step 2: Configure Project Settings

1. Go to **Settings → CI/CD → Variables**
2. Add required environment variables:
   - `AWS_ACCESS_KEY_ID` (if using AWS)
   - `AWS_SECRET_ACCESS_KEY` (if using AWS)
   - `AZURE_CLIENT_ID` (if using Azure)
   - `AZURE_CLIENT_SECRET` (if using Azure)
   - `GCP_PROJECT_ID` (if using GCP)
   - `GCP_SA_KEY` (if using GCP)

#### Step 3: Configure Merge Request Rules

1. Go to **Settings → Merge requests**
2. Enable **Require approval on merge requests**
3. Set **Approval rules** if needed

### Option 2: Using Cultivator Docker Container

If you want to use pre-built Cultivator image:

```yaml
stages:
  - plan
  - apply

cultivator_plan:
  stage: plan
  image: cultivator:latest
  script:
    - cultivator plan
  artifacts:
    reports:
      dotenv: build.env
  only:
    - merge_requests

cultivator_apply:
  stage: apply
  image: cultivator:latest
  script:
    - cultivator apply
  when: manual
  only:
    - merge_requests
```

## Variables

Common GitLab CI variables you might use:

| Variable | Description |
|----------|-------------|
| `CI_COMMIT_SHA` | Commit SHA |
| `CI_COMMIT_BRANCH` | Branch name |
| `CI_MERGE_REQUEST_IID` | MR number |
| `CI_MERGE_REQUEST_APPROVED` | MR approval status |
| `CI_PROJECT_PATH` | Project path |
| `CI_PIPELINE_ID` | Pipeline ID |

## Artifacts & Caching

### Caching Terraform/Terragrunt

```yaml
cache:
  paths:
    - .terraform/
    - .terragrunt-cache/
    - .cache/
```

### Storing Plan Output

```yaml
artifacts:
  paths:
    - "**/*.tfplan"
    - "**/*.json"
  expire_in: 7 days
  reports:
    dotenv: plan.env
```

## Manual Pipeline Triggers

To trigger Cultivator manually on demand:

```yaml
cultivator_manual:
  stage: plan
  script:
    - cultivator plan
  when: manual
  only:
    - merge_requests
```

## Conditional Execution

Run Cultivator only on specific folder changes:

```yaml
cultivator_plan:
  stage: plan
  script:
    - cultivator plan
  only:
    - merge_requests
  changes:
    - infrastructure/**/*.tf
    - infrastructure/**/*.hcl
    - terragrunt.hcl
```

## GitLab Merge Request Integration

Cultivator will automatically post plan/apply results as MR comments:

```yaml
script:
  - |
    PLAN_OUTPUT=$(terragrunt run-all plan --terragrunt-non-interactive 2>&1)
    curl -X POST \
      "https://gitlab.com/api/v4/projects/$CI_PROJECT_ID/merge_requests/$CI_MERGE_REQUEST_IID/notes" \
      -H "PRIVATE-TOKEN: $CI_JOB_TOKEN" \
      -d "body=$PLAN_OUTPUT"
```

## Troubleshooting

### Pipeline doesn't trigger

- Check workflow rules in `.gitlab-ci.yml`
- Verify branch is not protected from CI
- Check **Settings → CI/CD → Pipeline** for errors

### Permission denied errors

- Verify Runner has access to AWS/Azure/GCP credentials
- Check **Settings → CI/CD → Variables** for correct secrets
- Ensure service account has required permissions

### Terraform/Terragrunt not found

- Verify `before_script` is running correctly
- Check Runner logs for installation errors
- Try using Docker image with pre-installed tools

## Next Steps

- [Workflows](workflows.md) - Learn about Cultivator commands
- [Configuration](../getting-started/configuration.md) - Configure Cultivator behavior
- [GitLab CI Docs](https://docs.gitlab.com/ee/ci/) - GitLab CI/CD documentation
