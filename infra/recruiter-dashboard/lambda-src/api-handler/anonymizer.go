package main

import (
	"fmt"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// AnonymizedItem represents a recruiter email with PII removed.
// Recruiter identity is shown only as "Recruiter at {Company}".
// Date is coarsened to month/year (e.g., "2026-02").
type AnonymizedItem struct {
	ID             string  `json:"id"`
	Company        string  `json:"company"`
	JobTitle       string  `json:"jobTitle"`
	Month          string  `json:"month"`
	RecruiterLabel string  `json:"recruiterLabel"`
	Confidence     float64 `json:"confidence"`
}

// anonymizeItem transforms a raw DynamoDB item into an anonymized response shape.
// PII fields (recruiter_email, first_name, last_name, phone, s3_key, s3_bucket, dedup_key)
// are never included in the output.
func anonymizeItem(item map[string]types.AttributeValue) AnonymizedItem {
	company := attributeValueString(item, "company", "Unknown")
	month := attributeValueString(item, "date_day", "")
	if len(month) >= 7 {
		month = month[:7] // "2026-03-15" -> "2026-03"
	}

	return AnonymizedItem{
		ID:             attributeValueString(item, "id", ""),
		Company:        company,
		JobTitle:       attributeValueString(item, "job_title", "Unknown"),
		Month:          month,
		RecruiterLabel: fmt.Sprintf("Recruiter at %s", company),
		Confidence:     attributeValueFloat(item, "confidence"),
	}
}

// anonymizeItems transforms a slice of DynamoDB items into anonymized responses.
func anonymizeItems(items []map[string]types.AttributeValue) []AnonymizedItem {
	result := make([]AnonymizedItem, len(items))
	for i, item := range items {
		result[i] = anonymizeItem(item)
	}
	return result
}

// attributeValueString extracts a string value from a DynamoDB attribute map.
func attributeValueString(item map[string]types.AttributeValue, key, defaultVal string) string {
	if val, ok := item[key]; ok {
		if s, ok := val.(*types.AttributeValueMemberS); ok {
			return s.Value
		}
	}
	return defaultVal
}

// attributeValueFloat extracts a numeric value from a DynamoDB attribute map.
func attributeValueFloat(item map[string]types.AttributeValue, key string) float64 {
	if val, ok := item[key]; ok {
		if n, ok := val.(*types.AttributeValueMemberN); ok {
			f, err := strconv.ParseFloat(n.Value, 64)
			if err != nil {
				return 0
			}
			return f
		}
	}
	return 0
}
