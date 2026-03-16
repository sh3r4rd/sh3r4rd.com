# AGENTS.md — Recruiter Dashboard Infrastructure

Subdirectory instructions for Codex CLI when working in `infra/recruiter-dashboard/`.
These supplement the root [AGENTS.md](../../AGENTS.md).

## Quick Reference

```bash
# Go tests
cd lambda-src/email-parser && go test -v -race ./...
cd lambda-src/email-parser && go vet ./...

# Lambda builds
make -C ../.. build-lambdas

# Terraform
terraform init
terraform validate
terraform fmt -check -recursive
terraform plan
```

## Email Parsing Pipeline

SES receives email → stores raw email in S3 → triggers email-parser Lambda → parses with `internal/parser` → extracts recruiter data via OpenAI (`internal/extractor`) → sanitizes fields (`internal/sanitizer`) → writes to DynamoDB (`internal/db`) → tags S3 object (`internal/tagger`)

## DynamoDB Schema

- **Table:** `recruiter_emails` (PROVISIONED billing, 15/15 RCU/WCU)
- **Primary key:** `id` (S, partition) + `received_at` (S, sort)
- **GSI `recruiter-index`:** `recruiter_email` (HASH) + `received_at` (RANGE)
- **GSI `date-index`:** `date_year` (HASH) + `date_day` (RANGE)

## Terraform Modules

| Module | Purpose |
|--------|---------|
| `modules/s3` | Email storage bucket (AES-256, 30-day lifecycle) |
| `modules/dynamodb` | Recruiter emails table + GSIs |
| `modules/lambda` | Lambda function deployment (arm64) |
| `modules/iam` | Roles, policies, log group ARN construction |
| `modules/ses` | Email receiving on `inbox.sh3r4rd.com` |
| `modules/api-gateway` | REST API at `api.sh3r4rd.com` |
| `modules/monitoring` | CloudWatch alarms, budget alerts via SNS |

## Terraform Conventions

- Module pattern: `modules/<name>/main.tf`, `variables.tf`, `outputs.tf`
- `snake_case` for resource names and variables
- Add `description` to all variables and outputs
- Use `validation` blocks for input constraints
- No hardcoded ARNs or account IDs — use data sources or variables
- IAM module constructs log group ARNs from function names (not module outputs) to avoid circular dependencies

## Go Conventions

- Entry points: `cmd/handler/main.go` (email-parser), `main.go` (api-handler)
- All internal logic in `internal/` packages
- Initialize AWS clients once on cold start in `main.go`, reuse across invocations
- Define interfaces for AWS service clients (testability)
- Custom error types in `internal/errors/`
- Test naming: `TestFunctionName_scenario`
- Test fixtures in `testdata/` — never modify without explicit instruction
