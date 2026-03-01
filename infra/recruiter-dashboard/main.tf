terraform {
  required_version = ">= 1.5.0"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = ">= 5.31.0"
    }
  }

  # Local state for single-developer project.
  # To migrate to S3 backend in the future, uncomment the block below:
  #
  # backend "s3" {
  #   bucket         = "sh3r4rd-terraform-state"
  #   key            = "recruiter-dashboard/terraform.tfstate"
  #   region         = "us-east-1"
  #   dynamodb_table = "terraform-locks"
  #   encrypt        = true
  # }
}

provider "aws" {
  region = var.aws_region

  default_tags {
    tags = var.default_tags
  }
}

data "aws_caller_identity" "current" {}

# ---------------------------------------------------------------------------
# Module: S3 — Raw email storage
# ---------------------------------------------------------------------------
module "s3" {
  source = "./modules/s3"

  bucket_name               = var.email_bucket_name
  lifecycle_expiration_days = var.s3_lifecycle_expiration_days
  aws_account_id            = data.aws_caller_identity.current.account_id
  aws_region                = var.aws_region
}

# ---------------------------------------------------------------------------
# Module: DynamoDB — Parsed recruiter data
# ---------------------------------------------------------------------------
module "dynamodb" {
  source = "./modules/dynamodb"

  table_name = var.dynamodb_table_name
}
