# Installation

Install Cultivator in your repository using GitHub Actions.

## Prerequisites

- **GitHub Organization/Repository** with admin access
- **.github/workflows** directory (will be created if doesn't exist)
- **Terragrunt v0.40.0+** (or latest)
- **Git** (for repository management)

## Installation Steps

### Option 1: Using GitHub Actions (Recommended)

#### Step 1: Create Workflow File

Create `.github/workflows/cultivator.yml`:

```yaml
name: Cultivator

on:
  pull_request:
    types: [opened, synchronize, reopened]
  issue_comment:
    types: [created]

permissions:
  contents: read
  pull-requests: write
  checks: write
  statuses: write

jobs:
  cultivator:
    if: |
      github.event_name == 'pull_request' ||
      (github.event_name == 'issue_comment' && github.event.issue.pull_request)
    
    runs-on: ubuntu-latest
    
    steps:
      - name: Checkout repository
        uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - name: Run Cultivator
        uses: weyderfs/cultivator@v1
        with:
          github-token: ${{ secrets.GITHUB_TOKEN }}
          terragrunt-version: 'latest'
```

#### Step 2: Commit and Push

```bash
git add .github/workflows/cultivator.yml
git commit -m "Add Cultivator GitHub Action"
git push origin main
```

### Option 2: GitLab CI/CD

If using GitLab, create `.gitlab-ci.yml` in your repository:

```yaml
stages:
  - plan
  - apply

variables:
  TERRAFORM_VERSION: "1.7.0"
  TERRAGRUNT_VERSION: "0.55.0"

cultivator_plan:
  stage: plan
  image: ubuntu:latest
  script:
    - apt-get update && apt-get install -y wget unzip
    - wget -q https://releases.hashicorp.com/terraform/${TERRAFORM_VERSION}/terraform_${TERRAFORM_VERSION}_linux_amd64.zip
    - unzip -q terraform_${TERRAFORM_VERSION}_linux_amd64.zip -d /usr/local/bin
    - wget -q https://github.com/gruntwork-io/terragrunt/releases/download/v${TERRAGRUNT_VERSION}/terragrunt_linux_amd64
    - chmod +x terragrunt_linux_amd64 && mv terragrunt_linux_amd64 /usr/local/bin/terragrunt
    - terragrunt run-all plan --terragrunt-non-interactive
  only:
    - merge_requests

cultivator_apply:
  stage: apply
  image: ubuntu:latest
  script:
    - apt-get update && apt-get install -y wget unzip
    - wget -q https://releases.hashicorp.com/terraform/${TERRAFORM_VERSION}/terraform_${TERRAFORM_VERSION}_linux_amd64.zip
    - unzip -q terraform_${TERRAFORM_VERSION}_linux_amd64.zip -d /usr/local/bin
    - wget -q https://github.com/gruntwork-io/terragrunt/releases/download/v${TERRAGRUNT_VERSION}/terragrunt_linux_amd64
    - chmod +x terragrunt_linux_amd64 && mv terragrunt_linux_amd64 /usr/local/bin/terragrunt
    - terragrunt run-all apply --terragrunt-non-interactive
  when: manual
  only:
    - merge_requests
```

See [GitLab Pipelines](../user-guide/gitlab-pipelines.md) for full configuration options.

### Option 3: Docker

Run Cultivator in a container:

```bash
docker run -v $(pwd):/work \
  -e GITHUB_TOKEN=$GITHUB_TOKEN \
  -e GITHUB_REPOSITORY=owner/repo \
  weyderfs/cultivator:latest
```

### Option 4: Local Installation (Development)

```bash
# Clone the repository
git clone https://github.com/Ops-Talks/cultivator.git
cd cultivator

# Build from source
go build -o cultivator ./cmd/cultivator

# Run
./cultivator --help
```

## Configuration

Create `cultivator.yml` in your repository root:

```yaml
version: 1

settings:
  # Automatically plan on PR open/update
  auto_plan: true
  
  # Require approval before apply
  require_approval: true
  
  # Lock timeout for concurrent applies
  lock_timeout: 30m
  
  # Parallel execution
  parallel_plan: false
  max_parallel: 1
```

See [Configuration Reference](configuration.md) for all options.

## Verify Installation

1. Open a Pull Request with infrastructure changes
2. Comment on the PR:
   ```
   cultivator plan
   ```
3. Wait for Cultivator to respond with plan results

If you see a formatted comment with Terragrunt plan output, installation is successful!

## Next Steps

- [Quick Start](quickstart.md) - Run your first commands
- [Configuration](configuration.md) - Customize your setup
- [User Guide](../user-guide/index.md) - Learn available commands
