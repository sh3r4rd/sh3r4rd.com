variable "ses_domain" {
  description = "Domain for the SES domain identity."
  type        = string
}

variable "ses_recipient" {
  description = "Email recipient address for the SES receipt rule."
  type        = string
}

variable "aws_region" {
  description = "AWS region for SES inbound SMTP endpoint."
  type        = string
}

variable "s3_bucket_name" {
  description = "Name of the S3 bucket for storing raw emails."
  type        = string
}

variable "email_parser_function_arn" {
  description = "ARN of the email parser Lambda function to invoke on receipt."
  type        = string
}

variable "hosted_zone_id" {
  description = "Route53 hosted zone ID for DNS record creation."
  type        = string
}

variable "s3_key_prefix" {
  description = "S3 key prefix for stored emails."
  type        = string
  default     = "incoming/"
}
