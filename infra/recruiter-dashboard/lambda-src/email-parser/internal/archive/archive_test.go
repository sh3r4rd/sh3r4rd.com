package archive

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"

	perrors "github.com/sh3r4rd/sh3r4rd.com/infra/recruiter-dashboard/lambda-src/email-parser/internal/errors"
	"github.com/sh3r4rd/sh3r4rd.com/infra/recruiter-dashboard/lambda-src/email-parser/internal/models"
)

const testKeyARN = "arn:aws:kms:us-east-1:111122223333:key/11111111-2222-3333-4444-555555555555"

// mockS3Client records the last CopyObject and PutObject inputs.
type mockS3Client struct {
	copyErr   error
	putErr    error
	copyCalls int
	putCalls  int
	lastCopy  *s3.CopyObjectInput
	lastPut   *s3.PutObjectInput
	putBody   []byte
}

func (m *mockS3Client) CopyObject(ctx context.Context, params *s3.CopyObjectInput, optFns ...func(*s3.Options)) (*s3.CopyObjectOutput, error) {
	m.copyCalls++
	m.lastCopy = params
	if m.copyErr != nil {
		return nil, m.copyErr
	}
	return &s3.CopyObjectOutput{}, nil
}

func (m *mockS3Client) PutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
	m.putCalls++
	m.lastPut = params
	if params.Body != nil {
		body, err := io.ReadAll(params.Body)
		if err != nil {
			return nil, err
		}
		m.putBody = body
	}
	if m.putErr != nil {
		return nil, m.putErr
	}
	return &s3.PutObjectOutput{}, nil
}

func TestValidateMessageID(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		wantErr bool
	}{
		{"ses message id", "0100018c9a2b3c4d-5e6f7a8b-1234-5678-9abc-def012345678-000000", false},
		{"lowercase alphanumeric", "abc123def456ghi789jk", false},
		{"dots and underscores", "a.b_c-d", false},
		{"empty", "", true},
		{"contains slash", "incoming/abc", true},
		{"parent traversal", "../abc", true},
		{"leading dot", ".abc", true},
		{"space", "abc def", true},
		{"too long", strings.Repeat("a", 257), true},
		{"max length", strings.Repeat("a", 256), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateMessageID(tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateMessageID(%q) error = %v, wantErr %v", tt.id, err, tt.wantErr)
			}
		})
	}
}

func TestKeys(t *testing.T) {
	if got := RawKey("abc"); got != "archive/raw/abc" {
		t.Errorf("RawKey = %q, want archive/raw/abc", got)
	}
	if got := CleanKey("abc"); got != "archive/clean/abc.json" {
		t.Errorf("CleanKey = %q, want archive/clean/abc.json", got)
	}
}

func TestCopySource(t *testing.T) {
	tests := []struct {
		name   string
		bucket string
		key    string
		want   string
	}{
		{"plain key", "my-bucket", "incoming/abc123", "my-bucket/incoming/abc123"},
		{"space and plus are escaped", "my-bucket", "incoming/a b+c", "my-bucket/incoming/a%20b%2Bc"},
		{"slashes kept as separators", "my-bucket", "a/b/c", "my-bucket/a/b/c"},
		{"percent escaped", "my-bucket", "incoming/50%", "my-bucket/incoming/50%25"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := copySource(tt.bucket, tt.key); got != tt.want {
				t.Errorf("copySource(%q, %q) = %q, want %q", tt.bucket, tt.key, got, tt.want)
			}
		})
	}
}

func TestCopyRaw(t *testing.T) {
	tests := []struct {
		name      string
		messageID string
		copyErr   error
		wantErr   bool
		wantCalls int
	}{
		{"success", "abc123", nil, false, 1},
		{"s3 error is returned", "abc123", errors.New("AccessDenied"), true, 1},
		{"invalid id never calls s3", "../abc", nil, true, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockS3Client{copyErr: tt.copyErr}
			a := NewArchiver(mock, testKeyARN)

			err := a.CopyRaw(context.Background(), "test-bucket", "incoming/"+tt.messageID, tt.messageID)
			if (err != nil) != tt.wantErr {
				t.Fatalf("CopyRaw() error = %v, wantErr %v", err, tt.wantErr)
			}
			if mock.copyCalls != tt.wantCalls {
				t.Fatalf("CopyObject calls = %d, want %d", mock.copyCalls, tt.wantCalls)
			}
			if err != nil {
				var storageErr *perrors.StorageError
				if !errors.As(err, &storageErr) || storageErr.Op != "archive_raw_copy" {
					t.Errorf("error = %v, want StorageError with Op archive_raw_copy", err)
				}
				if tt.copyErr != nil && !errors.Is(err, tt.copyErr) {
					t.Errorf("error %v does not wrap %v", err, tt.copyErr)
				}
			}
			if tt.wantCalls == 0 {
				return
			}

			in := mock.lastCopy
			if aws.ToString(in.Bucket) != "test-bucket" {
				t.Errorf("Bucket = %q, want test-bucket", aws.ToString(in.Bucket))
			}
			if aws.ToString(in.Key) != "archive/raw/"+tt.messageID {
				t.Errorf("Key = %q, want archive/raw/%s", aws.ToString(in.Key), tt.messageID)
			}
			if aws.ToString(in.CopySource) != "test-bucket/incoming/"+tt.messageID {
				t.Errorf("CopySource = %q", aws.ToString(in.CopySource))
			}
			if in.TaggingDirective != types.TaggingDirectiveReplace {
				t.Errorf("TaggingDirective = %q, want REPLACE (the copy must carry no tags)", in.TaggingDirective)
			}
			if in.Tagging != nil {
				t.Errorf("Tagging = %q, want nil (no tags on the copy)", aws.ToString(in.Tagging))
			}
			if in.MetadataDirective != types.MetadataDirectiveCopy {
				t.Errorf("MetadataDirective = %q, want COPY", in.MetadataDirective)
			}
			if in.ServerSideEncryption != types.ServerSideEncryptionAwsKms {
				t.Errorf("ServerSideEncryption = %q, want aws:kms", in.ServerSideEncryption)
			}
			if aws.ToString(in.SSEKMSKeyId) != testKeyARN {
				t.Errorf("SSEKMSKeyId = %q, want the key ARN", aws.ToString(in.SSEKMSKeyId))
			}
			if in.BucketKeyEnabled != nil {
				t.Errorf("BucketKeyEnabled = %v, want unset (the IAM condition relies on the object-ARN encryption context)", *in.BucketKeyEnabled)
			}
		})
	}
}

func testParsedEmail() *models.ParsedEmail {
	return &models.ParsedEmail{
		From:        "Jane Recruiter <jane@example.com>",
		To:          "me@example.org",
		Subject:     "Senior Go Engineer <remote>",
		Date:        time.Date(2026, 3, 3, 10, 0, 0, 0, time.FixedZone("EST", -5*3600)),
		Body:        "Hello & welcome <b>",
		HTMLBody:    "<p>ignored</p>",
		IsForwarded: true,
	}
}

func TestWriteClean(t *testing.T) {
	extraction := &models.ExtractionResult{
		RecruiterFirstName: "Jane",
		RecruiterLastName:  "Recruiter",
		RecruiterEmail:     "jane@example.com",
		Company:            "ExampleCorp",
		JobTitle:           "Senior Go Engineer",
		Phone:              "Unknown",
		Confidence:         0.9,
	}

	tests := []struct {
		name      string
		messageID string
		putErr    error
		wantErr   bool
		wantCalls int
	}{
		{"success", "abc123", nil, false, 1},
		{"s3 error is returned", "abc123", errors.New("AccessDenied"), true, 1},
		{"invalid id never calls s3", "a/b", nil, true, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockS3Client{putErr: tt.putErr}
			a := NewArchiver(mock, testKeyARN)
			rec := NewCleanRecord(tt.messageID, testParsedEmail(), extraction, ExtractionSourceLive)

			err := a.WriteClean(context.Background(), "test-bucket", rec)
			if (err != nil) != tt.wantErr {
				t.Fatalf("WriteClean() error = %v, wantErr %v", err, tt.wantErr)
			}
			if mock.putCalls != tt.wantCalls {
				t.Fatalf("PutObject calls = %d, want %d", mock.putCalls, tt.wantCalls)
			}
			if err != nil {
				var storageErr *perrors.StorageError
				if !errors.As(err, &storageErr) || storageErr.Op != "archive_clean_write" {
					t.Errorf("error = %v, want StorageError with Op archive_clean_write", err)
				}
			}
			if tt.wantCalls == 0 {
				return
			}

			in := mock.lastPut
			if aws.ToString(in.Key) != "archive/clean/"+tt.messageID+".json" {
				t.Errorf("Key = %q", aws.ToString(in.Key))
			}
			if aws.ToString(in.ContentType) != "application/json" {
				t.Errorf("ContentType = %q, want application/json", aws.ToString(in.ContentType))
			}
			if in.ServerSideEncryption != types.ServerSideEncryptionAwsKms {
				t.Errorf("ServerSideEncryption = %q, want aws:kms", in.ServerSideEncryption)
			}
			if aws.ToString(in.SSEKMSKeyId) != testKeyARN {
				t.Errorf("SSEKMSKeyId = %q, want the key ARN", aws.ToString(in.SSEKMSKeyId))
			}

			var got CleanRecord
			if err := json.Unmarshal(mock.putBody, &got); err != nil {
				t.Fatalf("body is not valid JSON: %v", err)
			}
			if got.MessageID != tt.messageID || got.Extraction == nil || got.Extraction.Company != "ExampleCorp" {
				t.Errorf("unexpected record round-trip: %+v", got)
			}
		})
	}
}

func TestNewCleanRecord(t *testing.T) {
	t.Run("all fields", func(t *testing.T) {
		extraction := &models.ExtractionResult{Company: "ExampleCorp", Confidence: 0.8}
		rec := NewCleanRecord("abc123", testParsedEmail(), extraction, ExtractionSourceLive)

		if rec.SchemaVersion != SchemaVersion || rec.ParserVersion != ParserVersion {
			t.Errorf("versions = %d/%q", rec.SchemaVersion, rec.ParserVersion)
		}
		if rec.RawKey != "archive/raw/abc123" {
			t.Errorf("RawKey = %q", rec.RawKey)
		}
		if rec.Headers.Date != "2026-03-03T15:00:00Z" {
			t.Errorf("Headers.Date = %q, want UTC RFC 3339", rec.Headers.Date)
		}
		if rec.Headers.From != "Jane Recruiter <jane@example.com>" || rec.Headers.Subject != "Senior Go Engineer <remote>" {
			t.Errorf("Headers = %+v", rec.Headers)
		}
		if !rec.IsForwarded || rec.Body != "Hello & welcome <b>" {
			t.Errorf("IsForwarded/Body = %v/%q", rec.IsForwarded, rec.Body)
		}
		if rec.ExtractionSource != ExtractionSourceLive || rec.Extraction.Company != "ExampleCorp" {
			t.Errorf("extraction = %+v (%q)", rec.Extraction, rec.ExtractionSource)
		}

		// The record holds its own copy of the extraction.
		extraction.Company = "Changed"
		if rec.Extraction.Company != "ExampleCorp" {
			t.Error("record aliases the caller's extraction")
		}
	})

	t.Run("zero date and nil extraction", func(t *testing.T) {
		parsed := testParsedEmail()
		parsed.Date = time.Time{}
		rec := NewCleanRecord("abc123", parsed, nil, ExtractionSourceNone)

		if rec.Headers.Date != "" {
			t.Errorf("Headers.Date = %q, want empty", rec.Headers.Date)
		}
		body, err := rec.Marshal()
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Contains(body, []byte(`"extraction":null`)) {
			t.Errorf("want extraction:null in %s", body)
		}
		if !bytes.Contains(body, []byte(`"extraction_source":"none"`)) {
			t.Errorf("want extraction_source none in %s", body)
		}
	})
}

func TestCleanRecordMarshal(t *testing.T) {
	rec := NewCleanRecord("abc123", testParsedEmail(), &models.ExtractionResult{Company: "A&B"}, ExtractionSourceLive)

	first, err := rec.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	second, err := rec.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(first, second) {
		t.Error("Marshal is not deterministic")
	}
	if !bytes.Contains(first, []byte(`<remote>`)) || !bytes.Contains(first, []byte(`A&B`)) {
		t.Errorf("HTML characters should not be escaped: %s", first)
	}
	if bytes.Contains(first, []byte(`HTMLBody`)) || bytes.Contains(first, []byte(`ignored`)) {
		t.Errorf("the raw HTML body must not be stored: %s", first)
	}
	for _, field := range []string{`"schema_version":1`, `"parser_version":"1"`, `"message_id":"abc123"`, `"raw_key":"archive/raw/abc123"`, `"is_forwarded":true`} {
		if !bytes.Contains(first, []byte(field)) {
			t.Errorf("missing %s in %s", field, first)
		}
	}
}
