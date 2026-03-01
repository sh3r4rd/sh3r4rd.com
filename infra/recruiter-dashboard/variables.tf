variable "aws_region" {
  description = "AWS region for all resources. Must support SES email receiving."
  type        = string
  default     = "us-east-1"

  validation {
    condition     = contains(["us-east-1", "us-west-2", "eu-west-1"], var.aws_region)
    error_message = "Region must support SES email receiving: us-east-1, us-west-2, or eu-west-1."
  }
}

variable "project_name" {
  description = "Project name used for resource naming and tagging."
  type        = string
  default     = "recruiter-dashboard"
}

variable "ses_domain" {
  description = "Domain for SES email receiving. Use a subdomain to preserve existing MX records."
  type        = string

  validation {
    condition     = can(regex("^[a-z0-9]([a-z0-9-]*[a-z0-9])?(\\.[a-z0-9]([a-z0-9-]*[a-z0-9])?)+$", var.ses_domain))
    error_message = "Must be a valid domain name."
  }
}

variable "ses_recipient" {
  description = "Email recipient address for the SES receipt rule."
  type        = string

  validation {
    condition     = can(regex("^[^@]+@[^@]+\\.[^@]+$", var.ses_recipient))
    error_message = "Must be a valid email address."
  }
}

variable "cors_allowed_origin" {
  description = "CORS allowed origin for API Gateway responses."
  type        = string

  validation {
    condition     = can(regex("^https?://", var.cors_allowed_origin))
    error_message = "Must be a valid URL starting with http:// or https://."
  }
}

variable "alert_email" {
  description = "Email address for budget and alarm notifications."
  type        = string
  sensitive   = true
}

variable "budget_limit" {
  description = "Monthly budget limit in USD."
  type        = string
  default     = "5.0"
}

variable "email_bucket_name" {
  description = "S3 bucket name for raw email storage. Must be globally unique."
  type        = string

  validation {
    condition     = can(regex("^[a-z0-9][a-z0-9.-]{1,61}[a-z0-9]$", var.email_bucket_name))
    error_message = "Must be a valid S3 bucket name (3-63 chars, lowercase, numbers, hyphens, periods)."
  }
}

variable "dynamodb_table_name" {
  description = "DynamoDB table name for parsed recruiter data."
  type        = string
  default     = "recruiter-emails"
}

variable "log_retention_days" {
  description = "CloudWatch log retention in days."
  type        = number
  default     = 7

  validation {
    condition     = contains([1, 3, 5, 7, 14, 30, 60, 90], var.log_retention_days)
    error_message = "Log retention must be one of: 1, 3, 5, 7, 14, 30, 60, 90 days."
  }
}

variable "s3_lifecycle_expiration_days" {
  description = "Number of days before raw emails are automatically deleted from S3."
  type        = number
  default     = 30
}

variable "default_tags" {
  description = "Default tags applied to all resources."
  type        = map(string)
  default = {
    Project   = "recruiter-dashboard"
    ManagedBy = "terraform"
  }
}
