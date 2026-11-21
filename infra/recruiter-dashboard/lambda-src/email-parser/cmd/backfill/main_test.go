package main

import (
	"io"
	"strings"
	"testing"
)

func TestParseFlags(t *testing.T) {
	dir := t.TempDir()
	const arn = "arn:aws:kms:us-east-1:111122223333:key/11111111-2222-3333-4444-555555555555"
	tests := []struct {
		name    string
		args    []string
		wantErr string
	}{
		{"write mode", []string{"--bucket", "b", "--table", "t", "--kms-key-arn", arn}, ""},
		{"write mode with reconcile and force", []string{"--bucket", "b", "--table", "t", "--kms-key-arn", arn, "--reconcile", "--force"}, ""},
		{"dry run needs no key", []string{"--bucket", "b", "--table", "t", "--dry-run", "--reconcile"}, ""},
		{"out dir needs no key", []string{"--bucket", "b", "--table", "t", "--out-dir", dir}, ""},
		{"missing bucket", []string{"--table", "t", "--dry-run"}, "--bucket and --table are required"},
		{"missing table", []string{"--bucket", "b", "--dry-run"}, "--bucket and --table are required"},
		{"missing key in write mode", []string{"--bucket", "b", "--table", "t"}, "full key ARN"},
		{"alias is rejected", []string{"--bucket", "b", "--table", "t", "--kms-key-arn", "alias/recruiter-dashboard-archive"}, "full key ARN"},
		{"bare key id is rejected", []string{"--bucket", "b", "--table", "t", "--kms-key-arn", "11111111-2222-3333-4444-555555555555"}, "full key ARN"},
		{"out dir with reconcile", []string{"--bucket", "b", "--table", "t", "--out-dir", dir, "--reconcile"}, "can't be combined"},
		{"out dir must exist", []string{"--bucket", "b", "--table", "t", "--out-dir", dir + "/missing"}, "existing directory"},
		{"incoming prefix needs slash", []string{"--bucket", "b", "--table", "t", "--dry-run", "--incoming-prefix", "incoming"}, "must end in /"},
		{"incoming prefix not under archive", []string{"--bucket", "b", "--table", "t", "--dry-run", "--incoming-prefix", "archive/raw/"}, "must not be under archive/"},
		{"stray argument", []string{"--bucket", "b", "--table", "t", "--dry-run", "extra"}, "unexpected arguments"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parseFlags(tt.args, io.Discard)
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("error = %v, want one containing %q", err, tt.wantErr)
			}
		})
	}
}
