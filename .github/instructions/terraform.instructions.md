---
applyTo: "infra/**"
---

# Terraform Conventions

## Module Structure

Each module follows a consistent file layout:

- `main.tf` — Resources and data sources
- `variables.tf` — Input variables with `description`, `type`, and `validation` where appropriate
- `outputs.tf` — Output values with `description`

## Resource Naming

- Use `snake_case` for all resource names, variables, and outputs
- Use descriptive resource names that reflect their purpose (e.g., `aws_s3_bucket.email_storage`)

## DynamoDB

- Table level: use `hash_key` and `range_key` attributes
- GSI level: use `key_schema` blocks (not deprecated `hash_key`/`range_key`)

## Tags

- Apply default tags via the provider `default_tags` block, not per-resource
- Only add per-resource tags when they differ from the defaults

## Variables

- Every variable must have a `description`
- Every variable must have an explicit `type`
- Add `validation` blocks for input constraints (e.g., allowed regions, valid email format)
- Use `terraform.tfvars.example` to document default/example values

## Outputs

- Every output must have a `description`

## General

- No hardcoded ARNs or account IDs — use data sources (e.g., `aws_caller_identity`) or variables
- Use `locals` for computed values that are reused
- Keep provider configuration in the root module only
