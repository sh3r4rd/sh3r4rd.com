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

output "custom_domain_url" {
  description = "HTTPS URL of the custom domain fronting the prod stage."
  value       = "https://${aws_api_gateway_domain_name.api.domain_name}"
}

output "custom_domain_regional_target" {
  description = "Regional domain name the custom domain alias points to (for diagnostics)."
  value       = aws_api_gateway_domain_name.api.regional_domain_name
}
