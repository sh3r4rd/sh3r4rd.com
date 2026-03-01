# ---------------------------------------------------------------------------
# SES Domain Identity & DKIM
# ---------------------------------------------------------------------------

resource "aws_ses_domain_identity" "this" {
  domain = var.ses_domain
}

resource "aws_ses_domain_dkim" "this" {
  domain = aws_ses_domain_identity.this.domain
}

# ---------------------------------------------------------------------------
# SES Receipt Rule Set
# ---------------------------------------------------------------------------

resource "aws_ses_receipt_rule_set" "this" {
  rule_set_name = "recruiter-dashboard-rules"
}

resource "aws_ses_active_receipt_rule_set" "this" {
  rule_set_name = aws_ses_receipt_rule_set.this.rule_set_name

  depends_on = [aws_ses_receipt_rule.store_and_parse]
}

# ---------------------------------------------------------------------------
# Lambda Permission — must exist BEFORE the receipt rule references it
# ---------------------------------------------------------------------------

resource "aws_lambda_permission" "ses_invoke" {
  statement_id   = "AllowSESInvoke"
  action         = "lambda:InvokeFunction"
  function_name  = var.email_parser_function_arn
  principal      = "ses.amazonaws.com"
  source_account = data.aws_caller_identity.current.account_id
  source_arn     = aws_ses_receipt_rule_set.this.arn
}

data "aws_caller_identity" "current" {}

# ---------------------------------------------------------------------------
# SES Receipt Rule — store to S3, then invoke Lambda
# ---------------------------------------------------------------------------

resource "aws_ses_receipt_rule" "store_and_parse" {
  name          = "store-and-parse-recruiter-email"
  rule_set_name = aws_ses_receipt_rule_set.this.rule_set_name
  recipients    = [var.ses_recipient]
  enabled       = true
  scan_enabled  = true

  s3_action {
    position          = 1
    bucket_name       = var.s3_bucket_name
    object_key_prefix = var.s3_key_prefix
  }

  lambda_action {
    position        = 2
    function_arn    = var.email_parser_function_arn
    invocation_type = "Event"
  }

  depends_on = [aws_lambda_permission.ses_invoke]
}
