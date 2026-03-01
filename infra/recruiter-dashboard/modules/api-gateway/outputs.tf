output "api_endpoint" {
  description = "Full invoke URL for the prod stage (e.g., https://xxx.execute-api.us-east-1.amazonaws.com/prod)."
  value       = aws_api_gateway_stage.prod.invoke_url
}

output "api_id" {
  description = "ID of the REST API."
  value       = aws_api_gateway_rest_api.api.id
}

output "stage_name" {
  description = "Name of the deployed stage."
  value       = aws_api_gateway_stage.prod.stage_name
}
