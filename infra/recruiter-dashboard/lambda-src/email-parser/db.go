package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// WriteResult reports the outcome of a single DynamoDB write.
type WriteResult struct {
	// Duplicate is true when the conditional check failed, meaning the item
	// already exists in the table and was intentionally skipped.
	Duplicate bool
}

// dynamoClientIface is satisfied by the AWS DynamoDB client and allows
// injection of a test double in unit tests.
type dynamoClientIface interface {
	PutItem(ctx context.Context, params *dynamodb.PutItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error)
}

// tableClient is the lazily-initialised DynamoDB client for the configured
// table. It is reused across warm Lambda invocations to reduce cold-start
// overhead.
var (
	tableOnce   sync.Once
	tableClient dynamoClientIface
	tableInitErr error
)

// initTableClient initialises the DynamoDB client exactly once per process.
func initTableClient(ctx context.Context) error {
	tableOnce.Do(func() {
		cfg, err := config.LoadDefaultConfig(ctx)
		if err != nil {
			tableInitErr = fmt.Errorf("load AWS config: %w", err)
			return
		}
		tableClient = dynamodb.NewFromConfig(cfg)
	})
	return tableInitErr
}

// WriteRecruiterEmail persists r to DynamoDB, skipping duplicates.
//
// A conditional expression ensures the write is skipped when an item with the
// same dedup_key already exists, making the operation idempotent.
//
//   - Returns WriteResult{Duplicate: true} if the item already exists.
//   - Returns an error for any other failure.
func WriteRecruiterEmail(ctx context.Context, r *RecruiterEmail) (*WriteResult, error) {
	// Allow test injection: only initialise if a client hasn't been set already.
	if tableClient == nil {
		if err := initTableClient(ctx); err != nil {
			return nil, err
		}
	}
	return writeRecruiterEmailWithClient(ctx, tableClient, r)
}

// writeRecruiterEmailWithClient is the testable core; it accepts an injected client.
func writeRecruiterEmailWithClient(ctx context.Context, client dynamoClientIface, r *RecruiterEmail) (*WriteResult, error) {
	tableName := os.Getenv("RECRUITER_TABLE")
	if tableName == "" {
		return nil, errors.New("RECRUITER_TABLE environment variable is not set")
	}

	item, err := r.ToDynamoDBItem()
	if err != nil {
		return nil, fmt.Errorf("marshal DynamoDB item: %w", err)
	}

	_, err = client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(tableName),
		Item:      item,
		// Skip the write when an item with the same dedup_key already exists.
		ConditionExpression: aws.String(
			"attribute_not_exists(dedup_key) OR dedup_key <> :dk",
		),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":dk": &types.AttributeValueMemberS{Value: r.DedupKey},
		},
	})
	if err != nil {
		var ccf *types.ConditionalCheckFailedException
		if errors.As(err, &ccf) {
			log.Printf("[WARN] duplicate email skipped: dedup_key=%s", r.DedupKey)
			return &WriteResult{Duplicate: true}, nil
		}
		return nil, fmt.Errorf("DynamoDB PutItem: %w", err)
	}

	return &WriteResult{}, nil
}
