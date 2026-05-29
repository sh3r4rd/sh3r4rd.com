#!/usr/bin/env bash
#
# Terraform format-check and validation for the recruiter-dashboard config.
#
# Uses `terraform init -backend=false` so it needs no AWS credentials and never
# touches remote state — safe to run in CI on every infra change. Resolves its
# own directory, so it works from any working directory (CI or local).
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TF_DIR="$(dirname "$SCRIPT_DIR")"

cd "$TF_DIR"

echo "==> terraform fmt -check -recursive"
terraform fmt -check -recursive

echo "==> terraform init -backend=false"
terraform init -backend=false

echo "==> terraform validate"
terraform validate

echo "==> Terraform validation passed"
