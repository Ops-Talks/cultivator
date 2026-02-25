#!/bin/bash

# Cultivator Setup Script
# Quickly initialize Cultivator in your Terragrunt repository

set -e

echo "Cultivator Setup Script"
echo "=========================="
echo ""

# Check if we're in a git repository
if ! git rev-parse --git-dir > /dev/null 2>&1; then
    echo "Error: Not a git repository"
    echo "Please run this script from the root of your git repository"
    exit 1
fi

# Check if GitHub Actions directory exists
if [ ! -d ".github/workflows" ]; then
    echo "Creating .github/workflows directory..."
    mkdir -p .github/workflows
fi

# Create cultivator.yml if it doesn't exist
if [ ! -f "cultivator.yml" ]; then
    echo "Creating cultivator.yml..."
    cat > cultivator.yml << 'EOF'
version: 1

projects:
  - name: infrastructure
    dir: .
    terragrunt_version: 0.55.0
    terraform_version: 1.7.0
    auto_plan: true
    apply_requirements:
      - approved

settings:
  auto_plan: true
  parallel_plan: true
  max_parallel: 5
  lock_timeout: 10m
EOF
    echo "Created cultivator.yml"
else
    echo "cultivator.yml already exists, skipping..."
fi

# Create GitHub workflow if it doesn't exist
if [ ! -f ".github/workflows/cultivator.yml" ]; then
    echo "Creating GitHub Actions workflow..."
    cat > .github/workflows/cultivator.yml << 'EOF'
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
          
      - name: Run Cultivator
        uses: cultivator-dev/cultivator-action@v1
        with:
          github-token: ${{ secrets.GITHUB_TOKEN }}
EOF
    echo "Created .github/workflows/cultivator.yml"
else
    echo "GitHub workflow already exists, skipping..."
fi

echo ""
echo "Setup complete!"
echo ""
echo "Next steps:"
echo "1. Review and customize cultivator.yml for your project"
echo "2. Commit the changes:"
echo "   git add cultivator.yml .github/workflows/cultivator.yml"
echo "   git commit -m 'Add Cultivator configuration'"
echo "   git push"
echo "3. Create a test PR to see Cultivator in action"
echo ""
echo "Documentation: https://github.com/cultivator-dev/cultivator"
echo "Issues: https://github.com/cultivator-dev/cultivator/issues"
echo ""
