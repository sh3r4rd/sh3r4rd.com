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
  default     = "https://sh3r4rd.com"
}

variable "throttling_rate_limit" {
  description = "Steady-state requests per second for stage throttling."
  type        = number
  default     = 5
}

variable "throttling_burst_limit" {
  description = "Maximum concurrent request burst for stage throttling."
  type        = number
  default     = 10
}
