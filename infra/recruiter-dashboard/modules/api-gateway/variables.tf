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
