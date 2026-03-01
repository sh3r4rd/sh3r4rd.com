// Email parser Lambda — placeholder for Phase 2 implementation.
package main

import (
	"context"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

func handler(ctx context.Context, event events.S3Event) error {
	// Process SES email events from S3.
	return nil
}

func main() {
	lambda.Start(handler)
}
