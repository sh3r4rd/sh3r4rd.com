# ---------------------------------------------------------------------------
# Data sources for constructing log group ARNs from function names.
# This avoids a circular dependency between IAM and Lambda modules.
# ---------------------------------------------------------------------------

data "aws_region" "current" {}
data "aws_caller_identity" "current" {}

locals {
  email_parser_log_group_arn = "arn:aws:logs:${data.aws_region.current.id}:${data.aws_caller_identity.current.account_id}:log-group:/aws/lambda/${var.email_parser_function_name}"
  api_handler_log_group_arn  = "arn:aws:logs:${data.aws_region.current.id}:${data.aws_caller_identity.current.account_id}:log-group:/aws/lambda/${var.api_handler_function_name}"
}

# ---------------------------------------------------------------------------
# Email Parser Lambda Role
# Permissions: S3 read (raw emails), DynamoDB write (parsed data), CloudWatch logs
# ---------------------------------------------------------------------------

data "aws_iam_policy_document" "lambda_assume_role" {
  statement {
    actions = ["sts:AssumeRole"]
    principals {
      type        = "Service"
      identifiers = ["lambda.amazonaws.com"]
    }
  }
}

resource "aws_iam_role" "email_parser" {
  name               = "${var.project_name}-email-parser-role"
  assume_role_policy = data.aws_iam_policy_document.lambda_assume_role.json
}

data "aws_iam_policy_document" "email_parser" {
  # S3: Read raw emails only
  statement {
    sid     = "S3ReadRawEmails"
    actions = ["s3:GetObject"]
    resources = [
      "${var.s3_bucket_arn}/incoming/*"
    ]
  }

  # DynamoDB: Write parsed data only
  statement {
    sid = "DynamoDBWriteParsedData"
    actions = [
      "dynamodb:PutItem",
      "dynamodb:UpdateItem",
    ]
    resources = [var.dynamodb_table_arn]
  }

  # CloudWatch Logs
  statement {
    sid = "CloudWatchLogs"
    actions = [
      "logs:CreateLogStream",
      "logs:PutLogEvents",
    ]
    resources = [
      local.email_parser_log_group_arn,
      "${local.email_parser_log_group_arn}:*",
    ]
  }
}

resource "aws_iam_role_policy" "email_parser" {
  name   = "${var.project_name}-email-parser-policy"
  role   = aws_iam_role.email_parser.id
  policy = data.aws_iam_policy_document.email_parser.json
}

# ---------------------------------------------------------------------------
# API Handler Lambda Role
# Permissions: DynamoDB read-only (table + GSIs), CloudWatch logs
# ---------------------------------------------------------------------------

resource "aws_iam_role" "api_handler" {
  name               = "${var.project_name}-api-handler-role"
  assume_role_policy = data.aws_iam_policy_document.lambda_assume_role.json
}

data "aws_iam_policy_document" "api_handler" {
  # DynamoDB: Read-only access to table and GSIs
  statement {
    sid = "DynamoDBReadOnly"
    actions = [
      "dynamodb:GetItem",
      "dynamodb:Query",
      "dynamodb:Scan",
    ]
    resources = concat(
      [var.dynamodb_table_arn],
      var.dynamodb_gsi_arns
    )
  }

  # CloudWatch Logs
  statement {
    sid = "CloudWatchLogs"
    actions = [
      "logs:CreateLogStream",
      "logs:PutLogEvents",
    ]
    resources = [
      local.api_handler_log_group_arn,
      "${local.api_handler_log_group_arn}:*",
    ]
  }
}

resource "aws_iam_role_policy" "api_handler" {
  name   = "${var.project_name}-api-handler-policy"
  role   = aws_iam_role.api_handler.id
  policy = data.aws_iam_policy_document.api_handler.json
}
