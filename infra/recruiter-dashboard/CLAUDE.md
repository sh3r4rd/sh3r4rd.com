# Recruiter Dashboard Infrastructure

Scoped instructions for Claude Code when working in `infra/recruiter-dashboard/`.

## Commands

```bash
# Go tests — email-parser
cd lambda-src/email-parser && go test -v -race ./...

# Go tests — api-handler
cd lambda-src/api-handler && RECRUITER_TABLE=test CORS_ALLOW_ORIGIN=http://localhost DATE_INDEX_NAME=date-index AWS_ACCESS_KEY_ID=test AWS_SECRET_ACCESS_KEY=test AWS_REGION=us-east-1 go test -v -race ./...

# Lambda builds
make -C ../.. build-lambdas

# Terraform
terraform init
terraform validate
terraform fmt -check -recursive

# terraform.tfvars sync (SSM is the source of truth; the file is git-ignored)
make -C ../.. tf-vars-pull   # before plan/apply
make -C ../.. tf-vars-push   # after editing terraform.tfvars
make -C ../.. tf-vars-diff   # check drift vs SSM (non-zero exit == drift)
```

## terraform.tfvars (sourced from SSM)

`terraform.tfvars` is git-ignored; its durable, versioned source of truth is the SSM SecureString parameter `/recruiter-dashboard/tfvars` (us-east-1). `scripts/tfvars.sh` mirrors the file to/from SSM (`pull`/`push`/`diff`) — Terraform code is unaware of it. Run `make tf-vars-pull` before `plan`/`apply` and `make tf-vars-push` after editing the file.

- `diff` exits 0 for no drift, 1 for drift, 2 for usage errors — so it can gate a script or CI check.
- Push normalizes the file: CRs are stripped and trailing blank lines collapse to one newline (SSM stores LF only, and this keeps round-trips byte-stable).
- `pull` writes a timestamped `terraform.tfvars.<UTC>.bak` before overwriting; git-ignored, never auto-pruned.
- `TFVARS_ARGS=--force` skips confirmation prompts, e.g. `make tf-vars-push TFVARS_ARGS=--force`.

## Lambda Functions

### email-parser (`lambda-src/email-parser/`)

SES email -> S3 -> this Lambda -> parse -> extract via OpenAI -> sanitize -> write to DynamoDB -> tag S3 object.

- Entry: `cmd/handler/main.go`
- 9 internal packages: `handler`, `ssm`, `sanitizer`, `extractor`, `models`, `parser`, `db`, `errors`, `tagger`
- Env vars: `RECRUITER_TABLE`, `EMAIL_BUCKET`, `S3_KEY_PREFIX`, `SSM_OPENAI_KEY_NAME`

### api-handler (`lambda-src/api-handler/`)

REST API serving anonymized recruiter data. PII is stripped from all responses.

- Source: `main.go` (entry), `handler.go` (routing/DynamoDB), `anonymizer.go` (PII stripping)
- Endpoints: `GET /recruiters` (`?company=`, `?month=YYYY-MM`), `GET /recruiters/{id}`, `GET /stats`
- Env vars: `RECRUITER_TABLE`, `CORS_ALLOW_ORIGIN`, `DATE_INDEX_NAME`

## Terraform Modules

7 modules in `modules/`: s3, dynamodb, lambda, iam, ses, api-gateway, monitoring.

Module pattern: `main.tf`, `variables.tf`, `outputs.tf`. Use `snake_case`. Add `description` to all variables/outputs. No hardcoded ARNs. IAM constructs log group ARNs from function names to avoid circular deps.

## DynamoDB Schema

- **Table:** `recruiter_emails` (PROVISIONED, 15/15 RCU/WCU)
- **PK:** `id` (S) + `received_at` (S)
- **GSI `recruiter-index`:** `recruiter_email` + `received_at`
- **GSI `date-index`:** `date_year` + `date_day`
