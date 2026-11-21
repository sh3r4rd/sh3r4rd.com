output "budget_name" {
  description = "Name of the monthly cost budget."
  value       = aws_budgets_budget.monthly.name
}

output "budget_arn" {
  description = "ARN of the monthly cost budget."
  value       = aws_budgets_budget.monthly.arn
}

output "bedrock_cost_category_arn" {
  description = "ARN of the cost category that groups Bedrock spend."
  value       = aws_ce_cost_category.bedrock.arn
}

output "anomaly_monitor_arn" {
  description = "ARN of the Bedrock Cost Anomaly Detection monitor."
  value       = aws_ce_anomaly_monitor.bedrock.arn
}

output "anomaly_subscription_arn" {
  description = "ARN of the Bedrock anomaly alert subscription."
  value       = aws_ce_anomaly_subscription.bedrock.arn
}
