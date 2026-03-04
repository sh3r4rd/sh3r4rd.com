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
# Route53 DNS Records — domain verification, DKIM, and MX
# ---------------------------------------------------------------------------

resource "aws_route53_record" "ses_verification" {
  zone_id = var.hosted_zone_id
  name    = "_amazonses.${var.ses_domain}"
  type    = "TXT"
  ttl     = 600
  records = [aws_ses_domain_identity.this.verification_token]
}

resource "aws_route53_record" "ses_dkim" {
  count   = 3
  zone_id = var.hosted_zone_id
  name    = "${aws_ses_domain_dkim.this.dkim_tokens[count.index]}._domainkey.${var.ses_domain}"
  type    = "CNAME"
  ttl     = 600
  records = ["${aws_ses_domain_dkim.this.dkim_tokens[count.index]}.dkim.amazonses.com"]
}

resource "aws_route53_record" "ses_mx" {
  zone_id = var.hosted_zone_id
  name    = var.ses_domain
  type    = "MX"
  ttl     = 600
  records = ["10 inbound-smtp.${var.aws_region}.amazonaws.com"]
}

resource "aws_ses_domain_identity_verification" "this" {
  domain = aws_ses_domain_identity.this.id

  depends_on = [aws_route53_record.ses_verification]
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
