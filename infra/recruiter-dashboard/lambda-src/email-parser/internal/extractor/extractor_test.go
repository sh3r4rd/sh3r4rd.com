package extractor

import (
	"context"
	"fmt"
	"testing"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

// mockOpenAIClient implements OpenAIClient for testing.
type mockOpenAIClient struct {
	newFn     func(ctx context.Context, body openai.ChatCompletionNewParams, opts ...option.RequestOption) (*openai.ChatCompletion, error)
	callCount int
}

func (m *mockOpenAIClient) New(ctx context.Context, body openai.ChatCompletionNewParams, opts ...option.RequestOption) (*openai.ChatCompletion, error) {
	m.callCount++
	return m.newFn(ctx, body, opts...)
}

func TestExtract_Success(t *testing.T) {
	mock := &mockOpenAIClient{
		newFn: func(ctx context.Context, body openai.ChatCompletionNewParams, opts ...option.RequestOption) (*openai.ChatCompletion, error) {
			return &openai.ChatCompletion{
				Choices: []openai.ChatCompletionChoice{
					{
						Message: openai.ChatCompletionMessage{
							Content: `{"recruiter_name":"Jane Smith","recruiter_email":"jane@google.com","company":"Google","job_title":"Senior Software Engineer","phone":"+16502530000","confidence":0.95}`,
						},
						FinishReason: "stop",
					},
				},
			}, nil
		},
	}

	ext := NewExtractor(mock)
	result := ext.Extract(context.Background(), "Test email body")

	if result.RecruiterName != "Jane Smith" {
		t.Errorf("expected RecruiterName=Jane Smith, got %s", result.RecruiterName)
	}
	if result.RecruiterEmail != "jane@google.com" {
		t.Errorf("expected RecruiterEmail=jane@google.com, got %s", result.RecruiterEmail)
	}
	if result.Company != "Google" {
		t.Errorf("expected Company=Google, got %s", result.Company)
	}
	if result.JobTitle != "Senior Software Engineer" {
		t.Errorf("expected JobTitle=Senior Software Engineer, got %s", result.JobTitle)
	}
	if result.Confidence != 0.95 {
		t.Errorf("expected Confidence=0.95, got %f", result.Confidence)
	}
	if mock.callCount != 1 {
		t.Errorf("expected 1 API call, got %d", mock.callCount)
	}
}

func TestExtract_RetryThenSuccess(t *testing.T) {
	attempt := 0
	mock := &mockOpenAIClient{
		newFn: func(ctx context.Context, body openai.ChatCompletionNewParams, opts ...option.RequestOption) (*openai.ChatCompletion, error) {
			attempt++
			if attempt < 3 {
				return nil, fmt.Errorf("rate limited")
			}
			return &openai.ChatCompletion{
				Choices: []openai.ChatCompletionChoice{
					{
						Message: openai.ChatCompletionMessage{
							Content: `{"recruiter_name":"Bob","recruiter_email":"bob@test.com","company":"TestCorp","job_title":"Engineer","phone":"","confidence":0.8}`,
						},
						FinishReason: "stop",
					},
				},
			}, nil
		},
	}

	ext := NewExtractor(mock)
	result := ext.Extract(context.Background(), "Test email")

	if result.RecruiterName != "Bob" {
		t.Errorf("expected RecruiterName=Bob after retry, got %s", result.RecruiterName)
	}
	if mock.callCount != 3 {
		t.Errorf("expected 3 API calls (2 failures + 1 success), got %d", mock.callCount)
	}
}

func TestExtract_RetryExhaustion(t *testing.T) {
	mock := &mockOpenAIClient{
		newFn: func(ctx context.Context, body openai.ChatCompletionNewParams, opts ...option.RequestOption) (*openai.ChatCompletion, error) {
			return nil, fmt.Errorf("service unavailable")
		},
	}

	ext := NewExtractor(mock)
	result := ext.Extract(context.Background(), "Test email")

	if result.RecruiterName != "Unknown" {
		t.Errorf("expected RecruiterName=Unknown on exhaustion, got %s", result.RecruiterName)
	}
	if result.Confidence != 0.0 {
		t.Errorf("expected Confidence=0.0 on exhaustion, got %f", result.Confidence)
	}
	if mock.callCount != 3 {
		t.Errorf("expected 3 API calls (all failed), got %d", mock.callCount)
	}
}

func TestExtract_NonRetryableFinishReason(t *testing.T) {
	mock := &mockOpenAIClient{
		newFn: func(ctx context.Context, body openai.ChatCompletionNewParams, opts ...option.RequestOption) (*openai.ChatCompletion, error) {
			return &openai.ChatCompletion{
				Choices: []openai.ChatCompletionChoice{
					{
						Message: openai.ChatCompletionMessage{
							Content: `{"truncated": true`,
						},
						FinishReason: "content_filter",
					},
				},
			}, nil
		},
	}

	ext := NewExtractor(mock)
	result := ext.Extract(context.Background(), "Test email")

	// Should break immediately on content_filter, return Unknown
	if result.RecruiterName != "Unknown" {
		t.Errorf("expected Unknown on content_filter, got %s", result.RecruiterName)
	}
	if mock.callCount != 1 {
		t.Errorf("expected 1 API call (no retry on content_filter), got %d", mock.callCount)
	}
}

func TestExtract_EmptyChoices(t *testing.T) {
	mock := &mockOpenAIClient{
		newFn: func(ctx context.Context, body openai.ChatCompletionNewParams, opts ...option.RequestOption) (*openai.ChatCompletion, error) {
			return &openai.ChatCompletion{
				Choices: []openai.ChatCompletionChoice{},
			}, nil
		},
	}

	ext := NewExtractor(mock)
	result := ext.Extract(context.Background(), "Test email")

	if result.RecruiterName != "Unknown" {
		t.Errorf("expected Unknown on empty choices, got %s", result.RecruiterName)
	}
}

func TestExtract_InvalidJSON(t *testing.T) {
	mock := &mockOpenAIClient{
		newFn: func(ctx context.Context, body openai.ChatCompletionNewParams, opts ...option.RequestOption) (*openai.ChatCompletion, error) {
			return &openai.ChatCompletion{
				Choices: []openai.ChatCompletionChoice{
					{
						Message: openai.ChatCompletionMessage{
							Content: `not valid json`,
						},
						FinishReason: "stop",
					},
				},
			}, nil
		},
	}

	ext := NewExtractor(mock)
	result := ext.Extract(context.Background(), "Test email")

	// All 3 attempts should fail parsing, then return Unknown
	if result.RecruiterName != "Unknown" {
		t.Errorf("expected Unknown on invalid JSON, got %s", result.RecruiterName)
	}
	if mock.callCount != 3 {
		t.Errorf("expected 3 API calls (all parse failures), got %d", mock.callCount)
	}
}
