output "domain_verification_token" {
  description = "TXT record value for _amazonses.inbox.sh3r4rd.com domain verification."
  value       = aws_ses_domain_identity.this.verification_token
}

output "dkim_tokens" {
  description = "DKIM CNAME record tokens for email signing."
  value       = aws_ses_domain_dkim.this.dkim_tokens
}

output "mx_record" {
  description = "MX record value to route inbound email to SES."
  value       = "10 inbound-smtp.us-east-1.amazonaws.com"
}

output "ses_domain" {
  description = "The SES domain name."
  value       = aws_ses_domain_identity.this.domain
}
