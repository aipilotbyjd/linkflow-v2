package resilience

import (
	"context"
	"errors"
	"math"
	"math/rand"
	"time"
)

// RetryConfig holds retry configuration
type RetryConfig struct {
	MaxAttempts     int
	InitialInterval time.Duration
	MaxInterval     time.Duration
	Multiplier      float64
	RandomFactor    float64
}

// DefaultRetryConfig returns default retry configuration
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxAttempts:     3,
		InitialInterval: 100 * time.Millisecond,
		MaxInterval:     10 * time.Second,
		Multiplier:      2.0,
		RandomFactor:    0.1,
	}
}

// Retry executes the given function with exponential backoff retry
func Retry(ctx context.Context, config RetryConfig, fn func() error) error {
	var lastErr error

	for attempt := 0; attempt < config.MaxAttempts; attempt++ {
		if err := ctx.Err(); err != nil {
			return err
		}

		lastErr = fn()
		if lastErr == nil {
			return nil
		}

		// Don't sleep after the last attempt
		if attempt < config.MaxAttempts-1 {
			interval := calculateInterval(config, attempt)
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(interval):
			}
		}
	}

	return lastErr
}

// RetryWithResult executes the given function with exponential backoff retry and returns a result
func RetryWithResult[T any](ctx context.Context, config RetryConfig, fn func() (T, error)) (T, error) {
	var lastErr error
	var result T

	for attempt := 0; attempt < config.MaxAttempts; attempt++ {
		if err := ctx.Err(); err != nil {
			return result, err
		}

		result, lastErr = fn()
		if lastErr == nil {
			return result, nil
		}

		// Don't sleep after the last attempt
		if attempt < config.MaxAttempts-1 {
			interval := calculateInterval(config, attempt)
			select {
			case <-ctx.Done():
				return result, ctx.Err()
			case <-time.After(interval):
			}
		}
	}

	return result, lastErr
}

func calculateInterval(config RetryConfig, attempt int) time.Duration {
	interval := float64(config.InitialInterval) * math.Pow(config.Multiplier, float64(attempt))

	// Add jitter
	if config.RandomFactor > 0 {
		delta := config.RandomFactor * interval
		minInterval := interval - delta
		maxInterval := interval + delta
		interval = minInterval + (rand.Float64() * (maxInterval - minInterval))
	}

	// Cap at max interval
	if interval > float64(config.MaxInterval) {
		interval = float64(config.MaxInterval)
	}

	return time.Duration(interval)
}

// IsRetryable checks if an error should be retried
func IsRetryable(err error) bool {
	if err == nil {
		return false
	}

	// Check for specific non-retryable errors
	var nonRetryable *NonRetryableError
	if errors.As(err, &nonRetryable) {
		return false
	}

	// Check for context cancellation
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return false
	}

	return true
}

// NonRetryableError marks an error as non-retryable
type NonRetryableError struct {
	Err error
}

func (e *NonRetryableError) Error() string {
	return e.Err.Error()
}

func (e *NonRetryableError) Unwrap() error {
	return e.Err
}

// WrapNonRetryable wraps an error to mark it as non-retryable
func WrapNonRetryable(err error) error {
	if err == nil {
		return nil
	}
	return &NonRetryableError{Err: err}
}

// RetryableFunc wraps a function with retry logic
type RetryableFunc struct {
	config RetryConfig
}

// NewRetryableFunc creates a new retryable function wrapper
func NewRetryableFunc(config RetryConfig) *RetryableFunc {
	return &RetryableFunc{config: config}
}

// Run executes the function with retry
func (r *RetryableFunc) Run(ctx context.Context, fn func() error) error {
	return Retry(ctx, r.config, fn)
}

// RunWithResult executes the function with retry and returns a result
func RunWithResult[T any](ctx context.Context, config RetryConfig, fn func() (T, error)) (T, error) {
	return RetryWithResult(ctx, config, fn)
}
