# ---------------------------------------------------------------------------
# AWS Budgets — Monthly cost budget with ACTUAL-spend alerts (D-09)
# ---------------------------------------------------------------------------
# The budget was created in the console before it was managed here; the root
# module imports it (see the import block in ../../main.tf). cost_types is
# pinned to the values the console budget already has, so the import does not
# change what the budget measures. Notifications go to the shared alerts SNS
# topic (monitoring module), whose email subscription delivers to alert_email.
resource "aws_budgets_budget" "monthly" {
  name         = var.budget_name
  budget_type  = "COST"
  limit_amount = tostring(var.monthly_budget_usd)
  limit_unit   = "USD"
  time_unit    = "MONTHLY"

  cost_types {
    include_credit             = true
    include_discount           = true
    include_other_subscription = true
    include_recurring          = true
    include_refund             = true
    include_subscription       = true
    include_support            = true
    include_tax                = true
    include_upfront            = true
    use_amortized              = false
    use_blended                = false
  }

  dynamic "notification" {
    for_each = toset(var.alert_thresholds_percent)

    content {
      notification_type         = "ACTUAL"
      comparison_operator       = "GREATER_THAN"
      threshold                 = notification.value
      threshold_type            = "PERCENTAGE"
      subscriber_sns_topic_arns = [var.sns_topic_arn]
    }
  }

  # Budgets validates that it can publish to the topic when the notification
  # is saved, so the topic policy must exist first.
  depends_on = [aws_sns_topic_policy.alerts]
}

# ---------------------------------------------------------------------------
# SNS topic policy — Let AWS Budgets and Cost Anomaly Detection publish
# ---------------------------------------------------------------------------
# This resource owns the alerts topic's whole access policy, so it restates
# the default owner statement alongside the two service grants. Keep any
# future statements for this topic here rather than in a second policy.
data "aws_iam_policy_document" "alerts_topic" {
  statement {
    sid    = "DefaultOwnerAccess"
    effect = "Allow"
    actions = [
      "SNS:GetTopicAttributes",
      "SNS:SetTopicAttributes",
      "SNS:AddPermission",
      "SNS:RemovePermission",
      "SNS:DeleteTopic",
      "SNS:Subscribe",
      "SNS:ListSubscriptionsByTopic",
      "SNS:Publish",
    ]
    resources = [var.sns_topic_arn]

    principals {
      type        = "AWS"
      identifiers = ["*"]
    }

    condition {
      test     = "StringEquals"
      variable = "AWS:SourceOwner"
      values   = [var.aws_account_id]
    }
  }

  statement {
    sid       = "AllowBudgetsPublish"
    effect    = "Allow"
    actions   = ["SNS:Publish"]
    resources = [var.sns_topic_arn]

    principals {
      type        = "Service"
      identifiers = ["budgets.amazonaws.com"]
    }

    condition {
      test     = "StringEquals"
      variable = "aws:SourceAccount"
      values   = [var.aws_account_id]
    }

    condition {
      test     = "ArnLike"
      variable = "aws:SourceArn"
      values   = ["arn:aws:budgets::${var.aws_account_id}:*"]
    }
  }

  statement {
    sid       = "AllowCostAnomalyDetectionPublish"
    effect    = "Allow"
    actions   = ["SNS:Publish"]
    resources = [var.sns_topic_arn]

    principals {
      type        = "Service"
      identifiers = ["costalerts.amazonaws.com"]
    }

    condition {
      test     = "StringEquals"
      variable = "aws:SourceAccount"
      values   = [var.aws_account_id]
    }
  }
}

resource "aws_sns_topic_policy" "alerts" {
  arn    = var.sns_topic_arn
  policy = data.aws_iam_policy_document.alerts_topic.json
}

# ---------------------------------------------------------------------------
# Cost category — Groups Amazon Bedrock spend under one value
# ---------------------------------------------------------------------------
# Customer managed (CUSTOM) anomaly monitors cannot filter on the SERVICE
# dimension directly; they accept linked accounts, cost allocation tags or a
# cost category value. This cost category is the supported way to scope a
# CUSTOM monitor to Bedrock. Cost category rules match on SERVICE_CODE, not on
# the SERVICE display name.
resource "aws_ce_cost_category" "bedrock" {
  name         = var.bedrock_cost_category_name
  rule_version = "CostCategoryExpression.v1"

  rule {
    value = var.bedrock_cost_category_value

    rule {
      dimension {
        key           = "SERVICE_CODE"
        values        = var.bedrock_service_codes
        match_options = ["EQUALS"]
      }
    }
  }
}

# ---------------------------------------------------------------------------
# Cost Anomaly Detection — CUSTOM monitor scoped to Bedrock
# ---------------------------------------------------------------------------
resource "aws_ce_anomaly_monitor" "bedrock" {
  name         = var.anomaly_monitor_name
  monitor_type = "CUSTOM"

  monitor_specification = jsonencode({
    CostCategories = {
      Key    = aws_ce_cost_category.bedrock.name
      Values = [var.bedrock_cost_category_value]
    }
  })
}

# IMMEDIATE frequency requires an SNS subscriber (email subscribers only get
# DAILY or WEEKLY summaries). It reuses the alerts topic, so anomalies reach
# alert_email as soon as they are detected.
resource "aws_ce_anomaly_subscription" "bedrock" {
  name             = var.anomaly_subscription_name
  frequency        = "IMMEDIATE"
  monitor_arn_list = [aws_ce_anomaly_monitor.bedrock.arn]

  subscriber {
    type    = "SNS"
    address = var.sns_topic_arn
  }

  threshold_expression {
    dimension {
      key           = "ANOMALY_TOTAL_IMPACT_ABSOLUTE"
      match_options = ["GREATER_THAN_OR_EQUAL"]
      values        = [tostring(var.anomaly_threshold_usd)]
    }
  }

  depends_on = [aws_sns_topic_policy.alerts]
}
