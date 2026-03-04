package main

import (
	"regexp"
	"strings"
)

// genericEmailDomains lists common free/consumer email providers whose domains
// are not useful for inferring a company name.
var genericEmailDomains = map[string]bool{
	"gmail.com":      true,
	"yahoo.com":      true,
	"outlook.com":    true,
	"hotmail.com":    true,
	"icloud.com":     true,
	"me.com":         true,
	"aol.com":        true,
	"protonmail.com": true,
	"proton.me":      true,
	"live.com":       true,
}

// companyPatterns is an ordered list of (compiled regexp, capture-group index)
// pairs for company name extraction from body text.
var companyPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)i(?:'m| am) a recruiter (?:at|with) ([A-Z][A-Za-z0-9 &.,'-]+?)(?:[,.]|$)`),
	regexp.MustCompile(`(?i)reaching out (?:from|on behalf of) ([A-Z][A-Za-z0-9 &.,'-]+?)(?:[,.]|$)`),
	regexp.MustCompile(`(?i)on behalf of ([A-Z][A-Za-z0-9 &.,'-]+?)(?:[,.]|$)`),
	regexp.MustCompile(`(?i)work(?:ing)? (?:at|for) ([A-Z][A-Za-z0-9 &.,'-]+?)(?:[,.]|$)`),
	regexp.MustCompile(`(?i)representing ([A-Z][A-Za-z0-9 &.,'-]+?)(?:[,.]|$)`),
}

// jobTitlePatterns matches common job-title phrasings.
var jobTitlePatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)regarding the ([A-Za-z ]+?) (?:position|role)\b`),
	regexp.MustCompile(`(?i)about the ([A-Za-z ]+?) (?:position|role)\b`),
	regexp.MustCompile(`(?i)position:\s*([A-Za-z ]+?)(?:[,.\n]|$)`),
	regexp.MustCompile(`(?i)role of ([A-Za-z ]+?)(?:[,.\n]|$)`),
	// Composite titles: (Senior|Staff|Lead) (Software|Data|Cloud|DevOps) (Engineer|Developer|Architect)
	regexp.MustCompile(`(?i)((?:Senior|Staff|Lead|Principal)\s+(?:Software|Data|Cloud|DevOps|ML|AI|Backend|Frontend|Full.Stack)\s+(?:Engineer|Developer|Architect|Scientist))`),
}

// signaturePatterns matches common email sign-off phrasings used to extract
// the recruiter's name.
var signaturePatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)best regards,\s*\n([A-Z][a-z]+(?: [A-Z][a-z]+)+)`),
	regexp.MustCompile(`(?i)kind regards,\s*\n([A-Z][a-z]+(?: [A-Z][a-z]+)+)`),
	regexp.MustCompile(`(?i)thanks(?:,| and regards,)\s*\n([A-Z][a-z]+(?: [A-Z][a-z]+)+)`),
	regexp.MustCompile(`(?i)sincerely,\s*\n([A-Z][a-z]+(?: [A-Z][a-z]+)+)`),
	regexp.MustCompile(`(?i)cheers,\s*\n([A-Z][a-z]+(?: [A-Z][a-z]+)+)`),
}

// phonePattern matches E.164 and common North-American / international formats.
var phonePattern = regexp.MustCompile(
	`(?:\+?1[-.\s]?)?` +
		`\(?([0-9]{3})\)?[-.\s]?` +
		`([0-9]{3})[-.\s]?` +
		`([0-9]{4})`,
)

// ExtractRecruiterData applies regex heuristics to a ParsedEmail and returns
// an ExtractionResult with nil fields for values that could not be extracted.
func ExtractRecruiterData(parsed *ParsedEmail) *ExtractionResult {
	result := &ExtractionResult{}

	result.Company = extractCompany(parsed.Body, parsed.From)
	result.JobTitle = extractJobTitle(parsed.Subject, parsed.Body)
	result.Name = extractName(parsed.From, parsed.Body)
	result.Phone = extractPhone(parsed.Body)

	return result
}

// extractCompany tries each company pattern in order, then falls back to the
// sender's email domain.
func extractCompany(body, from string) *string {
	for _, re := range companyPatterns {
		if m := re.FindStringSubmatch(body); len(m) > 1 {
			v := strings.TrimSpace(m[1])
			return &v
		}
	}

	// Domain fallback: parse the email address.
	domain := domainFromAddress(from)
	if domain != "" && !genericEmailDomains[domain] {
		// Capitalise the first label (e.g., "acme.com" → "Acme").
		label := strings.SplitN(domain, ".", 2)[0]
		v := strings.ToUpper(label[:1]) + label[1:]
		return &v
	}

	return nil
}

// extractJobTitle searches the subject first, then the body.
func extractJobTitle(subject, body string) *string {
	for _, src := range []string{subject, body} {
		for _, re := range jobTitlePatterns {
			if m := re.FindStringSubmatch(src); len(m) > 1 {
				v := strings.TrimSpace(m[1])
				return &v
			}
		}
	}
	return nil
}

// extractName first tries the display name from the From header, then looks
// for signature-block patterns in the body.
func extractName(from, body string) *string {
	// Attempt to parse display name from "Firstname Lastname <email@example.com>".
	if idx := strings.Index(from, "<"); idx > 0 {
		display := strings.TrimSpace(from[:idx])
		display = strings.Trim(display, `"'`)
		if display != "" {
			return &display
		}
	}

	for _, re := range signaturePatterns {
		if m := re.FindStringSubmatch(body); len(m) > 1 {
			v := strings.TrimSpace(m[1])
			return &v
		}
	}

	return nil
}

// extractPhone returns the first phone number found in the body, or nil.
func extractPhone(body string) *string {
	if m := phonePattern.FindString(body); m != "" {
		normalized := normalizePhone(m)
		return &normalized
	}
	return nil
}

// normalizePhone strips non-digit characters and formats as +1XXXXXXXXXX for
// 10-digit North American numbers, or +XXXXXXXXXXX for others.
func normalizePhone(raw string) string {
	digits := regexp.MustCompile(`\D`).ReplaceAllString(raw, "")
	switch len(digits) {
	case 10:
		return "+1" + digits
	case 11:
		return "+" + digits
	default:
		return digits
	}
}

// domainFromAddress extracts the domain portion of an email address.
// It understands "Display Name <user@domain>" and plain "user@domain" forms.
func domainFromAddress(addr string) string {
	// Strip angle-bracket form.
	if start := strings.Index(addr, "<"); start >= 0 {
		end := strings.Index(addr, ">")
		if end > start {
			addr = addr[start+1 : end]
		}
	}
	addr = strings.TrimSpace(addr)
	if at := strings.LastIndex(addr, "@"); at >= 0 {
		return strings.ToLower(strings.TrimSpace(addr[at+1:]))
	}
	return ""
}
