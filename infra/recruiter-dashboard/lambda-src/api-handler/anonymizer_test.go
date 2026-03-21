package main

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

func newTestDynamoDBItem() map[string]types.AttributeValue {
	return map[string]types.AttributeValue{
		"id":              &types.AttributeValueMemberS{Value: "msg-001"},
		"received_at":     &types.AttributeValueMemberS{Value: "2026-03-15T10:30:00Z"},
		"first_name":      &types.AttributeValueMemberS{Value: "Jane"},
		"last_name":       &types.AttributeValueMemberS{Value: "Smith"},
		"recruiter_email": &types.AttributeValueMemberS{Value: "jane.smith@google.com"},
		"company":         &types.AttributeValueMemberS{Value: "Google"},
		"job_title":       &types.AttributeValueMemberS{Value: "Senior Engineer"},
		"phone":           &types.AttributeValueMemberS{Value: "+16502530000"},
		"subject":         &types.AttributeValueMemberS{Value: "Senior Engineer at Google"},
		"confidence":      &types.AttributeValueMemberN{Value: "0.95"},
		"s3_bucket":       &types.AttributeValueMemberS{Value: "email-bucket"},
		"s3_key":          &types.AttributeValueMemberS{Value: "incoming/msg-001"},
		"dedup_key":       &types.AttributeValueMemberS{Value: "abc123def456"},
		"date_year":       &types.AttributeValueMemberS{Value: "2026"},
		"date_day":        &types.AttributeValueMemberS{Value: "2026-03-15"},
	}
}

func TestAnonymizeItem_IncludesExpectedFields(t *testing.T) {
	item := newTestDynamoDBItem()
	result := anonymizeItem(item)

	if result.ID != "msg-001" {
		t.Errorf("expected ID msg-001, got %s", result.ID)
	}
	if result.Company != "Google" {
		t.Errorf("expected Company Google, got %s", result.Company)
	}
	if result.JobTitle != "Senior Engineer" {
		t.Errorf("expected JobTitle Senior Engineer, got %s", result.JobTitle)
	}
	if result.Month != "2026-03" {
		t.Errorf("expected Month 2026-03, got %s", result.Month)
	}
	if result.RecruiterLabel != "Recruiter at Google" {
		t.Errorf("expected RecruiterLabel 'Recruiter at Google', got %s", result.RecruiterLabel)
	}
	if result.Confidence != 0.95 {
		t.Errorf("expected Confidence 0.95, got %f", result.Confidence)
	}
}

func TestAnonymizeItem_PIIFieldsAbsentFromJSON(t *testing.T) {
	item := newTestDynamoDBItem()
	result := anonymizeItem(item)

	b, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("JSON marshal error: %v", err)
	}
	jsonStr := string(b)

	piiFields := []struct {
		field string
		value string
	}{
		{"recruiter_email", "jane.smith@google.com"},
		{"first_name", "Jane"},
		{"last_name", "Smith"},
		{"phone", "+16502530000"},
		{"s3_key", "incoming/msg-001"},
		{"s3_bucket", "email-bucket"},
		{"dedup_key", "abc123def456"},
	}

	for _, pii := range piiFields {
		if contains(jsonStr, pii.value) {
			t.Errorf("JSON response must NOT contain PII value %q (field: %s), but found it in: %s", pii.value, pii.field, jsonStr)
		}
	}

	// Also verify the JSON keys are absent
	piiKeys := []string{"recruiterEmail", "firstName", "lastName", "phone", "s3Key", "s3Bucket", "dedupKey", "receivedAt"}
	for _, key := range piiKeys {
		if contains(jsonStr, `"`+key+`"`) {
			t.Errorf("JSON response must NOT contain PII key %q, but found it in: %s", key, jsonStr)
		}
	}
}

func TestAnonymizeItem_DateCoarsenedToMonth(t *testing.T) {
	item := newTestDynamoDBItem()
	result := anonymizeItem(item)

	if result.Month != "2026-03" {
		t.Errorf("expected date coarsened to 2026-03, got %s", result.Month)
	}

	// Verify exact date is not in JSON
	b, _ := json.Marshal(result)
	jsonStr := string(b)
	if contains(jsonStr, "2026-03-15") {
		t.Errorf("JSON response must NOT contain exact date 2026-03-15, but found it")
	}
}

func TestAnonymizeItem_MissingFields(t *testing.T) {
	item := map[string]types.AttributeValue{
		"id": &types.AttributeValueMemberS{Value: "msg-002"},
	}
	result := anonymizeItem(item)

	if result.ID != "msg-002" {
		t.Errorf("expected ID msg-002, got %s", result.ID)
	}
	if result.Company != "Unknown" {
		t.Errorf("expected default Company Unknown, got %s", result.Company)
	}
	if result.JobTitle != "Unknown" {
		t.Errorf("expected default JobTitle Unknown, got %s", result.JobTitle)
	}
	if result.Month != "" {
		t.Errorf("expected empty Month, got %s", result.Month)
	}
	if result.RecruiterLabel != "Recruiter at Unknown" {
		t.Errorf("expected RecruiterLabel 'Recruiter at Unknown', got %s", result.RecruiterLabel)
	}
	if result.Confidence != 0 {
		t.Errorf("expected default Confidence 0, got %f", result.Confidence)
	}
}

func TestAnonymizeItem_ShortDateDay(t *testing.T) {
	item := map[string]types.AttributeValue{
		"id":       &types.AttributeValueMemberS{Value: "msg-003"},
		"company":  &types.AttributeValueMemberS{Value: "Meta"},
		"date_day": &types.AttributeValueMemberS{Value: "2026"},
	}
	result := anonymizeItem(item)

	if result.Month != "" {
		t.Errorf("expected empty Month for short date_day, got %s", result.Month)
	}
}

func TestAnonymizeItems_MultipleItems(t *testing.T) {
	items := []map[string]types.AttributeValue{
		{
			"id":       &types.AttributeValueMemberS{Value: "msg-001"},
			"company":  &types.AttributeValueMemberS{Value: "Google"},
			"date_day": &types.AttributeValueMemberS{Value: "2026-03-15"},
		},
		{
			"id":       &types.AttributeValueMemberS{Value: "msg-002"},
			"company":  &types.AttributeValueMemberS{Value: "Meta"},
			"date_day": &types.AttributeValueMemberS{Value: "2026-02-10"},
		},
	}

	results := anonymizeItems(items)
	if len(results) != 2 {
		t.Fatalf("expected 2 items, got %d", len(results))
	}
	if results[0].ID != "msg-001" {
		t.Errorf("expected first item ID msg-001, got %s", results[0].ID)
	}
	if results[1].Company != "Meta" {
		t.Errorf("expected second item Company Meta, got %s", results[1].Company)
	}
}

func TestAnonymizeItems_EmptySlice(t *testing.T) {
	results := anonymizeItems([]map[string]types.AttributeValue{})
	if len(results) != 0 {
		t.Errorf("expected empty slice, got %d items", len(results))
	}
}

func TestAttributeValueString_StringType(t *testing.T) {
	item := map[string]types.AttributeValue{
		"key": &types.AttributeValueMemberS{Value: "hello"},
	}
	if v := attributeValueString(item, "key", "default"); v != "hello" {
		t.Errorf("expected hello, got %s", v)
	}
}

func TestAttributeValueString_MissingKey(t *testing.T) {
	item := map[string]types.AttributeValue{}
	if v := attributeValueString(item, "key", "default"); v != "default" {
		t.Errorf("expected default, got %s", v)
	}
}

func TestAttributeValueString_WrongType(t *testing.T) {
	item := map[string]types.AttributeValue{
		"key": &types.AttributeValueMemberN{Value: "42"},
	}
	if v := attributeValueString(item, "key", "default"); v != "default" {
		t.Errorf("expected default for non-string type, got %s", v)
	}
}

func TestAttributeValueFloat_NumberType(t *testing.T) {
	item := map[string]types.AttributeValue{
		"key": &types.AttributeValueMemberN{Value: "0.95"},
	}
	if v := attributeValueFloat(item, "key"); v != 0.95 {
		t.Errorf("expected 0.95, got %f", v)
	}
}

func TestAttributeValueFloat_MissingKey(t *testing.T) {
	item := map[string]types.AttributeValue{}
	if v := attributeValueFloat(item, "key"); v != 0 {
		t.Errorf("expected 0 for missing key, got %f", v)
	}
}

func TestAttributeValueFloat_WrongType(t *testing.T) {
	item := map[string]types.AttributeValue{
		"key": &types.AttributeValueMemberS{Value: "not-a-number"},
	}
	if v := attributeValueFloat(item, "key"); v != 0 {
		t.Errorf("expected 0 for wrong type, got %f", v)
	}
}

// contains checks if s contains substr (helper for PII checks).
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}
