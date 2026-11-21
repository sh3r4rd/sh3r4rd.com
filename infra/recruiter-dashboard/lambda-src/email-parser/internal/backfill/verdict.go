package backfill

import (
	"bytes"
	"net/mail"
	"regexp"
	"strings"
)

// Verdict is what the SES verdict headers in a stored email say about it. It mirrors
// the Lambda's rule (handler.evaluateVerdicts), which reads the same verdicts from the
// SES event instead of the headers.
type Verdict int

const (
	// VerdictPass: no verdict header says FAIL (or the headers are absent).
	VerdictPass Verdict = iota
	// VerdictRejected: SPF, DKIM or spam FAIL. Archived in archive/raw/, never parsed.
	VerdictRejected
	// VerdictVirus: virus FAIL. Never archived.
	VerdictVirus
)

// String returns a short label for counts and logs.
func (v Verdict) String() string {
	switch v {
	case VerdictRejected:
		return "rejected"
	case VerdictVirus:
		return "virus"
	default:
		return "pass"
	}
}

// authResultRe matches "spf=<result>" and "dkim=<result>" in an Authentication-Results value.
var authResultRe = regexp.MustCompile(`(?i)\b(spf|dkim)=([a-z]+)`)

// VerdictFromHeaders reads the verdicts SES writes into every message it stores when
// scanning is enabled (scan_enabled = true in modules/ses):
//
//   - X-SES-Spam-Verdict and X-SES-Virus-Verdict: PASS, FAIL, GRAY or PROCESSING_FAILED.
//   - Authentication-Results whose authserv-id is amazonses.com: spf=… and dkim=….
//
// Only an explicit FAIL rejects, as in the Lambda. SPF fails on spf=fail. DKIM fails when
// at least one signature is dkim=fail and none is dkim=pass. Authentication-Results
// headers added by other servers (for example inside forwarded mail) are ignored.
// A message whose headers can't be read is treated as passing; the parser reports it.
func VerdictFromHeaders(raw []byte) Verdict {
	msg, err := mail.ReadMessage(bytes.NewReader(raw))
	if err != nil {
		return VerdictPass
	}

	if strings.EqualFold(strings.TrimSpace(msg.Header.Get("X-SES-Virus-Verdict")), "FAIL") {
		return VerdictVirus
	}

	rejected := strings.EqualFold(strings.TrimSpace(msg.Header.Get("X-SES-Spam-Verdict")), "FAIL")

	for _, value := range msg.Header["Authentication-Results"] {
		authservID, _, _ := strings.Cut(strings.TrimSpace(value), ";")
		if !strings.EqualFold(strings.TrimSpace(authservID), "amazonses.com") {
			continue
		}

		dkimPass, dkimFail := false, false
		for _, m := range authResultRe.FindAllStringSubmatch(value, -1) {
			method, result := strings.ToLower(m[1]), strings.ToLower(m[2])
			switch {
			case method == "spf" && result == "fail":
				rejected = true
			case method == "dkim" && result == "pass":
				dkimPass = true
			case method == "dkim" && result == "fail":
				dkimFail = true
			}
		}
		if dkimFail && !dkimPass {
			rejected = true
		}
	}

	if rejected {
		return VerdictRejected
	}
	return VerdictPass
}
