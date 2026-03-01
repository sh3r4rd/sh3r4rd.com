variable "project_name" {
  description = "Project name used for resource naming."
  type        = string
}

variable "alert_email" {
  description = "Email address for alarm and budget notifications."
  type        = string
  sensitive   = true

  validation {
    condition     = can(regex("^[^@]+@[^@]+\\.[^@]+$", var.alert_email))
    error_message = "alert_email must be a valid email address (e.g., user@example.com)."
  }
}

variable "email_parser_function_name" {
  description = "Name of the email parser Lambda function (for CloudWatch alarms)."
  type        = string
}

variable "api_handler_function_name" {
  description = "Name of the API handler Lambda function (for CloudWatch alarms)."
  type        = string
}
