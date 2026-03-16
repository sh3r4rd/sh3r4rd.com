output "email_parser_role_arn" {
  description = "ARN of the IAM role for the email parser Lambda."
  value       = aws_iam_role.email_parser.arn
}

output "email_parser_role_name" {
  description = "Name of the IAM role for the email parser Lambda."
  value       = aws_iam_role.email_parser.name
}

output "api_handler_role_arn" {
  description = "ARN of the IAM role for the API handler Lambda."
  value       = aws_iam_role.api_handler.arn
}

output "api_handler_role_name" {
  description = "Name of the IAM role for the API handler Lambda."
  value       = aws_iam_role.api_handler.name
}
