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

variable "ssm_openai_key_name" {
  description = "SSM Parameter Store path for the OpenAI API key (used to scope IAM permissions)."
  type        = string

  validation {
    condition     = startswith(var.ssm_openai_key_name, "/")
    error_message = "SSM parameter name must begin with a leading slash (e.g. /recruiter-dashboard/openai-api-key)."
  }
}
