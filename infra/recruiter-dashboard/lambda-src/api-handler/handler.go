package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// DynamoDBAPI defines the DynamoDB operations used by the handler.
type DynamoDBAPI interface {
	GetItem(ctx context.Context, params *dynamodb.GetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error)
	Query(ctx context.Context, params *dynamodb.QueryInput, optFns ...func(*dynamodb.Options)) (*dynamodb.QueryOutput, error)
	Scan(ctx context.Context, params *dynamodb.ScanInput, optFns ...func(*dynamodb.Options)) (*dynamodb.ScanOutput, error)
}

// Handler routes API Gateway proxy requests to the appropriate handler function.
type Handler struct {
	db            DynamoDBAPI
	tableName     string
	corsOrigin    string
	dateIndexName string
}

// Handle routes the incoming API Gateway request by resource path and HTTP method.
func (h *Handler) Handle(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	if req.HTTPMethod == http.MethodOptions {
		return h.respond(http.StatusNoContent, nil), nil
	}

	if req.HTTPMethod != http.MethodGet {
		return h.respondError(http.StatusMethodNotAllowed, "Method not allowed"), nil
	}

	switch req.Resource {
	case "/recruiters":
		return h.listRecruiters(ctx, req.QueryStringParameters)
	case "/recruiters/{id}":
		return h.getRecruiter(ctx, req.PathParameters["id"])
	case "/stats":
		return h.getStats(ctx)
	default:
		return h.respondError(http.StatusNotFound, "Not found"), nil
	}
}

// listRecruiters returns a list of anonymized recruiter emails.
// Supports optional filters: ?company=X (Scan with filter) or ?month=YYYY-MM (date-index GSI query).
func (h *Handler) listRecruiters(ctx context.Context, params map[string]string) (events.APIGatewayProxyResponse, error) {
	company := params["company"]
	month := params["month"]

	var items []map[string]types.AttributeValue
	var err error

	switch {
	case month != "":
		items, err = h.queryByMonth(ctx, month)
		if err == nil && company != "" {
			items = filterByCompany(items, company)
		}
	case company != "":
		items, err = h.scanByCompany(ctx, company)
	default:
		items, err = h.scanAll(ctx)
	}

	if err != nil {
		log.Printf("DynamoDB error in listRecruiters: %v", err)
		return h.respondError(http.StatusInternalServerError, "Internal server error"), nil
	}

	return h.respond(http.StatusOK, anonymizeItems(items)), nil
}

// getRecruiter returns a single anonymized recruiter email by ID.
// Since the table uses a composite key (id + received_at), we query by id alone
// and return the first (most recent) result.
func (h *Handler) getRecruiter(ctx context.Context, id string) (events.APIGatewayProxyResponse, error) {
	if id == "" {
		return h.respondError(http.StatusBadRequest, "Missing id parameter"), nil
	}

	out, err := h.db.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(h.tableName),
		KeyConditionExpression: aws.String("id = :id"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":id": &types.AttributeValueMemberS{Value: id},
		},
		ScanIndexForward: aws.Bool(false),
		Limit:            aws.Int32(1),
	})
	if err != nil {
		log.Printf("DynamoDB error in getRecruiter: %v", err)
		return h.respondError(http.StatusInternalServerError, "Internal server error"), nil
	}

	if len(out.Items) == 0 {
		return h.respondError(http.StatusNotFound, "Recruiter not found"), nil
	}

	return h.respond(http.StatusOK, anonymizeItem(out.Items[0])), nil
}

// getStats returns aggregate statistics. Placeholder — implemented in issue #32.
func (h *Handler) getStats(ctx context.Context) (events.APIGatewayProxyResponse, error) {
	return h.respondError(http.StatusNotImplemented, "Not implemented"), nil
}

// queryByMonth queries the date-index GSI for a specific month (e.g., "2026-02").
func (h *Handler) queryByMonth(ctx context.Context, month string) ([]map[string]types.AttributeValue, error) {
	parts := strings.SplitN(month, "-", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid month format: %s", month)
	}
	year := parts[0]

	out, err := h.db.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(h.tableName),
		IndexName:              aws.String(h.dateIndexName),
		KeyConditionExpression: aws.String("date_year = :year AND begins_with(date_day, :month)"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":year":  &types.AttributeValueMemberS{Value: year},
			":month": &types.AttributeValueMemberS{Value: month},
		},
		ScanIndexForward: aws.Bool(false),
	})
	if err != nil {
		return nil, err
	}
	return out.Items, nil
}

// scanByCompany performs a table scan filtered by company name (case-insensitive contains).
func (h *Handler) scanByCompany(ctx context.Context, company string) ([]map[string]types.AttributeValue, error) {
	out, err := h.db.Scan(ctx, &dynamodb.ScanInput{
		TableName:        aws.String(h.tableName),
		FilterExpression: aws.String("contains(company, :company)"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":company": &types.AttributeValueMemberS{Value: company},
		},
		Limit: aws.Int32(100),
	})
	if err != nil {
		return nil, err
	}
	return out.Items, nil
}

// scanAll performs an unfiltered table scan with a limit.
func (h *Handler) scanAll(ctx context.Context) ([]map[string]types.AttributeValue, error) {
	out, err := h.db.Scan(ctx, &dynamodb.ScanInput{
		TableName: aws.String(h.tableName),
		Limit:     aws.Int32(100),
	})
	if err != nil {
		return nil, err
	}
	return out.Items, nil
}

// filterByCompany filters items in-memory by company name (case-insensitive contains).
func filterByCompany(items []map[string]types.AttributeValue, company string) []map[string]types.AttributeValue {
	lowerCompany := strings.ToLower(company)
	var filtered []map[string]types.AttributeValue
	for _, item := range items {
		val := attributeValueString(item, "company", "")
		if strings.Contains(strings.ToLower(val), lowerCompany) {
			filtered = append(filtered, item)
		}
	}
	return filtered
}

// respond builds an API Gateway response with CORS headers and JSON body.
func (h *Handler) respond(statusCode int, body any) events.APIGatewayProxyResponse {
	var jsonBody string
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			log.Printf("JSON marshal error: %v", err)
			return h.respondError(http.StatusInternalServerError, "Internal server error")
		}
		jsonBody = string(b)
	}

	return events.APIGatewayProxyResponse{
		StatusCode: statusCode,
		Headers: map[string]string{
			"Content-Type":                 "application/json",
			"Access-Control-Allow-Origin":  h.corsOrigin,
			"Access-Control-Allow-Methods": "GET, OPTIONS",
			"Access-Control-Allow-Headers": "Content-Type",
		},
		Body: jsonBody,
	}
}

// respondError builds an error response with a JSON error message.
func (h *Handler) respondError(statusCode int, message string) events.APIGatewayProxyResponse {
	return h.respond(statusCode, map[string]string{"error": message})
}
