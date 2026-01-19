package gateway

import (
	"context"
	"time"

	"github.com/linkflow-ai/linkflow/internal/core/domain/ai"
)

// FallbackChain handles retries and fallbacks to other providers
type FallbackChain struct {
	maxRetries int
	retryDelay time.Duration
}

// NewFallbackChain creates a new fallback chain
func NewFallbackChain(maxRetries int, retryDelay time.Duration) *FallbackChain {
	if maxRetries <= 0 {
		maxRetries = 3
	}
	if retryDelay <= 0 {
		retryDelay = 1 * time.Second
	}

	return &FallbackChain{
		maxRetries: maxRetries,
		retryDelay: retryDelay,
	}
}

// Execute executes a request with retries and fallbacks
func (f *FallbackChain) Execute(
	ctx context.Context,
	primary func(ctx context.Context) (*ai.ChatResponse, error),
	getFallbacks func() []ai.ProviderAdapter,
) (*ai.ChatResponse, error) {
	var lastErr error

	// Try primary provider with retries
	for attempt := 0; attempt < f.maxRetries; attempt++ {
		resp, err := primary(ctx)
		if err == nil {
			return resp, nil
		}

		lastErr = err

		// Don't retry on non-retryable errors
		if !ai.IsRetryableError(err) {
			break
		}

		// Wait before retry (with exponential backoff)
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(f.retryDelay * time.Duration(attempt+1)):
		}
	}

	// Try fallback providers
	fallbacks := getFallbacks()
	for _, fallback := range fallbacks {
		// We need a way to call the fallback with a ChatRequest
		// This is simplified - in practice we'd need the request
		_ = fallback // Fallback logic would go here
	}

	return nil, lastErr
}

// ExecuteWithFallbacks executes with explicit fallback adapters
func (f *FallbackChain) ExecuteWithFallbacks(
	ctx context.Context,
	req *ai.ChatRequest,
	primary ai.ProviderAdapter,
	fallbacks []ai.ProviderAdapter,
) (*ai.ChatResponse, error) {
	var lastErr error

	// Try primary provider with retries
	for attempt := 0; attempt < f.maxRetries; attempt++ {
		resp, err := primary.Chat(ctx, req)
		if err == nil {
			return resp, nil
		}

		lastErr = err

		// Don't retry on non-retryable errors
		if !ai.IsRetryableError(err) && !ai.IsRateLimited(err) {
			break
		}

		// Wait before retry
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(f.retryDelay * time.Duration(attempt+1)):
		}
	}

	// Try fallback providers
	for _, adapter := range fallbacks {
		// Check if the fallback supports the capability
		if !adapter.SupportsCapability(ai.CapabilityChat) {
			continue
		}

		// Map model to equivalent in fallback provider
		mappedReq := f.mapRequestToProvider(req, adapter.Provider())

		resp, err := adapter.Chat(ctx, mappedReq)
		if err == nil {
			return resp, nil
		}

		lastErr = err
	}

	return nil, lastErr
}

// mapRequestToProvider maps a request to an equivalent model in another provider
func (f *FallbackChain) mapRequestToProvider(req *ai.ChatRequest, provider ai.Provider) *ai.ChatRequest {
	// Clone the request
	newReq := *req

	// Map to equivalent model
	newReq.Model = f.getEquivalentModel(req.Model, provider)

	return &newReq
}

// getEquivalentModel returns an equivalent model in another provider
func (f *FallbackChain) getEquivalentModel(model string, provider ai.Provider) string {
	// Define model equivalents
	equivalents := map[string]map[ai.Provider]string{
		// Fast/cheap models
		"gpt-4o-mini": {
			ai.ProviderAnthropic: "claude-3-5-haiku-20241022",
			ai.ProviderGoogle:    "gemini-1.5-flash",
		},
		"claude-3-5-haiku-20241022": {
			ai.ProviderOpenAI: "gpt-4o-mini",
			ai.ProviderGoogle: "gemini-1.5-flash",
		},
		"gemini-1.5-flash": {
			ai.ProviderOpenAI:    "gpt-4o-mini",
			ai.ProviderAnthropic: "claude-3-5-haiku-20241022",
		},

		// Standard models
		"gpt-4o": {
			ai.ProviderAnthropic: "claude-3-5-sonnet-20241022",
			ai.ProviderGoogle:    "gemini-1.5-pro",
		},
		"claude-3-5-sonnet-20241022": {
			ai.ProviderOpenAI: "gpt-4o",
			ai.ProviderGoogle: "gemini-1.5-pro",
		},
		"gemini-1.5-pro": {
			ai.ProviderOpenAI:    "gpt-4o",
			ai.ProviderAnthropic: "claude-3-5-sonnet-20241022",
		},

		// Premium models
		"o1": {
			ai.ProviderAnthropic: "claude-3-opus-20240229",
			ai.ProviderGoogle:    "gemini-1.5-pro",
		},
		"claude-3-opus-20240229": {
			ai.ProviderOpenAI: "o1",
			ai.ProviderGoogle: "gemini-1.5-pro",
		},
	}

	if eq, ok := equivalents[model]; ok {
		if equivalent, ok := eq[provider]; ok {
			return equivalent
		}
	}

	// Default fallback models by provider
	defaults := map[ai.Provider]string{
		ai.ProviderOpenAI:    "gpt-4o-mini",
		ai.ProviderAnthropic: "claude-3-5-haiku-20241022",
		ai.ProviderGoogle:    "gemini-1.5-flash",
	}

	if defaultModel, ok := defaults[provider]; ok {
		return defaultModel
	}

	return model
}
