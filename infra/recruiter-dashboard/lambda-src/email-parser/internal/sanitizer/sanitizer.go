package sanitizer

import (
	"html"
	"regexp"
	"strings"
	"unicode"
)

var (
	scriptRe     = regexp.MustCompile(`(?is)<script[^>]*>.*?</script>`)
	styleRe      = regexp.MustCompile(`(?is)<style[^>]*>.*?</style>`)
	blockTagRe   = regexp.MustCompile(`(?i)<(?:br|p|div|tr|li|h[1-6])[^>]*/??>`)
	tagRe        = regexp.MustCompile(`<[^>]*>`)
	whitespaceRe = regexp.MustCompile(`\s{2,}`)
	blankLinesRe = regexp.MustCompile(`\n{3,}`)

	// Phone patterns: US formats with required area code, international with +, extensions
	phoneRe = regexp.MustCompile(`(?:(?:\+?1[-.\s]?)?\(?\d{3}\)?[-.\s]?\d{3}[-.\s]?\d{4}(?:\s*(?:ext|x|ext\.)\s*\d{1,5})?)`)

	// Email address pattern
	emailRe = regexp.MustCompile(`[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}`)

	// Digits only for phone normalization
	digitsRe = regexp.MustCompile(`\d`)

	// Angle brackets and quotes for name cleaning
	angleBracketRe = regexp.MustCompile(`[<>]`)
	quotesRe       = regexp.MustCompile(`^["']|["']$`)
)

// StripHTML removes <script>, <style>, all HTML tags, decodes entities, and normalizes whitespace.
func StripHTML(rawHTML string) string {
	// Remove script and style blocks entirely
	result := scriptRe.ReplaceAllString(rawHTML, "")
	result = styleRe.ReplaceAllString(result, "")

	// Replace block-level elements with newlines for readability
	result = blockTagRe.ReplaceAllString(result, "\n")

	// Remove remaining tags
	result = tagRe.ReplaceAllString(result, "")

	// Decode HTML entities
	result = html.UnescapeString(result)

	// Normalize whitespace within lines
	lines := strings.Split(result, "\n")
	for i, line := range lines {
		lines[i] = whitespaceRe.ReplaceAllString(strings.TrimSpace(line), " ")
	}
	result = strings.Join(lines, "\n")

	// Collapse excessive blank lines
	result = blankLinesRe.ReplaceAllString(result, "\n\n")

	return strings.TrimSpace(result)
}

// FindPhoneNumbers extracts phone numbers from text.
func FindPhoneNumbers(text string) []string {
	matches := phoneRe.FindAllString(text, -1)
	var result []string
	for _, m := range matches {
		digits := digitsRe.FindAllString(m, -1)
		// Require at least 7 digits for a valid phone number
		if len(digits) >= 7 {
			result = append(result, strings.TrimSpace(m))
		}
	}
	return result
}

// NormalizePhone converts a phone string to E.164 format (+1XXXXXXXXXX) for US numbers.
// Non-US numbers are returned with digits only prefixed by +.
func NormalizePhone(raw string) string {
	digits := strings.Join(digitsRe.FindAllString(raw, -1), "")
	if len(digits) == 0 {
		return ""
	}

	// US number: 10 digits or 11 digits starting with 1
	switch {
	case len(digits) == 10:
		return "+1" + digits
	case len(digits) == 11 && digits[0] == '1':
		return "+" + digits
	default:
		return "+" + digits
	}
}

// FindEmailAddresses extracts email addresses from text.
func FindEmailAddresses(text string) []string {
	return emailRe.FindAllString(text, -1)
}

// IsValidEmail performs structural validation of an email address.
func IsValidEmail(addr string) bool {
	if addr == "" {
		return false
	}
	parts := strings.SplitN(addr, "@", 2)
	if len(parts) != 2 {
		return false
	}
	local, domain := parts[0], parts[1]
	if local == "" || domain == "" {
		return false
	}
	// Validate local part: no whitespace or control characters
	for _, r := range local {
		if unicode.IsSpace(r) || unicode.IsControl(r) {
			return false
		}
	}
	if !strings.Contains(domain, ".") {
		return false
	}
	// Check domain part has valid characters
	for _, r := range domain {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '.' && r != '-' {
			return false
		}
	}
	return true
}

// Truncate cuts text at the given rune limit, respecting word boundaries.
func Truncate(text string, maxChars int) string {
	if maxChars <= 0 {
		return ""
	}
	runes := []rune(text)
	if len(runes) <= maxChars {
		return text
	}
	truncated := string(runes[:maxChars])
	lastSpace := strings.LastIndex(truncated, " ")
	if lastSpace > len(truncated)/2 {
		truncated = truncated[:lastSpace]
	}
	return truncated + "..."
}

// CleanName strips quotes, angle brackets, and normalizes whitespace from a name.
func CleanName(raw string) string {
	result := angleBracketRe.ReplaceAllString(raw, "")
	result = quotesRe.ReplaceAllString(result, "")
	result = strings.TrimSpace(result)
	result = whitespaceRe.ReplaceAllString(result, " ")
	return result
}
