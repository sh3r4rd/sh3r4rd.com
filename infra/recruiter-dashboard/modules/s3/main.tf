resource "aws_s3_bucket" "email_storage" {
  bucket = var.bucket_name
}

# ---------------------------------------------------------------------------
# Archive KMS key (D-06)
#
# Customer-managed key for the durable prefixes, archive/ and evals/. The bucket
# default stays AES256 so SES delivery to incoming/ never depends on this key;
# every archive/ and evals/ write names the key explicitly, and the bucket policy
# below denies writes to those prefixes that don't.
# ---------------------------------------------------------------------------
data "aws_iam_policy_document" "archive_key" {
  # Delegate access control to IAM, so only principals whose IAM policies grant
  # it (the email-parser role, account administrators) can use the key.
  statement {
    sid       = "EnableIAMPolicies"
    actions   = ["kms:*"]
    resources = ["*"]

    principals {
      type        = "AWS"
      identifiers = ["arn:aws:iam::${var.aws_account_id}:root"]
    }
  }

  # Cryptographic use only through S3 in this region...
  statement {
    sid    = "DenyUseOutsideS3"
    effect = "Deny"
    actions = [
      "kms:Encrypt",
      "kms:Decrypt",
      "kms:ReEncrypt*",
      "kms:GenerateDataKey*",
    ]
    resources = ["*"]

    principals {
      type        = "*"
      identifiers = ["*"]
    }

    condition {
      test     = "StringNotEquals"
      variable = "kms:ViaService"
      values   = ["s3.${var.aws_region}.amazonaws.com"]
    }
  }

  # ...and only for objects in this bucket. S3 uses the object ARN as the
  # encryption context when S3 Bucket Keys are off, which they are.
  statement {
    sid    = "DenyUseForOtherBuckets"
    effect = "Deny"
    actions = [
      "kms:Encrypt",
      "kms:Decrypt",
      "kms:ReEncrypt*",
      "kms:GenerateDataKey*",
    ]
    resources = ["*"]

    principals {
      type        = "*"
      identifiers = ["*"]
    }

    condition {
      test     = "StringNotLike"
      variable = "kms:EncryptionContext:aws:s3:arn"
      values   = ["${aws_s3_bucket.email_storage.arn}/*"]
    }
  }
}

resource "aws_kms_key" "archive" {
  description             = "Encrypts the durable archive/ and evals/ prefixes of ${var.bucket_name}"
  deletion_window_in_days = 30
  enable_key_rotation     = true
  policy                  = data.aws_iam_policy_document.archive_key.json

  # Destroying the key makes every archived email unreadable.
  lifecycle {
    prevent_destroy = true
  }
}

resource "aws_kms_alias" "archive" {
  name          = var.archive_kms_alias
  target_key_id = aws_kms_key.archive.key_id
}

resource "aws_s3_bucket_server_side_encryption_configuration" "email_storage" {
  bucket = aws_s3_bucket.email_storage.id

  rule {
    # Keep blocking uploads that use customer-provided keys (SSE-C). AWS set this on
    # the live bucket; declaring it stops every plan from proposing to remove it (#98).
    # Requires AWS provider >= 6.22.0.
    blocked_encryption_types = ["SSE-C"]

    # S3 Bucket Keys stay off. The archive KMS key policy and the parser's IAM grant
    # match on the object-ARN encryption context, which a Bucket Key would replace
    # with the bucket ARN, breaking every archive write. The archive writers also send
    # BucketKeyEnabled=false on each request.
    bucket_key_enabled = false

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

  # archive/ is kept indefinitely: no expiration, only a move to Intelligent-Tiering.
  # With the bucket's default 128 KB minimum transition size, objects under 128 KB
  # (most emails) stay in S3 Standard, which is what Intelligent-Tiering would charge
  # for them anyway, since it doesn't monitor or tier objects under 128 KB.
  rule {
    id     = "archive-intelligent-tiering"
    status = "Enabled"

    filter {
      prefix = "archive/"
    }

    transition {
      days          = 0
      storage_class = "INTELLIGENT_TIERING"
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

# Allow SES to write incoming emails to this bucket, and require the archive key
# for every write to the durable prefixes.
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
      },
      {
        # A missing header also matches StringNotEquals, so writes that rely on the
        # bucket default (AES256) are denied too. The value must be the full key ARN.
        # PutObject and CopyObject send the header; UploadPart doesn't, so multipart
        # uploads into these prefixes are denied as well. Single-part writes cover
        # objects up to 5 GB.
        Sid       = "DenyArchiveWritesWithoutArchiveKey"
        Effect    = "Deny"
        Principal = "*"
        Action    = "s3:PutObject"
        Resource = [
          "${aws_s3_bucket.email_storage.arn}/archive/*",
          "${aws_s3_bucket.email_storage.arn}/evals/*",
        ]
        Condition = {
          StringNotEquals = {
            "s3:x-amz-server-side-encryption-aws-kms-key-id" = aws_kms_key.archive.arn
          }
        }
      },
      {
        # archive/raw/ holds the only copy of each email once incoming/ expires again,
        # and the bucket isn't versioned. Deleting from it requires removing this
        # statement first. Lifecycle actions aren't subject to bucket policies, and
        # archive/ has no expiration anyway. archive/clean/ is regenerable, so it isn't
        # covered.
        Sid       = "DenyRawArchiveDeletes"
        Effect    = "Deny"
        Principal = "*"
        Action    = ["s3:DeleteObject", "s3:DeleteObjectVersion"]
        Resource  = "${aws_s3_bucket.email_storage.arn}/archive/raw/*"
      }
    ]
  })

  depends_on = [aws_s3_bucket_public_access_block.email_storage]
}
