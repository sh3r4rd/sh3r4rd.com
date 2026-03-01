output "s3_bucket_name" {
  description = "Name of the S3 bucket for raw email storage."
  value       = module.s3.bucket_name
}

output "s3_bucket_arn" {
  description = "ARN of the S3 bucket for raw email storage."
  value       = module.s3.bucket_arn
}

output "dynamodb_table_name" {
  description = "Name of the DynamoDB table for parsed recruiter data."
  value       = module.dynamodb.table_name
}

output "dynamodb_table_arn" {
  description = "ARN of the DynamoDB table."
  value       = module.dynamodb.table_arn
}

output "email_parser_function_name" {
  description = "Name of the email parser Lambda function."
  value       = module.lambda_email_parser.function_name
}

output "email_parser_function_arn" {
  description = "ARN of the email parser Lambda function."
  value       = module.lambda_email_parser.function_arn
}

output "api_handler_function_name" {
  description = "Name of the API handler Lambda function."
  value       = module.lambda_api_handler.function_name
}

output "api_handler_invoke_arn" {
  description = "Invoke ARN of the API handler Lambda (for API Gateway)."
  value       = module.lambda_api_handler.invoke_arn
}
