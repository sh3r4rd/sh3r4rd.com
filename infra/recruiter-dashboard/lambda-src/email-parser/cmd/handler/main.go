package main

import (
	"context"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	awsssm "github.com/aws/aws-sdk-go-v2/service/ssm"

	"github.com/sh3r4rd/sh3r4rd.com/infra/recruiter-dashboard/lambda-src/email-parser/internal/db"
	"github.com/sh3r4rd/sh3r4rd.com/infra/recruiter-dashboard/lambda-src/email-parser/internal/handler"
	"github.com/sh3r4rd/sh3r4rd.com/infra/recruiter-dashboard/lambda-src/email-parser/internal/ssm"
	"github.com/sh3r4rd/sh3r4rd.com/infra/recruiter-dashboard/lambda-src/email-parser/internal/tagger"
)

// h is initialized once on cold start and reused across warm invocations.
var h *handler.Handler

func init() {
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		log.Fatalf("failed to load AWS config: %v", err)
	}

	tableName := os.Getenv("RECRUITER_TABLE")
	if tableName == "" {
		log.Fatal("RECRUITER_TABLE environment variable is required")
	}

	s3Client := s3.NewFromConfig(cfg)
	dynamoClient := dynamodb.NewFromConfig(cfg)
	ssmClient := awsssm.NewFromConfig(cfg)

	store := db.NewStore(dynamoClient, tableName)
	t := tagger.NewTagger(s3Client)
	ssmFetcher := ssm.NewParameterFetcher(ssmClient)

	h = handler.NewHandler(s3Client, store, t, ssmFetcher)
}

func handleEvent(ctx context.Context, event events.SimpleEmailEvent) (handler.Summary, error) {
	return h.HandleSESEvent(ctx, event)
}

func main() {
	lambda.Start(handleEvent)
}
