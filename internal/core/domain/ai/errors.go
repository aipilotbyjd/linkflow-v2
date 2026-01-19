package ai

import (
	"errors"
	"fmt"
)

// Common AI errors
var (
	ErrProviderNotFound     = errors.New("ai provider not found")
	ErrProviderUnavailable  = errors.New("ai provider unavailable")
	ErrModelNotFound        = errors.New("ai model not found")
	ErrModelNotSupported    = errors.New("model not supported by provider")
	ErrInvalidAPIKey        = errors.New("invalid API key")
	ErrRateLimited          = errors.New("rate limited by provider")
	ErrContextLengthExceeded = errors.New("context length exceeded")
	ErrContentFiltered      = errors.New("content filtered by provider")
	ErrInvalidRequest       = errors.New("invalid request")
	ErrStreamingNotSupported = errors.New("streaming not supported")
	ErrToolsNotSupported    = errors.New("tools not supported by model")
	ErrVisionNotSupported   = errors.New("vision not supported by model")
	ErrTimeout              = errors.New("request timed out")
	ErrCacheNotFound        = errors.New("cache entry not found")
	ErrTemplateNotFound     = errors.New("prompt template not found")
	ErrTemplateRenderFailed = errors.New("failed to render prompt template")
)

// ProviderError represents an error from an AI provider
type ProviderError struct {
	Provider   Provider
	StatusCode int
	Code       string
	Message    string
	Retryable  bool
	Err        error
}

func (e *ProviderError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s error (code: %s, status: %d): %s: %v",
			e.Provider, e.Code, e.StatusCode, e.Message, e.Err)
	}
	return fmt.Sprintf("%s error (code: %s, status: %d): %s",
		e.Provider, e.Code, e.StatusCode, e.Message)
}

func (e *ProviderError) Unwrap() error {
	return e.Err
}

// IsRetryable returns true if the error is retryable
func (e *ProviderError) IsRetryable() bool {
	return e.Retryable
}

// NewProviderError creates a new provider error
func NewProviderError(provider Provider, statusCode int, code, message string, retryable bool) *ProviderError {
	return &ProviderError{
		Provider:   provider,
		StatusCode: statusCode,
		Code:       code,
		Message:    message,
		Retryable:  retryable,
	}
}

// ValidationError represents a validation error
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error on field %s: %s", e.Field, e.Message)
}

// NewValidationError creates a new validation error
func NewValidationError(field, message string) *ValidationError {
	return &ValidationError{
		Field:   field,
		Message: message,
	}
}

// IsRateLimited checks if the error is a rate limit error
func IsRateLimited(err error) bool {
	if errors.Is(err, ErrRateLimited) {
		return true
	}
	var providerErr *ProviderError
	if errors.As(err, &providerErr) {
		return providerErr.StatusCode == 429
	}
	return false
}

// IsContextLengthExceeded checks if the error is a context length error
func IsContextLengthExceeded(err error) bool {
	return errors.Is(err, ErrContextLengthExceeded)
}

// IsRetryableError checks if the error should be retried
func IsRetryableError(err error) bool {
	var providerErr *ProviderError
	if errors.As(err, &providerErr) {
		return providerErr.IsRetryable()
	}
	return false
}
