package main

import (
	"strings"
	"testing"
)

// rawSimpleEmail builds a minimal RFC 5322 email string.
func rawSimpleEmail(from, to, subject, body string) string {
	return "From: " + from + "\r\n" +
		"To: " + to + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"Content-Type: text/plain\r\n" +
		"\r\n" +
		body
}

// rawGmailForward wraps an inner message in a Gmail-style forwarding block.
func rawGmailForward(innerFrom, innerSubject, innerDate, innerBody string) string {
	return "From: me@example.com\r\n" +
		"To: archive@example.com\r\n" +
		"Subject: Fwd: " + innerSubject + "\r\n" +
		"Content-Type: text/plain\r\n" +
		"\r\n" +
		"---------- Forwarded message ----------\n" +
		"From: " + innerFrom + "\n" +
		"Date: " + innerDate + "\n" +
		"Subject: " + innerSubject + "\n" +
		"To: me@example.com\n" +
		"\n" +
		innerBody
}

func TestParseRawEmail_Simple(t *testing.T) {
	raw := rawSimpleEmail("recruiter@acme.com", "me@example.com", "Job opportunity", "Hello!")
	got, err := ParseRawEmail([]byte(raw))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.From != "recruiter@acme.com" {
		t.Errorf("From: got %q, want %q", got.From, "recruiter@acme.com")
	}
	if got.Subject != "Job opportunity" {
		t.Errorf("Subject: got %q, want %q", got.Subject, "Job opportunity")
	}
	if got.IsForwarded {
		t.Errorf("IsForwarded: expected false for direct email")
	}
}

func TestParseRawEmail_GmailForward(t *testing.T) {
	raw := rawGmailForward(
		"recruiter@acme.com",
		"Senior Go Engineer",
		"Mon, 3 Mar 2026 10:00:00 -0500",
		"I'm a recruiter at Acme and we have a great opportunity.",
	)
	got, err := ParseRawEmail([]byte(raw))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !got.IsForwarded {
		t.Error("IsForwarded: expected true for Gmail-forwarded email")
	}
	if got.From != "recruiter@acme.com" {
		t.Errorf("From: got %q, want %q", got.From, "recruiter@acme.com")
	}
	if got.Subject != "Senior Go Engineer" {
		t.Errorf("Subject: got %q, want %q", got.Subject, "Senior Go Engineer")
	}
}

func TestParseRawEmail_OutlookForward(t *testing.T) {
	raw := "From: me@example.com\r\n" +
		"To: archive@example.com\r\n" +
		"Subject: FW: DevOps role\r\n" +
		"Content-Type: text/plain\r\n" +
		"\r\n" +
		"-----Original Message-----\n" +
		"From: recruiter@bigcorp.com\n" +
		"Date: Tue, 4 Mar 2026\n" +
		"Subject: DevOps role\n" +
		"To: me@example.com\n" +
		"\n" +
		"We have a great DevOps position at BigCorp."
	got, err := ParseRawEmail([]byte(raw))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !got.IsForwarded {
		t.Error("IsForwarded: expected true for Outlook-forwarded email")
	}
	if got.From != "recruiter@bigcorp.com" {
		t.Errorf("From: got %q, want %q", got.From, "recruiter@bigcorp.com")
	}
}

func TestParseRawEmail_MalformedHeaders(t *testing.T) {
	// Missing headers should not panic; the parser must return an error or
	// a best-effort ParsedEmail with empty fields.
	raw := "not valid rfc822 content"
	_, err := ParseRawEmail([]byte(raw))
	// net/mail.ReadMessage returns an error when there are no headers; that's fine.
	if err != nil {
		// acceptable — just ensure no panic
		return
	}
}

func TestParseRawEmail_EmptyInput(t *testing.T) {
	_, err := ParseRawEmail([]byte(""))
	if err == nil {
		t.Error("expected error for empty input")
	}
}

// --- Parser helper tests ---

func TestContains(t *testing.T) {
	if !contains([]string{"a", "b", "c"}, "b") {
		t.Error("expected true")
	}
	if contains([]string{"a", "b"}, "z") {
		t.Error("expected false")
	}
}

func TestDomainFromAddress(t *testing.T) {
	tests := []struct {
		addr string
		want string
	}{
		{"recruiter@acme.com", "acme.com"},
		{"Alice Smith <alice@bigcorp.io>", "bigcorp.io"},
		{"noatsign", ""},
	}
	for _, tc := range tests {
		got := domainFromAddress(tc.addr)
		if got != tc.want {
			t.Errorf("domainFromAddress(%q) = %q, want %q", tc.addr, got, tc.want)
		}
	}
}

func TestNormalizePhone(t *testing.T) {
	tests := []struct {
		raw  string
		want string
	}{
		{"555-867-5309", "+15558675309"},
		{"15558675309", "+15558675309"},
		{"5558675309", "+15558675309"},
	}
	for _, tc := range tests {
		got := normalizePhone(tc.raw)
		if got != tc.want {
			t.Errorf("normalizePhone(%q) = %q, want %q", tc.raw, got, tc.want)
		}
	}
}

// --- Extractor tests ---

func TestExtractCompany_Pattern(t *testing.T) {
	body := "Hi, I'm a recruiter at Acme Corp and I think you'd be a great fit."
	parsed := &ParsedEmail{Body: body, From: "recruiter@acme.com"}
	result := ExtractRecruiterData(parsed)
	if result.Company == nil {
		t.Fatal("Company: expected non-nil")
	}
	if !strings.Contains(*result.Company, "Acme") {
		t.Errorf("Company: got %q, want something containing 'Acme'", *result.Company)
	}
}

func TestExtractCompany_DomainFallback(t *testing.T) {
	body := "Let me tell you about this role."
	parsed := &ParsedEmail{Body: body, From: "recruiter@bigcorp.com"}
	result := ExtractRecruiterData(parsed)
	if result.Company == nil {
		t.Fatal("Company: expected non-nil from domain fallback")
	}
	if *result.Company != "Bigcorp" {
		t.Errorf("Company: got %q, want %q", *result.Company, "Bigcorp")
	}
}

func TestExtractCompany_GenericDomain(t *testing.T) {
	body := "Let me tell you about this role."
	parsed := &ParsedEmail{Body: body, From: "recruiter@gmail.com"}
	result := ExtractRecruiterData(parsed)
	if result.Company != nil {
		t.Errorf("Company: expected nil for generic domain, got %q", *result.Company)
	}
}

func TestExtractJobTitle(t *testing.T) {
	parsed := &ParsedEmail{
		Subject: "Regarding the Senior Go Engineer position",
		Body:    "",
		From:    "r@acme.com",
	}
	result := ExtractRecruiterData(parsed)
	if result.JobTitle == nil {
		t.Fatal("JobTitle: expected non-nil")
	}
	if !strings.Contains(*result.JobTitle, "Senior Go Engineer") {
		t.Errorf("JobTitle: got %q", *result.JobTitle)
	}
}

func TestExtractCompositeTitle(t *testing.T) {
	parsed := &ParsedEmail{
		Subject: "Open role",
		Body:    "We are looking for a Senior Software Engineer to join the team.",
		From:    "r@acme.com",
	}
	result := ExtractRecruiterData(parsed)
	if result.JobTitle == nil {
		t.Fatal("JobTitle: expected non-nil for composite title")
	}
}

func TestExtractName_FromHeader(t *testing.T) {
	parsed := &ParsedEmail{From: "Alice Smith <alice@acme.com>", Body: ""}
	result := ExtractRecruiterData(parsed)
	if result.Name == nil {
		t.Fatal("Name: expected non-nil from header display name")
	}
	if *result.Name != "Alice Smith" {
		t.Errorf("Name: got %q, want %q", *result.Name, "Alice Smith")
	}
}

func TestExtractPhone(t *testing.T) {
	parsed := &ParsedEmail{
		From: "r@acme.com",
		Body: "Feel free to call me at 555-867-5309 anytime.",
	}
	result := ExtractRecruiterData(parsed)
	if result.Phone == nil {
		t.Fatal("Phone: expected non-nil")
	}
	if *result.Phone != "+15558675309" {
		t.Errorf("Phone: got %q, want +15558675309", *result.Phone)
	}
}

func TestExtractPhone_None(t *testing.T) {
	parsed := &ParsedEmail{From: "r@acme.com", Body: "No phone number here."}
	result := ExtractRecruiterData(parsed)
	if result.Phone != nil {
		t.Errorf("Phone: expected nil, got %q", *result.Phone)
	}
}

// --- Models tests ---

func TestBuildDedupKey_Deterministic(t *testing.T) {
	a := buildDedupKey("from@example.com", "Subject", "Mon 1 Jan 2024")
	b := buildDedupKey("from@example.com", "Subject", "Mon 1 Jan 2024")
	if a != b {
		t.Errorf("dedup key not deterministic: %q != %q", a, b)
	}
}

func TestBuildDedupKey_UniquenessOnDifferentInputs(t *testing.T) {
	a := buildDedupKey("a@a.com", "Hello", "2024-01-01")
	b := buildDedupKey("b@b.com", "Hello", "2024-01-01")
	if a == b {
		t.Error("different inputs produced same dedup key")
	}
}

func TestNewRecruiterEmail_Fallbacks(t *testing.T) {
	parsed := &ParsedEmail{
		From:    "r@acme.com",
		Subject: "Opportunity",
		Date:    "2024-01-01",
	}
	extracted := &ExtractionResult{} // all nil fields

	r := NewRecruiterEmail("my-bucket", "emails/001.eml", parsed, extracted)
	if r.Company != "" || r.JobTitle != "" || r.Name != "" || r.Phone != "" {
		t.Error("expected empty string fallbacks for nil extracted fields")
	}
	if r.DedupKey == "" {
		t.Error("DedupKey must not be empty")
	}
	if r.ProcessedAt == "" {
		t.Error("ProcessedAt must not be empty")
	}
}

func TestToDynamoDBItem(t *testing.T) {
	parsed := &ParsedEmail{From: "r@acme.com", Subject: "Hi", Date: "2024-01-01"}
	extracted := &ExtractionResult{}
	r := NewRecruiterEmail("bucket", "key", parsed, extracted)

	item, err := r.ToDynamoDBItem()
	if err != nil {
		t.Fatalf("ToDynamoDBItem error: %v", err)
	}
	if _, ok := item["dedup_key"]; !ok {
		t.Error("DynamoDB item missing 'dedup_key'")
	}
	if _, ok := item["processed_at"]; !ok {
		t.Error("DynamoDB item missing 'processed_at'")
	}
}
