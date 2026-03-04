package handler

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"github.com/sh3r4rd/sh3r4rd.com/infra/recruiter-dashboard/lambda-src/email-parser/internal/db"
	"github.com/sh3r4rd/sh3r4rd.com/infra/recruiter-dashboard/lambda-src/email-parser/internal/extractor"
	"github.com/sh3r4rd/sh3r4rd.com/infra/recruiter-dashboard/lambda-src/email-parser/internal/models"
	"github.com/sh3r4rd/sh3r4rd.com/infra/recruiter-dashboard/lambda-src/email-parser/internal/parser"
	"github.com/sh3r4rd/sh3r4rd.com/infra/recruiter-dashboard/lambda-src/email-parser/internal/sanitizer"
	"github.com/sh3r4rd/sh3r4rd.com/infra/recruiter-dashboard/lambda-src/email-parser/internal/ssm"
	"github.com/sh3r4rd/sh3r4rd.com/infra/recruiter-dashboard/lambda-src/email-parser/internal/tagger"
)

const maxEmailBytes = 10 * 1024 * 1024 // 10 MB

// S3Client defines the interface for S3 read operations.
type S3Client interface {
	GetObject(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error)
}

// Summary holds processing counts returned by the handler.
type Summary struct {
	Total     int `json:"total"`
	Success   int `json:"success"`
	Failed    int `json:"failed"`
	Duplicate int `json:"duplicate"`
	Rejected  int `json:"rejected"`
}

// Handler orchestrates the email parsing pipeline.
type Handler struct {
	s3Client   S3Client
	store      *db.Store
	tagger     *tagger.Tagger
	ssmFetcher *ssm.ParameterFetcher
	extractor  *extractor.Extractor
}

// NewHandler creates a new email parsing handler.
func NewHandler(s3Client S3Client, store *db.Store, t *tagger.Tagger, ssmFetcher *ssm.ParameterFetcher) *Handler {
	return &Handler{
		s3Client:   s3Client,
		store:      store,
		tagger:     t,
		ssmFetcher: ssmFetcher,
	}
}

// HandleSESEvent processes an SES SimpleEmailEvent.
func (h *Handler) HandleSESEvent(ctx context.Context, event events.SimpleEmailEvent) (Summary, error) {
	summary := Summary{Total: len(event.Records)}

	if len(event.Records) == 0 {
		return summary, fmt.Errorf("no records in SES event")
	}

	bucket := os.Getenv("EMAIL_BUCKET")
	keyPrefix := os.Getenv("S3_KEY_PREFIX")
	ssmKeyName := os.Getenv("SSM_OPENAI_KEY_NAME")

	// Lazily initialize the extractor on first invocation (or cold start)
	if h.extractor == nil {
		apiKey, err := h.ssmFetcher.GetSecureParameter(ctx, ssmKeyName)
		if err != nil {
			log.Printf("ERROR: failed to fetch OpenAI API key: %v", err)
			summary.Failed = summary.Total
			return summary, fmt.Errorf("failed to fetch OpenAI API key: %w", err)
		}
		client := extractor.NewOpenAIClient(apiKey)
		h.extractor = extractor.NewExtractor(client)
	}

	for _, record := range event.Records {
		h.processRecord(ctx, record, bucket, keyPrefix, &summary)
	}

	log.Printf("Processing complete: total=%d success=%d failed=%d duplicate=%d rejected=%d",
		summary.Total, summary.Success, summary.Failed, summary.Duplicate, summary.Rejected)

	return summary, nil
}

// processRecord handles a single SES event record.
func (h *Handler) processRecord(ctx context.Context, record events.SimpleEmailRecord, bucket, keyPrefix string, summary *Summary) {
	messageID := record.SES.Mail.MessageID
	if messageID == "" {
		log.Printf("ERROR: SES record has empty MessageID")
		summary.Failed++
		return
	}

	// Validate SES verdicts
	if !h.validateVerdicts(record, messageID) {
		summary.Rejected++
		return
	}

	// Derive S3 key
	s3Key := keyPrefix + "/" + messageID

	// Fetch raw email from S3
	rawEmail, err := h.fetchEmail(ctx, bucket, s3Key)
	if err != nil {
		log.Printf("ERROR: failed to fetch email from s3://%s/%s: %v", bucket, s3Key, err)
		h.tagFailure(ctx, bucket, s3Key, "S3 fetch failed: "+err.Error())
		summary.Failed++
		return
	}

	// Parse raw email
	parsed, err := parser.ParseRawEmail(rawEmail)
	if err != nil {
		log.Printf("ERROR: failed to parse email %s: %v", messageID, err)
		h.tagFailure(ctx, bucket, s3Key, "Parse failed: "+err.Error())
		summary.Failed++
		return
	}

	// Extract recruiter data via OpenAI
	extraction := h.extractor.Extract(ctx, parsed.Body)

	// Apply sanitization to extracted fields
	if extraction.Phone != "" && extraction.Phone != "Unknown" {
		phones := sanitizer.FindPhoneNumbers(extraction.Phone)
		if len(phones) > 0 {
			extraction.Phone = sanitizer.NormalizePhone(phones[0])
		}
	}
	if extraction.RecruiterFirstName != "Unknown" {
		extraction.RecruiterFirstName = sanitizer.CleanName(extraction.RecruiterFirstName)
	}
	if extraction.RecruiterLastName != "Unknown" {
		extraction.RecruiterLastName = sanitizer.CleanName(extraction.RecruiterLastName)
	}

	// Determine parse status
	status := "success"
	if extraction.IsEmpty() {
		status = "partial"
	}

	// Build the RecruiterEmail model
	now := time.Now().UTC()
	email := &models.RecruiterEmail{
		ID:         messageID,
		ReceivedAt: parsed.Date,
		FirstName:  extraction.RecruiterFirstName,
		LastName:   extraction.RecruiterLastName,
		Email:      extraction.RecruiterEmail,
		Company:    extraction.Company,
		JobTitle:   extraction.JobTitle,
		Phone:      extraction.Phone,
		Subject:    parsed.Subject,
		Confidence: extraction.Confidence,
		S3Bucket:   bucket,
		S3Key:      s3Key,
		DedupKey:   models.GenerateDedupKey(extraction.RecruiterEmail, extraction.JobTitle),
		DateYear:   now.Format("2006"),
		DateDay:    now.Format("2006-01-02"),
	}

	// Tag S3 object with parse results
	h.tagger.TagObject(ctx, bucket, s3Key, tagger.TagResult{
		Status:     status,
		Company:    extraction.Company,
		Sender:     extraction.RecruiterEmail,
		Confidence: extraction.Confidence,
	})

	// Persist to DynamoDB
	result, err := h.store.WriteRecruiterEmail(ctx, email)
	if err != nil {
		log.Printf("ERROR: failed to write email %s to DynamoDB: %v", messageID, err)
		summary.Failed++
		return
	}

	if result.Duplicate {
		summary.Duplicate++
	} else {
		summary.Success++
	}
}

// validateVerdicts checks SES receipt verdicts and rejects emails with FAIL status.
func (h *Handler) validateVerdicts(record events.SimpleEmailRecord, messageID string) bool {
	verdicts := map[string]string{
		"SPF":   record.SES.Receipt.SPFVerdict.Status,
		"DKIM":  record.SES.Receipt.DKIMVerdict.Status,
		"Spam":  record.SES.Receipt.SpamVerdict.Status,
		"Virus": record.SES.Receipt.VirusVerdict.Status,
	}

	for name, status := range verdicts {
		if strings.EqualFold(status, "FAIL") {
			log.Printf("Rejecting email %s: %s verdict FAIL", messageID, name)
			return false
		}
	}

	return true
}

// fetchEmail retrieves raw email bytes from S3.
func (h *Handler) fetchEmail(ctx context.Context, bucket, key string) ([]byte, error) {
	output, err := h.s3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, err
	}
	defer output.Body.Close()

	return io.ReadAll(io.LimitReader(output.Body, maxEmailBytes))
}

// tagFailure tags an S3 object with a failed parse status.
func (h *Handler) tagFailure(ctx context.Context, bucket, key, reason string) {
	h.tagger.TagObject(ctx, bucket, key, tagger.TagResult{
		Status:      "failed",
		ErrorReason: reason,
	})
}
