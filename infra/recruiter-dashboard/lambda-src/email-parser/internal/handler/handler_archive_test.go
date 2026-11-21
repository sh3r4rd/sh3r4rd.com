package handler

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"

	"github.com/sh3r4rd/sh3r4rd.com/infra/recruiter-dashboard/lambda-src/email-parser/internal/archive"
	"github.com/sh3r4rd/sh3r4rd.com/infra/recruiter-dashboard/lambda-src/email-parser/internal/db"
	"github.com/sh3r4rd/sh3r4rd.com/infra/recruiter-dashboard/lambda-src/email-parser/internal/extractor"
	localssm "github.com/sh3r4rd/sh3r4rd.com/infra/recruiter-dashboard/lambda-src/email-parser/internal/ssm"
	"github.com/sh3r4rd/sh3r4rd.com/infra/recruiter-dashboard/lambda-src/email-parser/internal/tagger"
)

const archiveTestOpenAIResp = `{"recruiter_first_name":"Test","recruiter_last_name":"Recruiter","recruiter_email":"test@test.com","company":"TestCorp","job_title":"Engineer","phone":"","confidence":0.9}`

const archiveTestEmail = "From: test@test.com\r\nTo: me@test.com\r\nSubject: Test\r\nDate: Mon, 03 Mar 2026 10:00:00 +0000\r\n\r\nHello, this is a test email."

// archiveHarness records the order of AWS calls made while processing an event, so
// tests can assert what ran, what didn't, and in which order.
type archiveHarness struct {
	calls []string // e.g. "copy archive/raw/x", "get incoming/x", "put archive/clean/x.json", "dynamo", "openai", "tag incoming/x"

	copyErr error // returned by CopyObject when set
	putErr  error // returned by PutObject when set

	// copyErrFor limits copyErr to one destination key; empty means every copy fails.
	copyErrFor string

	lastCopy *s3.CopyObjectInput
	cleanDoc []byte
	lastTags map[string]string
}

func (h *archiveHarness) count(prefix string) int {
	n := 0
	for _, c := range h.calls {
		if strings.HasPrefix(c, prefix) {
			n++
		}
	}
	return n
}

func (h *archiveHarness) indexOf(prefix string) int {
	for i, c := range h.calls {
		if strings.HasPrefix(c, prefix) {
			return i
		}
	}
	return -1
}

func newArchiveTestHandler(t *testing.T, hns *archiveHarness, emailBody string) *Handler {
	t.Helper()
	t.Setenv("EMAIL_BUCKET", "test-bucket")
	t.Setenv("S3_KEY_PREFIX", "incoming")
	t.Setenv("SSM_OPENAI_KEY_NAME", "/test/openai-key")

	s3Read := &mockS3ReadClient{
		getObjectFn: func(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
			hns.calls = append(hns.calls, "get "+aws.ToString(params.Key))
			return &s3.GetObjectOutput{Body: io.NopCloser(strings.NewReader(emailBody))}, nil
		},
	}
	s3Tag := &mockS3TagClient{
		putObjectTaggingFn: func(ctx context.Context, params *s3.PutObjectTaggingInput, optFns ...func(*s3.Options)) (*s3.PutObjectTaggingOutput, error) {
			hns.calls = append(hns.calls, "tag "+aws.ToString(params.Key))
			hns.lastTags = map[string]string{}
			for _, tag := range params.Tagging.TagSet {
				hns.lastTags[aws.ToString(tag.Key)] = aws.ToString(tag.Value)
			}
			return &s3.PutObjectTaggingOutput{}, nil
		},
	}
	s3Archive := &mockS3ArchiveClient{
		copyObjectFn: func(ctx context.Context, params *s3.CopyObjectInput, optFns ...func(*s3.Options)) (*s3.CopyObjectOutput, error) {
			key := aws.ToString(params.Key)
			hns.calls = append(hns.calls, "copy "+key)
			hns.lastCopy = params
			if hns.copyErr != nil && (hns.copyErrFor == "" || hns.copyErrFor == key) {
				return nil, hns.copyErr
			}
			return &s3.CopyObjectOutput{}, nil
		},
		putObjectFn: func(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
			hns.calls = append(hns.calls, "put "+aws.ToString(params.Key))
			body, err := io.ReadAll(params.Body)
			if err != nil {
				return nil, err
			}
			hns.cleanDoc = body
			if hns.putErr != nil {
				return nil, hns.putErr
			}
			return &s3.PutObjectOutput{}, nil
		},
	}
	dynamo := &mockDynamoDBClient{
		putItemFn: func(ctx context.Context, params *dynamodb.PutItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error) {
			hns.calls = append(hns.calls, "dynamo")
			return &dynamodb.PutItemOutput{}, nil
		},
	}
	openaiMock := &mockOpenAIClient{
		newFn: func(ctx context.Context, body openai.ChatCompletionNewParams, opts ...option.RequestOption) (*openai.ChatCompletion, error) {
			hns.calls = append(hns.calls, "openai")
			return &openai.ChatCompletion{
				Choices: []openai.ChatCompletionChoice{{
					Message:      openai.ChatCompletionMessage{Content: archiveTestOpenAIResp},
					FinishReason: "stop",
				}},
			}, nil
		},
	}

	h := NewHandler(
		s3Read,
		db.NewStore(dynamo, "test-table"),
		tagger.NewTagger(s3Tag),
		localssm.NewParameterFetcher(&mockSSMClient{}),
		archive.NewArchiver(s3Archive, testArchiveKMSKeyARN),
	)
	h.extractor = extractor.NewExtractor(openaiMock)
	return h
}

func TestHandleSESEvent_ArchiveByVerdict(t *testing.T) {
	tests := []struct {
		name         string
		spf          string
		dkim         string
		spam         string
		virus        string
		wantCopy     bool
		wantParse    bool
		wantRejected int
		wantSuccess  int
	}{
		{"all pass: archived and parsed", "PASS", "PASS", "PASS", "PASS", true, true, 0, 1},
		{"gray verdicts are not failures", "GRAY", "GRAY", "PASS", "PASS", true, true, 0, 1},
		{"spf fail: archived, not parsed", "FAIL", "PASS", "PASS", "PASS", true, false, 1, 0},
		{"dkim fail: archived, not parsed", "PASS", "FAIL", "PASS", "PASS", true, false, 1, 0},
		{"spam fail: archived, not parsed", "PASS", "PASS", "FAIL", "PASS", true, false, 1, 0},
		{"lowercase fail still rejects", "PASS", "PASS", "fail", "PASS", true, false, 1, 0},
		{"virus fail: not archived", "PASS", "PASS", "PASS", "FAIL", false, false, 1, 0},
		{"virus and spam fail: not archived", "PASS", "PASS", "FAIL", "FAIL", false, false, 1, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hns := &archiveHarness{}
			h := newArchiveTestHandler(t, hns, archiveTestEmail)
			event := newTestSESEvent("verdict-msg")
			r := &event.Records[0].SES.Receipt
			r.SPFVerdict.Status, r.DKIMVerdict.Status, r.SpamVerdict.Status, r.VirusVerdict.Status = tt.spf, tt.dkim, tt.spam, tt.virus

			summary, err := h.HandleSESEvent(context.Background(), event)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got := hns.count("copy archive/raw/verdict-msg") == 1; got != tt.wantCopy {
				t.Errorf("raw copied = %v, want %v (calls %v)", got, tt.wantCopy, hns.calls)
			}
			parsed := hns.count("get ") > 0 || hns.count("openai") > 0 || hns.count("put ") > 0 || hns.count("dynamo") > 0
			if parsed != tt.wantParse {
				t.Errorf("parsed = %v, want %v (calls %v)", parsed, tt.wantParse, hns.calls)
			}
			if summary.Rejected != tt.wantRejected || summary.Success != tt.wantSuccess {
				t.Errorf("summary = %+v, want rejected=%d success=%d", summary, tt.wantRejected, tt.wantSuccess)
			}
		})
	}
}

func TestHandleSESEvent_RawCopyRunsBeforeParsing(t *testing.T) {
	hns := &archiveHarness{}
	h := newArchiveTestHandler(t, hns, archiveTestEmail)

	if _, err := h.HandleSESEvent(context.Background(), newTestSESEvent("order-msg")); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := []string{
		"copy archive/raw/order-msg",
		"get incoming/order-msg",
		"openai",
		"put archive/clean/order-msg.json",
		"dynamo",
		"tag incoming/order-msg",
	}
	if strings.Join(hns.calls, "|") != strings.Join(want, "|") {
		t.Errorf("call order =\n  %v\nwant\n  %v", hns.calls, want)
	}

	in := hns.lastCopy
	if aws.ToString(in.CopySource) != "test-bucket/incoming/order-msg" {
		t.Errorf("CopySource = %q", aws.ToString(in.CopySource))
	}
	if in.TaggingDirective != types.TaggingDirectiveReplace || in.Tagging != nil {
		t.Errorf("copy must carry no tags: directive=%q tagging=%v", in.TaggingDirective, in.Tagging)
	}
	if in.ServerSideEncryption != types.ServerSideEncryptionAwsKms || aws.ToString(in.SSEKMSKeyId) != testArchiveKMSKeyARN {
		t.Errorf("copy must use SSE-KMS with the archive key: %q %q", in.ServerSideEncryption, aws.ToString(in.SSEKMSKeyId))
	}
}

func TestHandleSESEvent_RawCopyFailureFailsInvocation(t *testing.T) {
	copyErr := errors.New("AccessDenied: simulated")
	hns := &archiveHarness{copyErr: copyErr}
	h := newArchiveTestHandler(t, hns, archiveTestEmail)

	summary, err := h.HandleSESEvent(context.Background(), newTestSESEvent("copy-fail-msg"))
	if err == nil {
		t.Fatal("expected an error so Lambda retries the invocation")
	}
	if !errors.Is(err, copyErr) {
		t.Errorf("error %v does not wrap the S3 error", err)
	}
	if !strings.Contains(err.Error(), "copy-fail-msg") {
		t.Errorf("error %q should name the message id", err)
	}
	if summary.Failed != 1 || summary.Success != 0 {
		t.Errorf("summary = %+v, want failed=1", summary)
	}
	for _, step := range []string{"get ", "openai", "put ", "dynamo", "tag "} {
		if hns.count(step) != 0 {
			t.Errorf("%q ran after a failed raw copy (calls %v)", step, hns.calls)
		}
	}
}

func TestHandleSESEvent_RawCopyFailureOnRejectedEmailFailsInvocation(t *testing.T) {
	hns := &archiveHarness{copyErr: errors.New("AccessDenied: simulated")}
	h := newArchiveTestHandler(t, hns, archiveTestEmail)
	event := newTestSESEvent("spam-copy-fail")
	event.Records[0].SES.Receipt.SpamVerdict.Status = "FAIL"

	summary, err := h.HandleSESEvent(context.Background(), event)
	if err == nil {
		t.Fatal("expected an error: spam is archived, so a failed copy must be retried")
	}
	if summary.Failed != 1 || summary.Rejected != 0 {
		t.Errorf("summary = %+v, want failed=1 rejected=0", summary)
	}
}

func TestHandleSESEvent_CleanWriteFailureDoesNotFailParse(t *testing.T) {
	hns := &archiveHarness{putErr: errors.New("AccessDenied: simulated")}
	h := newArchiveTestHandler(t, hns, archiveTestEmail)

	summary, err := h.HandleSESEvent(context.Background(), newTestSESEvent("clean-fail-msg"))
	if err != nil {
		t.Fatalf("a clean write failure must not fail the invocation: %v", err)
	}
	if summary.Success != 1 || summary.Failed != 0 {
		t.Errorf("summary = %+v, want success=1", summary)
	}
	if hns.count("copy archive/raw/clean-fail-msg") != 1 {
		t.Error("the raw copy should exist so the backfill can regenerate the record")
	}
	if hns.count("dynamo") != 1 {
		t.Error("the DynamoDB write should still happen")
	}
	if hns.lastTags["parse-status"] != "success" {
		t.Errorf("parse-status tag = %q, want success", hns.lastTags["parse-status"])
	}
}

func TestHandleSESEvent_CleanRecordContents(t *testing.T) {
	fixture, err := os.ReadFile("../../testdata/gmail_forward.eml")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	hns := &archiveHarness{}
	h := newArchiveTestHandler(t, hns, string(fixture))

	if _, err := h.HandleSESEvent(context.Background(), newTestSESEvent("forwarded-msg")); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var rec archive.CleanRecord
	if err := json.Unmarshal(hns.cleanDoc, &rec); err != nil {
		t.Fatalf("clean record is not valid JSON: %v", err)
	}
	if rec.MessageID != "forwarded-msg" || rec.RawKey != "archive/raw/forwarded-msg" {
		t.Errorf("ids = %q / %q", rec.MessageID, rec.RawKey)
	}
	if rec.SchemaVersion != archive.SchemaVersion || rec.ParserVersion != archive.ParserVersion {
		t.Errorf("versions = %d / %q", rec.SchemaVersion, rec.ParserVersion)
	}
	if !rec.IsForwarded {
		t.Error("IsForwarded should be true for the Gmail forward fixture")
	}
	if strings.TrimSpace(rec.Body) == "" || rec.Headers.From == "" {
		t.Errorf("body/headers missing: from=%q body=%d bytes", rec.Headers.From, len(rec.Body))
	}
	if rec.ExtractionSource != archive.ExtractionSourceLive || rec.Extraction == nil {
		t.Fatalf("extraction = %+v (%q), want the live extraction", rec.Extraction, rec.ExtractionSource)
	}
	if rec.Extraction.Company != "TestCorp" || rec.Extraction.Confidence != 0.9 {
		t.Errorf("extraction = %+v", rec.Extraction)
	}
}

func TestHandleSESEvent_FatalErrorStillProcessesOtherRecords(t *testing.T) {
	hns := &archiveHarness{copyErr: errors.New("AccessDenied: simulated"), copyErrFor: "archive/raw/msg-bad"}
	h := newArchiveTestHandler(t, hns, archiveTestEmail)

	pass := events.SimpleEmailReceipt{
		SPFVerdict:   events.SimpleEmailVerdict{Status: "PASS"},
		DKIMVerdict:  events.SimpleEmailVerdict{Status: "PASS"},
		SpamVerdict:  events.SimpleEmailVerdict{Status: "PASS"},
		VirusVerdict: events.SimpleEmailVerdict{Status: "PASS"},
	}
	event := events.SimpleEmailEvent{Records: []events.SimpleEmailRecord{
		{SES: events.SimpleEmailService{Mail: events.SimpleEmailMessage{MessageID: "msg-bad"}, Receipt: pass}},
		{SES: events.SimpleEmailService{Mail: events.SimpleEmailMessage{MessageID: "msg-good"}, Receipt: pass}},
	}}

	summary, err := h.HandleSESEvent(context.Background(), event)
	if err == nil || !strings.Contains(err.Error(), "msg-bad") {
		t.Fatalf("error = %v, want one naming msg-bad", err)
	}
	if summary.Failed != 1 || summary.Success != 1 {
		t.Errorf("summary = %+v, want failed=1 success=1", summary)
	}
	if hns.count("put archive/clean/msg-good.json") != 1 {
		t.Errorf("msg-good should still be processed (calls %v)", hns.calls)
	}
}
