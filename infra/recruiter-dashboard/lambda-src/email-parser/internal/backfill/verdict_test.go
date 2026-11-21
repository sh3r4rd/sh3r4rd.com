package backfill

import "testing"

func TestVerdictFromHeaders(t *testing.T) {
	const body = "\r\nHello\r\n"
	const base = "From: a@example.com\r\nTo: b@example.org\r\nSubject: hi\r\n"
	tests := []struct {
		name    string
		headers string
		want    Verdict
	}{
		{"no SES headers", "", VerdictPass},
		{"all pass (real archive shape)", "X-SES-Spam-Verdict: PASS\r\nX-SES-Virus-Verdict: PASS\r\nAuthentication-Results: amazonses.com; spf=pass (spfCheck: domain allowed) smtp.mailfrom=example.com; dkim=pass header.i=@example.com; dmarc=pass header.from=example.com;\r\n", VerdictPass},
		{"spam fail", "X-SES-Spam-Verdict: FAIL\r\n", VerdictRejected},
		{"spam fail lowercase", "X-SES-Spam-Verdict: fail\r\n", VerdictRejected},
		{"spam gray is not a failure", "X-SES-Spam-Verdict: GRAY\r\n", VerdictPass},
		{"virus fail", "X-SES-Virus-Verdict: FAIL\r\n", VerdictVirus},
		{"virus beats spam", "X-SES-Spam-Verdict: FAIL\r\nX-SES-Virus-Verdict: FAIL\r\n", VerdictVirus},
		{"forged virus PASS above the real FAIL", "X-SES-Virus-Verdict: PASS\r\nX-SES-Virus-Verdict: FAIL\r\n", VerdictVirus},
		{"forged spam PASS above the real FAIL", "X-SES-Spam-Verdict: PASS\r\nX-SES-Spam-Verdict: FAIL\r\n", VerdictRejected},
		{"spf fail", "Authentication-Results: amazonses.com; spf=fail smtp.mailfrom=example.com; dkim=pass header.i=@example.com\r\n", VerdictRejected},
		{"spf softfail is not FAIL", "Authentication-Results: amazonses.com; spf=softfail smtp.mailfrom=example.com; dkim=pass\r\n", VerdictPass},
		{"dkim fail with no pass", "Authentication-Results: amazonses.com; spf=pass; dkim=fail header.i=@example.com\r\n", VerdictRejected},
		{"dkim fail plus a passing signature", "Authentication-Results: amazonses.com; spf=pass; dkim=fail header.i=@a.example; dkim=pass header.i=@b.example\r\n", VerdictPass},
		{"other server's results are ignored", "Authentication-Results: mx.google.com; spf=fail smtp.mailfrom=x; dkim=fail\r\nAuthentication-Results: amazonses.com; spf=pass; dkim=pass\r\n", VerdictPass},
		{"case-insensitive authserv-id", "Authentication-Results: AmazonSES.com; spf=FAIL\r\n", VerdictRejected},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			raw := []byte(tt.headers + base + body)
			if got := VerdictFromHeaders(raw); got != tt.want {
				t.Errorf("VerdictFromHeaders() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVerdictFromHeaders_UnreadableHeadersPass(t *testing.T) {
	if got := VerdictFromHeaders([]byte("not an email at all")); got != VerdictPass {
		t.Errorf("got %v, want pass (the parser reports unreadable mail)", got)
	}
}
