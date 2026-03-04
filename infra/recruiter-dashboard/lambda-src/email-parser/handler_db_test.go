package main

import (
	"context"
	"errors"
	"io"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	s3svc "github.com/aws/aws-sdk-go-v2/service/s3"
)

// --- Test doubles ---

type mockS3Client struct {
	// getObjectFn is called when GetObject is invoked.
	getObjectFn func(ctx context.Context, params *s3svc.GetObjectInput, optFns ...func(*s3svc.Options)) (*s3svc.GetObjectOutput, error)
}

func (m *mockS3Client) GetObject(ctx context.Context, params *s3svc.GetObjectInput, optFns ...func(*s3svc.Options)) (*s3svc.GetObjectOutput, error) {
	return m.getObjectFn(ctx, params, optFns...)
}

type mockDynamoClient struct {
	// putItemFn is called when PutItem is invoked.
	putItemFn func(ctx context.Context, params *dynamodb.PutItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error)
}

func (m *mockDynamoClient) PutItem(ctx context.Context, params *dynamodb.PutItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error) {
	return m.putItemFn(ctx, params, optFns...)
}

// rawEmail builds a minimal plain-text RFC 5322 email.
func rawEmail(from, subject, body string) string {
	return "From: " + from + "\r\nTo: archive@example.com\r\nSubject: " + subject + "\r\nContent-Type: text/plain\r\n\r\n" + body
}

// bodyReader wraps a string as an io.ReadCloser.
func bodyReader(s string) io.ReadCloser {
	return io.NopCloser(strings.NewReader(s))
}

// --- Handler tests ---

func TestHandler_NoRecords(t *testing.T) {
	results, err := Handler(context.Background(), events.S3Event{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if results != nil {
		t.Errorf("expected nil results for empty event, got %v", results)
	}
}

func TestHandler_S3FetchError(t *testing.T) {
	s3Client = &mockS3Client{
		getObjectFn: func(_ context.Context, _ *s3svc.GetObjectInput, _ ...func(*s3svc.Options)) (*s3svc.GetObjectOutput, error) {
			return nil, errors.New("access denied")
		},
	}
	t.Cleanup(func() { s3Client = nil })

	event := events.S3Event{
		Records: []events.S3EventRecord{
			{S3: events.S3Entity{Bucket: events.S3Bucket{Name: "my-bucket"}, Object: events.S3Object{Key: "email.eml"}}},
		},
	}
	results, err := Handler(context.Background(), event)
	if err != nil {
		t.Fatalf("Handler must not propagate record-level errors: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Error == "" {
		t.Error("expected non-empty error field for S3 fetch failure")
	}
}

func TestHandler_SuccessfulRecord(t *testing.T) {
	t.Setenv("RECRUITER_TABLE", "recruiter-emails-test")

	email := rawEmail("Alice Smith <alice@acme.com>", "Senior Go Engineer role", "I'm a recruiter at Acme.")
	s3Client = &mockS3Client{
		getObjectFn: func(_ context.Context, _ *s3svc.GetObjectInput, _ ...func(*s3svc.Options)) (*s3svc.GetObjectOutput, error) {
			return &s3svc.GetObjectOutput{Body: bodyReader(email)}, nil
		},
	}
	// Inject a DynamoDB client that succeeds.
	resetTableClient()
	tableClient = &mockDynamoClient{
		putItemFn: func(_ context.Context, _ *dynamodb.PutItemInput, _ ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error) {
			return &dynamodb.PutItemOutput{}, nil
		},
	}
	t.Cleanup(func() {
		s3Client = nil
		resetTableClient()
	})

	event := events.S3Event{
		Records: []events.S3EventRecord{
			{S3: events.S3Entity{Bucket: events.S3Bucket{Name: "my-bucket"}, Object: events.S3Object{Key: "email.eml"}}},
		},
	}
	results, err := Handler(context.Background(), event)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Error != "" {
		t.Errorf("unexpected error in result: %s", results[0].Error)
	}
	if results[0].Duplicate {
		t.Error("expected Duplicate=false for new email")
	}
}

func TestHandler_DuplicateRecord(t *testing.T) {
	t.Setenv("RECRUITER_TABLE", "recruiter-emails-test")

	email := rawEmail("r@acme.com", "Opportunity", "Hello!")
	s3Client = &mockS3Client{
		getObjectFn: func(_ context.Context, _ *s3svc.GetObjectInput, _ ...func(*s3svc.Options)) (*s3svc.GetObjectOutput, error) {
			return &s3svc.GetObjectOutput{Body: bodyReader(email)}, nil
		},
	}
	resetTableClient()
	tableClient = &mockDynamoClient{
		putItemFn: func(_ context.Context, _ *dynamodb.PutItemInput, _ ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error) {
			return nil, &types.ConditionalCheckFailedException{Message: strPtr("The conditional request failed")}
		},
	}
	t.Cleanup(func() {
		s3Client = nil
		resetTableClient()
	})

	event := events.S3Event{
		Records: []events.S3EventRecord{
			{S3: events.S3Entity{Bucket: events.S3Bucket{Name: "bucket"}, Object: events.S3Object{Key: "dup.eml"}}},
		},
	}
	results, err := Handler(context.Background(), event)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if !results[0].Duplicate {
		t.Error("expected Duplicate=true for ConditionalCheckFailedException")
	}
	if results[0].Error != "" {
		t.Errorf("duplicates must not be treated as errors, got: %s", results[0].Error)
	}
}

func TestHandler_MultipleRecordsIndependent(t *testing.T) {
	t.Setenv("RECRUITER_TABLE", "recruiter-emails-test")

	callCount := 0
	s3Client = &mockS3Client{
		getObjectFn: func(_ context.Context, params *s3svc.GetObjectInput, _ ...func(*s3svc.Options)) (*s3svc.GetObjectOutput, error) {
			callCount++
			if *params.Key == "bad.eml" {
				return nil, errors.New("not found")
			}
			email := rawEmail("r@acme.com", "Opportunity", "Hello!")
			return &s3svc.GetObjectOutput{Body: bodyReader(email)}, nil
		},
	}
	resetTableClient()
	tableClient = &mockDynamoClient{
		putItemFn: func(_ context.Context, _ *dynamodb.PutItemInput, _ ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error) {
			return &dynamodb.PutItemOutput{}, nil
		},
	}
	t.Cleanup(func() {
		s3Client = nil
		resetTableClient()
	})

	event := events.S3Event{
		Records: []events.S3EventRecord{
			{S3: events.S3Entity{Bucket: events.S3Bucket{Name: "b"}, Object: events.S3Object{Key: "ok.eml"}}},
			{S3: events.S3Entity{Bucket: events.S3Bucket{Name: "b"}, Object: events.S3Object{Key: "bad.eml"}}},
			{S3: events.S3Entity{Bucket: events.S3Bucket{Name: "b"}, Object: events.S3Object{Key: "ok2.eml"}}},
		},
	}
	results, err := Handler(context.Background(), event)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}
	if results[0].Error != "" {
		t.Errorf("first record should succeed, got error: %s", results[0].Error)
	}
	if results[1].Error == "" {
		t.Error("second record should fail")
	}
	if results[2].Error != "" {
		t.Errorf("third record should succeed despite second failing, got: %s", results[2].Error)
	}
}

// --- DB tests ---

func TestWriteRecruiterEmailWithClient_Success(t *testing.T) {
	t.Setenv("RECRUITER_TABLE", "test-table")

	called := false
	client := &mockDynamoClient{
		putItemFn: func(_ context.Context, in *dynamodb.PutItemInput, _ ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error) {
			called = true
			if *in.TableName != "test-table" {
				t.Errorf("table name: got %q, want %q", *in.TableName, "test-table")
			}
			if in.ConditionExpression == nil {
				t.Error("ConditionExpression must be set")
			}
			return &dynamodb.PutItemOutput{}, nil
		},
	}

	r := &RecruiterEmail{
		DedupKey:    "abc123",
		S3Bucket:    "b",
		S3Key:       "k",
		From:        "r@acme.com",
		Subject:     "Opportunity",
		ProcessedAt: time.Now().UTC().Format(time.RFC3339),
	}
	result, err := writeRecruiterEmailWithClient(context.Background(), client, r)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Error("PutItem was not called")
	}
	if result.Duplicate {
		t.Error("expected Duplicate=false")
	}
}

func TestWriteRecruiterEmailWithClient_Duplicate(t *testing.T) {
	t.Setenv("RECRUITER_TABLE", "test-table")

	client := &mockDynamoClient{
		putItemFn: func(_ context.Context, _ *dynamodb.PutItemInput, _ ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error) {
			return nil, &types.ConditionalCheckFailedException{Message: strPtr("The conditional request failed")}
		},
	}

	r := &RecruiterEmail{DedupKey: "dup", ProcessedAt: time.Now().UTC().Format(time.RFC3339)}
	result, err := writeRecruiterEmailWithClient(context.Background(), client, r)
	if err != nil {
		t.Fatalf("duplicate must not return an error: %v", err)
	}
	if !result.Duplicate {
		t.Error("expected Duplicate=true for ConditionalCheckFailedException")
	}
}

func TestWriteRecruiterEmailWithClient_NoTableEnv(t *testing.T) {
	t.Setenv("RECRUITER_TABLE", "")

	client := &mockDynamoClient{
		putItemFn: func(_ context.Context, _ *dynamodb.PutItemInput, _ ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error) {
			t.Error("PutItem should not be called when table name is missing")
			return nil, nil
		},
	}
	r := &RecruiterEmail{DedupKey: "x"}
	_, err := writeRecruiterEmailWithClient(context.Background(), client, r)
	if err == nil {
		t.Error("expected error when RECRUITER_TABLE is unset")
	}
}

func TestWriteRecruiterEmailWithClient_OtherError(t *testing.T) {
	t.Setenv("RECRUITER_TABLE", "test-table")

	client := &mockDynamoClient{
		putItemFn: func(_ context.Context, _ *dynamodb.PutItemInput, _ ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error) {
			return nil, errors.New("network timeout")
		},
	}
	r := &RecruiterEmail{DedupKey: "y", ProcessedAt: time.Now().UTC().Format(time.RFC3339)}
	_, err := writeRecruiterEmailWithClient(context.Background(), client, r)
	if err == nil {
		t.Error("expected error for non-conditional DynamoDB failure")
	}
}

// --- Helpers ---

// resetTableClient resets the lazy-once state so tests can inject their own client.
func resetTableClient() {
	tableClient = nil
	tableInitErr = nil
	tableOnce = sync.Once{}
}

func strPtr(s string) *string { return &s }
