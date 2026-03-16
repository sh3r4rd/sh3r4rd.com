package db

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	"github.com/sh3r4rd/sh3r4rd.com/infra/recruiter-dashboard/lambda-src/email-parser/internal/models"
)

// mockDynamoDBClient implements DynamoDBClient for testing.
type mockDynamoDBClient struct {
	putItemFn func(ctx context.Context, params *dynamodb.PutItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error)
	lastInput *dynamodb.PutItemInput
}

func (m *mockDynamoDBClient) PutItem(ctx context.Context, params *dynamodb.PutItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error) {
	m.lastInput = params
	return m.putItemFn(ctx, params, optFns...)
}

func newTestEmail() *models.RecruiterEmail {
	return &models.RecruiterEmail{
		ID:             "test-msg-001",
		ReceivedAt:     time.Date(2026, 3, 3, 15, 30, 0, 0, time.UTC),
		FirstName: "Jane",
		LastName:  "Smith",
		Email:     "jane@google.com",
		Company:        "Google",
		JobTitle:       "Senior Engineer",
		Phone:          "+16502530000",
		Subject:        "Senior Engineer at Google",
		Confidence:     0.95,
		S3Bucket:       "test-bucket",
		S3Key:          "incoming/test-msg-001",
		DedupKey:       "abc123",
		DateYear:       "2026",
		DateDay:        "2026-03-03",
	}
}

func TestWriteRecruiterEmail_Success(t *testing.T) {
	mock := &mockDynamoDBClient{
		putItemFn: func(ctx context.Context, params *dynamodb.PutItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error) {
			return &dynamodb.PutItemOutput{}, nil
		},
	}

	store := NewStore(mock, "test-table")
	result, err := store.WriteRecruiterEmail(context.Background(), newTestEmail())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Duplicate {
		t.Error("expected Duplicate=false for new item")
	}

	if *mock.lastInput.TableName != "test-table" {
		t.Errorf("expected table name test-table, got %s", *mock.lastInput.TableName)
	}
	if *mock.lastInput.ConditionExpression != "attribute_not_exists(id) AND attribute_not_exists(received_at)" {
		t.Errorf("expected condition expression attribute_not_exists(id) AND attribute_not_exists(received_at), got %s", *mock.lastInput.ConditionExpression)
	}
}

func TestWriteRecruiterEmail_Duplicate(t *testing.T) {
	mock := &mockDynamoDBClient{
		putItemFn: func(ctx context.Context, params *dynamodb.PutItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error) {
			return nil, &types.ConditionalCheckFailedException{
				Message: aws.String("The conditional request failed"),
			}
		},
	}

	store := NewStore(mock, "test-table")
	result, err := store.WriteRecruiterEmail(context.Background(), newTestEmail())
	if err != nil {
		t.Fatalf("duplicate should not return error, got: %v", err)
	}
	if !result.Duplicate {
		t.Error("expected Duplicate=true for conditional check failure")
	}
}

func TestWriteRecruiterEmail_Error(t *testing.T) {
	mock := &mockDynamoDBClient{
		putItemFn: func(ctx context.Context, params *dynamodb.PutItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error) {
			return nil, fmt.Errorf("DynamoDB service unavailable")
		},
	}

	store := NewStore(mock, "test-table")
	_, err := store.WriteRecruiterEmail(context.Background(), newTestEmail())
	if err == nil {
		t.Fatal("expected error for DynamoDB failure")
	}
}

func TestWriteRecruiterEmail_ItemContainsAllFields(t *testing.T) {
	mock := &mockDynamoDBClient{
		putItemFn: func(ctx context.Context, params *dynamodb.PutItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error) {
			return &dynamodb.PutItemOutput{}, nil
		},
	}

	store := NewStore(mock, "test-table")
	email := newTestEmail()
	_, err := store.WriteRecruiterEmail(context.Background(), email)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	item := mock.lastInput.Item
	requiredKeys := []string{"id", "received_at", "first_name", "last_name", "recruiter_email", "company",
		"job_title", "phone", "subject", "confidence", "s3_bucket", "s3_key", "dedup_key", "date_year", "date_day"}

	for _, key := range requiredKeys {
		if _, ok := item[key]; !ok {
			t.Errorf("missing required DynamoDB attribute: %s", key)
		}
	}
}
