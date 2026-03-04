package main

import (
	"context"
	"fmt"
	"io"
	"log"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// s3ClientIface is satisfied by the AWS S3 client and allows test injection.
type s3ClientIface interface {
	GetObject(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error)
}

// s3Client is the lazily-initialised S3 client reused across warm invocations.
var s3Client s3ClientIface

// RecordResult summarises the outcome of processing a single S3 record.
type RecordResult struct {
	Bucket    string `json:"bucket"`
	Key       string `json:"key"`
	Duplicate bool   `json:"duplicate,omitempty"`
	Error     string `json:"error,omitempty"`
}

// Handler is the Lambda entry point. It processes an S3 event, which may
// contain one or more records. Each record is processed independently so that
// a single failure does not block the rest of the batch.
func Handler(ctx context.Context, event events.S3Event) ([]RecordResult, error) {
	if len(event.Records) == 0 {
		log.Print("[WARN] received S3 event with no records")
		return nil, nil
	}

	// Initialise the S3 client once per cold start.
	if s3Client == nil {
		cfg, err := config.LoadDefaultConfig(ctx)
		if err != nil {
			return nil, fmt.Errorf("load AWS config: %w", err)
		}
		s3Client = s3.NewFromConfig(cfg)
	}

	results := make([]RecordResult, 0, len(event.Records))
	for _, record := range event.Records {
		result := processRecord(ctx, record)
		results = append(results, result)
	}
	return results, nil
}

// processRecord handles a single S3 event record: fetch → parse → extract → persist.
// All errors are captured in the returned RecordResult; they are never propagated
// so that sibling records in the same batch can still be processed.
func processRecord(ctx context.Context, record events.S3EventRecord) RecordResult {
	bucket := record.S3.Bucket.Name
	key := record.S3.Object.Key

	res := RecordResult{Bucket: bucket, Key: key}

	// Step 1 – Fetch raw email bytes from S3.
	// NOTE: raw email content is never logged to avoid exposing PII.
	raw, err := fetchS3Object(ctx, bucket, key)
	if err != nil {
		log.Printf("[ERROR] fetch S3 object bucket=%s key=%s error_type=%T", bucket, key, err)
		res.Error = "failed to fetch email from S3"
		return res
	}

	// Step 2 – Parse the raw email (MIME + forwarded detection).
	parsed, err := ParseRawEmail(raw)
	if err != nil {
		log.Printf("[ERROR] parse email bucket=%s key=%s error_type=%T", bucket, key, err)
		res.Error = "failed to parse email"
		return res
	}

	// Step 3 – Extract structured recruiter data.
	extracted := ExtractRecruiterData(parsed)

	// Step 4 – Assemble the canonical record.
	recruiterEmail := NewRecruiterEmail(bucket, key, parsed, extracted)

	// Step 5 – Persist to DynamoDB (idempotent; duplicates are skipped).
	writeResult, err := WriteRecruiterEmail(ctx, recruiterEmail)
	if err != nil {
		log.Printf("[ERROR] write to DynamoDB bucket=%s key=%s error_type=%T", bucket, key, err)
		res.Error = "failed to write to DynamoDB"
		return res
	}

	res.Duplicate = writeResult.Duplicate
	return res
}

// fetchS3Object retrieves the raw bytes for a single S3 object.
func fetchS3Object(ctx context.Context, bucket, key string) ([]byte, error) {
	out, err := s3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: &bucket,
		Key:    &key,
	})
	if err != nil {
		return nil, err
	}
	defer out.Body.Close()
	return io.ReadAll(out.Body)
}

func main() {
	lambda.Start(Handler)
}
