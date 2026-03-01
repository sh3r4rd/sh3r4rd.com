output "table_name" {
  description = "Name of the DynamoDB table."
  value       = aws_dynamodb_table.recruiter_emails.name
}

output "table_arn" {
  description = "ARN of the DynamoDB table."
  value       = aws_dynamodb_table.recruiter_emails.arn
}

output "table_id" {
  description = "ID of the DynamoDB table."
  value       = aws_dynamodb_table.recruiter_emails.id
}

output "company_index_arn" {
  description = "ARN of the company-index GSI."
  value       = "${aws_dynamodb_table.recruiter_emails.arn}/index/company-index"
}

output "date_index_arn" {
  description = "ARN of the date-index GSI."
  value       = "${aws_dynamodb_table.recruiter_emails.arn}/index/date-index"
}
