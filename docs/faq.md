# Frequently Asked Questions

## General Questions

### What is Cultivator?
Cultivator is a CI/CD automation tool for Terragrunt that runs plan and apply operations directly from Pull Requests, similar to Atlantis or Digger.

### How is it different from Atlantis?
Atlantis is designed for Terraform. Cultivator was built specifically for Terragrunt, supporting:
- Terragrunt dependencies between modules
- Run-all operations
- Hierarchical configurations with inheritance
- Impact detection when parent configs change

### Do I need a separate server?
No! Cultivator runs in your existing CI/CD (GitHub Actions, GitLab CI, etc). No separate infrastructure needed.

## Installation

### Can I use Cultivator with existing GitHub Actions?
Yes! Cultivator is designed to work alongside other workflows. Just add the Cultivator workflow to your repository.

### What versions of Terragrunt are supported?
Cultivator supports Terragrunt 0.40.0+. For best results, use recent versions (1.0+).

### What about OpenTofu/Terraform version?
Cultivator works with any OpenTofu or Terraform version that your Terragrunt version supports.

## Usage

### How do I run a plan?
Comment on a PR:
```
cultivator plan
```

### How do I apply changes?
Comment on a PR:
```
cultivator apply
```

### Can I run all modules?
Yes:
```
cultivator plan --all
cultivator apply --all
```

### What if a module fails?
Cultivator will post the error in the PR comment. Fix the issue and comment again to retry.

## Security

### Is my OpenTofu/Terraform state safe?
Yes. Cultivator requires proper AWS/Azure/GCP credentials in GitHub Secrets. Only authorized users can run operations.

### What about sensitive outputs?
Cultivator automatically redacts sensitive data marked in Terragrunt outputs. Additionally, outputs marked as `sensitive = true` are hidden.

### How do comments get validated?
GitHub webhook signatures are validated. Only authorized collaborators can trigger operations (configurable).

## Troubleshooting

### Cultivator doesn't respond to my comment
- Check your workflow file is in `.github/workflows/`
- Verify the repository has permissions for the workflow
- Check that you're commenting on a PR (not an issue)

### Plan/Apply fails with module not found
- Verify your Terragrunt structure
- Check file paths are relative to repo root
- Ensure `terragrunt.hcl` files exist

### Lock timeout errors
- Check if another apply is in progress
- Verify lock storage is accessible
- Adjust `lock_timeout` in `cultivator.yml` if needed

### Sensitive data appears in PR comment
- Mark outputs as `sensitive = true` in Terraform
- Update redaction patterns in configuration
- Report the issue if it's a known sensitive pattern

## Advanced Questions

### Can I customize the output format?
Currently, formatting is fixed. Custom formatting is a planned feature.

### How do I integrate with other tools?
Cultivator can be used alongside other GitHub Actions. It respects branch protection rules and requires reviews when configured.

### Can I run custom scripts?
Not directly. You can pre/post process with GitHub Actions before calling Cultivator.

### How do I monitor Cultivator runs?
Check GitHub Actions logs and PR comments. GitHub also provides workflow run history and analytics.

## Support

Still have questions?

- **Documentation**: Check our [full documentation](/)
- **Issues**: [GitHub Issues](https://github.com/Ops-Talks/cultivator/issues)
- **Discussions**: [GitHub Discussions](https://github.com/Ops-Talks/cultivator/discussions)
