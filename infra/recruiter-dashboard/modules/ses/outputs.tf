output "domain_verification_token" {
  description = "TXT record value for SES domain verification."
  value       = aws_ses_domain_identity.this.verification_token
}

output "dkim_tokens" {
  description = "DKIM CNAME record tokens for email signing."
  value       = aws_ses_domain_dkim.this.dkim_tokens
}

output "mx_record" {
  description = "MX record value to route inbound email to SES."
  value       = "10 inbound-smtp.${var.aws_region}.amazonaws.com"
}

output "ses_domain" {
  description = "The SES domain name."
  value       = aws_ses_domain_identity.this.domain
}

output "domain_verification_id" {
  description = "The verified SES domain name (resource exists only after verification succeeds)."
  value       = aws_ses_domain_identity_verification.this.id
}
