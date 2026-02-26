# Quick Start Guide

This guide will help you get started with Cultivator in 5 minutes.

## Step 1: Prerequisites

Ensure you have:
- A GitHub repository with Terragrunt configuration
- Terraform and Terragrunt installed locally (for testing)
- GitHub Actions enabled

## Step 2: Add GitHub Workflow

Create `.github/workflows/cultivator.yml` in your repository:

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

jobs:
  cultivator:
    runs-on: ubuntu-latest
    if: |
      (github.event_name == 'pull_request') ||
      (github.event_name == 'issue_comment' && 
       github.event.issue.pull_request &&
       startsWith(github.event.comment.body, '/cultivator'))
    
    steps:
      - uses: actions/checkout@v4
        with:
          ref: ${{ github.event.pull_request.head.sha }}
          fetch-depth: 0
          
      - name: Setup Terraform
        uses: hashicorp/setup-terraform@v3
        with:
          terraform_version: 1.7.0
          
      - name: Setup Terragrunt
        run: |
          wget -q https://github.com/gruntwork-io/terragrunt/releases/download/v0.55.0/terragrunt_linux_amd64
          chmod +x terragrunt_linux_amd64
          sudo mv terragrunt_linux_amd64 /usr/local/bin/terragrunt
          
      - name: Configure AWS Credentials
        uses: aws-actions/configure-aws-credentials@v4
        with:
          aws-access-key-id: ${{ secrets.AWS_ACCESS_KEY_ID }}
          aws-secret-access-key: ${{ secrets.AWS_SECRET_ACCESS_KEY }}
          aws-region: us-east-1
          
      - name: Run Cultivator
        uses: cultivator-dev/cultivator-action@v1
        with:
          github-token: ${{ secrets.GITHUB_TOKEN }}
```

## Step 3: Configure Cultivator

Create `cultivator.yml` in your repository root:

```yaml
version: 1

projects:
  - name: infrastructure
    dir: infrastructure/
    terragrunt_version: 0.55.0
    terraform_version: 1.7.0
    auto_plan: true
    apply_requirements:
      - approved

settings:
  auto_plan: true
  parallel_plan: true
  max_parallel: 5
```

## Step 4: Test It Out

1. **Create a Pull Request** that modifies a Terragrunt file
2. **Cultivator will automatically run `plan`** and comment on your PR
3. **Review the plan output** in the PR comment
4. **Apply changes** by commenting `/cultivator apply`

## Available Commands

Comment these on your PR to trigger actions:

- `/cultivator plan` - Run plan on changed modules
- `/cultivator apply` - Apply changes
- `/cultivator plan-all` - Plan all modules
- `/cultivator apply-all` - Apply all modules
- `/cultivator unlock` - Remove locks

## Example PR Workflow

1. **Open PR** with infrastructure changes
2. Cultivator **automatically runs plan**
3. Review the **plan output** in PR comments
4. Get **approval** from team
5. Comment `/cultivator apply` to **apply changes**
6. Cultivator **applies** and comments with results
7. **Merge** the PR

## Next Steps

- Read the [Configuration Guide](configuration.md)
- Check out [Examples](../examples/)
- Learn about advanced features in the [User Guide](user-guide/features.md)

## Troubleshooting

### Plan doesn't run automatically
- Check if `auto_plan: true` is set in `cultivator.yml`
- Verify the GitHub workflow has correct permissions

### Apply fails
- Ensure PR is approved if `apply_requirements: [approved]` is set
- Check AWS credentials are configured correctly
- Verify Terragrunt/Terraform versions match

### Module not detected
- Ensure the changed file is in a directory with `terragrunt.hcl`
- Check file extensions (`.hcl`, `.tf`, `.tfvars`)

## Need Help?

- Check the [FAQ](faq.md)
- Open an [Issue](https://github.com/cultivator-dev/cultivator/issues)
- View [Contributing Guide](/CONTRIBUTING.md)
