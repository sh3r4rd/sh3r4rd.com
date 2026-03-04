package ssm

import (
	"context"
	"fmt"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
)

// SSMClient defines the interface for SSM operations (enables testing with mocks).
type SSMClient interface {
	GetParameter(ctx context.Context, params *ssm.GetParameterInput, optFns ...func(*ssm.Options)) (*ssm.GetParameterOutput, error)
}

// ParameterFetcher retrieves and caches SSM parameters.
type ParameterFetcher struct {
	client SSMClient
	mu     sync.RWMutex
	cache  map[string]string
}

// NewParameterFetcher creates a new ParameterFetcher with the given SSM client.
func NewParameterFetcher(client SSMClient) *ParameterFetcher {
	return &ParameterFetcher{
		client: client,
		cache:  make(map[string]string),
	}
}

// GetSecureParameter fetches a SecureString parameter from SSM.
// The value is cached in memory for reuse across warm Lambda invocations.
func (f *ParameterFetcher) GetSecureParameter(ctx context.Context, name string) (string, error) {
	// Check cache first (read lock)
	f.mu.RLock()
	if val, ok := f.cache[name]; ok {
		f.mu.RUnlock()
		return val, nil
	}
	f.mu.RUnlock()

	// Fetch from SSM (write lock)
	f.mu.Lock()
	defer f.mu.Unlock()

	// Double-check after acquiring write lock
	if val, ok := f.cache[name]; ok {
		return val, nil
	}

	output, err := f.client.GetParameter(ctx, &ssm.GetParameterInput{
		Name:           aws.String(name),
		WithDecryption: aws.Bool(true),
	})
	if err != nil {
		return "", fmt.Errorf("failed to fetch SSM parameter %q: %w", name, err)
	}

	if output.Parameter == nil || output.Parameter.Value == nil {
		return "", fmt.Errorf("SSM parameter %q has no value", name)
	}

	val := *output.Parameter.Value
	f.cache[name] = val
	return val, nil
}
