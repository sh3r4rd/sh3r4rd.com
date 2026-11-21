// Package backfill regenerates archive/clean/ records from the raw emails in
// archive/raw/, deterministically and with no LLM calls. With Reconcile it first copies
// any incoming/ email missing from archive/raw/, which is the recovery path when the
// Lambda's raw copy failed and its async retries ran out (#133).
package backfill

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	dbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"github.com/sh3r4rd/sh3r4rd.com/infra/recruiter-dashboard/lambda-src/email-parser/internal/archive"
	perrors "github.com/sh3r4rd/sh3r4rd.com/infra/recruiter-dashboard/lambda-src/email-parser/internal/errors"
	"github.com/sh3r4rd/sh3r4rd.com/infra/recruiter-dashboard/lambda-src/email-parser/internal/models"
	"github.com/sh3r4rd/sh3r4rd.com/infra/recruiter-dashboard/lambda-src/email-parser/internal/parser"
)

// maxEmailBytes matches the Lambda's limit (handler.maxEmailBytes).
const maxEmailBytes = 10 * 1024 * 1024

// S3API is the subset of the S3 API the backfill reads with.
type S3API interface {
	ListObjectsV2(ctx context.Context, params *s3.ListObjectsV2Input, optFns ...func(*s3.Options)) (*s3.ListObjectsV2Output, error)
	GetObject(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error)
}

// DynamoDBAPI is the subset of the DynamoDB API used to look up existing extractions.
type DynamoDBAPI interface {
	Query(ctx context.Context, params *dynamodb.QueryInput, optFns ...func(*dynamodb.Options)) (*dynamodb.QueryOutput, error)
}

// RawCopier copies an incoming/ email to archive/raw/ (archive.Archiver implements it).
type RawCopier interface {
	CopyRaw(ctx context.Context, bucket, sourceKey, messageID string) error
}

// CleanSink is where clean records go: archive/clean/ in S3, or a local directory.
type CleanSink interface {
	// Existing returns the message ids that already have a clean record.
	Existing(ctx context.Context) (map[string]bool, error)
	// Write stores one record, replacing any existing one.
	Write(ctx context.Context, rec archive.CleanRecord) error
}

// Options controls a run.
type Options struct {
	Bucket         string // the email bucket
	IncomingPrefix string // where SES writes, e.g. "incoming/"
	Table          string // DynamoDB table holding the extractions
	Force          bool   // rewrite records that already exist
	Reconcile      bool   // first copy incoming/ emails missing from archive/raw/
	DryRun         bool   // read and parse everything, write nothing
}

// Failure is one object that couldn't be processed. Reason never contains email
// content: it is an error category, an operation name or an AWS error code.
type Failure struct {
	Key    string
	Reason string
}

// Result holds the counts from a run.
type Result struct {
	// Reconcile phase (zero unless Options.Reconcile).
	IncomingCount  int // emails under incoming/
	MissingFromRaw int // incoming/ emails not in archive/raw/ before the run
	Copied         int // copied to archive/raw/ (or would be, in a dry run)
	SkippedVirus   int // missing emails not copied because the virus verdict is FAIL

	// Backfill phase.
	RawCount        int // emails in archive/raw/ (including ones just copied)
	Written         int // clean records written (or would be, in a dry run)
	WithExtraction  int // of those, records whose extraction came from a DynamoDB row
	SkippedExisting int // records that already existed (without Force)
	Rejected        int // emails SES rejected (SPF, DKIM, spam or virus FAIL): never parsed

	Failures []Failure // sorted by key
}

// Runner performs a backfill.
type Runner struct {
	s3     S3API
	ddb    DynamoDBAPI
	copier RawCopier
	sink   CleanSink
	opts   Options
}

// NewRunner creates a Runner. copier may be nil unless opts.Reconcile is set.
func NewRunner(s3Client S3API, ddb DynamoDBAPI, copier RawCopier, sink CleanSink, opts Options) *Runner {
	return &Runner{s3: s3Client, ddb: ddb, copier: copier, sink: sink, opts: opts}
}

// Run executes the reconcile phase (if enabled) and then the backfill. It returns an
// error only when a listing fails; per-object problems are reported in Result.Failures.
func (r *Runner) Run(ctx context.Context) (Result, error) {
	var res Result

	rawIDs, listFailures, err := r.listIDs(ctx, archive.RawPrefix, "")
	if err != nil {
		return res, fmt.Errorf("list %s: %w", archive.RawPrefix, err)
	}
	res.Failures = append(res.Failures, listFailures...)

	// sources maps each message id to the key its raw email is read from.
	sources := make(map[string]string, len(rawIDs))
	for id := range rawIDs {
		sources[id] = archive.RawKey(id)
	}

	if r.opts.Reconcile {
		if err := r.reconcile(ctx, rawIDs, sources, &res); err != nil {
			return res, err
		}
	}

	existing, err := r.sink.Existing(ctx)
	if err != nil {
		return res, fmt.Errorf("list existing clean records: %w", err)
	}

	ids := sortedKeys(sources)
	res.RawCount = len(ids)
	for _, id := range ids {
		if existing[id] && !r.opts.Force {
			res.SkippedExisting++
			continue
		}
		r.backfillOne(ctx, id, sources[id], &res)
	}

	sort.Slice(res.Failures, func(i, j int) bool { return res.Failures[i].Key < res.Failures[j].Key })
	return res, nil
}

// reconcile copies every incoming/ email that's missing from archive/raw/, except
// virus-flagged ones, and adds the copies to sources.
func (r *Runner) reconcile(ctx context.Context, rawIDs map[string]bool, sources map[string]string, res *Result) error {
	incomingIDs, listFailures, err := r.listIDs(ctx, r.opts.IncomingPrefix, "")
	if err != nil {
		return fmt.Errorf("list %s: %w", r.opts.IncomingPrefix, err)
	}
	res.Failures = append(res.Failures, listFailures...)
	res.IncomingCount = len(incomingIDs)

	for _, id := range sortedKeys(incomingIDs) {
		if rawIDs[id] {
			continue
		}
		res.MissingFromRaw++
		incomingKey := r.opts.IncomingPrefix + id

		raw, err := r.getObject(ctx, incomingKey)
		if err != nil {
			res.Failures = append(res.Failures, Failure{Key: incomingKey, Reason: describeErr(err)})
			continue
		}
		if VerdictFromHeaders(raw) == VerdictVirus {
			res.SkippedVirus++
			continue
		}

		if r.opts.DryRun {
			// Nothing was copied, so the backfill phase reads the original instead.
			sources[id] = incomingKey
		} else {
			if err := r.copier.CopyRaw(ctx, r.opts.Bucket, incomingKey, id); err != nil {
				res.Failures = append(res.Failures, Failure{Key: incomingKey, Reason: describeErr(err)})
				continue
			}
			sources[id] = archive.RawKey(id)
		}
		res.Copied++
	}
	return nil
}

// backfillOne parses one raw email and writes its clean record.
func (r *Runner) backfillOne(ctx context.Context, id, sourceKey string, res *Result) {
	raw, err := r.getObject(ctx, sourceKey)
	if err != nil {
		res.Failures = append(res.Failures, Failure{Key: sourceKey, Reason: describeErr(err)})
		return
	}

	// Same rule as the Lambda: anything SES rejected is never parsed.
	if VerdictFromHeaders(raw) != VerdictPass {
		res.Rejected++
		return
	}

	parsed, err := parser.ParseRawEmail(raw)
	if err != nil {
		res.Failures = append(res.Failures, Failure{Key: sourceKey, Reason: describeErr(err)})
		return
	}

	extraction, source, err := r.lookupExtraction(ctx, id)
	if err != nil {
		res.Failures = append(res.Failures, Failure{Key: sourceKey, Reason: "dynamodb lookup: " + describeErr(err)})
		return
	}

	rec := archive.NewCleanRecord(id, parsed, extraction, source)
	if !r.opts.DryRun {
		if err := r.sink.Write(ctx, rec); err != nil {
			res.Failures = append(res.Failures, Failure{Key: archive.CleanKey(id), Reason: describeErr(err)})
			return
		}
	}
	res.Written++
	if source == archive.ExtractionSourceDynamoDB {
		res.WithExtraction++
	}
}

// lookupExtraction returns the extraction the Lambda stored in DynamoDB for this message,
// or (nil, "none") when there's no row. The rows hold the sanitized values, the same
// ones the Lambda puts in a live clean record.
func (r *Runner) lookupExtraction(ctx context.Context, id string) (*models.ExtractionResult, string, error) {
	out, err := r.ddb.Query(ctx, &dynamodb.QueryInput{
		TableName: aws.String(r.opts.Table),
		// A row the Lambda just wrote must be seen, or the record would be written with
		// extraction "none" and skipped by every later run without --force.
		ConsistentRead:         aws.Bool(true),
		KeyConditionExpression: aws.String("id = :id"),
		ExpressionAttributeValues: map[string]dbtypes.AttributeValue{
			":id": &dbtypes.AttributeValueMemberS{Value: id},
		},
	})
	if err != nil {
		return nil, "", err
	}
	if len(out.Items) == 0 {
		return nil, archive.ExtractionSourceNone, nil
	}

	// One row per message id in practice; Query returns rows in sort-key (received_at)
	// order, so taking the first is deterministic either way.
	item := out.Items[0]
	confidence, err := numberAttr(item, "confidence")
	if err != nil {
		return nil, "", err
	}
	return &models.ExtractionResult{
		RecruiterFirstName: stringAttr(item, "first_name"),
		RecruiterLastName:  stringAttr(item, "last_name"),
		RecruiterEmail:     stringAttr(item, "recruiter_email"),
		Company:            stringAttr(item, "company"),
		JobTitle:           stringAttr(item, "job_title"),
		Phone:              stringAttr(item, "phone"),
		Confidence:         confidence,
	}, archive.ExtractionSourceDynamoDB, nil
}

// listIDs lists the message ids under prefix whose keys end in suffix. Keys that can't
// be message ids (nested paths, invalid characters) are returned as failures.
func (r *Runner) listIDs(ctx context.Context, prefix, suffix string) (map[string]bool, []Failure, error) {
	keys, err := listKeys(ctx, r.s3, r.opts.Bucket, prefix)
	if err != nil {
		return nil, nil, err
	}
	return idsFromKeys(keys, prefix, suffix)
}

func idsFromKeys(keys []string, prefix, suffix string) (map[string]bool, []Failure, error) {
	ids := make(map[string]bool, len(keys))
	var failures []Failure
	for _, key := range keys {
		if key == prefix {
			continue // a zero-byte "folder" marker
		}
		name := strings.TrimPrefix(key, prefix)
		if suffix != "" {
			if !strings.HasSuffix(name, suffix) {
				failures = append(failures, Failure{Key: key, Reason: "unexpected key: missing " + suffix + " suffix"})
				continue
			}
			name = strings.TrimSuffix(name, suffix)
		}
		if err := archive.ValidateMessageID(name); err != nil {
			failures = append(failures, Failure{Key: key, Reason: "unexpected key: not a message id"})
			continue
		}
		ids[name] = true
	}
	return ids, failures, nil
}

func listKeys(ctx context.Context, client S3API, bucket, prefix string) ([]string, error) {
	var keys []string
	var token *string
	for {
		out, err := client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
			Bucket:            aws.String(bucket),
			Prefix:            aws.String(prefix),
			ContinuationToken: token,
		})
		if err != nil {
			return nil, err
		}
		for _, obj := range out.Contents {
			keys = append(keys, aws.ToString(obj.Key))
		}
		if !aws.ToBool(out.IsTruncated) {
			return keys, nil
		}
		token = out.NextContinuationToken
	}
}

func (r *Runner) getObject(ctx context.Context, key string) ([]byte, error) {
	out, err := r.s3.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(r.opts.Bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, err
	}
	defer out.Body.Close()

	data, err := io.ReadAll(io.LimitReader(out.Body, maxEmailBytes+1))
	if err != nil {
		return nil, fmt.Errorf("read object body: %w", err)
	}
	if len(data) > maxEmailBytes {
		return nil, errTooLarge
	}
	return data, nil
}

var errTooLarge = errors.New("object exceeds 10 MB")

// describeErr reduces an error to a reason that's safe to print: parser errors can
// quote a malformed header line, so only their operation name is kept.
func describeErr(err error) string {
	var parseErr *perrors.ParseError
	if errors.As(err, &parseErr) {
		return "parse error [" + parseErr.Op + "]"
	}

	reason := "error"
	var storageErr *perrors.StorageError
	if errors.As(err, &storageErr) {
		reason = "storage error [" + storageErr.Op + "]"
	}
	var apiErr interface{ ErrorCode() string }
	if errors.As(err, &apiErr) {
		return reason + ": " + apiErr.ErrorCode()
	}
	if errors.Is(err, errTooLarge) {
		return errTooLarge.Error()
	}
	return reason
}

func stringAttr(item map[string]dbtypes.AttributeValue, name string) string {
	if v, ok := item[name].(*dbtypes.AttributeValueMemberS); ok {
		return v.Value
	}
	return ""
}

func numberAttr(item map[string]dbtypes.AttributeValue, name string) (float64, error) {
	v, ok := item[name].(*dbtypes.AttributeValueMemberN)
	if !ok {
		return 0, nil
	}
	f, err := strconv.ParseFloat(v.Value, 64)
	if err != nil {
		return 0, fmt.Errorf("attribute %s is not a number", name)
	}
	return f, nil
}

func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// S3Sink writes clean records to archive/clean/ with the archive key.
type S3Sink struct {
	s3       S3API
	archiver *archive.Archiver
	bucket   string
}

// NewS3Sink creates a sink that lists and writes archive/clean/ in bucket.
func NewS3Sink(s3Client S3API, archiver *archive.Archiver, bucket string) *S3Sink {
	return &S3Sink{s3: s3Client, archiver: archiver, bucket: bucket}
}

// Existing lists the message ids that have a record under archive/clean/.
func (s *S3Sink) Existing(ctx context.Context) (map[string]bool, error) {
	keys, err := listKeys(ctx, s.s3, s.bucket, archive.CleanPrefix)
	if err != nil {
		return nil, err
	}
	ids, _, err := idsFromKeys(keys, archive.CleanPrefix, ".json")
	return ids, err
}

// Write stores rec at archive/clean/{id}.json.
func (s *S3Sink) Write(ctx context.Context, rec archive.CleanRecord) error {
	return s.archiver.WriteClean(ctx, s.bucket, rec)
}

// DirSink writes clean records as {id}.json files in a local directory, for a run that
// must not write to the bucket. The files contain email content: keep the directory
// private and delete it when done.
type DirSink struct {
	dir string
}

// NewDirSink creates a sink that writes into dir, which must already exist.
func NewDirSink(dir string) *DirSink {
	return &DirSink{dir: dir}
}

// Existing lists the message ids that have a {id}.json file in the directory.
func (d *DirSink) Existing(ctx context.Context) (map[string]bool, error) {
	entries, err := os.ReadDir(d.dir)
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		if !e.IsDir() {
			names = append(names, e.Name())
		}
	}
	ids, _, err := idsFromKeys(names, "", ".json")
	return ids, err
}

// Write stores rec as {id}.json, readable only by the current user.
func (d *DirSink) Write(ctx context.Context, rec archive.CleanRecord) error {
	if err := archive.ValidateMessageID(rec.MessageID); err != nil {
		return err
	}
	body, err := rec.Marshal()
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(d.dir, rec.MessageID+".json"), body, 0o600)
}
