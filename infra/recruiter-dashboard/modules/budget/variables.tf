variable "sns_topic_arn" {
  description = "ARN of the shared alerts SNS topic (monitoring module). Budget and anomaly alerts publish here; its email subscription delivers them to alert_email. This module manages the topic's access policy."
  type        = string

  validation {
    condition     = can(regex("^arn:aws[a-z-]*:sns:[a-z0-9-]+:[0-9]{12}:[A-Za-z0-9_-]+$", var.sns_topic_arn))
    error_message = "sns_topic_arn must be a valid SNS topic ARN."
  }
}

variable "aws_account_id" {
  description = "AWS account ID, used to scope the SNS topic policy to this account's budgets and anomaly alerts."
  type        = string

  validation {
    condition     = can(regex("^[0-9]{12}$", var.aws_account_id))
    error_message = "aws_account_id must be a 12-digit AWS account ID."
  }
}

variable "budget_name" {
  description = "Name of the monthly cost budget. Defaults to the name of the console-created budget the root module imports; changing it forces a new budget."
  type        = string
  default     = "Monthly Budget"

  validation {
    condition     = length(var.budget_name) > 0 && length(var.budget_name) <= 100 && !can(regex("[:\\\\]", var.budget_name))
    error_message = "budget_name must be 1-100 characters and must not contain ':' or '\\'."
  }
}

variable "monthly_budget_usd" {
  description = "Monthly cost budget in USD (D-09)."
  type        = number
  default     = 10

  validation {
    condition     = var.monthly_budget_usd > 0
    error_message = "monthly_budget_usd must be greater than 0."
  }
}

variable "alert_thresholds_percent" {
  description = "Percentages of the monthly budget at which an ACTUAL-spend alert is emailed."
  type        = list(number)
  default     = [50, 80, 100]

  validation {
    condition     = length(var.alert_thresholds_percent) > 0 && alltrue([for t in var.alert_thresholds_percent : t > 0 && t <= 1000])
    error_message = "alert_thresholds_percent must be a non-empty list of percentages between 0 (exclusive) and 1000."
  }

  validation {
    condition     = length(var.alert_thresholds_percent) == length(distinct(var.alert_thresholds_percent))
    error_message = "alert_thresholds_percent must not contain duplicates."
  }
}

variable "bedrock_cost_category_name" {
  description = "Name of the cost category that groups Bedrock spend for the anomaly monitor."
  type        = string
  default     = "bedrock-spend"

  validation {
    condition     = can(regex("^[A-Za-z0-9][A-Za-z0-9 _-]{0,48}[A-Za-z0-9]$", var.bedrock_cost_category_name))
    error_message = "bedrock_cost_category_name must be 2-50 characters of letters, digits, spaces, '-' or '_', with no leading or trailing space."
  }
}

variable "bedrock_cost_category_value" {
  description = "Cost category value assigned to Bedrock spend; the anomaly monitor watches this value."
  type        = string
  default     = "Amazon Bedrock"

  validation {
    condition     = can(regex("^[A-Za-z0-9][A-Za-z0-9 _-]{0,48}[A-Za-z0-9]$", var.bedrock_cost_category_value))
    error_message = "bedrock_cost_category_value must be 2-50 characters of letters, digits, spaces, '-' or '_', with no leading or trailing space."
  }
}

variable "bedrock_service_codes" {
  description = "Cost Explorer SERVICE_CODE values counted as Bedrock spend. AmazonBedrock covers Amazon (Nova) and other first-party-billed models; third-party models such as Claude bill under their own per-model service, so add each one's SERVICE_CODE once it has spend."
  type        = list(string)
  default     = ["AmazonBedrock"]

  validation {
    condition     = length(var.bedrock_service_codes) > 0 && alltrue([for c in var.bedrock_service_codes : length(trimspace(c)) > 0])
    error_message = "bedrock_service_codes must contain at least one non-empty service code."
  }
}

variable "anomaly_monitor_name" {
  description = "Name of the Cost Anomaly Detection monitor for Bedrock."
  type        = string
  default     = "bedrock-spend"

  validation {
    condition     = length(var.anomaly_monitor_name) > 0 && length(var.anomaly_monitor_name) <= 1024
    error_message = "anomaly_monitor_name must be 1-1024 characters."
  }
}

variable "anomaly_subscription_name" {
  description = "Name of the Cost Anomaly Detection alert subscription for Bedrock."
  type        = string
  default     = "bedrock-spend-alerts"

  validation {
    condition     = length(var.anomaly_subscription_name) > 0 && length(var.anomaly_subscription_name) <= 1024
    error_message = "anomaly_subscription_name must be 1-1024 characters."
  }
}

variable "anomaly_threshold_usd" {
  description = "Minimum total cost impact in USD of a Bedrock anomaly before it is emailed."
  type        = number
  default     = 2

  validation {
    condition     = var.anomaly_threshold_usd > 0
    error_message = "anomaly_threshold_usd must be greater than 0."
  }
}
