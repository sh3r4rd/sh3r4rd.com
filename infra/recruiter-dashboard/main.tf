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

# ---------------------------------------------------------------------------
# Module: Lambda — Email Parser
# ---------------------------------------------------------------------------
module "lambda_email_parser" {
  source = "./modules/lambda"

  function_name                  = "${var.project_name}-email-parser"
  description                    = "Parses forwarded recruiter emails from S3 and stores data in DynamoDB"
  source_dir                     = "${path.module}/lambda-src/email-parser"
  handler                        = "email_parser.handler.lambda_handler"
  role_arn                       = module.iam.email_parser_role_arn
  memory_size                    = 128
  timeout                        = 30
  reserved_concurrent_executions = 2
  log_retention_days             = var.log_retention_days

  environment_variables = {
    RECRUITER_TABLE = module.dynamodb.table_name
    S3_BUCKET       = module.s3.bucket_name
  }
}

# ---------------------------------------------------------------------------
# Module: Lambda — API Handler
# ---------------------------------------------------------------------------
module "lambda_api_handler" {
  source = "./modules/lambda"

  function_name                  = "${var.project_name}-api-handler"
  description                    = "Serves the recruiter dashboard REST API with anonymized responses"
  source_dir                     = "${path.module}/lambda-src/api-handler"
  handler                        = "api_handler.handler.lambda_handler"
  role_arn                       = module.iam.api_handler_role_arn
  memory_size                    = 128
  timeout                        = 10
  reserved_concurrent_executions = 5
  log_retention_days             = var.log_retention_days

  environment_variables = {
    RECRUITER_TABLE    = module.dynamodb.table_name
    CORS_ALLOW_ORIGIN  = var.cors_allowed_origin
    COMPANY_INDEX_NAME = "company-index"
    DATE_INDEX_NAME    = "date-index"
  }
}

# ---------------------------------------------------------------------------
# Module: IAM — Least-privilege roles for both Lambdas
# ---------------------------------------------------------------------------
module "iam" {
  source = "./modules/iam"

  project_name       = var.project_name
  s3_bucket_arn      = module.s3.bucket_arn
  dynamodb_table_arn = module.dynamodb.table_arn
  dynamodb_gsi_arns = [
    module.dynamodb.company_index_arn,
    module.dynamodb.date_index_arn,
  ]
  email_parser_log_group_arn = module.lambda_email_parser.log_group_arn
  api_handler_log_group_arn  = module.lambda_api_handler.log_group_arn
}
