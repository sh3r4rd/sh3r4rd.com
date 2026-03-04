package db

import (
	"context"
	"errors"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	"github.com/sh3r4rd/sh3r4rd.com/infra/recruiter-dashboard/lambda-src/email-parser/internal/models"
)

// DynamoDBClient defines the interface for DynamoDB operations (enables testing with mocks).
type DynamoDBClient interface {
	PutItem(ctx context.Context, params *dynamodb.PutItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error)
}

// WriteResult holds the outcome of a DynamoDB write operation.
type WriteResult struct {
	Duplicate bool
}

// Store handles DynamoDB persistence for recruiter emails.
type Store struct {
	client    DynamoDBClient
	tableName string
}

// NewStore creates a new DynamoDB store.
func NewStore(client DynamoDBClient, tableName string) *Store {
	return &Store{
		client:    client,
		tableName: tableName,
	}
}

// WriteRecruiterEmail writes a parsed recruiter email to DynamoDB.
// Uses a conditional expression on the primary key to ensure idempotent writes
// (prevents re-processing the same SES message on retry/redelivery).
// Returns WriteResult with Duplicate=true if the item already exists (not an error).
func (s *Store) WriteRecruiterEmail(ctx context.Context, email *models.RecruiterEmail) (WriteResult, error) {
	item := email.ToDynamoDBItem()

	_, err := s.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName:           aws.String(s.tableName),
		Item:                item,
		ConditionExpression: aws.String("attribute_not_exists(id)"),
	})
	if err != nil {
		var condErr *types.ConditionalCheckFailedException
		if errors.As(err, &condErr) {
			log.Printf("Duplicate email detected (dedup_key: %s), skipping", email.DedupKey)
			return WriteResult{Duplicate: true}, nil
		}
		return WriteResult{}, err
	}

	return WriteResult{Duplicate: false}, nil
}
