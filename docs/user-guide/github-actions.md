# GitHub Actions Integration

This guide shows how to run Cultivator inside GitHub Actions workflows.

Like with GitLab CI, you need three binaries in the same job environment:

1. The `cultivator` binary
2. The `terragrunt` binary
3. `terraform` (or OpenTofu)

---

## Minimal example

```yaml
# .github/workflows/cultivator.yml

name: Terragrunt

on:
  pull_request:
    branches: [ main, develop ]
  push:
    branches: [ main ]

env:
  CULTIVATOR_VERSION: v0.2.0
  TERRAGRUNT_VERSION: 0.67.0
  TERRAFORM_VERSION: 1.10.0
  CULTIVATOR_ROOT: infrastructure
  CULTIVATOR_PARALLELISM: 4
  CULTIVATOR_OUTPUT_FORMAT: json

jobs:
  doctor:
    name: Doctor
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@v4

      - name: Install tools
        run: |
          # Terraform
          wget -q https://releases.hashicorp.com/terraform/${TERRAFORM_VERSION}/terraform_${TERRAFORM_VERSION}_linux_amd64.zip
          unzip -q terraform_${TERRAFORM_VERSION}_linux_amd64.zip -d /usr/local/bin
          rm terraform_${TERRAFORM_VERSION}_linux_amd64.zip
          
          # Terragrunt
          wget -q -O /usr/local/bin/terragrunt \
            https://github.com/gruntwork-io/terragrunt/releases/download/v${TERRAGRUNT_VERSION}/terragrunt_linux_amd64
          chmod +x /usr/local/bin/terragrunt
          
          # Cultivator
          wget -q -O /usr/local/bin/cultivator \
            https://github.com/Ops-Talks/cultivator/releases/download/${CULTIVATOR_VERSION}/cultivator-linux-amd64
          chmod +x /usr/local/bin/cultivator

      - name: Run doctor
        run: cultivator doctor

  plan:
    name: Plan
    runs-on: ubuntu-latest
    if: github.event_name == 'pull_request'
    steps:
      - name: Checkout
        uses: actions/checkout@v4

      - name: Install tools
        run: |
          wget -q https://releases.hashicorp.com/terraform/${TERRAFORM_VERSION}/terraform_${TERRAFORM_VERSION}_linux_amd64.zip
          unzip -q terraform_${TERRAFORM_VERSION}_linux_amd64.zip -d /usr/local/bin && rm terraform_${TERRAFORM_VERSION}_linux_amd64.zip
          
          wget -q -O /usr/local/bin/terragrunt \
            https://github.com/gruntwork-io/terragrunt/releases/download/v${TERRAGRUNT_VERSION}/terragrunt_linux_amd64
          chmod +x /usr/local/bin/terragrunt
          
          wget -q -O /usr/local/bin/cultivator \
            https://github.com/Ops-Talks/cultivator/releases/download/${CULTIVATOR_VERSION}/cultivator-linux-amd64
          chmod +x /usr/local/bin/cultivator

      - name: Determine environment
        id: env
        run: |
          if [ "${{ github.base_ref }}" == "main" ]; then
            echo "environment=prod" >> $GITHUB_OUTPUT
          else
            echo "environment=dev" >> $GITHUB_OUTPUT
          fi

      - name: Plan
        run: |
          cultivator plan \
            --root "${{ env.CULTIVATOR_ROOT }}" \
            --env "${{ steps.env.outputs.environment }}" \
            --parallelism "${{ env.CULTIVATOR_PARALLELISM }}" \
            --output-format "${{ env.CULTIVATOR_OUTPUT_FORMAT }}" \
            --non-interactive

      - name: Comment PR
        if: always()
        uses: actions/github-script@v7
        with:
          script: |
            const fs = require('fs');
            // Optional: post plan output to PR comment
            github.rest.issues.createComment({
              issue_number: context.issue.number,
              owner: context.repo.owner,
              repo: context.repo.repo,
              body: 'Cultivator plan completed. See workflow run for details.'
            });

  apply:
    name: Apply
    runs-on: ubuntu-latest
    needs: plan
    if: github.event_name == 'push' && github.ref == 'refs/heads/main'
    steps:
      - name: Checkout
        uses: actions/checkout@v4

      - name: Install tools
        run: |
          wget -q https://releases.hashicorp.com/terraform/${TERRAFORM_VERSION}/terraform_${TERRAFORM_VERSION}_linux_amd64.zip
          unzip -q terraform_${TERRAFORM_VERSION}_linux_amd64.zip -d /usr/local/bin && rm terraform_${TERRAFORM_VERSION}_linux_amd64.zip
          
          wget -q -O /usr/local/bin/terragrunt \
            https://github.com/gruntwork-io/terragrunt/releases/download/v${TERRAGRUNT_VERSION}/terragrunt_linux_amd64
          chmod +x /usr/local/bin/terragrunt
          
          wget -q -O /usr/local/bin/cultivator \
            https://github.com/Ops-Talks/cultivator/releases/download/${CULTIVATOR_VERSION}/cultivator-linux-amd64
          chmod +x /usr/local/bin/cultivator

      - name: Apply
        run: |
          cultivator apply \
            --root "${{ env.CULTIVATOR_ROOT }}" \
            --env prod \
            --parallelism "${{ env.CULTIVATOR_PARALLELISM }}" \
            --output-format "${{ env.CULTIVATOR_OUTPUT_FORMAT }}" \
            --non-interactive \
            --auto-approve
```

---

## DRY: Reusing tool installation

To avoid repeating the installation steps, use a composite action:

```yaml
# .github/actions/install-tools/action.yml

name: Install tools
description: Install terraform, terragrunt, and cultivator

inputs:
  terraform-version:
    required: true
  terragrunt-version:
    required: true
  cultivator-version:
    required: true

runs:
  using: composite
  steps:
    - run: |
        wget -q https://releases.hashicorp.com/terraform/${{ inputs.terraform-version }}/terraform_${{ inputs.terraform-version }}_linux_amd64.zip
        unzip -q terraform_${{ inputs.terraform-version }}_linux_amd64.zip -d /usr/local/bin
        rm terraform_${{ inputs.terraform-version }}_linux_amd64.zip
      shell: bash

    - run: |
        wget -q -O /usr/local/bin/terragrunt \
          https://github.com/gruntwork-io/terragrunt/releases/download/v${{ inputs.terragrunt-version }}/terragrunt_linux_amd64
        chmod +x /usr/local/bin/terragrunt
      shell: bash

    - run: |
        wget -q -O /usr/local/bin/cultivator \
          https://github.com/Ops-Talks/cultivator/releases/download/${{ inputs.cultivator-version }}/cultivator-linux-amd64
        chmod +x /usr/local/bin/cultivator
      shell: bash
```

Then in your workflow:

```yaml
jobs:
  plan:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - uses: ./.github/actions/install-tools
        with:
          terraform-version: 1.10.0
          terragrunt-version: 0.67.0
          cultivator-version: v0.2.0

      - run: cultivator plan --root infrastructure --env dev --non-interactive
```

---

## Using secrets for cloud credentials

Store credentials in **Settings → Secrets and variables → Actions**, then expose them:

```yaml
plan:
  runs-on: ubuntu-latest
  env:
    AWS_ACCESS_KEY_ID: ${{ secrets.AWS_ACCESS_KEY_ID }}
    AWS_SECRET_ACCESS_KEY: ${{ secrets.AWS_SECRET_ACCESS_KEY }}
    AWS_REGION: us-east-1
  steps:
    - uses: actions/checkout@v4
    - run: |
        wget -q -O /usr/local/bin/cultivator \
          https://github.com/Ops-Talks/cultivator/releases/download/v0.2.0/cultivator-linux-amd64
        chmod +x /usr/local/bin/cultivator
    - run: cultivator plan --root infrastructure --env dev --non-interactive
```

---

## Environment-based strategy matrix

Run plan/apply for multiple environments in parallel:

```yaml
plan:
  runs-on: ubuntu-latest
  strategy:
    matrix:
      environment: [ dev, staging, prod ]
    max-parallel: 2
  steps:
    - uses: actions/checkout@v4

    - uses: ./.github/actions/install-tools
      with:
        terraform-version: 1.10.0
        terragrunt-version: 0.67.0
        cultivator-version: v0.2.0

    - run: |
        cultivator plan \
          --root infrastructure \
          --env ${{ matrix.environment }} \
          --non-interactive
```

---

## Tag-based filtering

```yaml
plan:
  runs-on: ubuntu-latest
  steps:
    - uses: actions/checkout@v4
    - uses: ./.github/actions/install-tools
      with:
        terraform-version: 1.10.0
        terragrunt-version: 0.67.0
        cultivator-version: v0.2.0

    - run: |
        # Only plan modules tagged with "networking"
        cultivator plan \
          --root infrastructure \
          --tags networking \
          --non-interactive
```

---

## Caching

Cache the Terragrunt plugin directory to speed up runs:

```yaml
plan:
  runs-on: ubuntu-latest
  steps:
    - uses: actions/checkout@v4

    - name: Cache .terragrunt-cache
      uses: actions/cache@v4
      with:
        path: .terragrunt-cache
        key: terragrunt-${{ github.ref }}-${{ github.sha }}
        restore-keys: |
          terragrunt-${{ github.ref }}-
          terragrunt-
```

---

## Troubleshooting

### `cultivator: command not found`
The binary was not added to `PATH`. Verify the installation step ran and `/usr/local/bin` is in `PATH`.

### `terragrunt: command not found`
Cultivator needs both binaries. Add it to the install step and verify it's executable.

### Workflow not triggering
Check the `on` conditions in your workflow file. Example: `if: github.event_name == 'push'` won't trigger on PRs.

### Secrets not available
Ensure secrets are defined in **Settings → Secrets and variables → Actions** and referenced correctly as `${{ secrets.SECRET_NAME }}`.

---

## Further reading

- [Quickstart](../getting-started/quickstart.md) — local usage
- [Configuration](../getting-started/configuration.md) — all options
- [Workflows](workflows.md) — available commands and flags
- [GitHub Actions documentation](https://docs.github.com/en/actions)
- [GitLab CI equivalent](gitlab-pipelines.md)
