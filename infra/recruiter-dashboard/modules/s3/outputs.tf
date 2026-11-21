output "bucket_name" {
  description = "Name of the S3 bucket."
  value       = aws_s3_bucket.email_storage.id
}

output "bucket_arn" {
  description = "ARN of the S3 bucket."
  value       = aws_s3_bucket.email_storage.arn
}

output "archive_kms_key_arn" {
  description = "ARN of the KMS key that encrypts archive/ and evals/. Writers must pass this full ARN."
  value       = aws_kms_key.archive.arn
}

output "archive_kms_alias" {
  description = "Alias of the archive KMS key."
  value       = aws_kms_alias.archive.name
}
