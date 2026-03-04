package main

import (
	"bytes"
	"io"
	"mime"
	"mime/multipart"
	"net/mail"
	"strings"
)

// forwardedBoundaries is an ordered list of inline-forwarding header markers
// used by common email clients. Checked case-insensitively.
var forwardedBoundaries = []string{
	"---------- Forwarded message ----------",
	"-----Original Message-----",
	"Begin forwarded message:",
}

// ParseRawEmail parses raw RFC 5322 email bytes and returns a ParsedEmail.
// It detects forwarded messages from Gmail, Outlook, and Apple Mail.
func ParseRawEmail(raw []byte) (*ParsedEmail, error) {
	msg, err := mail.ReadMessage(bytes.NewReader(raw))
	if err != nil {
		return nil, err
	}

	outerFrom := msg.Header.Get("From")
	outerTo := msg.Header.Get("To")
	outerSubject := msg.Header.Get("Subject")
	outerDate := msg.Header.Get("Date")

	mediaType, params, err := mime.ParseMediaType(msg.Header.Get("Content-Type"))
	if err != nil {
		// Not a MIME message; treat the body as plain text.
		body, _ := io.ReadAll(msg.Body)
		return &ParsedEmail{
			From:    outerFrom,
			To:      outerTo,
			Subject: outerSubject,
			Date:    outerDate,
			Body:    strings.TrimSpace(string(body)),
		}, nil
	}

	// --- Strategy 1: MIME message/rfc822 attachment (Outlook) ---
	if strings.HasPrefix(mediaType, "multipart/") {
		if nested, ok := extractNestedRFC822(msg.Body, params["boundary"]); ok {
			return nested, nil
		}
	}

	// --- Strategy 2: Inline forwarded-message boundary marker ---
	body, err := readMIMEBody(msg.Body, mediaType, params)
	if err != nil {
		body = ""
	}
	if found, parsed := detectInlineForward(body, outerFrom, outerTo, outerSubject, outerDate); found {
		return parsed, nil
	}

	// --- Strategy 3: Fallback — use outer headers ---
	return &ParsedEmail{
		From:    outerFrom,
		To:      outerTo,
		Subject: outerSubject,
		Date:    outerDate,
		Body:    strings.TrimSpace(body),
	}, nil
}

// extractNestedRFC822 walks a multipart body looking for a message/rfc822 part
// (Outlook attaches the original as a nested MIME message).
func extractNestedRFC822(body io.Reader, boundary string) (*ParsedEmail, bool) {
	mr := multipart.NewReader(body, boundary)
	for {
		part, err := mr.NextPart()
		if err != nil {
			break
		}
		ct := part.Header.Get("Content-Type")
		if strings.Contains(strings.ToLower(ct), "message/rfc822") {
			raw, err := io.ReadAll(part)
			if err != nil {
				break
			}
			nested, err := ParseRawEmail(raw)
			if err != nil {
				break
			}
			nested.IsForwarded = true
			return nested, true
		}
	}
	return nil, false
}

// readMIMEBody extracts the plain-text body from a MIME entity.
func readMIMEBody(body io.Reader, mediaType string, params map[string]string) (string, error) {
	switch {
	case strings.HasPrefix(mediaType, "multipart/"):
		mr := multipart.NewReader(body, params["boundary"])
		for {
			part, err := mr.NextPart()
			if err != nil {
				break
			}
			ct := part.Header.Get("Content-Type")
			partType, _, _ := mime.ParseMediaType(ct)
			if partType == "text/plain" || partType == "" {
				data, err := io.ReadAll(part)
				if err != nil {
					continue
				}
				return string(data), nil
			}
		}
		return "", nil
	default:
		data, err := io.ReadAll(body)
		return string(data), err
	}
}

// detectInlineForward searches the body text for known forwarding boundary
// markers. When found it attempts to parse the pseudo-header block that follows.
func detectInlineForward(body, outerFrom, outerTo, outerSubject, outerDate string) (bool, *ParsedEmail) {
	lower := strings.ToLower(body)
	for _, marker := range forwardedBoundaries {
		idx := strings.Index(lower, strings.ToLower(marker))
		if idx < 0 {
			continue
		}
		after := body[idx+len(marker):]
		from, to, subject, date := parseForwardedHeaders(after)
		// Fall back to outer headers when the pseudo-block is missing fields.
		if from == "" {
			from = outerFrom
		}
		if to == "" {
			to = outerTo
		}
		if subject == "" {
			subject = outerSubject
		}
		if date == "" {
			date = outerDate
		}
		// Body is everything after the first blank line following the headers.
		innerBody := extractBodyAfterHeaders(after)
		return true, &ParsedEmail{
			From:        from,
			To:          to,
			Subject:     subject,
			Date:        date,
			Body:        strings.TrimSpace(innerBody),
			IsForwarded: true,
		}
	}
	return false, nil
}

// parseForwardedHeaders reads up to 20 lines after a forwarding marker and
// extracts From/To/Subject/Date values. Multi-language header names are
// supported (e.g., De/Von for From, Fecha/Datum for Date).
func parseForwardedHeaders(text string) (from, to, subject, date string) {
	lines := strings.SplitN(text, "\n", 21)
	fromAliases := []string{"from", "de", "von"}
	dateAliases := []string{"date", "sent", "fecha", "datum"}
	subjectAliases := []string{"subject", "asunto", "betreff"}
	toAliases := []string{"to", "a", "an"}

	headersStarted := false
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			// Skip blank lines before the header block begins; stop once we
			// have seen at least one header and encounter a blank line.
			if headersStarted {
				break
			}
			continue
		}
		colonIdx := strings.Index(line, ":")
		if colonIdx < 0 {
			continue
		}
		headersStarted = true
		key := strings.ToLower(strings.TrimSpace(line[:colonIdx]))
		val := strings.TrimSpace(line[colonIdx+1:])
		switch {
		case contains(fromAliases, key):
			from = val
		case contains(dateAliases, key):
			date = val
		case contains(subjectAliases, key):
			subject = val
		case contains(toAliases, key):
			to = val
		}
	}
	return
}

// extractBodyAfterHeaders returns the text after the first blank line.
func extractBodyAfterHeaders(text string) string {
	idx := strings.Index(text, "\n\n")
	if idx < 0 {
		return ""
	}
	return text[idx+2:]
}

// contains is a simple case-sensitive membership test.
func contains(haystack []string, needle string) bool {
	for _, s := range haystack {
		if s == needle {
			return true
		}
	}
	return false
}
