package handler

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	dbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	ssmtypes "github.com/aws/aws-sdk-go-v2/service/ssm/types"
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"

	"github.com/sh3r4rd/sh3r4rd.com/infra/recruiter-dashboard/lambda-src/email-parser/internal/db"
	"github.com/sh3r4rd/sh3r4rd.com/infra/recruiter-dashboard/lambda-src/email-parser/internal/extractor"
	localssm "github.com/sh3r4rd/sh3r4rd.com/infra/recruiter-dashboard/lambda-src/email-parser/internal/ssm"
	"github.com/sh3r4rd/sh3r4rd.com/infra/recruiter-dashboard/lambda-src/email-parser/internal/tagger"
)

// --- Mock implementations ---

type mockS3ReadClient struct {
	getObjectFn func(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error)
}

func (m *mockS3ReadClient) GetObject(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
	return m.getObjectFn(ctx, params, optFns...)
}

type mockS3TagClient struct {
	putObjectTaggingFn func(ctx context.Context, params *s3.PutObjectTaggingInput, optFns ...func(*s3.Options)) (*s3.PutObjectTaggingOutput, error)
}

func (m *mockS3TagClient) PutObjectTagging(ctx context.Context, params *s3.PutObjectTaggingInput, optFns ...func(*s3.Options)) (*s3.PutObjectTaggingOutput, error) {
	return m.putObjectTaggingFn(ctx, params, optFns...)
}

type mockDynamoDBClient struct {
	putItemFn func(ctx context.Context, params *dynamodb.PutItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error)
}

func (m *mockDynamoDBClient) PutItem(ctx context.Context, params *dynamodb.PutItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error) {
	return m.putItemFn(ctx, params, optFns...)
}

type mockSSMClient struct {
	getParameterFn func(ctx context.Context, params *ssm.GetParameterInput, optFns ...func(*ssm.Options)) (*ssm.GetParameterOutput, error)
}

func (m *mockSSMClient) GetParameter(ctx context.Context, params *ssm.GetParameterInput, optFns ...func(*ssm.Options)) (*ssm.GetParameterOutput, error) {
	return m.getParameterFn(ctx, params, optFns...)
}

type mockOpenAIClient struct {
	newFn func(ctx context.Context, body openai.ChatCompletionNewParams, opts ...option.RequestOption) (*openai.ChatCompletion, error)
}

func (m *mockOpenAIClient) New(ctx context.Context, body openai.ChatCompletionNewParams, opts ...option.RequestOption) (*openai.ChatCompletion, error) {
	return m.newFn(ctx, body, opts...)
}

// --- Test helpers ---

func newTestSESEvent(messageID string) events.SimpleEmailEvent {
	return events.SimpleEmailEvent{
		Records: []events.SimpleEmailRecord{
			{
				SES: events.SimpleEmailService{
					Mail: events.SimpleEmailMessage{
						MessageID: messageID,
					},
					Receipt: events.SimpleEmailReceipt{
						SPFVerdict:   events.SimpleEmailVerdict{Status: "PASS"},
						DKIMVerdict:  events.SimpleEmailVerdict{Status: "PASS"},
						SpamVerdict:  events.SimpleEmailVerdict{Status: "PASS"},
						VirusVerdict: events.SimpleEmailVerdict{Status: "PASS"},
					},
				},
			},
		},
	}
}

func newTestHandler(s3Body string, openAIResponse string) *Handler {
	s3Read := &mockS3ReadClient{
		getObjectFn: func(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
			return &s3.GetObjectOutput{
				Body: io.NopCloser(strings.NewReader(s3Body)),
			}, nil
		},
	}

	s3Tag := &mockS3TagClient{
		putObjectTaggingFn: func(ctx context.Context, params *s3.PutObjectTaggingInput, optFns ...func(*s3.Options)) (*s3.PutObjectTaggingOutput, error) {
			return &s3.PutObjectTaggingOutput{}, nil
		},
	}

	dynamo := &mockDynamoDBClient{
		putItemFn: func(ctx context.Context, params *dynamodb.PutItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error) {
			return &dynamodb.PutItemOutput{}, nil
		},
	}

	ssmMock := &mockSSMClient{
		getParameterFn: func(ctx context.Context, params *ssm.GetParameterInput, optFns ...func(*ssm.Options)) (*ssm.GetParameterOutput, error) {
			return &ssm.GetParameterOutput{
				Parameter: &ssmtypes.Parameter{
					Value: aws.String("test-openai-key"),
				},
			}, nil
		},
	}

	openaiMock := &mockOpenAIClient{
		newFn: func(ctx context.Context, body openai.ChatCompletionNewParams, opts ...option.RequestOption) (*openai.ChatCompletion, error) {
			return &openai.ChatCompletion{
				Choices: []openai.ChatCompletionChoice{
					{
						Message: openai.ChatCompletionMessage{
							Content: openAIResponse,
						},
						FinishReason: "stop",
					},
				},
			}, nil
		},
	}

	store := db.NewStore(dynamo, "test-table")
	t := tagger.NewTagger(s3Tag)
	ssmFetcher := localssm.NewParameterFetcher(ssmMock)
	ext := extractor.NewExtractor(openaiMock)

	h := NewHandler(s3Read, store, t, ssmFetcher)
	h.extractor = ext

	return h
}

// --- Tests ---

func TestHandleSESEvent_Success(t *testing.T) {
	os.Setenv("EMAIL_BUCKET", "test-bucket")
	os.Setenv("S3_KEY_PREFIX", "incoming")
	os.Setenv("SSM_OPENAI_KEY_NAME", "/test/openai-key")
	defer func() {
		os.Unsetenv("EMAIL_BUCKET")
		os.Unsetenv("S3_KEY_PREFIX")
		os.Unsetenv("SSM_OPENAI_KEY_NAME")
	}()

	emailBody := "From: test@test.com\r\nTo: me@test.com\r\nSubject: Test\r\nDate: Mon, 03 Mar 2026 10:00:00 +0000\r\n\r\nHello, this is a test email."
	openAIResp := `{"recruiter_first_name":"Test","recruiter_last_name":"Recruiter","recruiter_email":"test@test.com","company":"TestCorp","job_title":"Engineer","phone":"","confidence":0.9}`

	h := newTestHandler(emailBody, openAIResp)
	event := newTestSESEvent("test-msg-001")

	summary, err := h.HandleSESEvent(context.Background(), event)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if summary.Total != 1 {
		t.Errorf("expected Total=1, got %d", summary.Total)
	}
	if summary.Success != 1 {
		t.Errorf("expected Success=1, got %d", summary.Success)
	}
	if summary.Failed != 0 {
		t.Errorf("expected Failed=0, got %d", summary.Failed)
	}
}

func TestHandleSESEvent_SPFVerdictFail(t *testing.T) {
	os.Setenv("EMAIL_BUCKET", "test-bucket")
	os.Setenv("S3_KEY_PREFIX", "incoming")
	os.Setenv("SSM_OPENAI_KEY_NAME", "/test/openai-key")
	defer func() {
		os.Unsetenv("EMAIL_BUCKET")
		os.Unsetenv("S3_KEY_PREFIX")
		os.Unsetenv("SSM_OPENAI_KEY_NAME")
	}()

	h := newTestHandler("", "")
	event := newTestSESEvent("spf-fail-msg")
	event.Records[0].SES.Receipt.SPFVerdict.Status = "FAIL"

	summary, err := h.HandleSESEvent(context.Background(), event)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if summary.Rejected != 1 {
		t.Errorf("expected Rejected=1 for SPF FAIL, got %d", summary.Rejected)
	}
	if summary.Success != 0 {
		t.Errorf("expected Success=0, got %d", summary.Success)
	}
}

func TestHandleSESEvent_DKIMVerdictFail(t *testing.T) {
	os.Setenv("EMAIL_BUCKET", "test-bucket")
	os.Setenv("S3_KEY_PREFIX", "incoming")
	os.Setenv("SSM_OPENAI_KEY_NAME", "/test/openai-key")
	defer func() {
		os.Unsetenv("EMAIL_BUCKET")
		os.Unsetenv("S3_KEY_PREFIX")
		os.Unsetenv("SSM_OPENAI_KEY_NAME")
	}()

	h := newTestHandler("", "")
	event := newTestSESEvent("dkim-fail-msg")
	event.Records[0].SES.Receipt.DKIMVerdict.Status = "FAIL"

	summary, err := h.HandleSESEvent(context.Background(), event)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if summary.Rejected != 1 {
		t.Errorf("expected Rejected=1 for DKIM FAIL, got %d", summary.Rejected)
	}
}

func TestHandleSESEvent_EmptyEvent(t *testing.T) {
	os.Setenv("EMAIL_BUCKET", "test-bucket")
	os.Setenv("S3_KEY_PREFIX", "incoming")
	os.Setenv("SSM_OPENAI_KEY_NAME", "/test/openai-key")
	defer func() {
		os.Unsetenv("EMAIL_BUCKET")
		os.Unsetenv("S3_KEY_PREFIX")
		os.Unsetenv("SSM_OPENAI_KEY_NAME")
	}()

	h := newTestHandler("", "")
	event := events.SimpleEmailEvent{Records: []events.SimpleEmailRecord{}}

	_, err := h.HandleSESEvent(context.Background(), event)
	if err == nil {
		t.Fatal("expected error for empty event")
	}
}

func TestHandleSESEvent_S3FetchFailure(t *testing.T) {
	os.Setenv("EMAIL_BUCKET", "test-bucket")
	os.Setenv("S3_KEY_PREFIX", "incoming")
	os.Setenv("SSM_OPENAI_KEY_NAME", "/test/openai-key")
	defer func() {
		os.Unsetenv("EMAIL_BUCKET")
		os.Unsetenv("S3_KEY_PREFIX")
		os.Unsetenv("SSM_OPENAI_KEY_NAME")
	}()

	s3Read := &mockS3ReadClient{
		getObjectFn: func(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
			return nil, fmt.Errorf("NoSuchKey: The specified key does not exist")
		},
	}

	s3Tag := &mockS3TagClient{
		putObjectTaggingFn: func(ctx context.Context, params *s3.PutObjectTaggingInput, optFns ...func(*s3.Options)) (*s3.PutObjectTaggingOutput, error) {
			return &s3.PutObjectTaggingOutput{}, nil
		},
	}

	dynamo := &mockDynamoDBClient{
		putItemFn: func(ctx context.Context, params *dynamodb.PutItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error) {
			return &dynamodb.PutItemOutput{}, nil
		},
	}

	ssmMock := &mockSSMClient{
		getParameterFn: func(ctx context.Context, params *ssm.GetParameterInput, optFns ...func(*ssm.Options)) (*ssm.GetParameterOutput, error) {
			return &ssm.GetParameterOutput{
				Parameter: &ssmtypes.Parameter{Value: aws.String("key")},
			}, nil
		},
	}

	openaiMock := &mockOpenAIClient{
		newFn: func(ctx context.Context, body openai.ChatCompletionNewParams, opts ...option.RequestOption) (*openai.ChatCompletion, error) {
			return nil, fmt.Errorf("should not be called")
		},
	}

	store := db.NewStore(dynamo, "test-table")
	tgr := tagger.NewTagger(s3Tag)
	ssmFetcher := localssm.NewParameterFetcher(ssmMock)
	ext := extractor.NewExtractor(openaiMock)

	h := NewHandler(s3Read, store, tgr, ssmFetcher)
	h.extractor = ext

	event := newTestSESEvent("s3-fail-msg")
	summary, err := h.HandleSESEvent(context.Background(), event)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if summary.Failed != 1 {
		t.Errorf("expected Failed=1 for S3 fetch failure, got %d", summary.Failed)
	}
}

func TestHandleSESEvent_DynamoDBWriteFailure(t *testing.T) {
	os.Setenv("EMAIL_BUCKET", "test-bucket")
	os.Setenv("S3_KEY_PREFIX", "incoming")
	os.Setenv("SSM_OPENAI_KEY_NAME", "/test/openai-key")
	defer func() {
		os.Unsetenv("EMAIL_BUCKET")
		os.Unsetenv("S3_KEY_PREFIX")
		os.Unsetenv("SSM_OPENAI_KEY_NAME")
	}()

	emailBody := "From: test@test.com\r\nTo: me@test.com\r\nSubject: Test\r\nDate: Mon, 03 Mar 2026 10:00:00 +0000\r\n\r\nTest body."
	openAIResp := `{"recruiter_first_name":"Test","recruiter_last_name":"User","recruiter_email":"test@test.com","company":"Corp","job_title":"Eng","phone":"","confidence":0.5}`

	s3Read := &mockS3ReadClient{
		getObjectFn: func(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
			return &s3.GetObjectOutput{
				Body: io.NopCloser(strings.NewReader(emailBody)),
			}, nil
		},
	}

	s3Tag := &mockS3TagClient{
		putObjectTaggingFn: func(ctx context.Context, params *s3.PutObjectTaggingInput, optFns ...func(*s3.Options)) (*s3.PutObjectTaggingOutput, error) {
			return &s3.PutObjectTaggingOutput{}, nil
		},
	}

	dynamo := &mockDynamoDBClient{
		putItemFn: func(ctx context.Context, params *dynamodb.PutItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error) {
			return nil, fmt.Errorf("DynamoDB ProvisionedThroughputExceededException")
		},
	}

	ssmMock := &mockSSMClient{
		getParameterFn: func(ctx context.Context, params *ssm.GetParameterInput, optFns ...func(*ssm.Options)) (*ssm.GetParameterOutput, error) {
			return &ssm.GetParameterOutput{
				Parameter: &ssmtypes.Parameter{Value: aws.String("key")},
			}, nil
		},
	}

	openaiMock := &mockOpenAIClient{
		newFn: func(ctx context.Context, body openai.ChatCompletionNewParams, opts ...option.RequestOption) (*openai.ChatCompletion, error) {
			return &openai.ChatCompletion{
				Choices: []openai.ChatCompletionChoice{
					{Message: openai.ChatCompletionMessage{Content: openAIResp}, FinishReason: "stop"},
				},
			}, nil
		},
	}

	store := db.NewStore(dynamo, "test-table")
	tgr := tagger.NewTagger(s3Tag)
	ssmFetcher := localssm.NewParameterFetcher(ssmMock)
	ext := extractor.NewExtractor(openaiMock)

	h := NewHandler(s3Read, store, tgr, ssmFetcher)
	h.extractor = ext

	event := newTestSESEvent("dynamo-fail-msg")
	summary, err := h.HandleSESEvent(context.Background(), event)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if summary.Failed != 1 {
		t.Errorf("expected Failed=1 for DynamoDB failure, got %d", summary.Failed)
	}
}

func TestHandleSESEvent_DuplicateDetection(t *testing.T) {
	os.Setenv("EMAIL_BUCKET", "test-bucket")
	os.Setenv("S3_KEY_PREFIX", "incoming")
	os.Setenv("SSM_OPENAI_KEY_NAME", "/test/openai-key")
	defer func() {
		os.Unsetenv("EMAIL_BUCKET")
		os.Unsetenv("S3_KEY_PREFIX")
		os.Unsetenv("SSM_OPENAI_KEY_NAME")
	}()

	emailBody := "From: test@test.com\r\nTo: me@test.com\r\nSubject: Test\r\nDate: Mon, 03 Mar 2026 10:00:00 +0000\r\n\r\nTest body."
	openAIResp := `{"recruiter_first_name":"Test","recruiter_last_name":"User","recruiter_email":"test@test.com","company":"Corp","job_title":"Eng","phone":"","confidence":0.5}`

	s3Read := &mockS3ReadClient{
		getObjectFn: func(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
			return &s3.GetObjectOutput{
				Body: io.NopCloser(strings.NewReader(emailBody)),
			}, nil
		},
	}

	s3Tag := &mockS3TagClient{
		putObjectTaggingFn: func(ctx context.Context, params *s3.PutObjectTaggingInput, optFns ...func(*s3.Options)) (*s3.PutObjectTaggingOutput, error) {
			return &s3.PutObjectTaggingOutput{}, nil
		},
	}

	dynamo := &mockDynamoDBClient{
		putItemFn: func(ctx context.Context, params *dynamodb.PutItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error) {
			return nil, &dbtypes.ConditionalCheckFailedException{
				Message: aws.String("The conditional request failed"),
			}
		},
	}

	ssmMock := &mockSSMClient{
		getParameterFn: func(ctx context.Context, params *ssm.GetParameterInput, optFns ...func(*ssm.Options)) (*ssm.GetParameterOutput, error) {
			return &ssm.GetParameterOutput{
				Parameter: &ssmtypes.Parameter{Value: aws.String("key")},
			}, nil
		},
	}

	openaiMock := &mockOpenAIClient{
		newFn: func(ctx context.Context, body openai.ChatCompletionNewParams, opts ...option.RequestOption) (*openai.ChatCompletion, error) {
			return &openai.ChatCompletion{
				Choices: []openai.ChatCompletionChoice{
					{Message: openai.ChatCompletionMessage{Content: openAIResp}, FinishReason: "stop"},
				},
			}, nil
		},
	}

	store := db.NewStore(dynamo, "test-table")
	tgr := tagger.NewTagger(s3Tag)
	ssmFetcher := localssm.NewParameterFetcher(ssmMock)
	ext := extractor.NewExtractor(openaiMock)

	h := NewHandler(s3Read, store, tgr, ssmFetcher)
	h.extractor = ext

	event := newTestSESEvent("dup-msg")
	summary, err := h.HandleSESEvent(context.Background(), event)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if summary.Duplicate != 1 {
		t.Errorf("expected Duplicate=1, got %d", summary.Duplicate)
	}
	if summary.Failed != 0 {
		t.Errorf("expected Failed=0 (duplicate is not a failure), got %d", summary.Failed)
	}
}

func TestHandleSESEvent_MultipleRecords(t *testing.T) {
	os.Setenv("EMAIL_BUCKET", "test-bucket")
	os.Setenv("S3_KEY_PREFIX", "incoming")
	os.Setenv("SSM_OPENAI_KEY_NAME", "/test/openai-key")
	defer func() {
		os.Unsetenv("EMAIL_BUCKET")
		os.Unsetenv("S3_KEY_PREFIX")
		os.Unsetenv("SSM_OPENAI_KEY_NAME")
	}()

	emailBody := "From: test@test.com\r\nTo: me@test.com\r\nSubject: Test\r\nDate: Mon, 03 Mar 2026 10:00:00 +0000\r\n\r\nTest body."
	openAIResp := `{"recruiter_first_name":"Test","recruiter_last_name":"User","recruiter_email":"test@test.com","company":"Corp","job_title":"Eng","phone":"","confidence":0.5}`

	h := newTestHandler(emailBody, openAIResp)

	event := events.SimpleEmailEvent{
		Records: []events.SimpleEmailRecord{
			{SES: events.SimpleEmailService{
				Mail:    events.SimpleEmailMessage{MessageID: "msg-1"},
				Receipt: events.SimpleEmailReceipt{SPFVerdict: events.SimpleEmailVerdict{Status: "PASS"}, DKIMVerdict: events.SimpleEmailVerdict{Status: "PASS"}, SpamVerdict: events.SimpleEmailVerdict{Status: "PASS"}, VirusVerdict: events.SimpleEmailVerdict{Status: "PASS"}},
			}},
			{SES: events.SimpleEmailService{
				Mail:    events.SimpleEmailMessage{MessageID: "msg-2"},
				Receipt: events.SimpleEmailReceipt{SPFVerdict: events.SimpleEmailVerdict{Status: "PASS"}, DKIMVerdict: events.SimpleEmailVerdict{Status: "PASS"}, SpamVerdict: events.SimpleEmailVerdict{Status: "PASS"}, VirusVerdict: events.SimpleEmailVerdict{Status: "PASS"}},
			}},
		},
	}

	summary, err := h.HandleSESEvent(context.Background(), event)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if summary.Total != 2 {
		t.Errorf("expected Total=2, got %d", summary.Total)
	}
	if summary.Success != 2 {
		t.Errorf("expected Success=2, got %d", summary.Success)
	}
}

func TestHandleSESEvent_SummaryCountsCorrect(t *testing.T) {
	os.Setenv("EMAIL_BUCKET", "test-bucket")
	os.Setenv("S3_KEY_PREFIX", "incoming")
	os.Setenv("SSM_OPENAI_KEY_NAME", "/test/openai-key")
	defer func() {
		os.Unsetenv("EMAIL_BUCKET")
		os.Unsetenv("S3_KEY_PREFIX")
		os.Unsetenv("SSM_OPENAI_KEY_NAME")
	}()

	emailBody := "From: test@test.com\r\nTo: me@test.com\r\nSubject: Test\r\nDate: Mon, 03 Mar 2026 10:00:00 +0000\r\n\r\nTest body."
	openAIResp := `{"recruiter_first_name":"Test","recruiter_last_name":"User","recruiter_email":"test@test.com","company":"Corp","job_title":"Eng","phone":"","confidence":0.5}`

	h := newTestHandler(emailBody, openAIResp)
	event := newTestSESEvent("summary-msg")

	summary, err := h.HandleSESEvent(context.Background(), event)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	total := summary.Success + summary.Failed + summary.Duplicate + summary.Rejected
	if total != summary.Total {
		t.Errorf("summary counts don't add up: success(%d)+failed(%d)+duplicate(%d)+rejected(%d) = %d, expected total=%d",
			summary.Success, summary.Failed, summary.Duplicate, summary.Rejected, total, summary.Total)
	}
}

