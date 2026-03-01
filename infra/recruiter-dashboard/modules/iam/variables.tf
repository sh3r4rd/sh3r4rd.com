variable "project_name" {
  description = "Project name used for resource naming."
  type        = string
}

variable "s3_bucket_arn" {
  description = "ARN of the S3 bucket for raw email storage."
  type        = string
}

variable "dynamodb_table_arn" {
  description = "ARN of the DynamoDB table."
  type        = string
}

variable "dynamodb_gsi_arns" {
  description = "List of DynamoDB GSI ARNs for the API handler to query."
  type        = list(string)
}

variable "email_parser_function_name" {
  description = "Name of the email parser Lambda function (used to construct log group ARN)."
  type        = string
}

variable "api_handler_function_name" {
  description = "Name of the API handler Lambda function (used to construct log group ARN)."
  type        = string
}
