variable "bucket_name" {
  description = "Name of the S3 bucket for raw email storage."
  type        = string
}

variable "lifecycle_expiration_days" {
  description = "Number of days before objects under incoming/ are automatically deleted. archive/ never expires."
  type        = number
  default     = 30
}

variable "aws_account_id" {
  description = "AWS account ID for the bucket policy."
  type        = string
}

variable "aws_region" {
  description = "AWS region for the SES service principal."
  type        = string
}

variable "archive_kms_alias" {
  description = "Alias for the customer-managed KMS key that encrypts archive/ and evals/."
  type        = string

  validation {
    condition     = startswith(var.archive_kms_alias, "alias/") && !startswith(var.archive_kms_alias, "alias/aws/")
    error_message = "The alias must start with \"alias/\" and must not use the reserved \"alias/aws/\" prefix."
  }
}
