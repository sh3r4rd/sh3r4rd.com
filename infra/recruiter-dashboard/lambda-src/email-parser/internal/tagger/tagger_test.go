package tagger

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// mockS3Client implements S3Client for testing.
type mockS3Client struct {
	putObjectTaggingFn func(ctx context.Context, params *s3.PutObjectTaggingInput, optFns ...func(*s3.Options)) (*s3.PutObjectTaggingOutput, error)
	lastInput          *s3.PutObjectTaggingInput
}

func (m *mockS3Client) PutObjectTagging(ctx context.Context, params *s3.PutObjectTaggingInput, optFns ...func(*s3.Options)) (*s3.PutObjectTaggingOutput, error) {
	m.lastInput = params
	return m.putObjectTaggingFn(ctx, params, optFns...)
}

func TestTagObject_Success(t *testing.T) {
	mock := &mockS3Client{
		putObjectTaggingFn: func(ctx context.Context, params *s3.PutObjectTaggingInput, optFns ...func(*s3.Options)) (*s3.PutObjectTaggingOutput, error) {
			return &s3.PutObjectTaggingOutput{}, nil
		},
	}

	tgr := NewTagger(mock)
	tgr.TagObject(context.Background(), "test-bucket", "incoming/msg-001", TagResult{
		Status:     "success",
		Company:    "Google",
		Sender:     "jane@google.com",
		Confidence: 0.95,
	})

	if mock.lastInput == nil {
		t.Fatal("PutObjectTagging was not called")
	}

	if *mock.lastInput.Bucket != "test-bucket" {
		t.Errorf("expected bucket test-bucket, got %s", *mock.lastInput.Bucket)
	}

	tags := mock.lastInput.Tagging.TagSet
	tagMap := make(map[string]string)
	for _, tag := range tags {
		tagMap[*tag.Key] = *tag.Value
	}

	if tagMap["parse-status"] != "success" {
		t.Errorf("expected parse-status=success, got %s", tagMap["parse-status"])
	}
	if tagMap["company"] != "Google" {
		t.Errorf("expected company=Google, got %s", tagMap["company"])
	}
	if tagMap["sender"] != "jane@google.com" {
		t.Errorf("expected sender=jane@google.com, got %s", tagMap["sender"])
	}
	if _, ok := tagMap["parsed-at"]; !ok {
		t.Error("expected parsed-at tag to be present")
	}
	// No error-reason tag on success
	if _, ok := tagMap["error-reason"]; ok {
		t.Error("error-reason tag should not be present on success")
	}
}

func TestTagObject_WithErrorReason(t *testing.T) {
	mock := &mockS3Client{
		putObjectTaggingFn: func(ctx context.Context, params *s3.PutObjectTaggingInput, optFns ...func(*s3.Options)) (*s3.PutObjectTaggingOutput, error) {
			return &s3.PutObjectTaggingOutput{}, nil
		},
	}

	tgr := NewTagger(mock)
	tgr.TagObject(context.Background(), "test-bucket", "incoming/msg-002", TagResult{
		Status:      "failed",
		ErrorReason: "OpenAI timeout after 3 retries",
	})

	tags := mock.lastInput.Tagging.TagSet
	tagMap := make(map[string]string)
	for _, tag := range tags {
		tagMap[*tag.Key] = *tag.Value
	}

	if tagMap["parse-status"] != "failed" {
		t.Errorf("expected parse-status=failed, got %s", tagMap["parse-status"])
	}
	if tagMap["error-reason"] != "OpenAI timeout after 3 retries" {
		t.Errorf("expected error-reason, got %s", tagMap["error-reason"])
	}
}

func TestTagObject_ValueTruncation(t *testing.T) {
	mock := &mockS3Client{
		putObjectTaggingFn: func(ctx context.Context, params *s3.PutObjectTaggingInput, optFns ...func(*s3.Options)) (*s3.PutObjectTaggingOutput, error) {
			return &s3.PutObjectTaggingOutput{}, nil
		},
	}

	longCompany := strings.Repeat("A", 300) // Exceeds 256 char limit
	tgr := NewTagger(mock)
	tgr.TagObject(context.Background(), "test-bucket", "incoming/msg-003", TagResult{
		Status:  "success",
		Company: longCompany,
	})

	tags := mock.lastInput.Tagging.TagSet
	var companyTag *types.Tag
	for i := range tags {
		if *tags[i].Key == "company" {
			companyTag = &tags[i]
			break
		}
	}

	if companyTag == nil {
		t.Fatal("company tag not found")
	}
	if len([]rune(*companyTag.Value)) > 256 {
		t.Errorf("company tag value should be truncated to 256 runes, got %d", len([]rune(*companyTag.Value)))
	}
}

func TestTagObject_FailureDoesNotPanic(t *testing.T) {
	mock := &mockS3Client{
		putObjectTaggingFn: func(ctx context.Context, params *s3.PutObjectTaggingInput, optFns ...func(*s3.Options)) (*s3.PutObjectTaggingOutput, error) {
			return nil, fmt.Errorf("S3 service unavailable")
		},
	}

	tgr := NewTagger(mock)
	// Should log warning but not panic or return error
	tgr.TagObject(context.Background(), "test-bucket", "incoming/msg-004", TagResult{
		Status: "success",
	})
	// If we reach here without panic, the test passes
}
