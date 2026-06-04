variable "api_name" {
  description = "Name of the REST API."
  type        = string
}

variable "api_handler_invoke_arn" {
  description = "Invoke ARN of the API handler Lambda function."
  type        = string
}

variable "api_handler_function_name" {
  description = "Name of the API handler Lambda function."
  type        = string
}

variable "cors_allowed_origin" {
  description = "CORS allowed origin for API Gateway responses."
  type        = string

  validation {
    condition     = can(regex("^https?://", var.cors_allowed_origin))
    error_message = "Must be a valid URL starting with http:// or https://."
  }
}

variable "custom_domain_name" {
  description = "Custom domain name for the REST API (e.g. dashboard-api.sh3r4rd.com)."
  type        = string

  validation {
    condition     = can(regex("^[a-z0-9]([a-z0-9-]*[a-z0-9])?(\\.[a-z0-9]([a-z0-9-]*[a-z0-9])?)+$", var.custom_domain_name))
    error_message = "Must be a valid domain name."
  }
}

variable "hosted_zone_id" {
  description = "Route53 hosted zone ID for ACM DNS validation and the custom domain alias record."
  type        = string

  validation {
    condition     = can(regex("^Z[A-Z0-9]+$", var.hosted_zone_id))
    error_message = "Must be a valid Route53 hosted zone ID (starts with Z, alphanumeric)."
  }
}

variable "throttling_rate_limit" {
  description = "Steady-state requests per second for stage throttling."
  type        = number
  default     = 5

  validation {
    condition     = var.throttling_rate_limit > 0
    error_message = "Throttling rate limit must be a positive number."
  }
}

variable "throttling_burst_limit" {
  description = "Maximum concurrent request burst for stage throttling."
  type        = number
  default     = 10

  validation {
    condition     = var.throttling_burst_limit > 0
    error_message = "Throttling burst limit must be a positive number."
  }
}
