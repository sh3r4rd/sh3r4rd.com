package parser

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	perrors "github.com/sh3r4rd/sh3r4rd.com/infra/recruiter-dashboard/lambda-src/email-parser/internal/errors"
)

// testdataDir returns the absolute path to the testdata directory.
func testdataDir(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("failed to determine test file location")
	}
	return filepath.Join(filepath.Dir(file), "..", "..", "testdata")
}

func readFixture(t *testing.T, name string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(testdataDir(t), name))
	if err != nil {
		t.Fatalf("failed to read fixture %s: %v", name, err)
	}
	return data
}

func TestParseRawEmail_ForwardedEmails(t *testing.T) {
	tests := []struct {
		name          string
		fixture       string
		wantForwarded bool
		wantFromName  string
		wantFromEmail string
		wantSubject   string
	}{
		{
			name:          "gmail forward detects forwarded and extracts original sender",
			fixture:       "gmail_forward.eml",
			wantForwarded: true,
			wantFromName:  "Jane Smith",
			wantFromEmail: "jane.smith@google.com",
			wantSubject:   "Senior Software Engineer Opportunity at Google",
		},
		{
			name:          "outlook forward detects forwarded and extracts original sender",
			fixture:       "outlook_forward.eml",
			wantForwarded: true,
			wantFromName:  "Bob Johnson",
			wantFromEmail: "bob.johnson@microsoft.com",
			wantSubject:   "Staff Engineer Role - Microsoft Azure",
		},
		{
			name:          "apple mail forward detects forwarded and extracts original sender",
			fixture:       "apple_mail_forward.eml",
			wantForwarded: true,
			wantFromName:  "Alice Chen",
			wantFromEmail: "alice.chen@stripe.com",
			wantSubject:   "Platform Engineering Lead at Stripe",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			raw := readFixture(t, tt.fixture)
			parsed, err := ParseRawEmail(raw)
			if err != nil {
				t.Fatalf("ParseRawEmail returned error: %v", err)
			}

			if parsed.IsForwarded != tt.wantForwarded {
				t.Errorf("IsForwarded = %v, want %v", parsed.IsForwarded, tt.wantForwarded)
			}

			if !strings.Contains(parsed.From, tt.wantFromName) {
				t.Errorf("From = %q, want it to contain name %q", parsed.From, tt.wantFromName)
			}
			if !strings.Contains(parsed.From, tt.wantFromEmail) {
				t.Errorf("From = %q, want it to contain email %q", parsed.From, tt.wantFromEmail)
			}
			if parsed.Subject != tt.wantSubject {
				t.Errorf("Subject = %q, want %q", parsed.Subject, tt.wantSubject)
			}
		})
	}
}

func TestParseRawEmail_DirectEmail(t *testing.T) {
	raw := readFixture(t, "direct_email.eml")
	parsed, err := ParseRawEmail(raw)
	if err != nil {
		t.Fatalf("ParseRawEmail returned error: %v", err)
	}

	if parsed.IsForwarded {
		t.Error("IsForwarded = true, want false for direct email")
	}

	wantEmail := "david.lee@amazon.com"
	if !strings.Contains(parsed.From, wantEmail) {
		t.Errorf("From = %q, want it to contain %q", parsed.From, wantEmail)
	}

	if parsed.Subject != "Distinguished Engineer - AWS" {
		t.Errorf("Subject = %q, want %q", parsed.Subject, "Distinguished Engineer - AWS")
	}

	if !strings.Contains(parsed.Body, "David Lee") {
		t.Errorf("Body should contain 'David Lee', got: %s", truncateForLog(parsed.Body, 200))
	}
}

func TestParseRawEmail_HTMLStripsTagsFromBody(t *testing.T) {
	raw := readFixture(t, "html_xss.eml")
	parsed, err := ParseRawEmail(raw)
	if err != nil {
		t.Fatalf("ParseRawEmail returned error: %v", err)
	}

	if strings.Contains(parsed.Body, "<script>") {
		t.Error("Body should not contain <script> tags after sanitization")
	}
	if strings.Contains(parsed.Body, "<style>") {
		t.Error("Body should not contain <style> tags after sanitization")
	}
	if strings.Contains(parsed.Body, "alert(") {
		t.Error("Body should not contain JavaScript alert() calls")
	}
	if strings.Contains(parsed.Body, "document.cookie") {
		t.Error("Body should not contain document.cookie references")
	}
	if strings.Contains(parsed.Body, "<p>") {
		t.Error("Body should not contain <p> tags")
	}

	// Should still contain the actual text content
	if !strings.Contains(parsed.Body, "Great Opportunity") {
		t.Errorf("Body should contain text content 'Great Opportunity', got: %s", truncateForLog(parsed.Body, 200))
	}
	if !strings.Contains(parsed.Body, "Eve Hacker") {
		t.Errorf("Body should contain 'Eve Hacker', got: %s", truncateForLog(parsed.Body, 200))
	}

	// HTMLBody should retain the original HTML
	if parsed.HTMLBody == "" {
		t.Error("HTMLBody should not be empty for HTML emails")
	}
}

func TestParseRawEmail_EmptyInput(t *testing.T) {
	_, err := ParseRawEmail([]byte{})
	if err == nil {
		t.Fatal("expected error for empty input, got nil")
	}

	var parseErr *perrors.ParseError
	if !isParseError(err, &parseErr) {
		t.Errorf("expected *perrors.ParseError, got %T: %v", err, err)
	}
}

func TestParseRawEmail_GarbageBytes(t *testing.T) {
	garbage := []byte{0xFF, 0xFE, 0x00, 0x01, 0x02, 0x03, 0xAB, 0xCD}
	_, err := ParseRawEmail(garbage)
	if err == nil {
		t.Fatal("expected error for garbage bytes, got nil")
	}

	var parseErr *perrors.ParseError
	if !isParseError(err, &parseErr) {
		t.Errorf("expected *perrors.ParseError, got %T: %v", err, err)
	}
}

func TestParseRawEmail_MultipartTextPreferred(t *testing.T) {
	// Build a multipart email with both text/plain and text/html parts.
	// The parser should use text/plain when available.
	boundary := "BOUNDARY123"
	raw := fmt.Sprintf(`From: sender@example.com
To: recipient@example.com
Subject: Multipart Test
Date: Mon, 03 Mar 2026 10:00:00 +0000
MIME-Version: 1.0
Content-Type: multipart/alternative; boundary="%s"

--%s
Content-Type: text/plain; charset="UTF-8"

This is the plain text body.
--%s
Content-Type: text/html; charset="UTF-8"

<html><body><p>This is the HTML body.</p></body></html>
--%s--
`, boundary, boundary, boundary, boundary)

	parsed, err := ParseRawEmail([]byte(raw))
	if err != nil {
		t.Fatalf("ParseRawEmail returned error: %v", err)
	}

	if !strings.Contains(parsed.Body, "plain text body") {
		t.Errorf("Body should prefer text/plain content, got: %q", truncateForLog(parsed.Body, 200))
	}

	if parsed.HTMLBody == "" {
		t.Error("HTMLBody should be populated from the text/html part")
	}
}

func TestParseRawEmail_MaxDepthExceeded(t *testing.T) {
	// The parser guards depth via parseRawEmailWithDepth. Since multipart
	// handling silently continues on nested errors, we call the internal
	// function directly to verify the depth guard works. We do this by
	// constructing a valid email and calling parseRawEmailWithDepth at a
	// depth that exceeds the limit.
	raw := []byte(`From: deep@example.com
To: recipient@example.com
Subject: Depth Test
Date: Mon, 03 Mar 2026 10:00:00 +0000
Content-Type: text/plain

Content.`)

	// Calling at maxDepth+1 should trigger the depth guard.
	_, err := parseRawEmailWithDepth(raw, maxDepth+1)
	if err == nil {
		t.Fatal("expected error for depth exceeding maxDepth, got nil")
	}

	errStr := err.Error()
	if !strings.Contains(errStr, "depth") {
		t.Errorf("error should mention depth, got: %v", err)
	}

	var parseErr *perrors.ParseError
	if !isParseError(err, &parseErr) {
		t.Errorf("expected *perrors.ParseError, got %T", err)
	}
}

func TestParseRawEmail_PlainTextNoContentType(t *testing.T) {
	raw := `From: plain@example.com
To: recipient@example.com
Subject: No Content-Type Header
Date: Mon, 03 Mar 2026 10:00:00 +0000

This email has no Content-Type header.
It should default to text/plain.`

	parsed, err := ParseRawEmail([]byte(raw))
	if err != nil {
		t.Fatalf("ParseRawEmail returned error: %v", err)
	}

	if !strings.Contains(parsed.Body, "no Content-Type header") {
		t.Errorf("Body should contain plain text, got: %q", parsed.Body)
	}
}

func TestParseRawEmail_DateParsing(t *testing.T) {
	raw := `From: sender@example.com
To: recipient@example.com
Subject: Date Test
Date: Mon, 03 Mar 2026 10:30:00 -0500

Body content.`

	parsed, err := ParseRawEmail([]byte(raw))
	if err != nil {
		t.Fatalf("ParseRawEmail returned error: %v", err)
	}

	if parsed.Date.IsZero() {
		t.Error("Date should not be zero")
	}
	if parsed.Date.Year() != 2026 {
		t.Errorf("Date year = %d, want 2026", parsed.Date.Year())
	}
}

func TestParseRawEmail_MissingDate(t *testing.T) {
	raw := `From: sender@example.com
To: recipient@example.com
Subject: No Date

Body content.`

	parsed, err := ParseRawEmail([]byte(raw))
	if err != nil {
		t.Fatalf("ParseRawEmail returned error: %v", err)
	}

	if parsed.Date.IsZero() {
		t.Error("Date should default to current time, not zero")
	}
}

// isParseError checks if err is a *perrors.ParseError via type assertion.
func isParseError(err error, target **perrors.ParseError) bool {
	pe, ok := err.(*perrors.ParseError)
	if ok && target != nil {
		*target = pe
	}
	return ok
}

// truncateForLog truncates a string for log output readability.
func truncateForLog(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}

func TestParseRawEmail_MessageRFC822Nested(t *testing.T) {
	// Construct a multipart/mixed email with a nested message/rfc822 part (Outlook style)
	nestedEmail := "From: nested@recruiter.com\r\nTo: user@test.com\r\nSubject: Nested Job Opportunity\r\nDate: Mon, 03 Mar 2026 10:00:00 +0000\r\nContent-Type: text/plain\r\n\r\nThis is the nested recruiter email body."

	boundary := "----=_Part_12345"
	raw := fmt.Sprintf("From: forwarder@test.com\r\nTo: inbox@test.com\r\nSubject: FW: Nested Job Opportunity\r\nDate: Tue, 04 Mar 2026 12:00:00 +0000\r\nContent-Type: multipart/mixed; boundary=\"%s\"\r\n\r\n--%s\r\nContent-Type: text/plain\r\n\r\nForwarded email attached.\r\n--%s\r\nContent-Type: message/rfc822\r\n\r\n%s\r\n--%s--\r\n",
		boundary, boundary, boundary, nestedEmail, boundary)

	parsed, err := ParseRawEmail([]byte(raw))
	if err != nil {
		t.Fatalf("ParseRawEmail returned error: %v", err)
	}

	if !parsed.IsForwarded {
		t.Error("expected IsForwarded=true for message/rfc822 part")
	}
	if !strings.Contains(parsed.From, "nested@recruiter.com") {
		t.Errorf("expected From to contain nested@recruiter.com, got %q", parsed.From)
	}
	if !strings.Contains(parsed.Body, "nested recruiter email body") {
		t.Errorf("expected Body to contain nested email content, got %q", truncateForLog(parsed.Body, 100))
	}
}

func TestParseRawEmail_MultipartEmptyBoundary(t *testing.T) {
	raw := "From: test@test.com\r\nTo: me@test.com\r\nSubject: Bad Multipart\r\nDate: Mon, 03 Mar 2026 10:00:00 +0000\r\nContent-Type: multipart/mixed\r\n\r\nNo boundary parameter.\r\n"

	_, err := ParseRawEmail([]byte(raw))
	if err == nil {
		t.Fatal("expected error for multipart with no boundary")
	}
}

func TestParseRawEmail_MultipartHTMLOnly(t *testing.T) {
	boundary := "----=_Part_HTML"
	raw := fmt.Sprintf("From: html@test.com\r\nTo: me@test.com\r\nSubject: HTML Only\r\nDate: Mon, 03 Mar 2026 10:00:00 +0000\r\nContent-Type: multipart/alternative; boundary=\"%s\"\r\n\r\n--%s\r\nContent-Type: text/html\r\n\r\n<html><body><p>Hello <b>World</b></p></body></html>\r\n--%s--\r\n",
		boundary, boundary, boundary)

	parsed, err := ParseRawEmail([]byte(raw))
	if err != nil {
		t.Fatalf("ParseRawEmail returned error: %v", err)
	}

	if parsed.HTMLBody == "" {
		t.Error("expected HTMLBody to be populated")
	}
	if !strings.Contains(parsed.Body, "Hello") {
		t.Errorf("expected Body to contain stripped HTML content, got %q", parsed.Body)
	}
	if strings.Contains(parsed.Body, "<b>") {
		t.Error("Body should not contain HTML tags")
	}
}

func TestParseRawEmail_NestedMultipart(t *testing.T) {
	innerBoundary := "inner_bound"
	outerBoundary := "outer_bound"

	raw := fmt.Sprintf("From: nested@test.com\r\nTo: me@test.com\r\nSubject: Nested Multipart\r\nDate: Mon, 03 Mar 2026 10:00:00 +0000\r\nContent-Type: multipart/mixed; boundary=\"%s\"\r\n\r\n--%s\r\nContent-Type: multipart/alternative; boundary=\"%s\"\r\n\r\n--%s\r\nContent-Type: text/plain\r\n\r\nPlain text from nested multipart.\r\n--%s\r\nContent-Type: text/html\r\n\r\n<p>HTML from nested multipart</p>\r\n--%s--\r\n--%s--\r\n",
		outerBoundary, outerBoundary, innerBoundary, innerBoundary, innerBoundary, innerBoundary, outerBoundary)

	parsed, err := ParseRawEmail([]byte(raw))
	if err != nil {
		t.Fatalf("ParseRawEmail returned error: %v", err)
	}

	if !strings.Contains(parsed.Body, "Plain text from nested multipart") {
		t.Errorf("expected Body to contain plain text from nested multipart, got %q", parsed.Body)
	}
}

func TestParseRawEmail_RFC822EmptyFrom(t *testing.T) {
	// Test that empty From in nested message doesn't overwrite valid outer From
	nestedEmail := "To: user@test.com\r\nSubject: No From Header\r\nDate: Mon, 03 Mar 2026 10:00:00 +0000\r\nContent-Type: text/plain\r\n\r\nBody without From."

	boundary := "----=_Part_NoFrom"
	raw := fmt.Sprintf("From: outer@test.com\r\nTo: inbox@test.com\r\nSubject: FW: Test\r\nDate: Tue, 04 Mar 2026 12:00:00 +0000\r\nContent-Type: multipart/mixed; boundary=\"%s\"\r\n\r\n--%s\r\nContent-Type: message/rfc822\r\n\r\n%s\r\n--%s--\r\n",
		boundary, boundary, nestedEmail, boundary)

	parsed, err := ParseRawEmail([]byte(raw))
	if err != nil {
		t.Fatalf("ParseRawEmail returned error: %v", err)
	}

	// From should remain from outer email since nested has no From
	if parsed.From == "" {
		t.Error("From should not be empty when nested message has no From header")
	}
}
