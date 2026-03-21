package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strconv"
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
// Supports optional filters: ?month=YYYY-MM (date-index GSI query) and/or ?company=X
// (case-insensitive in-memory filtering after DynamoDB read).
func (h *Handler) listRecruiters(ctx context.Context, params map[string]string) (events.APIGatewayProxyResponse, error) {
	company := params["company"]
	month := params["month"]

	var items []map[string]types.AttributeValue
	var err error

	switch {
	case month != "":
		if err := validateMonth(month); err != nil {
			return h.respondError(http.StatusBadRequest, err.Error()), nil
		}
		items, err = h.queryByMonth(ctx, month)
		if err == nil && company != "" {
			items = filterByCompany(items, company)
		}
	case company != "":
		items, err = h.scanAll(ctx)
		if err == nil {
			items = filterByCompany(items, company)
		}
	default:
		items, err = h.scanAll(ctx)
	}

	if err != nil {
		log.Printf("DynamoDB error in listRecruiters: %v", err)
		return h.respondError(http.StatusInternalServerError, "Internal server error"), nil
	}

	sortByReceivedAtDesc(items)
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

// StatsResponse holds the aggregate statistics for the dashboard.
type StatsResponse struct {
	TotalEmails     int            `json:"totalEmails"`
	UniqueCompanies int            `json:"uniqueCompanies"`
	ByMonth         map[string]int `json:"byMonth"`
	TopJobTitles    map[string]int `json:"topJobTitles"`
}

// getStats scans the table with a ProjectionExpression to minimize RCU usage,
// then aggregates statistics in-memory.
func (h *Handler) getStats(ctx context.Context) (events.APIGatewayProxyResponse, error) {
	items, err := h.scanForStats(ctx)
	if err != nil {
		log.Printf("DynamoDB error in getStats: %v", err)
		return h.respondError(http.StatusInternalServerError, "Internal server error"), nil
	}

	companies := make(map[string]struct{})
	byMonth := make(map[string]int)
	jobTitles := make(map[string]int)

	for _, item := range items {
		company := attributeValueString(item, "company", "")
		if company != "" {
			companies[company] = struct{}{}
		}

		dateDay := attributeValueString(item, "date_day", "")
		if len(dateDay) >= 7 {
			month := dateDay[:7]
			byMonth[month]++
		}

		jobTitle := attributeValueString(item, "job_title", "")
		if jobTitle != "" {
			jobTitles[jobTitle]++
		}
	}

	stats := StatsResponse{
		TotalEmails:     len(items),
		UniqueCompanies: len(companies),
		ByMonth:         byMonth,
		TopJobTitles:    topN(jobTitles, 10),
	}

	return h.respond(http.StatusOK, stats), nil
}

// topN returns the n highest-count entries from a frequency map.
func topN(counts map[string]int, n int) map[string]int {
	type kv struct {
		key   string
		count int
	}

	entries := make([]kv, 0, len(counts))
	for k, v := range counts {
		entries = append(entries, kv{k, v})
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].count > entries[j].count
	})

	if n > len(entries) {
		n = len(entries)
	}
	result := make(map[string]int, n)
	for _, e := range entries[:n] {
		result[e.key] = e.count
	}
	return result
}

// scanForStats performs a paginated Scan with ProjectionExpression to fetch only
// the fields needed for aggregation, minimizing data transfer and RCU usage.
func (h *Handler) scanForStats(ctx context.Context) ([]map[string]types.AttributeValue, error) {
	var allItems []map[string]types.AttributeValue
	var exclusiveStartKey map[string]types.AttributeValue

	for {
		input := &dynamodb.ScanInput{
			TableName:            aws.String(h.tableName),
			ProjectionExpression: aws.String("company, job_title, date_day"),
		}
		if exclusiveStartKey != nil {
			input.ExclusiveStartKey = exclusiveStartKey
		}

		out, err := h.db.Scan(ctx, input)
		if err != nil {
			return nil, err
		}

		allItems = append(allItems, out.Items...)

		if out.LastEvaluatedKey == nil {
			break
		}
		exclusiveStartKey = out.LastEvaluatedKey
	}

	return allItems, nil
}

// validateMonth checks that the month parameter is in YYYY-MM format
// with a valid month value (01–12).
func validateMonth(month string) error {
	parts := strings.SplitN(month, "-", 2)
	if len(parts) != 2 || len(parts[0]) != 4 || len(parts[1]) != 2 {
		return fmt.Errorf("invalid month format: %s (expected YYYY-MM)", month)
	}
	monthNum, err := strconv.Atoi(parts[1])
	if err != nil || monthNum < 1 || monthNum > 12 {
		return fmt.Errorf("invalid month value: %s", parts[1])
	}
	return nil
}

// queryByMonth queries the date-index GSI for a specific month (e.g., "2026-02").
func (h *Handler) queryByMonth(ctx context.Context, month string) ([]map[string]types.AttributeValue, error) {
	year := strings.SplitN(month, "-", 2)[0]

	input := &dynamodb.QueryInput{
		TableName:              aws.String(h.tableName),
		IndexName:              aws.String(h.dateIndexName),
		KeyConditionExpression: aws.String("date_year = :year AND begins_with(date_day, :month)"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":year":  &types.AttributeValueMemberS{Value: year},
			":month": &types.AttributeValueMemberS{Value: month},
		},
		ScanIndexForward: aws.Bool(false),
	}

	var items []map[string]types.AttributeValue
	for {
		out, err := h.db.Query(ctx, input)
		if err != nil {
			return nil, err
		}
		items = append(items, out.Items...)
		if out.LastEvaluatedKey == nil {
			break
		}
		input.ExclusiveStartKey = out.LastEvaluatedKey
	}
	return items, nil
}

// scanAll performs a full table scan with pagination.
func (h *Handler) scanAll(ctx context.Context) ([]map[string]types.AttributeValue, error) {
	var items []map[string]types.AttributeValue
	var lastKey map[string]types.AttributeValue

	for {
		input := &dynamodb.ScanInput{
			TableName:         aws.String(h.tableName),
			ExclusiveStartKey: lastKey,
		}
		out, err := h.db.Scan(ctx, input)
		if err != nil {
			return nil, err
		}
		items = append(items, out.Items...)
		if out.LastEvaluatedKey == nil {
			break
		}
		lastKey = out.LastEvaluatedKey
	}
	return items, nil
}

// filterByCompany filters items by case-insensitive substring match on the company attribute.
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

// sortByReceivedAtDesc sorts items by received_at in descending order.
func sortByReceivedAtDesc(items []map[string]types.AttributeValue) {
	sort.Slice(items, func(i, j int) bool {
		a := attributeValueString(items[i], "received_at", "")
		b := attributeValueString(items[j], "received_at", "")
		return a > b
	})
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
