package handler

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"github.com/sh3r4rd/sh3r4rd.com/infra/recruiter-dashboard/lambda-src/email-parser/internal/archive"
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
	archiver   *archive.Archiver
}

// NewHandler creates a new email parsing handler.
func NewHandler(s3Client S3Client, store *db.Store, t *tagger.Tagger, ssmFetcher *ssm.ParameterFetcher, archiver *archive.Archiver) *Handler {
	return &Handler{
		s3Client:   s3Client,
		store:      store,
		tagger:     t,
		ssmFetcher: ssmFetcher,
		archiver:   archiver,
	}
}

// verdictOutcome is the result of checking a record's SES receipt verdicts.
type verdictOutcome int

const (
	// verdictPass: no verdict is FAIL. The email is archived and parsed.
	verdictPass verdictOutcome = iota
	// verdictRejected: SPF, DKIM or spam is FAIL (virus is not). The raw email is
	// archived but never parsed.
	verdictRejected
	// verdictVirus: the virus verdict is FAIL. The email is neither archived nor parsed;
	// it expires from incoming/ with the lifecycle rule.
	verdictVirus
)

// HandleSESEvent processes an SES SimpleEmailEvent.
//
// It returns an error, failing the invocation so Lambda retries it asynchronously, only
// when the event is unusable or a raw email could not be copied to archive/raw/: that
// copy is the one artifact that can't be regenerated. Every other per-record failure is
// counted in the Summary and logged, as before.
func (h *Handler) HandleSESEvent(ctx context.Context, event events.SimpleEmailEvent) (Summary, error) {
	summary := Summary{Total: len(event.Records)}

	if len(event.Records) == 0 {
		return summary, fmt.Errorf("no records in SES event")
	}

	bucket := os.Getenv("EMAIL_BUCKET")
	keyPrefix := os.Getenv("S3_KEY_PREFIX")
	ssmKeyName := os.Getenv("SSM_OPENAI_KEY_NAME")

	// Validate required environment variables
	if bucket == "" {
		err := fmt.Errorf("required environment variable EMAIL_BUCKET is not set or empty")
		log.Printf("ERROR: %v", err)
		summary.Failed = summary.Total
		return summary, err
	}
	if keyPrefix == "" {
		err := fmt.Errorf("required environment variable S3_KEY_PREFIX is not set or empty")
		log.Printf("ERROR: %v", err)
		summary.Failed = summary.Total
		return summary, err
	}
	if ssmKeyName == "" {
		err := fmt.Errorf("required environment variable SSM_OPENAI_KEY_NAME is not set or empty")
		log.Printf("ERROR: %v", err)
		summary.Failed = summary.Total
		return summary, err
	}

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

	// Process every record even if one fails fatally, then report all fatal errors.
	// SES sends one record per event; a retry re-runs any others idempotently
	// (the archive writes overwrite, and the DynamoDB write is conditional).
	var fatalErrs []error
	for _, record := range event.Records {
		if err := h.processRecord(ctx, record, bucket, keyPrefix, &summary); err != nil {
			fatalErrs = append(fatalErrs, err)
		}
	}

	log.Printf("Processing complete: total=%d success=%d failed=%d duplicate=%d rejected=%d",
		summary.Total, summary.Success, summary.Failed, summary.Duplicate, summary.Rejected)

	if len(fatalErrs) > 0 {
		return summary, errors.Join(fatalErrs...)
	}
	return summary, nil
}

// processRecord handles a single SES event record. It returns a non-nil error only for
// failures that must fail the invocation (see HandleSESEvent).
func (h *Handler) processRecord(ctx context.Context, record events.SimpleEmailRecord, bucket, keyPrefix string, summary *Summary) error {
	messageID := record.SES.Mail.MessageID
	if messageID == "" {
		log.Printf("ERROR: SES record has empty MessageID")
		summary.Failed++
		return nil
	}

	// Derive S3 key with normalized prefix to avoid double slashes
	normalizedPrefix := strings.TrimSuffix(keyPrefix, "/")
	s3Key := normalizedPrefix + "/" + messageID

	// 1. SES verdict check comes first.
	outcome := evaluateVerdicts(record, messageID)
	if outcome == verdictVirus {
		summary.Rejected++
		return nil
	}

	// 2. Copy the raw email to archive/raw/ before parsing. A failure here fails the
	// invocation so Lambda retries it (and email-parser-errors fires) rather than losing
	// the only copy that can't be regenerated once incoming/ expires.
	if err := h.archiver.CopyRaw(ctx, bucket, s3Key, messageID); err != nil {
		log.Printf("ERROR: failed to copy email %s to %s: %v", messageID, archive.RawPrefix, err)
		summary.Failed++
		return fmt.Errorf("archive raw email %s: %w", messageID, err)
	}

	// 3. SPF, DKIM or spam FAIL: archived above, never parsed.
	if outcome == verdictRejected {
		summary.Rejected++
		return nil
	}

	// Fetch raw email from S3
	rawEmail, err := h.fetchEmail(ctx, bucket, s3Key)
	if err != nil {
		log.Printf("ERROR: failed to fetch email from s3://%s/%s: %v", bucket, s3Key, err)
		h.tagFailure(ctx, bucket, s3Key, "S3 fetch failed: "+err.Error())
		summary.Failed++
		return nil
	}

	// Parse raw email
	parsed, err := parser.ParseRawEmail(rawEmail)
	if err != nil {
		log.Printf("ERROR: failed to parse email %s: %v", messageID, err)
		h.tagFailure(ctx, bucket, s3Key, "Parse failed: "+err.Error())
		summary.Failed++
		return nil
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

	// 4. Write the clean record. A failure only logs: the raw email is already in
	// archive/raw/, and the backfill command (cmd/backfill) regenerates the record.
	cleanRecord := archive.NewCleanRecord(messageID, parsed, &extraction, archive.ExtractionSourceLive)
	if err := h.archiver.WriteClean(ctx, bucket, cleanRecord); err != nil {
		log.Printf("WARNING: failed to write %s record for email %s: %v", archive.CleanPrefix, messageID, err)
	}

	// Determine parse status
	status := "success"
	if extraction.IsEmpty() {
		status = "partial"
	}

	// Build the RecruiterEmail model
	now := time.Now().UTC()
	receivedAt := parsed.Date
	if receivedAt.IsZero() {
		receivedAt = now
	}
	email := &models.RecruiterEmail{
		ID:         messageID,
		ReceivedAt: receivedAt,
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
		DateYear:   receivedAt.Format("2006"),
		DateDay:    receivedAt.Format("2006-01-02"),
	}

	// Persist to DynamoDB
	result, err := h.store.WriteRecruiterEmail(ctx, email)
	if err != nil {
		log.Printf("ERROR: failed to write email %s to DynamoDB: %v", messageID, err)
		h.tagFailure(ctx, bucket, s3Key, "dynamodb_write_failed")
		summary.Failed++
		return nil
	}

	// Tag S3 object with parse results only after successful persistence
	h.tagger.TagObject(ctx, bucket, s3Key, tagger.TagResult{
		Status:     status,
		Company:    extraction.Company,
		Sender:     extraction.RecruiterEmail,
		Confidence: extraction.Confidence,
	})

	if result.Duplicate {
		summary.Duplicate++
	} else {
		summary.Success++
	}
	return nil
}

// evaluateVerdicts checks the SES receipt verdicts. Any FAIL rejects the email; a virus
// FAIL additionally keeps it out of the archive. Every failing verdict is logged.
func evaluateVerdicts(record events.SimpleEmailRecord, messageID string) verdictOutcome {
	verdicts := []struct {
		name   string
		status string
	}{
		{"SPF", record.SES.Receipt.SPFVerdict.Status},
		{"DKIM", record.SES.Receipt.DKIMVerdict.Status},
		{"Spam", record.SES.Receipt.SpamVerdict.Status},
		{"Virus", record.SES.Receipt.VirusVerdict.Status},
	}

	outcome := verdictPass
	for _, v := range verdicts {
		if !strings.EqualFold(v.status, "FAIL") {
			continue
		}
		log.Printf("Rejecting email %s: %s verdict FAIL", messageID, v.name)
		if v.name == "Virus" {
			outcome = verdictVirus
		} else if outcome == verdictPass {
			outcome = verdictRejected
		}
	}
	return outcome
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

	data, err := io.ReadAll(io.LimitReader(output.Body, maxEmailBytes+1))
	if err != nil {
		return nil, err
	}
	if len(data) > maxEmailBytes {
		return nil, fmt.Errorf("email size exceeds maximum allowed %d bytes", maxEmailBytes)
	}
	return data, nil
}

// tagFailure tags an S3 object with a failed parse status.
func (h *Handler) tagFailure(ctx context.Context, bucket, key, reason string) {
	h.tagger.TagObject(ctx, bucket, key, tagger.TagResult{
		Status:      "failed",
		ErrorReason: reason,
	})
}
