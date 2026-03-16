resource "aws_s3_bucket" "email_storage" {
  bucket = var.bucket_name
}

resource "aws_s3_bucket_server_side_encryption_configuration" "email_storage" {
  bucket = aws_s3_bucket.email_storage.id

  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm = "AES256"
    }
  }
}

resource "aws_s3_bucket_lifecycle_configuration" "email_storage" {
  bucket = aws_s3_bucket.email_storage.id

  rule {
    id     = "expire-raw-emails"
    status = "Enabled"

    filter {
      prefix = "incoming/"
    }

    expiration {
      days = var.lifecycle_expiration_days
    }
  }
}

resource "aws_s3_bucket_public_access_block" "email_storage" {
  bucket = aws_s3_bucket.email_storage.id

  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

# Allow SES to write incoming emails to this bucket
resource "aws_s3_bucket_policy" "allow_ses" {
  bucket = aws_s3_bucket.email_storage.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid       = "AllowSESPut"
        Effect    = "Allow"
        Principal = { Service = "ses.amazonaws.com" }
        Action    = "s3:PutObject"
        Resource  = "${aws_s3_bucket.email_storage.arn}/incoming/*"
        Condition = {
          StringEquals = {
            "AWS:SourceAccount" = var.aws_account_id
          }
          ArnLike = {
            "AWS:SourceArn" = "arn:aws:ses:${var.aws_region}:${var.aws_account_id}:receipt-rule-set/*"
          }
        }
      }
    ]
  })

  depends_on = [aws_s3_bucket_public_access_block.email_storage]
}
