variable "bucket_name" {
  description = "Name of the S3 bucket for raw email storage."
  type        = string
}

variable "lifecycle_expiration_days" {
  description = "Number of days before objects are automatically deleted."
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
