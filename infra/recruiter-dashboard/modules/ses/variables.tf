variable "ses_domain" {
  description = "Domain for the SES domain identity."
  type        = string
  default     = "inbox.sh3r4rd.com"
}

variable "ses_recipient" {
  description = "Email recipient address for the SES receipt rule."
  type        = string
  default     = "recruiters@inbox.sh3r4rd.com"
}

variable "s3_bucket_name" {
  description = "Name of the S3 bucket for storing raw emails."
  type        = string
}

variable "email_parser_function_arn" {
  description = "ARN of the email parser Lambda function to invoke on receipt."
  type        = string
}

variable "s3_key_prefix" {
  description = "S3 key prefix for stored emails."
  type        = string
  default     = "incoming/"
}
