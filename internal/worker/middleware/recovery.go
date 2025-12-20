package middleware

import (
	"context"
	"fmt"
	"runtime/debug"
	"time"

	"github.com/linkflow-ai/linkflow/internal/worker/processor"
	"github.com/rs/zerolog/log"
)

// RecoveryMiddleware recovers from panics during node execution
type RecoveryMiddleware struct {
	logStackTrace bool
	onPanic       func(ctx context.Context, rctx *processor.RuntimeContext, node *processor.NodeDefinition, recovered interface{}, stack []byte)
}

// RecoveryConfig configures recovery middleware
type RecoveryConfig struct {
	LogStackTrace bool
	OnPanic       func(ctx context.Context, rctx *processor.RuntimeContext, node *processor.NodeDefinition, recovered interface{}, stack []byte)
}

// NewRecoveryMiddleware creates a new recovery middleware
func NewRecoveryMiddleware(cfg RecoveryConfig) *RecoveryMiddleware {
	return &RecoveryMiddleware{
		logStackTrace: cfg.LogStackTrace,
		onPanic:       cfg.OnPanic,
	}
}

// Execute implements Middleware
func (m *RecoveryMiddleware) Execute(ctx context.Context, rctx *processor.RuntimeContext, node *processor.NodeDefinition, next NextFunc) (output map[string]interface{}, err error) {
	defer func() {
		if r := recover(); r != nil {
			stack := debug.Stack()

			// Log the panic
			log.Error().
				Str("execution_id", rctx.ExecutionID.String()).
				Str("node_id", node.ID).
				Str("node_type", node.Type).
				Interface("panic", r).
				Msg("Node execution panicked")

			if m.logStackTrace {
				log.Error().Str("stack", string(stack)).Msg("Panic stack trace")
			}

			// Call custom handler
			if m.onPanic != nil {
				m.onPanic(ctx, rctx, node, r, stack)
			}

			// Return error instead of crashing
			err = fmt.Errorf("node panicked: %v", r)
			output = nil
		}
	}()

	result, err := next(ctx)
	if err != nil {
		return nil, err
	}
	return result.Output, nil
}

// RetryOnPanicMiddleware retries on panic
type RetryOnPanicMiddleware struct {
	maxRetries int
	onRetry    func(ctx context.Context, rctx *processor.RuntimeContext, node *processor.NodeDefinition, attempt int, recovered interface{})
}

// NewRetryOnPanicMiddleware creates a retry on panic middleware
func NewRetryOnPanicMiddleware(maxRetries int) *RetryOnPanicMiddleware {
	return &RetryOnPanicMiddleware{
		maxRetries: maxRetries,
	}
}

// Execute implements Middleware
func (m *RetryOnPanicMiddleware) Execute(ctx context.Context, rctx *processor.RuntimeContext, node *processor.NodeDefinition, next NextFunc) (map[string]interface{}, error) {
	var lastPanic interface{}
	var lastStack []byte

	for attempt := 0; attempt <= m.maxRetries; attempt++ {
		output, err, panicked := m.executeWithRecovery(ctx, rctx, node, next)
		if panicked == nil {
			return output, err
		}

		lastPanic = panicked
		lastStack = debug.Stack()

		if m.onRetry != nil {
			m.onRetry(ctx, rctx, node, attempt+1, panicked)
		}

		log.Warn().
			Str("execution_id", rctx.ExecutionID.String()).
			Str("node_id", node.ID).
			Int("attempt", attempt+1).
			Interface("panic", panicked).
			Msg("Retrying after panic")
	}

	log.Error().
		Str("execution_id", rctx.ExecutionID.String()).
		Str("node_id", node.ID).
		Str("stack", string(lastStack)).
		Interface("panic", lastPanic).
		Msg("Max retries exceeded after panic")

	return nil, fmt.Errorf("node panicked after %d retries: %v", m.maxRetries, lastPanic)
}

func (m *RetryOnPanicMiddleware) executeWithRecovery(ctx context.Context, rctx *processor.RuntimeContext, node *processor.NodeDefinition, next NextFunc) (output map[string]interface{}, err error, panicked interface{}) {
	defer func() {
		if r := recover(); r != nil {
			panicked = r
		}
	}()

	result, err := next(ctx)
	if err != nil {
		return nil, err, nil
	}
	return result.Output, nil, nil
}

// ErrorHandlingMiddleware provides comprehensive error handling
type ErrorHandlingMiddleware struct {
	onError           func(ctx context.Context, rctx *processor.RuntimeContext, node *processor.NodeDefinition, err error) error
	transformError    func(err error) error
	shouldRetry       func(err error) bool
	maxRetries        int
	retryDelay        func(attempt int) time.Duration
}

// ErrorHandlingConfig configures error handling
type ErrorHandlingConfig struct {
	OnError        func(ctx context.Context, rctx *processor.RuntimeContext, node *processor.NodeDefinition, err error) error
	TransformError func(err error) error
	ShouldRetry    func(err error) bool
	MaxRetries     int
	RetryDelay     func(attempt int) time.Duration
}

// NewErrorHandlingMiddleware creates an error handling middleware
func NewErrorHandlingMiddleware(cfg ErrorHandlingConfig) *ErrorHandlingMiddleware {
	m := &ErrorHandlingMiddleware{
		onError:        cfg.OnError,
		transformError: cfg.TransformError,
		shouldRetry:    cfg.ShouldRetry,
		maxRetries:     cfg.MaxRetries,
		retryDelay:     cfg.RetryDelay,
	}

	if m.retryDelay == nil {
		m.retryDelay = func(attempt int) time.Duration {
			return time.Duration(attempt*attempt) * 100 * time.Millisecond
		}
	}

	return m
}

// Execute implements Middleware
func (m *ErrorHandlingMiddleware) Execute(ctx context.Context, rctx *processor.RuntimeContext, node *processor.NodeDefinition, next NextFunc) (map[string]interface{}, error) {
	var lastErr error

	for attempt := 0; attempt <= m.maxRetries; attempt++ {
		result, err := next(ctx)
		if err == nil {
			return result.Output, nil
		}

		lastErr = err

		// Transform error if configured
		if m.transformError != nil {
			lastErr = m.transformError(err)
		}

		// Call error handler
		if m.onError != nil {
			if handledErr := m.onError(ctx, rctx, node, lastErr); handledErr != nil {
				lastErr = handledErr
			}
		}

		// Check if we should retry
		if m.shouldRetry == nil || !m.shouldRetry(lastErr) {
			break
		}

		if attempt < m.maxRetries {
			delay := m.retryDelay(attempt + 1)
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(delay):
				continue
			}
		}
	}

	return nil, lastErr
}

// CircuitState represents circuit breaker state
type CircuitState int

const (
	CircuitClosed CircuitState = iota
	CircuitOpen
	CircuitHalfOpen
)

// CircuitBreakerMiddleware implements circuit breaker pattern
type CircuitBreakerMiddleware struct {
	state          CircuitState
	failures       int
	successes      int
	threshold      int
	successThreshold int
	timeout        time.Duration
	lastFailure    time.Time
	onStateChange  func(from, to CircuitState)
}

// CircuitBreakerConfig configures circuit breaker
type CircuitBreakerConfig struct {
	FailureThreshold int
	SuccessThreshold int
	Timeout          time.Duration
	OnStateChange    func(from, to CircuitState)
}

// NewCircuitBreakerMiddleware creates a circuit breaker middleware
func NewCircuitBreakerMiddleware(cfg CircuitBreakerConfig) *CircuitBreakerMiddleware {
	if cfg.FailureThreshold == 0 {
		cfg.FailureThreshold = 5
	}
	if cfg.SuccessThreshold == 0 {
		cfg.SuccessThreshold = 2
	}
	if cfg.Timeout == 0 {
		cfg.Timeout = 30 * time.Second
	}

	return &CircuitBreakerMiddleware{
		state:            CircuitClosed,
		threshold:        cfg.FailureThreshold,
		successThreshold: cfg.SuccessThreshold,
		timeout:          cfg.Timeout,
		onStateChange:    cfg.OnStateChange,
	}
}

// Execute implements Middleware
func (m *CircuitBreakerMiddleware) Execute(ctx context.Context, rctx *processor.RuntimeContext, node *processor.NodeDefinition, next NextFunc) (map[string]interface{}, error) {
	if m.state == CircuitOpen {
		// Check if timeout has passed
		if time.Since(m.lastFailure) > m.timeout {
			m.transitionTo(CircuitHalfOpen)
		} else {
			return nil, fmt.Errorf("circuit breaker is open")
		}
	}

	result, err := next(ctx)

	if err != nil {
		m.recordFailure()
		return nil, err
	}

	m.recordSuccess()
	return result.Output, nil
}

func (m *CircuitBreakerMiddleware) recordFailure() {
	m.failures++
	m.lastFailure = time.Now()

	if m.state == CircuitHalfOpen || m.failures >= m.threshold {
		m.transitionTo(CircuitOpen)
	}
}

func (m *CircuitBreakerMiddleware) recordSuccess() {
	if m.state == CircuitHalfOpen {
		m.successes++
		if m.successes >= m.successThreshold {
			m.transitionTo(CircuitClosed)
		}
	} else {
		m.failures = 0
	}
}

func (m *CircuitBreakerMiddleware) transitionTo(newState CircuitState) {
	oldState := m.state
	m.state = newState

	if newState == CircuitClosed {
		m.failures = 0
		m.successes = 0
	} else if newState == CircuitHalfOpen {
		m.successes = 0
	}

	if m.onStateChange != nil {
		m.onStateChange(oldState, newState)
	}
}

// State returns current circuit state
func (m *CircuitBreakerMiddleware) State() CircuitState {
	return m.state
}
