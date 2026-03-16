// API handler Lambda — placeholder for Phase 3 implementation.
package main

import (
	"context"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

func handler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	corsOrigin := os.Getenv("CORS_ALLOW_ORIGIN")
	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Headers: map[string]string{
			"Content-Type":                 "application/json",
			"Access-Control-Allow-Origin":  corsOrigin,
			"Access-Control-Allow-Methods": "GET,OPTIONS",
			"Access-Control-Allow-Headers": "Content-Type",
		},
		Body: "{}",
	}, nil
}

func main() {
	lambda.Start(handler)
}
