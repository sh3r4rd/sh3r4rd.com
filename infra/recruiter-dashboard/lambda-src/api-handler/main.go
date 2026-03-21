// API handler Lambda — serves the recruiter dashboard REST API with anonymized responses.
package main

import (
	"context"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

var h *Handler

func init() {
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		log.Fatalf("unable to load AWS config: %v", err)
	}

	client := dynamodb.NewFromConfig(cfg)

	tableName := os.Getenv("RECRUITER_TABLE")
	if tableName == "" {
		log.Fatal("RECRUITER_TABLE environment variable is required")
	}
	corsOrigin := os.Getenv("CORS_ALLOW_ORIGIN")
	if corsOrigin == "" {
		log.Fatal("CORS_ALLOW_ORIGIN environment variable is required")
	}
	dateIndexName := os.Getenv("DATE_INDEX_NAME")
	if dateIndexName == "" {
		log.Fatal("DATE_INDEX_NAME environment variable is required")
	}

	h = &Handler{
		db:            client,
		tableName:     tableName,
		corsOrigin:    corsOrigin,
		dateIndexName: dateIndexName,
	}
}

func handleRequest(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	return h.Handle(ctx, request)
}

func main() {
	lambda.Start(handleRequest)
}
