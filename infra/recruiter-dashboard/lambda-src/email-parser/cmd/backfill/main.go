// Command backfill regenerates archive/clean/{id}.json records from the raw emails in
// archive/raw/, using the same parser as the email-parser Lambda and no LLM calls.
// The extraction in each record is copied from the email's DynamoDB row when one exists.
//
// It is run by hand from a workstation with AWS credentials, never in Lambda:
//
//	go run ./cmd/backfill --bucket <bucket> --table <table> --kms-key-arn <arn> [--reconcile] [--force]
//	go run ./cmd/backfill --bucket <bucket> --table <table> --dry-run [--reconcile]
//	go run ./cmd/backfill --bucket <bucket> --table <table> --out-dir <dir>
//
// It prints counts and failing keys only, never email content. The exit status is 1 if
// any object failed and 2 for usage or listing errors.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"github.com/sh3r4rd/sh3r4rd.com/infra/recruiter-dashboard/lambda-src/email-parser/internal/archive"
	"github.com/sh3r4rd/sh3r4rd.com/infra/recruiter-dashboard/lambda-src/email-parser/internal/backfill"
)

type flags struct {
	bucket         string
	table          string
	kmsKeyARN      string
	incomingPrefix string
	outDir         string
	force          bool
	reconcile      bool
	dryRun         bool
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	f, err := parseFlags(args, stderr)
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		fmt.Fprintf(stderr, "backfill: %v\n", err)
		return 2
	}

	ctx := context.Background()
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		fmt.Fprintf(stderr, "backfill: load AWS config: %v\n", err)
		return 2
	}
	s3Client := s3.NewFromConfig(cfg)
	ddbClient := dynamodb.NewFromConfig(cfg)

	var sink backfill.CleanSink
	var copier backfill.RawCopier
	mode := "write"
	switch {
	case f.outDir != "":
		mode = "out-dir"
		sink = backfill.NewDirSink(f.outDir)
	case f.dryRun:
		mode = "dry-run"
		// Listed but never written to in a dry run.
		sink = backfill.NewS3Sink(s3Client, nil, f.bucket)
	default:
		archiver := archive.NewArchiver(s3Client, f.kmsKeyARN)
		sink = backfill.NewS3Sink(s3Client, archiver, f.bucket)
		copier = archiver
	}

	runner := backfill.NewRunner(s3Client, ddbClient, copier, sink, backfill.Options{
		Bucket:         f.bucket,
		IncomingPrefix: f.incomingPrefix,
		Table:          f.table,
		Force:          f.force,
		Reconcile:      f.reconcile,
		DryRun:         f.dryRun,
	})

	fmt.Fprintf(stdout, "backfill: bucket=%s table=%s mode=%s force=%t reconcile=%t\n",
		f.bucket, f.table, mode, f.force, f.reconcile)

	res, err := runner.Run(ctx)
	if err != nil {
		fmt.Fprintf(stderr, "backfill: %v\n", err)
		return 2
	}
	printResult(stdout, f, res)

	if len(res.Failures) > 0 {
		return 1
	}
	return 0
}

func parseFlags(args []string, stderr io.Writer) (flags, error) {
	var f flags
	fs := flag.NewFlagSet("backfill", flag.ContinueOnError)
	fs.SetOutput(stderr)
	fs.StringVar(&f.bucket, "bucket", "", "email bucket (required)")
	fs.StringVar(&f.table, "table", "", "DynamoDB table with the extractions (required)")
	fs.StringVar(&f.kmsKeyARN, "kms-key-arn", "", "full ARN of the archive KMS key (required unless --dry-run or --out-dir); terraform output -raw archive_kms_key_arn")
	fs.StringVar(&f.incomingPrefix, "incoming-prefix", "incoming/", "prefix SES writes to")
	fs.StringVar(&f.outDir, "out-dir", "", "write records to this existing local directory instead of archive/clean/ (no bucket writes)")
	fs.BoolVar(&f.force, "force", false, "rewrite records that already exist")
	fs.BoolVar(&f.reconcile, "reconcile", false, "first copy incoming/ emails missing from archive/raw/ (tag-free, SSE-KMS, virus-flagged ones skipped)")
	fs.BoolVar(&f.dryRun, "dry-run", false, "read and parse everything, write nothing")
	if err := fs.Parse(args); err != nil {
		return f, err
	}
	if fs.NArg() > 0 {
		return f, fmt.Errorf("unexpected arguments: %v", fs.Args())
	}

	if f.bucket == "" || f.table == "" {
		return f, errors.New("--bucket and --table are required")
	}
	if !strings.HasSuffix(f.incomingPrefix, "/") || strings.HasPrefix(f.incomingPrefix, archive.RawPrefix[:len("archive/")]) {
		return f, errors.New("--incoming-prefix must end in / and must not be under archive/")
	}
	if f.outDir != "" {
		if f.reconcile || f.dryRun {
			return f, errors.New("--out-dir can't be combined with --reconcile or --dry-run: it never writes to the bucket")
		}
		info, err := os.Stat(f.outDir)
		if err != nil || !info.IsDir() {
			return f, fmt.Errorf("--out-dir %q must be an existing directory", f.outDir)
		}
		return f, nil
	}
	if f.dryRun {
		return f, nil
	}
	// The bucket policy compares the request's key id with the full key ARN, so an alias
	// or bare key id would be denied on every write.
	if !strings.HasPrefix(f.kmsKeyARN, "arn:aws:kms:") || !strings.Contains(f.kmsKeyARN, ":key/") {
		return f, errors.New("--kms-key-arn must be the full key ARN (arn:aws:kms:<region>:<account>:key/<id>), not an alias")
	}
	return f, nil
}

func printResult(w io.Writer, f flags, res backfill.Result) {
	if f.reconcile {
		fmt.Fprintf(w, "reconcile: incoming=%d missing_from_raw=%d copied=%d skipped_virus=%d\n",
			res.IncomingCount, res.MissingFromRaw, res.Copied, res.SkippedVirus)
	}
	fmt.Fprintf(w, "backfill: raw=%d written=%d (with_dynamodb_extraction=%d) skipped_existing=%d rejected=%d failed=%d\n",
		res.RawCount, res.Written, res.WithExtraction, res.SkippedExisting, res.Rejected, len(res.Failures))
	if f.dryRun {
		fmt.Fprintln(w, "dry run: nothing was written; copied/written are what a real run would do")
	}
	if len(res.Failures) > 0 {
		fmt.Fprintln(w, "failed keys:")
		for _, failure := range res.Failures {
			fmt.Fprintf(w, "  %s  %s\n", failure.Key, failure.Reason)
		}
	}
}
