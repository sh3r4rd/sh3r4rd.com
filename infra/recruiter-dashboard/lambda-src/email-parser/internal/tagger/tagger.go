package tagger

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

const maxTagValueLen = 256

// S3Client defines the interface for S3 tagging operations (enables testing with mocks).
type S3Client interface {
	PutObjectTagging(ctx context.Context, params *s3.PutObjectTaggingInput, optFns ...func(*s3.Options)) (*s3.PutObjectTaggingOutput, error)
}

// Tagger applies parse status tags to S3 objects.
type Tagger struct {
	client S3Client
}

// NewTagger creates a new S3 object tagger.
func NewTagger(client S3Client) *Tagger {
	return &Tagger{client: client}
}

// TagResult holds the metadata to tag on an S3 object after parsing.
type TagResult struct {
	Status      string  // "success", "partial", or "failed"
	Company     string  // Extracted company name
	Sender      string  // Recruiter email address
	Confidence  float64 // Extraction confidence score
	ErrorReason string  // Error description (only on failure)
}

// TagObject applies parse result tags to the specified S3 object.
// Tagging failures are logged but do not return an error.
func (t *Tagger) TagObject(ctx context.Context, bucket, key string, result TagResult) {
	tags := []types.Tag{
		{Key: aws.String("parse-status"), Value: aws.String(truncate(result.Status))},
		{Key: aws.String("company"), Value: aws.String(truncate(result.Company))},
		{Key: aws.String("sender"), Value: aws.String(truncate(result.Sender))},
		{Key: aws.String("confidence"), Value: aws.String(truncate(fmt.Sprintf("%.2f", result.Confidence)))},
		{Key: aws.String("parsed-at"), Value: aws.String(truncate(time.Now().UTC().Format(time.RFC3339)))},
	}

	if result.ErrorReason != "" {
		tags = append(tags, types.Tag{
			Key:   aws.String("error-reason"),
			Value: aws.String(truncate(result.ErrorReason)),
		})
	}

	_, err := t.client.PutObjectTagging(ctx, &s3.PutObjectTaggingInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Tagging: &types.Tagging{
			TagSet: tags,
		},
	})
	if err != nil {
		log.Printf("WARNING: failed to tag S3 object s3://%s/%s: %v", bucket, key, err)
	}
}

// truncate limits a string to the S3 tag value maximum length (256 characters).
func truncate(s string) string {
	runes := []rune(s)
	if len(runes) <= maxTagValueLen {
		return s
	}
	return string(runes[:maxTagValueLen])
}
