# Quick Start

Get Cultivator running in your repository in 5 minutes.

## Step 1: Add GitHub Action

Create `.github/workflows/cultivator.yml`:

```yaml
name: Cultivator

on:
  pull_request:
    types: [opened, synchronize, reopened]
  issue_comment:
    types: [created]

jobs:
  cultivator:
    runs-on: ubuntu-latest
    permissions:
      contents: read
      pull-requests: write
      checks: write
      statuses: write
      
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0
          
      - name: Run Cultivator
        uses: weyderfs/cultivator@v1
        with:
          github-token: ${{ secrets.GITHUB_TOKEN }}
          terragrunt-version: 0.55.0
```

## Step 2: Commit and Push

```bash
git add .github/workflows/cultivator.yml
git commit -m "Add Cultivator workflow"
git push origin your-branch
```

## Step 3: Create a Pull Request

1. Go to GitHub
2. Open a Pull Request with your changes
3. Comment on the PR:

```
cultivator plan
```

## Step 4: Wait for Results

Cultivator will:
1. Detect changed modules
2. Run `terragrunt plan` on affected modules
3. Post results as a PR comment

That's it! 

## Next Steps

- Learn about [available commands](../user-guide/workflows.md)
- Configure [advanced settings](configuration.md)
- Set up [locking and approvals](../user-guide/configuration.md)

## Common Commands

```bash
# Plan affected modules
cultivator plan

# Plan all modules
cultivator plan --all

# Apply changes
cultivator apply

# Apply with auto-approve
cultivator apply --all
```
