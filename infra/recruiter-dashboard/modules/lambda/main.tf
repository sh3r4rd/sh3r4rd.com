# CloudWatch log group (created before Lambda to ensure Lambda doesn't auto-create
# with infinite retention)
resource "aws_cloudwatch_log_group" "function" {
  name              = "/aws/lambda/${var.function_name}"
  retention_in_days = var.log_retention_days
}

# Package the Lambda source code
data "archive_file" "function" {
  type        = "zip"
  source_dir  = var.package_dir
  output_path = "${path.module}/../../.build/${var.function_name}.zip"
}

resource "aws_lambda_function" "function" {
  function_name = var.function_name
  description   = var.description
  role          = var.role_arn

  filename         = data.archive_file.function.output_path
  source_code_hash = data.archive_file.function.output_base64sha256

  handler       = var.handler
  runtime       = var.runtime
  architectures = var.architectures

  memory_size = var.memory_size
  timeout     = var.timeout

  reserved_concurrent_executions = var.reserved_concurrent_executions

  dynamic "environment" {
    for_each = length(var.environment_variables) > 0 ? [1] : []
    content {
      variables = var.environment_variables
    }
  }

  depends_on = [aws_cloudwatch_log_group.function]
}
