package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// mockDynamoDB implements DynamoDBAPI for testing.
type mockDynamoDB struct {
	getItemFn func(ctx context.Context, params *dynamodb.GetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error)
	queryFn   func(ctx context.Context, params *dynamodb.QueryInput, optFns ...func(*dynamodb.Options)) (*dynamodb.QueryOutput, error)
	scanFn    func(ctx context.Context, params *dynamodb.ScanInput, optFns ...func(*dynamodb.Options)) (*dynamodb.ScanOutput, error)
}

func (m *mockDynamoDB) GetItem(ctx context.Context, params *dynamodb.GetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error) {
	if m.getItemFn != nil {
		return m.getItemFn(ctx, params, optFns...)
	}
	return &dynamodb.GetItemOutput{}, nil
}

func (m *mockDynamoDB) Query(ctx context.Context, params *dynamodb.QueryInput, optFns ...func(*dynamodb.Options)) (*dynamodb.QueryOutput, error) {
	if m.queryFn != nil {
		return m.queryFn(ctx, params, optFns...)
	}
	return &dynamodb.QueryOutput{}, nil
}

func (m *mockDynamoDB) Scan(ctx context.Context, params *dynamodb.ScanInput, optFns ...func(*dynamodb.Options)) (*dynamodb.ScanOutput, error) {
	if m.scanFn != nil {
		return m.scanFn(ctx, params, optFns...)
	}
	return &dynamodb.ScanOutput{}, nil
}

func newTestHandler(mock *mockDynamoDB) *Handler {
	return &Handler{
		db:            mock,
		tableName:     "test-table",
		corsOrigin:    "https://sh3r4rd.com",
		dateIndexName: "date-index",
	}
}

func sampleItems() []map[string]types.AttributeValue {
	return []map[string]types.AttributeValue{
		newTestDynamoDBItem(),
		{
			"id":              &types.AttributeValueMemberS{Value: "msg-002"},
			"received_at":     &types.AttributeValueMemberS{Value: "2026-02-10T08:00:00Z"},
			"first_name":      &types.AttributeValueMemberS{Value: "John"},
			"last_name":       &types.AttributeValueMemberS{Value: "Doe"},
			"recruiter_email": &types.AttributeValueMemberS{Value: "john@meta.com"},
			"company":         &types.AttributeValueMemberS{Value: "Meta"},
			"job_title":       &types.AttributeValueMemberS{Value: "Staff Engineer"},
			"phone":           &types.AttributeValueMemberS{Value: "+14155551234"},
			"subject":         &types.AttributeValueMemberS{Value: "Staff Engineer at Meta"},
			"confidence":      &types.AttributeValueMemberN{Value: "0.87"},
			"s3_bucket":       &types.AttributeValueMemberS{Value: "email-bucket"},
			"s3_key":          &types.AttributeValueMemberS{Value: "incoming/msg-002"},
			"dedup_key":       &types.AttributeValueMemberS{Value: "xyz789"},
			"date_year":       &types.AttributeValueMemberS{Value: "2026"},
			"date_day":        &types.AttributeValueMemberS{Value: "2026-02-10"},
		},
	}
}

// --- CORS Tests ---

func TestCORSHeaders_OnSuccess(t *testing.T) {
	mock := &mockDynamoDB{
		scanFn: func(ctx context.Context, params *dynamodb.ScanInput, optFns ...func(*dynamodb.Options)) (*dynamodb.ScanOutput, error) {
			return &dynamodb.ScanOutput{Items: []map[string]types.AttributeValue{}}, nil
		},
	}
	h := newTestHandler(mock)

	resp, err := h.Handle(context.Background(), events.APIGatewayProxyRequest{
		HTTPMethod: "GET",
		Resource:   "/recruiters",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assertCORSHeaders(t, resp)
}

func TestCORSHeaders_On404(t *testing.T) {
	mock := &mockDynamoDB{
		queryFn: func(ctx context.Context, params *dynamodb.QueryInput, optFns ...func(*dynamodb.Options)) (*dynamodb.QueryOutput, error) {
			return &dynamodb.QueryOutput{Items: []map[string]types.AttributeValue{}}, nil
		},
	}
	h := newTestHandler(mock)

	resp, err := h.Handle(context.Background(), events.APIGatewayProxyRequest{
		HTTPMethod:     "GET",
		Resource:       "/recruiters/{id}",
		PathParameters: map[string]string{"id": "nonexistent"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404, got %d", resp.StatusCode)
	}
	assertCORSHeaders(t, resp)
}

func TestCORSHeaders_On405(t *testing.T) {
	h := newTestHandler(&mockDynamoDB{})

	resp, err := h.Handle(context.Background(), events.APIGatewayProxyRequest{
		HTTPMethod: "POST",
		Resource:   "/recruiters",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", resp.StatusCode)
	}
	assertCORSHeaders(t, resp)
}

func TestCORSHeaders_On500(t *testing.T) {
	mock := &mockDynamoDB{
		scanFn: func(ctx context.Context, params *dynamodb.ScanInput, optFns ...func(*dynamodb.Options)) (*dynamodb.ScanOutput, error) {
			return nil, fmt.Errorf("DynamoDB unavailable")
		},
	}
	h := newTestHandler(mock)

	resp, err := h.Handle(context.Background(), events.APIGatewayProxyRequest{
		HTTPMethod: "GET",
		Resource:   "/recruiters",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", resp.StatusCode)
	}
	assertCORSHeaders(t, resp)
}

func TestCORSHeaders_OnOptions(t *testing.T) {
	h := newTestHandler(&mockDynamoDB{})

	resp, err := h.Handle(context.Background(), events.APIGatewayProxyRequest{
		HTTPMethod: "OPTIONS",
		Resource:   "/recruiters",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("expected 204, got %d", resp.StatusCode)
	}
	assertCORSHeaders(t, resp)
}

// --- Route Tests ---

func TestPOSTReturns405(t *testing.T) {
	h := newTestHandler(&mockDynamoDB{})

	resp, _ := h.Handle(context.Background(), events.APIGatewayProxyRequest{
		HTTPMethod: "POST",
		Resource:   "/recruiters",
	})
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", resp.StatusCode)
	}
}

func TestPUTReturns405(t *testing.T) {
	h := newTestHandler(&mockDynamoDB{})

	resp, _ := h.Handle(context.Background(), events.APIGatewayProxyRequest{
		HTTPMethod: "PUT",
		Resource:   "/recruiters",
	})
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", resp.StatusCode)
	}
}

func TestDELETEReturns405(t *testing.T) {
	h := newTestHandler(&mockDynamoDB{})

	resp, _ := h.Handle(context.Background(), events.APIGatewayProxyRequest{
		HTTPMethod: "DELETE",
		Resource:   "/recruiters",
	})
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", resp.StatusCode)
	}
}

func TestUnknownResourceReturns404(t *testing.T) {
	h := newTestHandler(&mockDynamoDB{})

	resp, _ := h.Handle(context.Background(), events.APIGatewayProxyRequest{
		HTTPMethod: "GET",
		Resource:   "/unknown",
	})
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404, got %d", resp.StatusCode)
	}
}

// --- GET /recruiters Tests ---

func TestListRecruiters_ReturnsAnonymizedData(t *testing.T) {
	mock := &mockDynamoDB{
		scanFn: func(ctx context.Context, params *dynamodb.ScanInput, optFns ...func(*dynamodb.Options)) (*dynamodb.ScanOutput, error) {
			return &dynamodb.ScanOutput{Items: sampleItems()}, nil
		},
	}
	h := newTestHandler(mock)

	resp, err := h.Handle(context.Background(), events.APIGatewayProxyRequest{
		HTTPMethod: "GET",
		Resource:   "/recruiters",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var items []AnonymizedItem
	if err := json.Unmarshal([]byte(resp.Body), &items); err != nil {
		t.Fatalf("JSON unmarshal error: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
	if items[0].RecruiterLabel != "Recruiter at Google" {
		t.Errorf("expected 'Recruiter at Google', got %s", items[0].RecruiterLabel)
	}
}

func TestListRecruiters_ResponseDoesNotContainPII(t *testing.T) {
	mock := &mockDynamoDB{
		scanFn: func(ctx context.Context, params *dynamodb.ScanInput, optFns ...func(*dynamodb.Options)) (*dynamodb.ScanOutput, error) {
			return &dynamodb.ScanOutput{Items: sampleItems()}, nil
		},
	}
	h := newTestHandler(mock)

	resp, _ := h.Handle(context.Background(), events.APIGatewayProxyRequest{
		HTTPMethod: "GET",
		Resource:   "/recruiters",
	})

	piiValues := []string{
		"jane.smith@google.com", "john@meta.com",
		"+16502530000", "+14155551234",
		"incoming/msg-001", "incoming/msg-002",
		"email-bucket",
	}
	for _, pii := range piiValues {
		if contains(resp.Body, pii) {
			t.Errorf("response body must NOT contain PII value %q", pii)
		}
	}
}

func TestListRecruiters_EmptyTable(t *testing.T) {
	mock := &mockDynamoDB{
		scanFn: func(ctx context.Context, params *dynamodb.ScanInput, optFns ...func(*dynamodb.Options)) (*dynamodb.ScanOutput, error) {
			return &dynamodb.ScanOutput{Items: []map[string]types.AttributeValue{}}, nil
		},
	}
	h := newTestHandler(mock)

	resp, _ := h.Handle(context.Background(), events.APIGatewayProxyRequest{
		HTTPMethod: "GET",
		Resource:   "/recruiters",
	})
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}

	var items []AnonymizedItem
	if err := json.Unmarshal([]byte(resp.Body), &items); err != nil {
		t.Fatalf("JSON unmarshal error: %v", err)
	}
	if len(items) != 0 {
		t.Errorf("expected empty list, got %d items", len(items))
	}
}

func TestListRecruiters_DynamoDBError(t *testing.T) {
	mock := &mockDynamoDB{
		scanFn: func(ctx context.Context, params *dynamodb.ScanInput, optFns ...func(*dynamodb.Options)) (*dynamodb.ScanOutput, error) {
			return nil, fmt.Errorf("service unavailable")
		},
	}
	h := newTestHandler(mock)

	resp, _ := h.Handle(context.Background(), events.APIGatewayProxyRequest{
		HTTPMethod: "GET",
		Resource:   "/recruiters",
	})
	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", resp.StatusCode)
	}
}

// --- GET /recruiters?company=X Tests ---

func TestListRecruiters_CompanyFilter(t *testing.T) {
	var capturedInput *dynamodb.ScanInput
	mock := &mockDynamoDB{
		scanFn: func(ctx context.Context, params *dynamodb.ScanInput, optFns ...func(*dynamodb.Options)) (*dynamodb.ScanOutput, error) {
			capturedInput = params
			// Return only matching items since DynamoDB applies the FilterExpression
			return &dynamodb.ScanOutput{Items: sampleItems()[:1]}, nil
		},
	}
	h := newTestHandler(mock)

	resp, _ := h.Handle(context.Background(), events.APIGatewayProxyRequest{
		HTTPMethod:            "GET",
		Resource:              "/recruiters",
		QueryStringParameters: map[string]string{"company": "Google"},
	})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	// Verify FilterExpression was passed to DynamoDB
	if capturedInput == nil {
		t.Fatal("expected Scan to be called for company filter")
	}
	if capturedInput.FilterExpression == nil || *capturedInput.FilterExpression != "contains(company, :company)" {
		t.Errorf("expected FilterExpression 'contains(company, :company)', got %v", capturedInput.FilterExpression)
	}
	companyVal, ok := capturedInput.ExpressionAttributeValues[":company"]
	if !ok {
		t.Fatal("expected :company in ExpressionAttributeValues")
	}
	if sv, ok := companyVal.(*types.AttributeValueMemberS); !ok || sv.Value != "Google" {
		t.Errorf("expected :company value 'Google', got %v", companyVal)
	}

	var items []AnonymizedItem
	if err := json.Unmarshal([]byte(resp.Body), &items); err != nil {
		t.Fatalf("JSON unmarshal error: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item after company filter, got %d", len(items))
	}
	if items[0].Company != "Google" {
		t.Errorf("expected Company Google, got %s", items[0].Company)
	}
}

func TestListRecruiters_CompanyFilter_PassesValueVerbatim(t *testing.T) {
	var capturedInput *dynamodb.ScanInput
	mock := &mockDynamoDB{
		scanFn: func(ctx context.Context, params *dynamodb.ScanInput, optFns ...func(*dynamodb.Options)) (*dynamodb.ScanOutput, error) {
			capturedInput = params
			return &dynamodb.ScanOutput{Items: []map[string]types.AttributeValue{}}, nil
		},
	}
	h := newTestHandler(mock)

	resp, _ := h.Handle(context.Background(), events.APIGatewayProxyRequest{
		HTTPMethod:            "GET",
		Resource:              "/recruiters",
		QueryStringParameters: map[string]string{"company": "google"},
	})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	// DynamoDB contains() is case-sensitive; verify the value is passed verbatim
	if capturedInput == nil {
		t.Fatal("expected Scan to be called")
	}
	companyVal := capturedInput.ExpressionAttributeValues[":company"]
	if sv, ok := companyVal.(*types.AttributeValueMemberS); !ok || sv.Value != "google" {
		t.Errorf("expected :company value 'google' (verbatim), got %v", companyVal)
	}
}

// --- GET /recruiters?month=YYYY-MM Tests ---

func TestListRecruiters_MonthFilter_UsesGSI(t *testing.T) {
	var capturedInput *dynamodb.QueryInput
	mock := &mockDynamoDB{
		queryFn: func(ctx context.Context, params *dynamodb.QueryInput, optFns ...func(*dynamodb.Options)) (*dynamodb.QueryOutput, error) {
			capturedInput = params
			return &dynamodb.QueryOutput{Items: sampleItems()[:1]}, nil
		},
	}
	h := newTestHandler(mock)

	resp, _ := h.Handle(context.Background(), events.APIGatewayProxyRequest{
		HTTPMethod:            "GET",
		Resource:              "/recruiters",
		QueryStringParameters: map[string]string{"month": "2026-03"},
	})
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}

	if capturedInput == nil {
		t.Fatal("expected Query to be called for month filter")
	}
	if *capturedInput.IndexName != "date-index" {
		t.Errorf("expected date-index GSI, got %s", *capturedInput.IndexName)
	}
	// Month-only query should NOT have a FilterExpression
	if capturedInput.FilterExpression != nil {
		t.Errorf("expected no FilterExpression for month-only query, got %s", *capturedInput.FilterExpression)
	}
}

func TestListRecruiters_MonthAndCompanyFilter(t *testing.T) {
	var capturedInput *dynamodb.QueryInput
	mock := &mockDynamoDB{
		queryFn: func(ctx context.Context, params *dynamodb.QueryInput, optFns ...func(*dynamodb.Options)) (*dynamodb.QueryOutput, error) {
			capturedInput = params
			// Return only matching items since DynamoDB applies the FilterExpression
			return &dynamodb.QueryOutput{Items: sampleItems()[:1]}, nil
		},
	}
	h := newTestHandler(mock)

	resp, _ := h.Handle(context.Background(), events.APIGatewayProxyRequest{
		HTTPMethod:            "GET",
		Resource:              "/recruiters",
		QueryStringParameters: map[string]string{"month": "2026-03", "company": "Google"},
	})
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}

	// Verify FilterExpression was passed to the GSI query
	if capturedInput == nil {
		t.Fatal("expected Query to be called for month+company filter")
	}
	if capturedInput.FilterExpression == nil || *capturedInput.FilterExpression != "contains(company, :company)" {
		t.Errorf("expected FilterExpression 'contains(company, :company)', got %v", capturedInput.FilterExpression)
	}
	companyVal, ok := capturedInput.ExpressionAttributeValues[":company"]
	if !ok {
		t.Fatal("expected :company in ExpressionAttributeValues")
	}
	if sv, ok := companyVal.(*types.AttributeValueMemberS); !ok || sv.Value != "Google" {
		t.Errorf("expected :company value 'Google', got %v", companyVal)
	}

	var items []AnonymizedItem
	if err := json.Unmarshal([]byte(resp.Body), &items); err != nil {
		t.Fatalf("JSON unmarshal error: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item after month+company filter, got %d", len(items))
	}
	if items[0].Company != "Google" {
		t.Errorf("expected only Google items after company filter, got %s", items[0].Company)
	}
}

// --- GET /recruiters/{id} Tests ---

func TestGetRecruiter_Success(t *testing.T) {
	mock := &mockDynamoDB{
		queryFn: func(ctx context.Context, params *dynamodb.QueryInput, optFns ...func(*dynamodb.Options)) (*dynamodb.QueryOutput, error) {
			return &dynamodb.QueryOutput{Items: sampleItems()[:1]}, nil
		},
	}
	h := newTestHandler(mock)

	resp, _ := h.Handle(context.Background(), events.APIGatewayProxyRequest{
		HTTPMethod:     "GET",
		Resource:       "/recruiters/{id}",
		PathParameters: map[string]string{"id": "msg-001"},
	})
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}

	var item AnonymizedItem
	if err := json.Unmarshal([]byte(resp.Body), &item); err != nil {
		t.Fatalf("JSON unmarshal error: %v", err)
	}
	if item.ID != "msg-001" {
		t.Errorf("expected ID msg-001, got %s", item.ID)
	}
	if item.RecruiterLabel != "Recruiter at Google" {
		t.Errorf("expected 'Recruiter at Google', got %s", item.RecruiterLabel)
	}
}

func TestGetRecruiter_NotFound(t *testing.T) {
	mock := &mockDynamoDB{
		queryFn: func(ctx context.Context, params *dynamodb.QueryInput, optFns ...func(*dynamodb.Options)) (*dynamodb.QueryOutput, error) {
			return &dynamodb.QueryOutput{Items: []map[string]types.AttributeValue{}}, nil
		},
	}
	h := newTestHandler(mock)

	resp, _ := h.Handle(context.Background(), events.APIGatewayProxyRequest{
		HTTPMethod:     "GET",
		Resource:       "/recruiters/{id}",
		PathParameters: map[string]string{"id": "nonexistent"},
	})
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404, got %d", resp.StatusCode)
	}
}

func TestGetRecruiter_EmptyID(t *testing.T) {
	h := newTestHandler(&mockDynamoDB{})

	resp, _ := h.Handle(context.Background(), events.APIGatewayProxyRequest{
		HTTPMethod:     "GET",
		Resource:       "/recruiters/{id}",
		PathParameters: map[string]string{"id": ""},
	})
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}
}

func TestGetRecruiter_DynamoDBError(t *testing.T) {
	mock := &mockDynamoDB{
		queryFn: func(ctx context.Context, params *dynamodb.QueryInput, optFns ...func(*dynamodb.Options)) (*dynamodb.QueryOutput, error) {
			return nil, fmt.Errorf("DynamoDB unavailable")
		},
	}
	h := newTestHandler(mock)

	resp, _ := h.Handle(context.Background(), events.APIGatewayProxyRequest{
		HTTPMethod:     "GET",
		Resource:       "/recruiters/{id}",
		PathParameters: map[string]string{"id": "msg-001"},
	})
	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", resp.StatusCode)
	}
}

func TestGetRecruiter_ResponseDoesNotContainPII(t *testing.T) {
	mock := &mockDynamoDB{
		queryFn: func(ctx context.Context, params *dynamodb.QueryInput, optFns ...func(*dynamodb.Options)) (*dynamodb.QueryOutput, error) {
			return &dynamodb.QueryOutput{Items: sampleItems()[:1]}, nil
		},
	}
	h := newTestHandler(mock)

	resp, _ := h.Handle(context.Background(), events.APIGatewayProxyRequest{
		HTTPMethod:     "GET",
		Resource:       "/recruiters/{id}",
		PathParameters: map[string]string{"id": "msg-001"},
	})

	piiValues := []string{"jane.smith@google.com", "+16502530000", "incoming/msg-001", "email-bucket", "Jane", "Smith"}
	for _, pii := range piiValues {
		if contains(resp.Body, pii) {
			t.Errorf("response body must NOT contain PII value %q", pii)
		}
	}
}

// --- GET /stats Tests ---

func TestGetStats_Success(t *testing.T) {
	mock := &mockDynamoDB{
		scanFn: func(ctx context.Context, params *dynamodb.ScanInput, optFns ...func(*dynamodb.Options)) (*dynamodb.ScanOutput, error) {
			return &dynamodb.ScanOutput{
				Items: []map[string]types.AttributeValue{
					{
						"company":   &types.AttributeValueMemberS{Value: "Google"},
						"job_title": &types.AttributeValueMemberS{Value: "Senior Engineer"},
						"date_day":  &types.AttributeValueMemberS{Value: "2026-03-15"},
					},
					{
						"company":   &types.AttributeValueMemberS{Value: "Google"},
						"job_title": &types.AttributeValueMemberS{Value: "Staff Engineer"},
						"date_day":  &types.AttributeValueMemberS{Value: "2026-03-10"},
					},
					{
						"company":   &types.AttributeValueMemberS{Value: "Meta"},
						"job_title": &types.AttributeValueMemberS{Value: "Senior Engineer"},
						"date_day":  &types.AttributeValueMemberS{Value: "2026-02-05"},
					},
				},
			}, nil
		},
	}
	h := newTestHandler(mock)

	resp, _ := h.Handle(context.Background(), events.APIGatewayProxyRequest{
		HTTPMethod: "GET",
		Resource:   "/stats",
	})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var stats StatsResponse
	if err := json.Unmarshal([]byte(resp.Body), &stats); err != nil {
		t.Fatalf("JSON unmarshal error: %v", err)
	}
	if stats.TotalEmails != 3 {
		t.Errorf("expected totalEmails 3, got %d", stats.TotalEmails)
	}
	if stats.UniqueCompanies != 2 {
		t.Errorf("expected uniqueCompanies 2, got %d", stats.UniqueCompanies)
	}
	if stats.ByMonth["2026-03"] != 2 {
		t.Errorf("expected byMonth[2026-03]=2, got %d", stats.ByMonth["2026-03"])
	}
	if stats.ByMonth["2026-02"] != 1 {
		t.Errorf("expected byMonth[2026-02]=1, got %d", stats.ByMonth["2026-02"])
	}
	if stats.TopJobTitles["Senior Engineer"] != 2 {
		t.Errorf("expected topJobTitles[Senior Engineer]=2, got %d", stats.TopJobTitles["Senior Engineer"])
	}
	if stats.TopJobTitles["Staff Engineer"] != 1 {
		t.Errorf("expected topJobTitles[Staff Engineer]=1, got %d", stats.TopJobTitles["Staff Engineer"])
	}
}

func TestGetStats_EmptyTable(t *testing.T) {
	mock := &mockDynamoDB{
		scanFn: func(ctx context.Context, params *dynamodb.ScanInput, optFns ...func(*dynamodb.Options)) (*dynamodb.ScanOutput, error) {
			return &dynamodb.ScanOutput{Items: []map[string]types.AttributeValue{}}, nil
		},
	}
	h := newTestHandler(mock)

	resp, _ := h.Handle(context.Background(), events.APIGatewayProxyRequest{
		HTTPMethod: "GET",
		Resource:   "/stats",
	})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var stats StatsResponse
	if err := json.Unmarshal([]byte(resp.Body), &stats); err != nil {
		t.Fatalf("JSON unmarshal error: %v", err)
	}
	if stats.TotalEmails != 0 {
		t.Errorf("expected totalEmails 0, got %d", stats.TotalEmails)
	}
	if stats.UniqueCompanies != 0 {
		t.Errorf("expected uniqueCompanies 0, got %d", stats.UniqueCompanies)
	}
	if len(stats.ByMonth) != 0 {
		t.Errorf("expected empty byMonth, got %v", stats.ByMonth)
	}
	if len(stats.TopJobTitles) != 0 {
		t.Errorf("expected empty topJobTitles, got %v", stats.TopJobTitles)
	}
}

func TestGetStats_UsesProjectionExpression(t *testing.T) {
	var capturedInput *dynamodb.ScanInput
	mock := &mockDynamoDB{
		scanFn: func(ctx context.Context, params *dynamodb.ScanInput, optFns ...func(*dynamodb.Options)) (*dynamodb.ScanOutput, error) {
			capturedInput = params
			return &dynamodb.ScanOutput{Items: []map[string]types.AttributeValue{}}, nil
		},
	}
	h := newTestHandler(mock)

	h.Handle(context.Background(), events.APIGatewayProxyRequest{
		HTTPMethod: "GET",
		Resource:   "/stats",
	})

	if capturedInput == nil {
		t.Fatal("expected Scan to be called")
	}
	if capturedInput.ProjectionExpression == nil {
		t.Fatal("expected ProjectionExpression to be set")
	}
	if *capturedInput.ProjectionExpression != "company, job_title, date_day" {
		t.Errorf("expected ProjectionExpression 'company, job_title, date_day', got %s", *capturedInput.ProjectionExpression)
	}
}

func TestGetStats_NoPIIInResponse(t *testing.T) {
	mock := &mockDynamoDB{
		scanFn: func(ctx context.Context, params *dynamodb.ScanInput, optFns ...func(*dynamodb.Options)) (*dynamodb.ScanOutput, error) {
			return &dynamodb.ScanOutput{
				Items: []map[string]types.AttributeValue{
					{
						"company":   &types.AttributeValueMemberS{Value: "Google"},
						"job_title": &types.AttributeValueMemberS{Value: "Engineer"},
						"date_day":  &types.AttributeValueMemberS{Value: "2026-03-15"},
					},
				},
			}, nil
		},
	}
	h := newTestHandler(mock)

	resp, _ := h.Handle(context.Background(), events.APIGatewayProxyRequest{
		HTTPMethod: "GET",
		Resource:   "/stats",
	})

	piiKeys := []string{"recruiter_email", "first_name", "last_name", "phone", "s3_key", "s3_bucket"}
	for _, key := range piiKeys {
		if contains(resp.Body, key) {
			t.Errorf("stats response must NOT contain PII key %q", key)
		}
	}
}

func TestGetStats_DynamoDBError(t *testing.T) {
	mock := &mockDynamoDB{
		scanFn: func(ctx context.Context, params *dynamodb.ScanInput, optFns ...func(*dynamodb.Options)) (*dynamodb.ScanOutput, error) {
			return nil, fmt.Errorf("DynamoDB unavailable")
		},
	}
	h := newTestHandler(mock)

	resp, _ := h.Handle(context.Background(), events.APIGatewayProxyRequest{
		HTTPMethod: "GET",
		Resource:   "/stats",
	})
	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", resp.StatusCode)
	}
}

func TestGetStats_PaginatedScan(t *testing.T) {
	callCount := 0
	mock := &mockDynamoDB{
		scanFn: func(ctx context.Context, params *dynamodb.ScanInput, optFns ...func(*dynamodb.Options)) (*dynamodb.ScanOutput, error) {
			callCount++
			if callCount == 1 {
				return &dynamodb.ScanOutput{
					Items: []map[string]types.AttributeValue{
						{
							"company":   &types.AttributeValueMemberS{Value: "Google"},
							"job_title": &types.AttributeValueMemberS{Value: "Engineer"},
							"date_day":  &types.AttributeValueMemberS{Value: "2026-03-15"},
						},
					},
					LastEvaluatedKey: map[string]types.AttributeValue{
						"id": &types.AttributeValueMemberS{Value: "cursor"},
					},
				}, nil
			}
			return &dynamodb.ScanOutput{
				Items: []map[string]types.AttributeValue{
					{
						"company":   &types.AttributeValueMemberS{Value: "Meta"},
						"job_title": &types.AttributeValueMemberS{Value: "Engineer"},
						"date_day":  &types.AttributeValueMemberS{Value: "2026-02-10"},
					},
				},
			}, nil
		},
	}
	h := newTestHandler(mock)

	resp, _ := h.Handle(context.Background(), events.APIGatewayProxyRequest{
		HTTPMethod: "GET",
		Resource:   "/stats",
	})

	var stats StatsResponse
	if err := json.Unmarshal([]byte(resp.Body), &stats); err != nil {
		t.Fatalf("JSON unmarshal error: %v", err)
	}

	if callCount != 2 {
		t.Errorf("expected 2 Scan calls for pagination, got %d", callCount)
	}
	if stats.TotalEmails != 2 {
		t.Errorf("expected totalEmails 2 from paginated scan, got %d", stats.TotalEmails)
	}
}

// --- Helper ---

func assertCORSHeaders(t *testing.T, resp events.APIGatewayProxyResponse) {
	t.Helper()

	expected := map[string]string{
		"Access-Control-Allow-Origin":  "https://sh3r4rd.com",
		"Access-Control-Allow-Methods": "GET, OPTIONS",
		"Access-Control-Allow-Headers": "Content-Type",
	}
	for key, val := range expected {
		if resp.Headers[key] != val {
			t.Errorf("expected header %s=%s, got %s", key, val, resp.Headers[key])
		}
	}
}
