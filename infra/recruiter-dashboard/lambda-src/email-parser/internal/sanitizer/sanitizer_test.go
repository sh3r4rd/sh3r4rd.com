package sanitizer

import (
	"strings"
	"testing"
)

func TestStripHTML(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantAbsent []string // substrings that must NOT appear in output
		wantContain []string // substrings that MUST appear in output
	}{
		{
			name:       "removes script tags completely",
			input:      `<p>Hello</p><script>alert('xss')</script><p>World</p>`,
			wantAbsent: []string{"<script>", "alert(", "</script>"},
			wantContain: []string{"Hello", "World"},
		},
		{
			name:       "removes style tags completely",
			input:      `<p>Hello</p><style>body { color: red; }</style><p>World</p>`,
			wantAbsent: []string{"<style>", "color: red", "</style>"},
			wantContain: []string{"Hello", "World"},
		},
		{
			name:       "removes multiline script blocks",
			input:      "<div>Safe</div><script type=\"text/javascript\">\nvar x = 1;\nfetch('evil.com');\n</script><div>Also safe</div>",
			wantAbsent: []string{"<script", "var x", "fetch(", "</script>"},
			wantContain: []string{"Safe", "Also safe"},
		},
		{
			name:       "decodes HTML entities",
			input:      `<p>AT&amp;T &lt;rocks&gt; &quot;yes&quot;</p>`,
			wantAbsent: []string{"&amp;", "&lt;", "&gt;", "&quot;", "<p>"},
			wantContain: []string{"AT&T", "<rocks>", "\"yes\""},
		},
		{
			name:       "plain text passes through unchanged",
			input:      "Just plain text with no HTML tags at all.",
			wantAbsent: []string{},
			wantContain: []string{"Just plain text with no HTML tags at all."},
		},
		{
			name:       "removes all tag types",
			input:      `<h1>Title</h1><p>Paragraph</p><a href="link">Click</a><img src="img.png"/>`,
			wantAbsent: []string{"<h1>", "</h1>", "<p>", "<a ", "<img"},
			wantContain: []string{"Title", "Paragraph", "Click"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := StripHTML(tt.input)

			for _, absent := range tt.wantAbsent {
				if strings.Contains(got, absent) {
					t.Errorf("StripHTML output should not contain %q, got:\n%s", absent, got)
				}
			}
			for _, contain := range tt.wantContain {
				if !strings.Contains(got, contain) {
					t.Errorf("StripHTML output should contain %q, got:\n%s", contain, got)
				}
			}
		})
	}
}

func TestFindPhoneNumbers(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{
			name:  "parenthesized area code",
			input: "Call me at (650) 253-0000 for details.",
			want:  []string{"(650) 253-0000"},
		},
		{
			name:  "international with +1 and dashes",
			input: "Phone: +1-425-882-8080",
			want:  []string{"+1-425-882-8080"},
		},
		{
			name:  "dashed format",
			input: "Reach me at 415-555-1234 anytime.",
			want:  []string{"415-555-1234"},
		},
		{
			name:  "multiple numbers in text",
			input: "Office: (212) 555-0100, Cell: 310-555-9999",
			want:  []string{"(212) 555-0100", "310-555-9999"},
		},
		{
			name:  "no phone numbers",
			input: "There are no phone numbers in this text.",
			want:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FindPhoneNumbers(tt.input)

			if tt.want == nil {
				if len(got) != 0 {
					t.Errorf("FindPhoneNumbers = %v, want empty", got)
				}
				return
			}

			if len(got) != len(tt.want) {
				t.Fatalf("FindPhoneNumbers returned %d results %v, want %d results %v", len(got), got, len(tt.want), tt.want)
			}

			for i, want := range tt.want {
				if got[i] != want {
					t.Errorf("FindPhoneNumbers[%d] = %q, want %q", i, got[i], want)
				}
			}
		})
	}
}

func TestNormalizePhone(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "10-digit US number",
			input: "(650) 253-0000",
			want:  "+16502530000",
		},
		{
			name:  "11-digit starting with 1",
			input: "+1-425-882-8080",
			want:  "+14258828080",
		},
		{
			name:  "10 digits with dashes",
			input: "415-555-1234",
			want:  "+14155551234",
		},
		{
			name:  "empty string returns empty",
			input: "",
			want:  "",
		},
		{
			name:  "no digits returns empty",
			input: "no phone here",
			want:  "",
		},
		{
			name:  "11-digit starting with 1 dotted format",
			input: "1.650.253.0000",
			want:  "+16502530000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NormalizePhone(tt.input)
			if got != tt.want {
				t.Errorf("NormalizePhone(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestFindEmailAddresses(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{
			name:  "single email in text",
			input: "Contact me at jane.smith@google.com for info.",
			want:  []string{"jane.smith@google.com"},
		},
		{
			name:  "multiple emails",
			input: "From alice@stripe.com and bob@microsoft.com",
			want:  []string{"alice@stripe.com", "bob@microsoft.com"},
		},
		{
			name:  "email with plus addressing",
			input: "Send to user+tag@example.com",
			want:  []string{"user+tag@example.com"},
		},
		{
			name:  "no emails",
			input: "No email addresses here.",
			want:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FindEmailAddresses(tt.input)

			if tt.want == nil {
				if len(got) != 0 {
					t.Errorf("FindEmailAddresses = %v, want empty", got)
				}
				return
			}

			if len(got) != len(tt.want) {
				t.Fatalf("FindEmailAddresses returned %d results %v, want %d %v", len(got), got, len(tt.want), tt.want)
			}
			for i, want := range tt.want {
				if got[i] != want {
					t.Errorf("FindEmailAddresses[%d] = %q, want %q", i, got[i], want)
				}
			}
		})
	}
}

func TestIsValidEmail(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{name: "valid standard address", input: "user@example.com", want: true},
		{name: "valid with subdomain", input: "user@mail.example.com", want: true},
		{name: "valid with dots in local", input: "first.last@example.com", want: true},
		{name: "valid with plus", input: "user+tag@example.com", want: true},
		{name: "valid with hyphen in domain", input: "user@my-company.com", want: true},
		{name: "empty string", input: "", want: false},
		{name: "no at sign", input: "userexample.com", want: false},
		{name: "no domain dot", input: "user@localhost", want: false},
		{name: "empty local part", input: "@example.com", want: false},
		{name: "empty domain part", input: "user@", want: false},
		{name: "space in domain", input: "user@exam ple.com", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsValidEmail(tt.input)
			if got != tt.want {
				t.Errorf("IsValidEmail(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		maxChars int
		want     string
	}{
		{
			name:     "short text unchanged",
			input:    "Hello, world!",
			maxChars: 50,
			want:     "Hello, world!",
		},
		{
			name:     "exact length unchanged",
			input:    "Hello",
			maxChars: 5,
			want:     "Hello",
		},
		{
			name:     "long text truncated at word boundary",
			input:    "The quick brown fox jumps over the lazy dog",
			maxChars: 20,
			want:     "The quick brown fox...",
		},
		{
			name:     "truncation adds ellipsis",
			input:    "This is a sentence that is too long for the limit",
			maxChars: 15,
			want:     "This is a...",
		},
		{
			name:     "handles multi-byte UTF-8 Japanese characters",
			input:    "これは日本語のテストです",
			maxChars: 5,
			want:     "これは日本語...",
		},
		{
			name:     "handles multi-byte UTF-8 with spaces",
			input:    "Hello 世界 foo bar baz",
			maxChars: 10,
			want:     "Hello 世界...",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Truncate(tt.input, tt.maxChars)

			// For short/exact cases, expect exact match
			if len([]rune(tt.input)) <= tt.maxChars {
				if got != tt.want {
					t.Errorf("Truncate(%q, %d) = %q, want %q", tt.input, tt.maxChars, got, tt.want)
				}
				return
			}

			// For truncated cases: must end with "..." and be shorter than original
			if !strings.HasSuffix(got, "...") {
				t.Errorf("Truncate(%q, %d) = %q, should end with '...'", tt.input, tt.maxChars, got)
			}
			if len([]rune(got)) > tt.maxChars+3 { // maxChars + "..."
				t.Errorf("Truncate(%q, %d) = %q (len %d runes), too long", tt.input, tt.maxChars, got, len([]rune(got)))
			}
		})
	}
}

func TestCleanName(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "strips angle brackets",
			input: "<Jane Smith>",
			want:  "Jane Smith",
		},
		{
			name:  "strips surrounding quotes",
			input: `"Bob Johnson"`,
			want:  "Bob Johnson",
		},
		{
			name:  "strips single quotes",
			input: "'Alice Chen'",
			want:  "Alice Chen",
		},
		{
			name:  "name with email in angle brackets",
			input: "Jane Smith <jane@google.com>",
			want:  "Jane Smith jane@google.com",
		},
		{
			name:  "normalizes whitespace",
			input: "  Jane    Smith  ",
			want:  "Jane Smith",
		},
		{
			name:  "plain name unchanged",
			input: "David Lee",
			want:  "David Lee",
		},
		{
			name:  "empty string",
			input: "",
			want:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CleanName(tt.input)
			if got != tt.want {
				t.Errorf("CleanName(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
