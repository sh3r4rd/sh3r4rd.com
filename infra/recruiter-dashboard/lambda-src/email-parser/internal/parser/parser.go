package parser

import (
	"bytes"
	"errors"
	"io"
	"mime"
	"mime/multipart"
	"net/mail"
	"regexp"
	"strings"
	"time"

	perrors "github.com/sh3r4rd/sh3r4rd.com/infra/recruiter-dashboard/lambda-src/email-parser/internal/errors"
	"github.com/sh3r4rd/sh3r4rd.com/infra/recruiter-dashboard/lambda-src/email-parser/internal/models"
	"github.com/sh3r4rd/sh3r4rd.com/infra/recruiter-dashboard/lambda-src/email-parser/internal/sanitizer"
)

const (
	maxBodyBytes = 10 * 1024 * 1024 // 10 MB limit per body read
	maxDepth     = 5                // Maximum MIME nesting depth
)

// Forwarded email boundary patterns (priority order after MIME message/rfc822).
var forwardPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?m)^-{5,}\s*Forwarded message\s*-{5,}`),         // Gmail
	regexp.MustCompile(`(?m)^-{5,}\s*Original Message\s*-{5,}`),          // Outlook inline
	regexp.MustCompile(`(?mi)^Begin forwarded message\s*:`),               // Apple Mail
	regexp.MustCompile(`(?m)^-{5,}\s*Mensaje reenviado\s*-{5,}`),         // Gmail Spanish
	regexp.MustCompile(`(?m)^-{5,}\s*Weitergeleitete Nachricht\s*-{5,}`), // Gmail German
}

// Pseudo-header patterns found after forward boundaries (multi-language support).
var (
	fromHeaderRe    = regexp.MustCompile(`(?mi)^(?:From|De|Von|Da)\s*:\s*(.+)$`)
	dateHeaderRe    = regexp.MustCompile(`(?mi)^(?:Date|Sent|Fecha|Datum|Data)\s*:\s*(.+)$`)
	subjectHeaderRe = regexp.MustCompile(`(?mi)^(?:Subject|Asunto|Betreff|Oggetto|Objet)\s*:\s*(.+)$`)
)

// ParseRawEmail parses raw email bytes into a ParsedEmail.
// It handles plain text, HTML, and multipart MIME emails,
// and detects forwarded content from Gmail, Outlook, and Apple Mail.
func ParseRawEmail(raw []byte) (*models.ParsedEmail, error) {
	return parseRawEmailWithDepth(raw, 0)
}

func parseRawEmailWithDepth(raw []byte, depth int) (*models.ParsedEmail, error) {
	if depth > maxDepth {
		return nil, &perrors.ParseError{Op: "depth_limit", Err: errors.New("max MIME nesting depth exceeded")}
	}

	msg, err := mail.ReadMessage(bytes.NewReader(raw))
	if err != nil {
		return nil, &perrors.ParseError{Op: "read_message", Err: err}
	}

	parsed := &models.ParsedEmail{
		From:    msg.Header.Get("From"),
		To:      msg.Header.Get("To"),
		Subject: msg.Header.Get("Subject"),
	}

	if dateStr := msg.Header.Get("Date"); dateStr != "" {
		if t, err := mail.ParseDate(dateStr); err == nil {
			parsed.Date = t
		}
	}
	if parsed.Date.IsZero() {
		parsed.Date = time.Now().UTC()
	}

	contentType := msg.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "text/plain"
	}

	body, err := io.ReadAll(io.LimitReader(msg.Body, maxBodyBytes))
	if err != nil {
		return nil, &perrors.ParseError{Op: "read_body", Err: err}
	}

	mediaType, params, err := mime.ParseMediaType(contentType)
	if err != nil {
		// Treat as plain text on parse error
		parsed.Body = string(body)
		detectForwarded(parsed)
		return parsed, nil
	}

	switch {
	case mediaType == "text/plain":
		parsed.Body = string(body)
	case mediaType == "text/html":
		parsed.HTMLBody = string(body)
		parsed.Body = sanitizer.StripHTML(string(body))
	case strings.HasPrefix(mediaType, "multipart/"):
		if err := parseMultipart(parsed, body, params["boundary"], depth); err != nil {
			return nil, err
		}
	default:
		parsed.Body = string(body)
	}

	detectForwarded(parsed)
	return parsed, nil
}

// parseMultipart extracts text and HTML parts from a multipart MIME message.
// It also detects nested message/rfc822 parts (Outlook-style forwarded emails).
func parseMultipart(parsed *models.ParsedEmail, body []byte, boundary string, depth int) error {
	if boundary == "" {
		return &perrors.ParseError{Op: "multipart", Err: io.EOF}
	}
	if depth > maxDepth {
		return &perrors.ParseError{Op: "multipart_depth", Err: errors.New("max MIME nesting depth exceeded")}
	}

	reader := multipart.NewReader(bytes.NewReader(body), boundary)
	for {
		part, err := reader.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			return &perrors.ParseError{Op: "multipart_next", Err: err}
		}

		partData, err := io.ReadAll(io.LimitReader(part, maxBodyBytes))
		if err != nil {
			continue
		}

		partType := part.Header.Get("Content-Type")
		mediaType, params, _ := mime.ParseMediaType(partType)

		switch {
		case mediaType == "message/rfc822":
			// Outlook-style: original email attached as nested MIME
			nested, err := parseRawEmailWithDepth(partData, depth+1)
			if err == nil {
				parsed.IsForwarded = true
				// Use the nested email's headers, guarding against empty fields
				if nested.From != "" {
					parsed.From = nested.From
				}
				if nested.Subject != "" {
					parsed.Subject = nested.Subject
				}
				if !nested.Date.IsZero() {
					parsed.Date = nested.Date
				}
				parsed.Body = nested.Body
				if parsed.Body == "" {
					parsed.Body = nested.HTMLBody
				}
				// Prepend pseudo-headers for extraction
				var sb strings.Builder
				sb.WriteString("From: " + parsed.From + "\n")
				sb.WriteString("Date: " + parsed.Date.Format(time.RFC1123Z) + "\n")
				sb.WriteString("Subject: " + parsed.Subject + "\n")
				sb.WriteString("\n")
				sb.WriteString(parsed.Body)
				parsed.Body = sb.String()
			}
		case mediaType == "text/plain":
			if parsed.Body == "" {
				parsed.Body = string(partData)
			}
		case mediaType == "text/html":
			parsed.HTMLBody = string(partData)
			if parsed.Body == "" {
				parsed.Body = sanitizer.StripHTML(string(partData))
			}
		case strings.HasPrefix(mediaType, "multipart/"):
			// Nested multipart — recurse with depth tracking
			if err := parseMultipart(parsed, partData, params["boundary"], depth+1); err != nil {
				continue
			}
		}
	}

	return nil
}

// detectForwarded checks the email body for inline forward boundaries
// and extracts the original sender's headers if found.
func detectForwarded(parsed *models.ParsedEmail) {
	if parsed.IsForwarded {
		return // Already detected via message/rfc822
	}

	for _, pattern := range forwardPatterns {
		loc := pattern.FindStringIndex(parsed.Body)
		if loc == nil {
			continue
		}

		parsed.IsForwarded = true
		// Extract the forwarded content after the boundary
		forwardedContent := parsed.Body[loc[1]:]

		// Parse pseudo-headers in the next ~20 lines
		lines := strings.SplitN(forwardedContent, "\n", 25)
		headerBlock := strings.Join(lines, "\n")

		if match := fromHeaderRe.FindStringSubmatch(headerBlock); len(match) > 1 {
			parsed.From = strings.TrimSpace(match[1])
		}
		if match := dateHeaderRe.FindStringSubmatch(headerBlock); len(match) > 1 {
			if t, err := parseFlexibleDate(strings.TrimSpace(match[1])); err == nil {
				parsed.Date = t
			}
		}
		if match := subjectHeaderRe.FindStringSubmatch(headerBlock); len(match) > 1 {
			parsed.Subject = strings.TrimSpace(match[1])
		}

		// Set body to the forwarded content after the pseudo-header block
		bodyStart := strings.Index(forwardedContent, "\n\n")
		if bodyStart != -1 {
			parsed.Body = strings.TrimSpace(forwardedContent[bodyStart:])
		} else {
			parsed.Body = strings.TrimSpace(forwardedContent)
		}
		return
	}
}

// parseFlexibleDate tries multiple date formats common in forwarded email headers.
func parseFlexibleDate(dateStr string) (time.Time, error) {
	formats := []string{
		time.RFC1123Z,
		time.RFC1123,
		time.RFC822Z,
		time.RFC822,
		"Mon, 2 Jan 2006 15:04:05 -0700",
		"Mon, 2 Jan 2006 15:04:05 MST",
		"January 2, 2006 at 3:04:05 PM MST",
		"Jan 2, 2006, at 3:04 PM",
		"Jan 2, 2006 at 3:04 PM",
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05Z07:00",
		"Mon, Jan 2, 2006 at 3:04 PM",
	}

	for _, fmt := range formats {
		if t, err := time.Parse(fmt, dateStr); err == nil {
			return t, nil
		}
	}

	// Try mail.ParseDate as a last resort
	return mail.ParseDate(dateStr)
}
