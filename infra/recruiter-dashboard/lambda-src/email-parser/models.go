package main

import (
	"crypto/sha256"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// ParsedEmail holds the result of MIME parsing an incoming (possibly forwarded) email.
type ParsedEmail struct {
	// Headers from the original/forwarded message.
	From    string
	To      string
	Subject string
	Date    string

	// Decoded plain-text body of the original message.
	Body string

	// True when the handler determined the email was a forwarded recruiter email.
	IsForwarded bool
}

// ExtractionResult holds structured recruiter data extracted from a ParsedEmail.
type ExtractionResult struct {
	Company  *string
	JobTitle *string
	Name     *string
	Phone    *string
}

// RecruiterEmail is the canonical record stored in DynamoDB.
type RecruiterEmail struct {
	// DedupKey is a SHA-256 hash of (From + Subject + Date) used for conditional writes.
	DedupKey string `dynamodbav:"dedup_key"`

	// S3 coordinates of the raw email.
	S3Bucket string `dynamodbav:"s3_bucket"`
	S3Key    string `dynamodbav:"s3_key"`

	// Headers sourced from ParsedEmail.
	From    string `dynamodbav:"from"`
	To      string `dynamodbav:"to"`
	Subject string `dynamodbav:"subject"`
	Date    string `dynamodbav:"date"`

	// Extracted fields (may be empty strings when extraction failed).
	Company  string `dynamodbav:"company"`
	JobTitle string `dynamodbav:"job_title"`
	Name     string `dynamodbav:"name"`
	Phone    string `dynamodbav:"phone"`

	// ISO-8601 timestamp of when the Lambda processed this record.
	ProcessedAt string `dynamodbav:"processed_at"`
}

// NewRecruiterEmail assembles a RecruiterEmail from parsed and extracted data.
// Missing extracted fields fall back to empty strings.
func NewRecruiterEmail(bucket, key string, parsed *ParsedEmail, extracted *ExtractionResult) *RecruiterEmail {
	r := &RecruiterEmail{
		DedupKey:    buildDedupKey(parsed.From, parsed.Subject, parsed.Date),
		S3Bucket:    bucket,
		S3Key:       key,
		From:        parsed.From,
		To:          parsed.To,
		Subject:     parsed.Subject,
		Date:        parsed.Date,
		ProcessedAt: time.Now().UTC().Format(time.RFC3339),
	}

	if extracted.Company != nil {
		r.Company = *extracted.Company
	}
	if extracted.JobTitle != nil {
		r.JobTitle = *extracted.JobTitle
	}
	if extracted.Name != nil {
		r.Name = *extracted.Name
	}
	if extracted.Phone != nil {
		r.Phone = *extracted.Phone
	}

	return r
}

// ToDynamoDBItem serialises the RecruiterEmail into a DynamoDB attribute map.
func (r *RecruiterEmail) ToDynamoDBItem() (map[string]types.AttributeValue, error) {
	return attributevalue.MarshalMap(r)
}

// buildDedupKey returns a hex-encoded SHA-256 hash of the concatenated fields.
func buildDedupKey(from, subject, date string) string {
	h := sha256.New()
	h.Write([]byte(from + "|" + subject + "|" + date))
	return fmt.Sprintf("%x", h.Sum(nil))
}
