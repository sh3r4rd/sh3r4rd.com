# ---------------------------------------------------------------------------
# SNS Topic — Alert notifications
# ---------------------------------------------------------------------------
resource "aws_sns_topic" "alerts" {
  name = "${var.project_name}-alerts"
}

resource "aws_sns_topic_subscription" "email" {
  topic_arn = aws_sns_topic.alerts.arn
  protocol  = "email"
  endpoint  = var.alert_email
}

# ---------------------------------------------------------------------------
# CloudWatch Alarm — Email parser errors
# ---------------------------------------------------------------------------
resource "aws_cloudwatch_metric_alarm" "email_parser_errors" {
  alarm_name          = "${var.project_name}-email-parser-errors"
  alarm_description   = "Triggers when email parser Lambda has more than 5 errors in 1 hour"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = 1
  metric_name         = "Errors"
  namespace           = "AWS/Lambda"
  period              = 3600
  statistic           = "Sum"
  threshold           = 5
  treat_missing_data  = "notBreaching"

  dimensions = {
    FunctionName = var.email_parser_function_name
  }

  alarm_actions = [aws_sns_topic.alerts.arn]
  ok_actions    = [aws_sns_topic.alerts.arn]
}

# ---------------------------------------------------------------------------
# CloudWatch Alarm — Email loop detection (invocation spike)
# ---------------------------------------------------------------------------
resource "aws_cloudwatch_metric_alarm" "email_loop_detection" {
  alarm_name          = "${var.project_name}-email-loop-detection"
  alarm_description   = "Triggers when email parser Lambda exceeds 100 invocations in 1 hour (potential email loop)"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = 1
  metric_name         = "Invocations"
  namespace           = "AWS/Lambda"
  period              = 3600
  statistic           = "Sum"
  threshold           = 100
  treat_missing_data  = "notBreaching"

  dimensions = {
    FunctionName = var.email_parser_function_name
  }

  alarm_actions = [aws_sns_topic.alerts.arn]
  ok_actions    = [aws_sns_topic.alerts.arn]
}

# ---------------------------------------------------------------------------
# CloudWatch Alarm — Email parse failures (custom metric)
# ---------------------------------------------------------------------------
resource "aws_cloudwatch_metric_alarm" "email_parse_failures" {
  alarm_name          = "${var.project_name}-email-parse-failures"
  alarm_description   = "Triggers when email parser has more than 3 parse failures in 1 hour"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = 1
  metric_name         = "ParseFailures"
  namespace           = "RecruiterDashboard"
  period              = 3600
  statistic           = "Sum"
  threshold           = 3
  treat_missing_data  = "notBreaching"

  dimensions = {
    FunctionName = var.email_parser_function_name
  }

  alarm_actions = [aws_sns_topic.alerts.arn]
  ok_actions    = [aws_sns_topic.alerts.arn]
}

# ---------------------------------------------------------------------------
# CloudWatch Alarm — API handler errors
# ---------------------------------------------------------------------------
resource "aws_cloudwatch_metric_alarm" "api_handler_errors" {
  alarm_name          = "${var.project_name}-api-handler-errors"
  alarm_description   = "Triggers when API handler Lambda has more than 10 errors in 1 hour"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = 1
  metric_name         = "Errors"
  namespace           = "AWS/Lambda"
  period              = 3600
  statistic           = "Sum"
  threshold           = 10
  treat_missing_data  = "notBreaching"

  dimensions = {
    FunctionName = var.api_handler_function_name
  }

  alarm_actions = [aws_sns_topic.alerts.arn]
  ok_actions    = [aws_sns_topic.alerts.arn]
}
