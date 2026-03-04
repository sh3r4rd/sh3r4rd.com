package ssm

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	ssmtypes "github.com/aws/aws-sdk-go-v2/service/ssm/types"
)

// mockSSMClient implements SSMClient for testing.
type mockSSMClient struct {
	getParameterFn func(ctx context.Context, params *ssm.GetParameterInput, optFns ...func(*ssm.Options)) (*ssm.GetParameterOutput, error)
	callCount      int
}

func (m *mockSSMClient) GetParameter(ctx context.Context, params *ssm.GetParameterInput, optFns ...func(*ssm.Options)) (*ssm.GetParameterOutput, error) {
	m.callCount++
	return m.getParameterFn(ctx, params, optFns...)
}

func TestGetSecureParameter_Success(t *testing.T) {
	mock := &mockSSMClient{
		getParameterFn: func(ctx context.Context, params *ssm.GetParameterInput, optFns ...func(*ssm.Options)) (*ssm.GetParameterOutput, error) {
			if *params.Name != "/test/key" {
				t.Errorf("expected parameter name /test/key, got %s", *params.Name)
			}
			if !*params.WithDecryption {
				t.Error("expected WithDecryption=true")
			}
			return &ssm.GetParameterOutput{
				Parameter: &ssmtypes.Parameter{
					Value: aws.String("test-api-key-12345"),
				},
			}, nil
		},
	}

	fetcher := NewParameterFetcher(mock)
	val, err := fetcher.GetSecureParameter(context.Background(), "/test/key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != "test-api-key-12345" {
		t.Errorf("expected test-api-key-12345, got %s", val)
	}
}

func TestGetSecureParameter_Caching(t *testing.T) {
	mock := &mockSSMClient{
		getParameterFn: func(ctx context.Context, params *ssm.GetParameterInput, optFns ...func(*ssm.Options)) (*ssm.GetParameterOutput, error) {
			return &ssm.GetParameterOutput{
				Parameter: &ssmtypes.Parameter{
					Value: aws.String("cached-key"),
				},
			}, nil
		},
	}

	fetcher := NewParameterFetcher(mock)

	// First call fetches from SSM
	val1, err := fetcher.GetSecureParameter(context.Background(), "/test/cached")
	if err != nil {
		t.Fatalf("first call failed: %v", err)
	}

	// Second call should use cache
	val2, err := fetcher.GetSecureParameter(context.Background(), "/test/cached")
	if err != nil {
		t.Fatalf("second call failed: %v", err)
	}

	if val1 != val2 {
		t.Errorf("cached value mismatch: %s != %s", val1, val2)
	}

	if mock.callCount != 1 {
		t.Errorf("expected 1 SSM API call (cached), got %d", mock.callCount)
	}
}

func TestGetSecureParameter_NotFound(t *testing.T) {
	mock := &mockSSMClient{
		getParameterFn: func(ctx context.Context, params *ssm.GetParameterInput, optFns ...func(*ssm.Options)) (*ssm.GetParameterOutput, error) {
			return nil, &ssmtypes.ParameterNotFound{}
		},
	}

	fetcher := NewParameterFetcher(mock)
	_, err := fetcher.GetSecureParameter(context.Background(), "/missing/key")
	if err == nil {
		t.Fatal("expected error for missing parameter")
	}
}

func TestGetSecureParameter_AccessDenied(t *testing.T) {
	mock := &mockSSMClient{
		getParameterFn: func(ctx context.Context, params *ssm.GetParameterInput, optFns ...func(*ssm.Options)) (*ssm.GetParameterOutput, error) {
			return nil, fmt.Errorf("AccessDeniedException: User is not authorized")
		},
	}

	fetcher := NewParameterFetcher(mock)
	_, err := fetcher.GetSecureParameter(context.Background(), "/restricted/key")
	if err == nil {
		t.Fatal("expected error for access denied")
	}
}

func TestGetSecureParameter_NilValue(t *testing.T) {
	mock := &mockSSMClient{
		getParameterFn: func(ctx context.Context, params *ssm.GetParameterInput, optFns ...func(*ssm.Options)) (*ssm.GetParameterOutput, error) {
			return &ssm.GetParameterOutput{
				Parameter: &ssmtypes.Parameter{
					Value: nil,
				},
			}, nil
		},
	}

	fetcher := NewParameterFetcher(mock)
	_, err := fetcher.GetSecureParameter(context.Background(), "/nil/key")
	if err == nil {
		t.Fatal("expected error for nil parameter value")
	}
}
