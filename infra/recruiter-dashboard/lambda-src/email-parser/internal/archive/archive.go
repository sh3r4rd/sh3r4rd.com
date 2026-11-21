// Package archive writes the durable copies of each received email:
//
//   - archive/raw/{messageId}: the original .eml, copied from incoming/ without tags.
//   - archive/clean/{messageId}.json: the parsed, cleaned record (CleanRecord).
//
// Every write names the archive KMS key explicitly (SSE-KMS) and turns S3 Bucket Keys
// off, because the key policy and IAM grant match on the object-ARN encryption context.
// The bucket's default
// encryption stays AES256 so SES delivery to incoming/ never depends on the key, and a
// bucket policy denies archive/ writes that don't name the key (D-06).
package archive

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"

	perrors "github.com/sh3r4rd/sh3r4rd.com/infra/recruiter-dashboard/lambda-src/email-parser/internal/errors"
	"github.com/sh3r4rd/sh3r4rd.com/infra/recruiter-dashboard/lambda-src/email-parser/internal/models"
)

const (
	// RawPrefix holds the original emails, one object per SES message id.
	RawPrefix = "archive/raw/"
	// CleanPrefix holds one CleanRecord JSON document per parsed email.
	CleanPrefix = "archive/clean/"

	// SchemaVersion is the version of the CleanRecord JSON layout. Bump it when a field
	// is added, removed or changes meaning.
	SchemaVersion = 1
	// ParserVersion identifies the parsing logic that produced a CleanRecord. Bump it when
	// a change to internal/parser would change the cleaned body or headers of an existing
	// email, so a backfill run with --force can regenerate stale records.
	ParserVersion = "1"
)

// Values for CleanRecord.ExtractionSource.
const (
	// ExtractionSourceLive means the extraction was produced by the Lambda that wrote the record.
	ExtractionSourceLive = "live"
	// ExtractionSourceDynamoDB means the backfill copied the extraction from the DynamoDB row.
	ExtractionSourceDynamoDB = "dynamodb"
	// ExtractionSourceNone means no extraction was available (Extraction is null).
	ExtractionSourceNone = "none"
)

// messageIDPattern matches SES message ids (lowercase letters, digits and hyphens in
// practice). It deliberately excludes "/" so an id can never address another prefix.
var messageIDPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]{0,255}$`)

// S3Client is the subset of the S3 API the archiver uses.
type S3Client interface {
	CopyObject(ctx context.Context, params *s3.CopyObjectInput, optFns ...func(*s3.Options)) (*s3.CopyObjectOutput, error)
	PutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error)
}

// Archiver writes raw and clean archive objects, encrypted with the archive KMS key.
type Archiver struct {
	client    S3Client
	kmsKeyARN string
}

// NewArchiver creates an Archiver. kmsKeyARN must be the full key ARN: the bucket policy
// compares the request's key id against the ARN, so an alias or bare key id is denied.
func NewArchiver(client S3Client, kmsKeyARN string) *Archiver {
	return &Archiver{client: client, kmsKeyARN: kmsKeyARN}
}

// ValidateMessageID returns an error if id can't be used as an archive object name.
func ValidateMessageID(id string) error {
	if !messageIDPattern.MatchString(id) {
		return fmt.Errorf("invalid message id %q: want letters, digits, '.', '_' or '-' (max 256)", id)
	}
	return nil
}

// RawKey returns the archive/raw/ key for a message id.
func RawKey(messageID string) string {
	return RawPrefix + messageID
}

// CleanKey returns the archive/clean/ key for a message id.
func CleanKey(messageID string) string {
	return CleanPrefix + messageID + ".json"
}

// CopyRaw copies bucket/sourceKey to archive/raw/{messageID} with no tags, keeping the
// source's metadata and content type, encrypted with the archive key.
//
// TaggingDirective=REPLACE with no tags means the copy never inherits the parse-status
// tags the Lambda puts on incoming/ objects, even on a retry after tagging.
func (a *Archiver) CopyRaw(ctx context.Context, bucket, sourceKey, messageID string) error {
	if err := ValidateMessageID(messageID); err != nil {
		return &perrors.StorageError{Op: "archive_raw_copy", Err: err}
	}

	_, err := a.client.CopyObject(ctx, &s3.CopyObjectInput{
		Bucket:               aws.String(bucket),
		Key:                  aws.String(RawKey(messageID)),
		CopySource:           aws.String(copySource(bucket, sourceKey)),
		MetadataDirective:    types.MetadataDirectiveCopy,
		TaggingDirective:     types.TaggingDirectiveReplace,
		ServerSideEncryption: types.ServerSideEncryptionAwsKms,
		SSEKMSKeyId:          aws.String(a.kmsKeyARN),
		BucketKeyEnabled:     aws.Bool(false),
	})
	if err != nil {
		return &perrors.StorageError{Op: "archive_raw_copy", Err: err}
	}
	return nil
}

// WriteClean writes rec to archive/clean/{rec.MessageID}.json, encrypted with the archive key.
func (a *Archiver) WriteClean(ctx context.Context, bucket string, rec CleanRecord) error {
	if err := ValidateMessageID(rec.MessageID); err != nil {
		return &perrors.StorageError{Op: "archive_clean_write", Err: err}
	}

	body, err := rec.Marshal()
	if err != nil {
		return &perrors.StorageError{Op: "archive_clean_write", Err: err}
	}

	_, err = a.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:               aws.String(bucket),
		Key:                  aws.String(CleanKey(rec.MessageID)),
		Body:                 bytes.NewReader(body),
		ContentType:          aws.String("application/json"),
		ServerSideEncryption: types.ServerSideEncryptionAwsKms,
		SSEKMSKeyId:          aws.String(a.kmsKeyARN),
		BucketKeyEnabled:     aws.Bool(false),
	})
	if err != nil {
		return &perrors.StorageError{Op: "archive_clean_write", Err: err}
	}
	return nil
}

// copySource builds the URL-encoded "bucket/key" value CopyObject expects, escaping each
// key segment but keeping the "/" separators. url.PathEscape leaves "+" alone, but S3 can
// decode a literal "+" in x-amz-copy-source as a space, so it is escaped explicitly.
func copySource(bucket, key string) string {
	segments := strings.Split(key, "/")
	for i, s := range segments {
		segments[i] = strings.ReplaceAll(url.PathEscape(s), "+", "%2B")
	}
	return bucket + "/" + strings.Join(segments, "/")
}

// Headers holds the parsed message headers kept in a CleanRecord. For a forwarded email
// these are the original message's headers (the parser replaces them).
type Headers struct {
	From    string `json:"from"`
	To      string `json:"to"`
	Subject string `json:"subject"`
	// Date is RFC 3339 in UTC, or "" when the email had no parseable date.
	Date string `json:"date"`
}

// CleanRecord is the archive/clean/{messageId}.json document. It has no wall-clock
// fields, so the same raw email and extraction always produce identical bytes; S3's
// LastModified records when it was written.
type CleanRecord struct {
	SchemaVersion int     `json:"schema_version"`
	ParserVersion string  `json:"parser_version"`
	MessageID     string  `json:"message_id"`
	RawKey        string  `json:"raw_key"`
	Headers       Headers `json:"headers"`
	IsForwarded   bool    `json:"is_forwarded"`
	// Body is the cleaned body the extractor sees: HTML stripped, and for a forwarded
	// email the forwarded content only.
	Body string `json:"body"`
	// Extraction is the sanitized extraction result (the same values written to
	// DynamoDB), or null when ExtractionSource is "none".
	Extraction       *models.ExtractionResult `json:"extraction"`
	ExtractionSource string                   `json:"extraction_source"`
}

// NewCleanRecord builds a CleanRecord from a parsed email and its extraction.
// extraction may be nil, in which case source should be ExtractionSourceNone.
func NewCleanRecord(messageID string, parsed *models.ParsedEmail, extraction *models.ExtractionResult, source string) CleanRecord {
	date := ""
	if !parsed.Date.IsZero() {
		date = parsed.Date.UTC().Format(time.RFC3339)
	}

	var ext *models.ExtractionResult
	if extraction != nil {
		copied := *extraction
		ext = &copied
	}

	return CleanRecord{
		SchemaVersion: SchemaVersion,
		ParserVersion: ParserVersion,
		MessageID:     messageID,
		RawKey:        RawKey(messageID),
		Headers: Headers{
			From:    parsed.From,
			To:      parsed.To,
			Subject: parsed.Subject,
			Date:    date,
		},
		IsForwarded:      parsed.IsForwarded,
		Body:             parsed.Body,
		Extraction:       ext,
		ExtractionSource: source,
	}
}

// Marshal encodes the record as JSON without HTML escaping, so bodies stay readable.
// The output is deterministic for a given record.
func (r CleanRecord) Marshal() ([]byte, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(r); err != nil {
		return nil, fmt.Errorf("encode clean record: %w", err)
	}
	return buf.Bytes(), nil
}
