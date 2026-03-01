variable "function_name" {
  description = "Name of the Lambda function."
  type        = string
}

variable "description" {
  description = "Description of the Lambda function."
  type        = string
  default     = ""
}

variable "source_dir" {
  description = "Path to the directory containing the Lambda source code."
  type        = string
}

variable "handler" {
  description = "Lambda function handler (e.g., 'email_parser.handler.lambda_handler')."
  type        = string
}

variable "runtime" {
  description = "Lambda runtime."
  type        = string
  default     = "python3.12"
}

variable "architectures" {
  description = "Lambda instruction set architecture."
  type        = list(string)
  default     = ["arm64"]
}

variable "memory_size" {
  description = "Amount of memory in MB for the Lambda function."
  type        = number
  default     = 128
}

variable "timeout" {
  description = "Timeout in seconds for the Lambda function."
  type        = number
  default     = 30
}

variable "reserved_concurrent_executions" {
  description = "Reserved concurrent executions for the Lambda function. Use -1 for unreserved."
  type        = number
  default     = -1
}

variable "role_arn" {
  description = "ARN of the IAM role for the Lambda function."
  type        = string
}

variable "environment_variables" {
  description = "Environment variables for the Lambda function."
  type        = map(string)
  default     = {}
}

variable "log_retention_days" {
  description = "CloudWatch log group retention in days."
  type        = number
  default     = 7
}
