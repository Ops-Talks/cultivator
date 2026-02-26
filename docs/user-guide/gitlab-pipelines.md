# GitLab Pipelines Integration

This guide shows how to integrate Cultivator with GitLab CI/CD pipelines.

> **Note**: By default, this configuration uses **OpenTofu**, an open-source fork of Terraform that maintains full compatibility. If you prefer HashiCorp Terraform, simply set `IAC_TOOL: "terraform"` in the variables section.

## Setup

### Option 1: Using Docker Image (Recommended)

The easiest way to use Cultivator in GitLab CI is via Docker image.

#### Step 1: Create `.gitlab-ci.yml`

Create `.gitlab-ci.yml` in your repository root:

```yaml
stages:
  - plan
  - apply

variables:
  OPENTOFU_VERSION: "1.7.0"
  TERRAGRUNT_VERSION: "0.55.0"
  # Set to 'terraform' if you prefer HashiCorp Terraform instead of OpenTofu
  IAC_TOOL: "opentofu"

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
    - |
      if [ "$IAC_TOOL" = "terraform" ]; then
        wget -q https://releases.hashicorp.com/terraform/${OPENTOFU_VERSION}/terraform_${OPENTOFU_VERSION}_linux_amd64.zip
        unzip -q terraform_${OPENTOFU_VERSION}_linux_amd64.zip -d /usr/local/bin && rm terraform_${OPENTOFU_VERSION}_linux_amd64.zip
      else
        wget -q https://github.com/opentofu/opentofu/releases/download/v${OPENTOFU_VERSION}/tofu_${OPENTOFU_VERSION}_linux_amd64.zip
        unzip -q tofu_${OPENTOFU_VERSION}_linux_amd64.zip -d /usr/local/bin && rm tofu_${OPENTOFU_VERSION}_linux_amd64.zip
      fi
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
    - |
      if [ "$IAC_TOOL" = "terraform" ]; then
        wget -q https://releases.hashicorp.com/terraform/${OPENTOFU_VERSION}/terraform_${OPENTOFU_VERSION}_linux_amd64.zip
        unzip -q terraform_${OPENTOFU_VERSION}_linux_amd64.zip -d /usr/local/bin && rm terraform_${OPENTOFU_VERSION}_linux_amd64.zip
      else
        wget -q https://github.com/opentofu/opentofu/releases/download/v${OPENTOFU_VERSION}/tofu_${OPENTOFU_VERSION}_linux_amd64.zip
        unzip -q tofu_${OPENTOFU_VERSION}_linux_amd64.zip -d /usr/local/bin && rm tofu_${OPENTOFU_VERSION}_linux_amd64.zip
      fi
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

### Caching OpenTofu/Terraform and Terragrunt

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

### OpenTofu/Terraform not found

- Verify `before_script` is running correctly
- Check Runner logs for installation errors
- Try using Docker image with pre-installed tools

## Next Steps

- [Workflows](workflows.md) - Learn about Cultivator commands
- [Configuration](../getting-started/configuration.md) - Configure Cultivator behavior
- [GitLab CI Docs](https://docs.gitlab.com/ee/ci/) - GitLab CI/CD documentation
