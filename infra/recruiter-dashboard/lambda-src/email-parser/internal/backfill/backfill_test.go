package backfill

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	dbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"

	"github.com/sh3r4rd/sh3r4rd.com/infra/recruiter-dashboard/lambda-src/email-parser/internal/archive"
)

const (
	testBucket = "test-bucket"
	testTable  = "test-table"
	testKeyARN = "arn:aws:kms:us-east-1:111122223333:key/11111111-2222-3333-4444-555555555555"
)

// fakeObject is one object in the fake bucket.
type fakeObject struct {
	body    []byte
	tags    int // number of tags (incoming/ objects are tagged by the Lambda)
	sse     types.ServerSideEncryption
	kmsKey  string
	written bool // created or replaced during the test run
}

// fakeS3 is an in-memory bucket implementing S3API and archive.S3Client. It never
// touches a real bucket. ListObjectsV2 pages two keys at a time to exercise pagination.
type fakeS3 struct {
	objects map[string]*fakeObject
	getErr  map[string]error // GetObject errors by key
	putErr  map[string]error // PutObject/CopyObject errors by destination key
	gets    int
}

func newFakeS3() *fakeS3 {
	return &fakeS3{objects: map[string]*fakeObject{}, getErr: map[string]error{}, putErr: map[string]error{}}
}

func (f *fakeS3) put(key string, body []byte) {
	f.objects[key] = &fakeObject{body: body}
}

func (f *fakeS3) keys(prefix string) []string {
	var keys []string
	for k := range f.objects {
		if strings.HasPrefix(k, prefix) {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)
	return keys
}

func (f *fakeS3) writtenKeys() []string {
	var keys []string
	for k, o := range f.objects {
		if o.written {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)
	return keys
}

func (f *fakeS3) resetWritten() {
	for _, o := range f.objects {
		o.written = false
	}
}

func (f *fakeS3) ListObjectsV2(ctx context.Context, in *s3.ListObjectsV2Input, optFns ...func(*s3.Options)) (*s3.ListObjectsV2Output, error) {
	keys := f.keys(aws.ToString(in.Prefix))
	start := 0
	if in.ContinuationToken != nil {
		fmt.Sscanf(aws.ToString(in.ContinuationToken), "%d", &start)
	}
	end := min(start+2, len(keys))
	out := &s3.ListObjectsV2Output{IsTruncated: aws.Bool(end < len(keys))}
	for _, k := range keys[start:end] {
		out.Contents = append(out.Contents, types.Object{Key: aws.String(k)})
	}
	if end < len(keys) {
		out.NextContinuationToken = aws.String(fmt.Sprintf("%d", end))
	}
	return out, nil
}

func (f *fakeS3) GetObject(ctx context.Context, in *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
	f.gets++
	key := aws.ToString(in.Key)
	if err := f.getErr[key]; err != nil {
		return nil, err
	}
	obj, ok := f.objects[key]
	if !ok {
		return nil, &types.NoSuchKey{}
	}
	return &s3.GetObjectOutput{Body: io.NopCloser(bytes.NewReader(obj.body))}, nil
}

func (f *fakeS3) CopyObject(ctx context.Context, in *s3.CopyObjectInput, optFns ...func(*s3.Options)) (*s3.CopyObjectOutput, error) {
	dst := aws.ToString(in.Key)
	if err := f.putErr[dst]; err != nil {
		return nil, err
	}
	src, err := url.PathUnescape(strings.TrimPrefix(aws.ToString(in.CopySource), testBucket+"/"))
	if err != nil {
		return nil, err
	}
	obj, ok := f.objects[src]
	if !ok {
		return nil, &types.NoSuchKey{}
	}
	tags := obj.tags
	if in.TaggingDirective == types.TaggingDirectiveReplace {
		tags = 0
	}
	f.objects[dst] = &fakeObject{
		body:    append([]byte(nil), obj.body...),
		tags:    tags,
		sse:     in.ServerSideEncryption,
		kmsKey:  aws.ToString(in.SSEKMSKeyId),
		written: true,
	}
	return &s3.CopyObjectOutput{}, nil
}

func (f *fakeS3) PutObject(ctx context.Context, in *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
	key := aws.ToString(in.Key)
	if err := f.putErr[key]; err != nil {
		return nil, err
	}
	body, err := io.ReadAll(in.Body)
	if err != nil {
		return nil, err
	}
	f.objects[key] = &fakeObject{
		body:    body,
		sse:     in.ServerSideEncryption,
		kmsKey:  aws.ToString(in.SSEKMSKeyId),
		written: true,
	}
	return &s3.PutObjectOutput{}, nil
}

// fakeDynamoDB returns one row per known message id.
type fakeDynamoDB struct {
	rows map[string]map[string]dbtypes.AttributeValue
	err  error
}

func (f *fakeDynamoDB) Query(ctx context.Context, in *dynamodb.QueryInput, optFns ...func(*dynamodb.Options)) (*dynamodb.QueryOutput, error) {
	if f.err != nil {
		return nil, f.err
	}
	if !aws.ToBool(in.ConsistentRead) {
		return nil, errors.New("test: lookups must use ConsistentRead")
	}
	id := in.ExpressionAttributeValues[":id"].(*dbtypes.AttributeValueMemberS).Value
	if row, ok := f.rows[id]; ok {
		return &dynamodb.QueryOutput{Items: []map[string]dbtypes.AttributeValue{row}}, nil
	}
	return &dynamodb.QueryOutput{}, nil
}

func ddbRow(company string, confidence string) map[string]dbtypes.AttributeValue {
	return map[string]dbtypes.AttributeValue{
		"first_name":      &dbtypes.AttributeValueMemberS{Value: "Jane"},
		"last_name":       &dbtypes.AttributeValueMemberS{Value: "Doe"},
		"recruiter_email": &dbtypes.AttributeValueMemberS{Value: "jane@example.com"},
		"company":         &dbtypes.AttributeValueMemberS{Value: company},
		"job_title":       &dbtypes.AttributeValueMemberS{Value: "Engineer"},
		"phone":           &dbtypes.AttributeValueMemberS{Value: "Unknown"},
		"confidence":      &dbtypes.AttributeValueMemberN{Value: confidence},
	}
}

// fixtureEmails loads every .eml in testdata, keyed by a synthetic message id.
func fixtureEmails(t *testing.T) map[string][]byte {
	t.Helper()
	paths, err := filepath.Glob("../../testdata/*.eml")
	if err != nil || len(paths) == 0 {
		t.Fatalf("no fixtures found: %v", err)
	}
	emails := map[string][]byte{}
	for _, p := range paths {
		body, err := os.ReadFile(p)
		if err != nil {
			t.Fatal(err)
		}
		id := "msg-" + strings.ReplaceAll(strings.TrimSuffix(filepath.Base(p), ".eml"), "_", "-")
		emails[id] = body
	}
	return emails
}

func newS3Runner(fs *fakeS3, ddb DynamoDBAPI, opts Options) *Runner {
	archiver := archive.NewArchiver(fs, testKeyARN)
	opts.Bucket = testBucket
	opts.Table = testTable
	if opts.IncomingPrefix == "" {
		opts.IncomingPrefix = "incoming/"
	}
	return NewRunner(fs, ddb, archiver, NewS3Sink(fs, archiver, testBucket), opts)
}

func readRecord(t *testing.T, fs *fakeS3, id string) archive.CleanRecord {
	t.Helper()
	obj, ok := fs.objects[archive.CleanKey(id)]
	if !ok {
		t.Fatalf("no clean record for %s", id)
	}
	var rec archive.CleanRecord
	if err := json.Unmarshal(obj.body, &rec); err != nil {
		t.Fatalf("record %s is not valid JSON: %v", id, err)
	}
	return rec
}

func TestRun_FixturesTwiceWritesNothingTheSecondTime(t *testing.T) {
	fs := newFakeS3()
	emails := fixtureEmails(t)
	for id, body := range emails {
		fs.put(archive.RawKey(id), body)
	}
	ddb := &fakeDynamoDB{rows: map[string]map[string]dbtypes.AttributeValue{"msg-gmail-forward": ddbRow("ExampleCorp", "0.85")}}

	res, err := newS3Runner(fs, ddb, Options{}).Run(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if res.RawCount != len(emails) || res.Written != len(emails) || len(res.Failures) != 0 {
		t.Fatalf("first run = %+v, want all %d written", res, len(emails))
	}
	if res.WithExtraction != 1 {
		t.Errorf("WithExtraction = %d, want 1 (only msg-gmail-forward has a row)", res.WithExtraction)
	}
	for id := range emails {
		obj := fs.objects[archive.CleanKey(id)]
		if obj.sse != types.ServerSideEncryptionAwsKms || obj.kmsKey != testKeyARN {
			t.Errorf("%s: sse=%q key=%q, want SSE-KMS with the archive key", id, obj.sse, obj.kmsKey)
		}
	}

	// The DynamoDB row supplies the extraction; ids without a row get none.
	joined := readRecord(t, fs, "msg-gmail-forward")
	if joined.ExtractionSource != archive.ExtractionSourceDynamoDB || joined.Extraction == nil ||
		joined.Extraction.Company != "ExampleCorp" || joined.Extraction.Confidence != 0.85 {
		t.Errorf("joined record extraction = %+v (%q)", joined.Extraction, joined.ExtractionSource)
	}
	if !joined.IsForwarded {
		t.Error("gmail_forward fixture should be IsForwarded")
	}
	plain := readRecord(t, fs, "msg-direct-email")
	if plain.ExtractionSource != archive.ExtractionSourceNone || plain.Extraction != nil {
		t.Errorf("record without a row = %+v (%q), want null/none", plain.Extraction, plain.ExtractionSource)
	}

	// Second run: everything exists, nothing is written.
	fs.resetWritten()
	res, err = newS3Runner(fs, ddb, Options{}).Run(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if res.Written != 0 || res.SkippedExisting != len(emails) {
		t.Errorf("second run = %+v, want 0 written and %d skipped", res, len(emails))
	}
	if w := fs.writtenKeys(); len(w) != 0 {
		t.Errorf("second run wrote %v", w)
	}
}

func TestRun_ForceRewritesIdenticalBytes(t *testing.T) {
	fs := newFakeS3()
	emails := fixtureEmails(t)
	for id, body := range emails {
		fs.put(archive.RawKey(id), body)
	}
	ddb := &fakeDynamoDB{}

	if _, err := newS3Runner(fs, ddb, Options{}).Run(context.Background()); err != nil {
		t.Fatal(err)
	}
	first := map[string][]byte{}
	for id := range emails {
		first[id] = fs.objects[archive.CleanKey(id)].body
	}

	fs.resetWritten()
	res, err := newS3Runner(fs, ddb, Options{Force: true}).Run(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if res.Written != len(emails) || res.SkippedExisting != 0 {
		t.Fatalf("forced run = %+v", res)
	}
	for id := range emails {
		if !bytes.Equal(first[id], fs.objects[archive.CleanKey(id)].body) {
			t.Errorf("%s: output differs between runs (not deterministic)", id)
		}
	}
}

func TestRun_ReconcileRestoresMissingRawAndClean(t *testing.T) {
	// incoming/ has A, B and C. C's raw copy is missing (its retries ran out), and B's
	// clean record is missing (the Lambda's archive/clean write failed).
	fs := newFakeS3()
	emails := fixtureEmails(t)
	a, b, c := emails["msg-direct-email"], emails["msg-gmail-forward"], emails["msg-outlook-forward"]
	fs.put("incoming/aaa", a)
	fs.put("incoming/bbb", b)
	fs.put("incoming/ccc", c)
	for _, k := range []string{"incoming/aaa", "incoming/bbb", "incoming/ccc"} {
		fs.objects[k].tags = 5 // the Lambda tags incoming/ objects
	}
	fs.put("archive/raw/aaa", a)
	fs.put("archive/raw/bbb", b)
	fs.put("archive/clean/aaa.json", []byte(`{"message_id":"aaa"}`))

	res, err := newS3Runner(fs, &fakeDynamoDB{}, Options{Reconcile: true}).Run(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	if res.IncomingCount != 3 || res.MissingFromRaw != 1 || res.Copied != 1 || res.SkippedVirus != 0 {
		t.Errorf("reconcile counts = %+v", res)
	}
	if res.RawCount != 3 || res.Written != 2 || res.SkippedExisting != 1 || len(res.Failures) != 0 {
		t.Errorf("backfill counts = %+v, want raw=3 written=2 skipped=1", res)
	}

	raw := fs.objects["archive/raw/ccc"]
	if raw == nil {
		t.Fatal("archive/raw/ccc was not restored")
	}
	if raw.tags != 0 {
		t.Errorf("restored raw copy carries %d tags, want 0", raw.tags)
	}
	if raw.sse != types.ServerSideEncryptionAwsKms || raw.kmsKey != testKeyARN {
		t.Errorf("restored raw copy sse=%q key=%q", raw.sse, raw.kmsKey)
	}
	if !bytes.Equal(raw.body, c) {
		t.Error("restored raw copy differs from the original")
	}
	for _, id := range []string{"bbb", "ccc"} {
		if rec := readRecord(t, fs, id); rec.MessageID != id || rec.RawKey != "archive/raw/"+id {
			t.Errorf("record %s = %+v", id, rec)
		}
	}
	if got := fs.writtenKeys(); strings.Join(got, ",") != "archive/clean/bbb.json,archive/clean/ccc.json,archive/raw/ccc" {
		t.Errorf("written keys = %v", got)
	}
}

func withHeaders(headers string, body []byte) []byte {
	return append([]byte(headers), body...)
}

func TestRun_ReconcileSkipsVirusFlaggedMail(t *testing.T) {
	fs := newFakeS3()
	body := fixtureEmails(t)["msg-direct-email"]
	fs.put("incoming/virus", withHeaders("X-SES-Virus-Verdict: FAIL\r\n", body))
	fs.put("incoming/spam", withHeaders("X-SES-Spam-Verdict: FAIL\r\n", body))

	res, err := newS3Runner(fs, &fakeDynamoDB{}, Options{Reconcile: true}).Run(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if res.MissingFromRaw != 2 || res.SkippedVirus != 1 || res.Copied != 1 {
		t.Errorf("reconcile counts = %+v, want 2 missing, 1 virus skipped, 1 copied", res)
	}
	if _, ok := fs.objects["archive/raw/virus"]; ok {
		t.Error("virus-flagged mail must never be copied to archive/raw/")
	}
	if _, ok := fs.objects["archive/raw/spam"]; !ok {
		t.Error("spam is archived (like the Lambda does), so it should be copied")
	}
	// Spam is archived but never parsed.
	if res.Rejected != 1 || res.Written != 0 {
		t.Errorf("backfill counts = %+v, want the spam counted as rejected and not written", res)
	}
	if _, ok := fs.objects["archive/clean/spam.json"]; ok {
		t.Error("rejected mail must not get a clean record")
	}
}

func TestRun_CountsRejectedAndFailures(t *testing.T) {
	fs := newFakeS3()
	emails := fixtureEmails(t)
	good := emails["msg-direct-email"]
	fs.put("archive/raw/good", good)
	fs.put("archive/raw/spam", withHeaders("X-SES-Spam-Verdict: FAIL\r\n", good))
	fs.put("archive/raw/oldvirus", withHeaders("X-SES-Virus-Verdict: FAIL\r\n", good)) // e.g. a week 0 copy
	fs.put("archive/raw/garbled", []byte("Content-Type: multipart/mixed\r\nSecret Name: Jane Doe at 555-0100\r\nno-colon-line here\r\n\r\nbody"))
	fs.put("archive/raw/unreadable", good)
	fs.getErr["archive/raw/unreadable"] = errors.New("simulated read failure with secret content")
	fs.put("archive/raw/nested/key", good)
	fs.put("archive/raw/writefails", good)
	fs.putErr["archive/clean/writefails.json"] = errors.New("simulated AccessDenied")

	res, err := newS3Runner(fs, &fakeDynamoDB{}, Options{}).Run(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if res.Written != 1 || res.Rejected != 2 {
		t.Errorf("counts = %+v, want written=1 rejected=2", res)
	}

	gotKeys := make([]string, len(res.Failures))
	for i, f := range res.Failures {
		gotKeys[i] = f.Key
		if strings.Contains(f.Reason, "Jane") || strings.Contains(f.Reason, "555") || strings.Contains(f.Reason, "secret") {
			t.Errorf("failure reason leaks content: %q", f.Reason)
		}
	}
	wantKeys := []string{"archive/clean/writefails.json", "archive/raw/garbled", "archive/raw/nested/key", "archive/raw/unreadable"}
	if strings.Join(gotKeys, ",") != strings.Join(wantKeys, ",") {
		t.Errorf("failure keys = %v, want %v (sorted)", gotKeys, wantKeys)
	}

	// archive/clean count = archive/raw count - rejected - failures (#133 acceptance).
	cleanCount := len(fs.keys(archive.CleanPrefix))
	parseableRaw := res.RawCount // nested/key isn't a message id, so it isn't counted
	if cleanCount != parseableRaw-res.Rejected-3 {
		t.Errorf("clean=%d, raw=%d, rejected=%d, per-object failures=3", cleanCount, parseableRaw, res.Rejected)
	}
}

func TestRun_DynamoDBErrorFailsTheKeyInsteadOfWritingNone(t *testing.T) {
	fs := newFakeS3()
	fs.put("archive/raw/abc", fixtureEmails(t)["msg-direct-email"])

	res, err := newS3Runner(fs, &fakeDynamoDB{err: errors.New("throttled")}, Options{}).Run(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if res.Written != 0 || len(res.Failures) != 1 || res.Failures[0].Key != "archive/raw/abc" {
		t.Errorf("result = %+v", res)
	}
	if _, ok := fs.objects["archive/clean/abc.json"]; ok {
		t.Error("a record with a missing extraction would be skipped forever on later runs")
	}
}

func TestRun_DryRunWritesNothing(t *testing.T) {
	fs := newFakeS3()
	emails := fixtureEmails(t)
	fs.put("incoming/aaa", emails["msg-direct-email"])
	fs.put("incoming/bbb", emails["msg-gmail-forward"])
	fs.put("archive/raw/aaa", emails["msg-direct-email"])

	runner := NewRunner(fs, &fakeDynamoDB{}, nil, NewS3Sink(fs, nil, testBucket), Options{
		Bucket: testBucket, Table: testTable, IncomingPrefix: "incoming/", Reconcile: true, DryRun: true,
	})
	res, err := runner.Run(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if res.Copied != 1 || res.Written != 2 || len(res.Failures) != 0 {
		t.Errorf("dry run = %+v, want copied=1 written=2 (what a real run would do)", res)
	}
	if w := fs.writtenKeys(); len(w) != 0 {
		t.Errorf("dry run wrote %v", w)
	}
}

func TestRun_DirSink(t *testing.T) {
	fs := newFakeS3()
	emails := fixtureEmails(t)
	for id, body := range emails {
		fs.put(archive.RawKey(id), body)
	}
	dir := t.TempDir()
	runner := NewRunner(fs, &fakeDynamoDB{}, nil, NewDirSink(dir), Options{Bucket: testBucket, Table: testTable, IncomingPrefix: "incoming/"})

	res, err := runner.Run(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if res.Written != len(emails) {
		t.Fatalf("result = %+v", res)
	}
	if w := fs.writtenKeys(); len(w) != 0 {
		t.Errorf("dir sink wrote to the bucket: %v", w)
	}
	for id := range emails {
		info, err := os.Stat(filepath.Join(dir, id+".json"))
		if err != nil {
			t.Fatalf("missing %s.json: %v", id, err)
		}
		if info.Mode().Perm() != 0o600 {
			t.Errorf("%s.json mode = %v, want 0600", id, info.Mode().Perm())
		}
	}

	// Existing files are skipped on the next run.
	res, err = runner.Run(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if res.Written != 0 || res.SkippedExisting != len(emails) {
		t.Errorf("second dir run = %+v", res)
	}
}

func TestDescribeErr(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want string
	}{
		{"aws error code", &types.NoSuchKey{}, "error: NoSuchKey"},
		{"too large", errTooLarge, "object exceeds 10 MB"},
		{"plain error is withheld", errors.New("mail: malformed header line: Jane Doe"), "error"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := describeErr(tt.err); got != tt.want {
				t.Errorf("describeErr() = %q, want %q", got, tt.want)
			}
		})
	}
}
