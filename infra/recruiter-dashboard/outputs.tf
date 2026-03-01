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
