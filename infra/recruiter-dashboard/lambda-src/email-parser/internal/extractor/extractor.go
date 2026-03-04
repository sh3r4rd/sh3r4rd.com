package extractor

import (
	"context"
	"encoding/json"
	"log"
	"math"
	"math/rand/v2"
	"time"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"

	"github.com/sh3r4rd/sh3r4rd.com/infra/recruiter-dashboard/lambda-src/email-parser/internal/models"
)

const (
	modelName   = "gpt-4.1-nano"
	maxRetries  = 3
	baseBackoff = 500 * time.Millisecond
)

var systemPrompt = `You are a data extraction assistant. Extract the following fields from the recruiter email body text provided. If a field cannot be determined, use "Unknown". Set confidence between 0.0 and 1.0 based on how certain you are about the extraction accuracy.

Extract:
- recruiter_name: The full name of the recruiter who sent the email
- recruiter_email: The email address of the recruiter
- company: The company the recruiter works for or is recruiting for
- job_title: The job title or position being discussed
- phone: The recruiter's phone number (if present)
- confidence: Your confidence in the overall extraction accuracy (0.0-1.0)`

// extractionSchema defines the structured output JSON schema for OpenAI.
var extractionSchema = map[string]any{
	"type": "object",
	"properties": map[string]any{
		"recruiter_name":  map[string]any{"type": "string"},
		"recruiter_email": map[string]any{"type": "string"},
		"company":         map[string]any{"type": "string"},
		"job_title":       map[string]any{"type": "string"},
		"phone":           map[string]any{"type": "string"},
		"confidence":      map[string]any{"type": "number"},
	},
	"required":             []string{"recruiter_name", "recruiter_email", "company", "job_title", "phone", "confidence"},
	"additionalProperties": false,
}

// OpenAIClient defines the interface for OpenAI chat completion calls (enables testing with mocks).
type OpenAIClient interface {
	New(ctx context.Context, body openai.ChatCompletionNewParams, opts ...option.RequestOption) (*openai.ChatCompletion, error)
}

// Extractor handles recruiter data extraction using OpenAI structured outputs.
type Extractor struct {
	client OpenAIClient
}

// NewExtractor creates an Extractor with the given OpenAI client.
func NewExtractor(client OpenAIClient) *Extractor {
	return &Extractor{client: client}
}

// NewOpenAIClient creates an OpenAI client configured with the given API key.
func NewOpenAIClient(apiKey string) OpenAIClient {
	client := openai.NewClient(option.WithAPIKey(apiKey))
	return &client.Chat.Completions
}

// Extract sends sanitized email body text to OpenAI and returns the extracted recruiter data.
// Retries up to 3 times with exponential backoff + jitter on transient failures.
// On retry exhaustion, returns an ExtractionResult with all fields set to "Unknown".
func (e *Extractor) Extract(ctx context.Context, emailBody string) models.ExtractionResult {
	params := openai.ChatCompletionNewParams{
		Model: modelName,
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(systemPrompt),
			openai.UserMessage(emailBody),
		},
		ResponseFormat: openai.ChatCompletionNewParamsResponseFormatUnion{
			OfJSONSchema: &openai.ResponseFormatJSONSchemaParam{
				JSONSchema: openai.ResponseFormatJSONSchemaJSONSchemaParam{
					Name:   "recruiter_extraction",
					Strict: openai.Bool(true),
					Schema: extractionSchema,
				},
			},
		},
	}

	for attempt := range maxRetries {
		completion, err := e.client.New(ctx, params)
		if err != nil {
			log.Printf("OpenAI API attempt %d/%d failed: %v", attempt+1, maxRetries, err)
			sleepBackoff(attempt)
			continue
		}

		if len(completion.Choices) == 0 {
			log.Printf("OpenAI API attempt %d/%d: no choices returned", attempt+1, maxRetries)
			sleepBackoff(attempt)
			continue
		}

		choice := completion.Choices[0]

		// Check for non-retryable finish reasons
		finishReason := string(choice.FinishReason)
		if finishReason != "" && finishReason != "stop" {
			log.Printf("OpenAI API attempt %d/%d: non-retryable finish_reason %q", attempt+1, maxRetries, finishReason)
			break
		}

		content := choice.Message.Content
		var result models.ExtractionResult
		if err := json.Unmarshal([]byte(content), &result); err != nil {
			log.Printf("OpenAI API attempt %d/%d: failed to parse response: %v", attempt+1, maxRetries, err)
			sleepBackoff(attempt)
			continue
		}

		return result
	}

	log.Printf("OpenAI API: all %d attempts exhausted, returning Unknown result", maxRetries)
	return models.UnknownResult()
}

// sleepBackoff applies exponential backoff with jitter before the next retry attempt.
func sleepBackoff(attempt int) {
	if attempt >= maxRetries-1 {
		return // No sleep after last attempt
	}
	backoff := time.Duration(math.Pow(2, float64(attempt))) * baseBackoff
	jitter := time.Duration(rand.Int64N(int64(backoff / 2)))
	time.Sleep(backoff + jitter)
}

// ExtractWithKey creates a temporary OpenAI client with the given API key and extracts data.
func ExtractWithKey(ctx context.Context, apiKey, emailBody string) models.ExtractionResult {
	client := NewOpenAIClient(apiKey)
	ext := NewExtractor(client)
	return ext.Extract(ctx, emailBody)
}
