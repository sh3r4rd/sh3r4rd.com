package models

import (
	"crypto/sha256"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// RecruiterEmail represents a parsed recruiter email ready for DynamoDB storage.
type RecruiterEmail struct {
	ID         string    `json:"id"`
	ReceivedAt time.Time `json:"received_at"`
	FirstName  string    `json:"first_name"`
	LastName   string    `json:"last_name"`
	Email      string    `json:"email"`
	Company    string    `json:"company"`
	JobTitle   string    `json:"job_title"`
	Phone      string    `json:"phone"`
	Subject    string    `json:"subject"`
	Confidence float64   `json:"confidence"`
	S3Bucket   string    `json:"s3_bucket"`
	S3Key      string    `json:"s3_key"`
	DedupKey   string    `json:"dedup_key"`
	DateYear   string    `json:"date_year"`
	DateDay    string    `json:"date_day"`
}

// GenerateDedupKey produces a deterministic dedup key from recruiter email + job title.
// Uses SHA-256 hash to avoid special characters in DynamoDB key expressions.
func GenerateDedupKey(recruiterEmail, jobTitle string) string {
	normalizedEmail := strings.ToLower(strings.TrimSpace(recruiterEmail))
	normalizedTitle := strings.ToLower(strings.TrimSpace(jobTitle))
	hash := sha256.Sum256([]byte(normalizedEmail + "|" + normalizedTitle))
	return fmt.Sprintf("%x", hash[:16])
}

// ComputeDedupKey generates and populates the DedupKey field from Email and JobTitle.
func (r *RecruiterEmail) ComputeDedupKey() string {
	r.DedupKey = GenerateDedupKey(r.Email, r.JobTitle)
	return r.DedupKey
}

// ToDynamoDBItem converts the RecruiterEmail to a DynamoDB attribute value map.
func (r *RecruiterEmail) ToDynamoDBItem() map[string]types.AttributeValue {
	item := map[string]types.AttributeValue{
		"id":          &types.AttributeValueMemberS{Value: r.ID},
		"received_at": &types.AttributeValueMemberS{Value: r.ReceivedAt.Format(time.RFC3339)},
		"first_name":  &types.AttributeValueMemberS{Value: r.FirstName},
		"last_name":   &types.AttributeValueMemberS{Value: r.LastName},
		"recruiter_email": &types.AttributeValueMemberS{Value: r.Email},
		"company":     &types.AttributeValueMemberS{Value: r.Company},
		"job_title":   &types.AttributeValueMemberS{Value: r.JobTitle},
		"phone":       &types.AttributeValueMemberS{Value: r.Phone},
		"subject":     &types.AttributeValueMemberS{Value: r.Subject},
		"confidence":  &types.AttributeValueMemberN{Value: fmt.Sprintf("%.2f", r.Confidence)},
		"s3_bucket":   &types.AttributeValueMemberS{Value: r.S3Bucket},
		"s3_key":      &types.AttributeValueMemberS{Value: r.S3Key},
		"dedup_key":   &types.AttributeValueMemberS{Value: r.DedupKey},
		"date_year":   &types.AttributeValueMemberS{Value: r.DateYear},
		"date_day":    &types.AttributeValueMemberS{Value: r.DateDay},
	}
	return item
}

// ParsedEmail holds the intermediate result of MIME parsing before extraction.
type ParsedEmail struct {
	From        string
	To          string
	Subject     string
	Date        time.Time
	Body        string
	HTMLBody    string
	IsForwarded bool
}
