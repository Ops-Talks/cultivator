# PowerShell Setup Script for Cultivator
# Quickly initialize Cultivator in your Terragrunt repository

Write-Host "Cultivator Setup Script" -ForegroundColor Green
Write-Host "==========================" -ForegroundColor Green
Write-Host ""

# Check if we're in a git repository
if (-not (Test-Path .git)) {
    Write-Host "Error: Not a git repository" -ForegroundColor Red
    Write-Host "Please run this script from the root of your git repository"
    exit 1
}

# Check if GitHub Actions directory exists
if (-not (Test-Path .github\workflows)) {
    Write-Host "Creating .github/workflows directory..." -ForegroundColor Yellow
    New-Item -ItemType Directory -Path .github\workflows -Force | Out-Null
}

# Create cultivator.yml if it doesn't exist
if (-not (Test-Path cultivator.yml)) {
    Write-Host "Creating cultivator.yml..." -ForegroundColor Yellow
    
    $cultivatorConfig = @"
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
"@
    
    Set-Content -Path cultivator.yml -Value $cultivatorConfig
    Write-Host "Created cultivator.yml" -ForegroundColor Green
} else {
    Write-Host "cultivator.yml already exists, skipping..." -ForegroundColor Cyan
}

# Create GitHub workflow if it doesn't exist
if (-not (Test-Path .github\workflows\cultivator.yml)) {
    Write-Host "Creating GitHub Actions workflow..." -ForegroundColor Yellow
    
    $workflowConfig = @"
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
          ref: `${{ github.event.pull_request.head.sha }}
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
          github-token: `${{ secrets.GITHUB_TOKEN }}
"@
    
    Set-Content -Path .github\workflows\cultivator.yml -Value $workflowConfig
    Write-Host "Created .github/workflows/cultivator.yml" -ForegroundColor Green
} else {
    Write-Host "GitHub workflow already exists, skipping..." -ForegroundColor Cyan
}

Write-Host ""
Write-Host "Setup complete!" -ForegroundColor Green
Write-Host ""
Write-Host "Next steps:" -ForegroundColor Yellow
Write-Host "1. Review and customize cultivator.yml for your project"
Write-Host "2. Commit the changes:"
Write-Host "   git add cultivator.yml .github/workflows/cultivator.yml"
Write-Host "   git commit -m 'Add Cultivator configuration'"
Write-Host "   git push"
Write-Host "3. Create a test PR to see Cultivator in action"
Write-Host ""
Write-Host "Documentation: https://github.com/cultivator-dev/cultivator" -ForegroundColor Cyan
Write-Host "Issues: https://github.com/cultivator-dev/cultivator/issues" -ForegroundColor Cyan
Write-Host ""
