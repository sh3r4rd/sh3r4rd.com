package errors

import "fmt"

// ParseError indicates a failure during MIME parsing.
type ParseError struct {
	Op  string
	Err error
}

func (e *ParseError) Error() string {
	return fmt.Sprintf("parse error [%s]: %v", e.Op, e.Err)
}

func (e *ParseError) Unwrap() error { return e.Err }

// ExtractionError indicates a failure during recruiter data extraction.
type ExtractionError struct {
	Op  string
	Err error
}

func (e *ExtractionError) Error() string {
	return fmt.Sprintf("extraction error [%s]: %v", e.Op, e.Err)
}

func (e *ExtractionError) Unwrap() error { return e.Err }

// StorageError indicates a failure during DynamoDB or S3 operations.
type StorageError struct {
	Op  string
	Err error
}

func (e *StorageError) Error() string {
	return fmt.Sprintf("storage error [%s]: %v", e.Op, e.Err)
}

func (e *StorageError) Unwrap() error { return e.Err }

// VerdictError indicates an email failed SES verdict checks.
type VerdictError struct {
	MessageID string
	Verdict   string
	Status    string
}

func (e *VerdictError) Error() string {
	return fmt.Sprintf("verdict failed for message %s: %s=%s", e.MessageID, e.Verdict, e.Status)
}

// DuplicateError indicates a duplicate email was detected.
type DuplicateError struct {
	DedupKey string
}

func (e *DuplicateError) Error() string {
	return fmt.Sprintf("duplicate email: %s", e.DedupKey)
}
