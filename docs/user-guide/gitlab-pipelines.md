# GitLab CI/CD Integration

This guide shows how to run Cultivator in GitLab CI/CD using the same model as `examples/.gitlab-ci.yml` in this repository.

Cultivator runs as a CLI inside your job environment. You need:

1. `cultivator`
2. `terragrunt`
3. `tofu` or `terraform`

---

## Recommended pipeline

```yaml
# .gitlab-ci.yml

stages:
  - validate
  - plan
  - apply

variables:
  CULTIVATOR_VERSION: "v0.2.7"
  TOFU_VERSION: "1.11.5"
  TERRAGRUNT_VERSION: "0.99.0"

  CULTIVATOR_ROOT: "providers"
  CULTIVATOR_ENV: ""
  CULTIVATOR_PARALLELISM: "4"
  CULTIVATOR_OUTPUT_FORMAT: "text"

workflow:
  rules:
    - if: '$CI_PIPELINE_SOURCE == "merge_request_event"'
    - if: '$CI_COMMIT_BRANCH == "main"'
    - if: '$CI_PIPELINE_SOURCE == "web"'

.install_tools: &install_tools
  before_script:
    - apk add --no-cache wget unzip curl jq ca-certificates
    - wget -q https://github.com/opentofu/opentofu/releases/download/v${TOFU_VERSION}/tofu_${TOFU_VERSION}_linux_amd64.zip
    - unzip -q tofu_${TOFU_VERSION}_linux_amd64.zip -d /usr/local/bin
    - rm tofu_${TOFU_VERSION}_linux_amd64.zip
    - wget -q -O /usr/local/bin/terragrunt https://github.com/gruntwork-io/terragrunt/releases/download/v${TERRAGRUNT_VERSION}/terragrunt_linux_amd64
    - chmod +x /usr/local/bin/terragrunt
    - wget -q -O /usr/local/bin/cultivator https://github.com/Ops-Talks/cultivator/releases/download/${CULTIVATOR_VERSION}/cultivator-linux-amd64
    - chmod +x /usr/local/bin/cultivator

doctor:
  stage: validate
  image: alpine:3.21
  <<: *install_tools
  script:
    - cultivator doctor --root "$CULTIVATOR_ROOT"
  rules:
    - if: '$CI_PIPELINE_SOURCE == "merge_request_event"'
    - if: '$CI_COMMIT_BRANCH == "main"'

plan:
  stage: plan
  image: alpine:3.21
  <<: *install_tools
  script:
    - |
      set -- \
        --root "$CULTIVATOR_ROOT" \
        --parallelism "$CULTIVATOR_PARALLELISM" \
        --output-format "$CULTIVATOR_OUTPUT_FORMAT" \
        --non-interactive=true
      if [ -n "$CULTIVATOR_ENV" ]; then
        set -- "$@" --env "$CULTIVATOR_ENV"
      fi
      cultivator plan "$@" | tee plan_output.txt
    - |
      if [ -z "$GITLAB_TOKEN" ]; then
        echo "GITLAB_TOKEN not set; skipping MR comment"
        exit 0
      fi

      if [ -n "$CI_MERGE_REQUEST_IID" ]; then
        MR_IID="$CI_MERGE_REQUEST_IID"
      else
        MR_IID=$(curl --silent --header "PRIVATE-TOKEN: $GITLAB_TOKEN" \
          "${CI_API_V4_URL}/projects/${CI_PROJECT_ID}/merge_requests?state=opened&source_branch=${CI_COMMIT_REF_NAME}" \
          | jq -r '.[0].iid // empty')
      fi

      if [ -z "$MR_IID" ]; then
        echo "No MR found; skipping comment"
        exit 0
      fi

      PLAN_OUTPUT=$(cat plan_output.txt)
      COMMENT="## Cultivator Plan\n\n\`\`\`\n${PLAN_OUTPUT}\n\`\`\`"
      curl --silent --show-error --fail --request POST \
        --header "PRIVATE-TOKEN: $GITLAB_TOKEN" \
        "${CI_API_V4_URL}/projects/${CI_PROJECT_ID}/merge_requests/${MR_IID}/notes" \
        --data-urlencode "body=${COMMENT}"
  artifacts:
    when: always
    paths:
      - plan_output.txt
    expire_in: 1 day
  rules:
    - if: '$CI_PIPELINE_SOURCE == "merge_request_event"'
    - if: '$CI_PIPELINE_SOURCE == "web"'

apply:
  stage: apply
  image: alpine:3.21
  <<: *install_tools
  script:
    - |
      set -- \
        --root "$CULTIVATOR_ROOT" \
        --parallelism "$CULTIVATOR_PARALLELISM" \
        --output-format "$CULTIVATOR_OUTPUT_FORMAT" \
        --non-interactive=true \
        --auto-approve=true
      if [ -n "$CULTIVATOR_ENV" ]; then
        set -- "$@" --env "$CULTIVATOR_ENV"
      fi
      cultivator apply "$@" | tee apply_output.txt
    - |
      if [ -z "$GITLAB_TOKEN" ]; then
        echo "GITLAB_TOKEN not set; skipping MR comment"
        exit 0
      fi

      MR_IID=$(curl --silent --header "PRIVATE-TOKEN: $GITLAB_TOKEN" \
        "${CI_API_V4_URL}/projects/${CI_PROJECT_ID}/merge_requests?state=merged&order_by=updated_at&sort=desc" \
        | jq -r ".[] | select(.merge_commit_sha==\"${CI_COMMIT_SHA}\") | .iid" | head -1)

      if [ -z "$MR_IID" ]; then
        echo "No merged MR found for this commit; skipping comment"
        exit 0
      fi

      APPLY_OUTPUT=$(cat apply_output.txt)
      COMMENT="## Cultivator Apply Result\n\n\`\`\`\n${APPLY_OUTPUT}\n\`\`\`"
      curl --silent --show-error --fail --request POST \
        --header "PRIVATE-TOKEN: $GITLAB_TOKEN" \
        "${CI_API_V4_URL}/projects/${CI_PROJECT_ID}/merge_requests/${MR_IID}/notes" \
        --data-urlencode "body=${COMMENT}"
  artifacts:
    when: always
    paths:
      - apply_output.txt
    expire_in: 1 day
  rules:
    - if: '$CI_COMMIT_BRANCH == "main"'
      when: manual
  environment:
    name: production
```

---

## Optional: using `cultivator.yml`

A config file is optional. If you use one, pass it explicitly with `--config`.

```yaml
# cultivator.yml
root: providers
parallelism: 4
output_format: text
non_interactive: true
```

```yaml
script:
  - cultivator plan --config=cultivator.yml
```

---

## Notes

- Keep `CULTIVATOR_ENV` empty to run all stacks under `CULTIVATOR_ROOT`.
- Set `CULTIVATOR_ENV` when your repository layout maps environments to folders.
- Ensure `GITLAB_TOKEN` has permission to create MR notes.
- If using OpenTofu, ensure Terragrunt is configured to use `tofu`.

---

## Troubleshooting

### `cultivator: command not found`
Confirm binary installation and `PATH`; keep `doctor` job enabled.

### `terragrunt: command not found`
Cultivator delegates execution to Terragrunt. Install both binaries in the same job.

### No stacks discovered
Check `CULTIVATOR_ROOT` path and optional `CULTIVATOR_ENV` filter.

---

## Further reading

- [Quickstart](../getting-started/quickstart.md)
- [Configuration](../getting-started/configuration.md)
- [CLI Reference](cli-reference.md)
- [GitLab CI/CD documentation](https://docs.gitlab.com/ee/ci/)
