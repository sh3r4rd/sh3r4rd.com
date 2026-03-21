---
paths:
  - "infra/**/*.tf"
---

# Terraform Conventions

## Module Structure

Each module in `infra/recruiter-dashboard/modules/<name>/`:
- `main.tf` — Resources and data sources
- `variables.tf` — Input variables with `description`, `type`, and `validation`
- `outputs.tf` — Output values with `description`

## 7 Modules

| Module | Purpose |
|--------|---------|
| `s3` | Email storage bucket (AES-256, 30-day lifecycle) |
| `dynamodb` | Recruiter emails table + GSIs |
| `lambda` | Lambda function deployment (arm64) |
| `iam` | Roles, policies, log group ARN construction |
| `ses` | Email receiving on `inbox.sh3r4rd.com` |
| `api-gateway` | REST API at `api.sh3r4rd.com` |
| `monitoring` | CloudWatch alarms, budget alerts via SNS |

## DynamoDB Schema

- **Table:** `recruiter_emails` (PROVISIONED billing, 15/15 RCU/WCU)
- **Primary key:** `id` (S) + `received_at` (S)
- **GSI `recruiter-index`:** `recruiter_email` (HASH) + `received_at` (RANGE)
- **GSI `date-index`:** `date_year` (HASH) + `date_day` (RANGE)

## Key Rules

- `snake_case` for all resource names, variables, outputs
- No hardcoded ARNs or account IDs — use data sources or variables
- Default tags via provider `default_tags` block, not per-resource
- GSI: use `key_schema` blocks (not deprecated `hash_key`/`range_key`)
- IAM module constructs log group ARNs from function names (not module outputs) to avoid circular dependencies
- Budget notifications route through SNS (not direct email) for extensibility

## Validation Commands

```bash
terraform -chdir=infra/recruiter-dashboard fmt -check -recursive
terraform -chdir=infra/recruiter-dashboard validate
```
