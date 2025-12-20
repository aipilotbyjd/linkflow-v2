package middleware

import (
	"context"
	"fmt"
	"time"

	"github.com/linkflow-ai/linkflow/internal/worker/processor"
)

// TimeoutMiddleware enforces execution timeouts on nodes
type TimeoutMiddleware struct {
	defaultTimeout time.Duration
	nodeTimeouts   map[string]time.Duration // node type -> timeout
}

// TimeoutConfig configures the timeout middleware
type TimeoutConfig struct {
	DefaultTimeout time.Duration
	NodeTimeouts   map[string]time.Duration
}

// NewTimeoutMiddleware creates a new timeout middleware
func NewTimeoutMiddleware(cfg TimeoutConfig) *TimeoutMiddleware {
	if cfg.DefaultTimeout == 0 {
		cfg.DefaultTimeout = 5 * time.Minute
	}

	return &TimeoutMiddleware{
		defaultTimeout: cfg.DefaultTimeout,
		nodeTimeouts:   cfg.NodeTimeouts,
	}
}

// Execute implements Middleware
func (m *TimeoutMiddleware) Execute(ctx context.Context, rctx *processor.RuntimeContext, node *processor.NodeDefinition, next NextFunc) (map[string]interface{}, error) {
	// Determine timeout
	timeout := m.defaultTimeout

	// Check node-specific timeout from config
	if node.Timeout > 0 {
		timeout = node.Timeout
	}

	// Check node type specific timeout
	if t, ok := m.nodeTimeouts[node.Type]; ok {
		timeout = t
	}

	// Check if node has timeout in its config
	if t, ok := node.Config["timeout"].(float64); ok && t > 0 {
		timeout = time.Duration(t) * time.Millisecond
	}

	// Create timeout context
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Execute with timeout
	done := make(chan struct {
		result *processor.NodeResult
		err    error
	}, 1)

	go func() {
		result, err := next(ctx)
		done <- struct {
			result *processor.NodeResult
			err    error
		}{result, err}
	}()

	select {
	case <-ctx.Done():
		if ctx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("node execution timed out after %v", timeout)
		}
		return nil, ctx.Err()
	case res := <-done:
		if res.err != nil {
			return nil, res.err
		}
		return res.result.Output, nil
	}
}

// SetTimeout sets a timeout for a specific node type
func (m *TimeoutMiddleware) SetTimeout(nodeType string, timeout time.Duration) {
	if m.nodeTimeouts == nil {
		m.nodeTimeouts = make(map[string]time.Duration)
	}
	m.nodeTimeouts[nodeType] = timeout
}

// DefaultTimeouts returns common timeout configurations
func DefaultTimeouts() map[string]time.Duration {
	return map[string]time.Duration{
		// Quick operations
		"logic.condition":   10 * time.Second,
		"logic.switch":      10 * time.Second,
		"logic.merge":       10 * time.Second,
		"action.set":        10 * time.Second,

		// Medium operations
		"action.http":        30 * time.Second,
		"action.code":        30 * time.Second,
		"logic.loop":         2 * time.Minute,

		// Slow operations
		"action.sub_workflow": 10 * time.Minute,
		"integration.openai":  2 * time.Minute,
		"integration.anthropic": 2 * time.Minute,

		// Wait nodes
		"logic.wait": 24 * time.Hour,
	}
}

// GracefulTimeoutMiddleware provides graceful timeout handling
type GracefulTimeoutMiddleware struct {
	defaultTimeout time.Duration
	gracePeriod    time.Duration
	onTimeout      func(ctx context.Context, rctx *processor.RuntimeContext, node *processor.NodeDefinition)
}

// GracefulTimeoutConfig configures graceful timeout
type GracefulTimeoutConfig struct {
	DefaultTimeout time.Duration
	GracePeriod    time.Duration
	OnTimeout      func(ctx context.Context, rctx *processor.RuntimeContext, node *processor.NodeDefinition)
}

// NewGracefulTimeoutMiddleware creates a graceful timeout middleware
func NewGracefulTimeoutMiddleware(cfg GracefulTimeoutConfig) *GracefulTimeoutMiddleware {
	if cfg.DefaultTimeout == 0 {
		cfg.DefaultTimeout = 5 * time.Minute
	}
	if cfg.GracePeriod == 0 {
		cfg.GracePeriod = 5 * time.Second
	}

	return &GracefulTimeoutMiddleware{
		defaultTimeout: cfg.DefaultTimeout,
		gracePeriod:    cfg.GracePeriod,
		onTimeout:      cfg.OnTimeout,
	}
}

// Execute implements Middleware
func (m *GracefulTimeoutMiddleware) Execute(ctx context.Context, rctx *processor.RuntimeContext, node *processor.NodeDefinition, next NextFunc) (map[string]interface{}, error) {
	timeout := m.defaultTimeout
	if node.Timeout > 0 {
		timeout = node.Timeout
	}

	// Main execution context with timeout
	execCtx, execCancel := context.WithTimeout(ctx, timeout)
	defer execCancel()

	// Result channel
	resultChan := make(chan struct {
		result *processor.NodeResult
		err    error
	}, 1)

	go func() {
		result, err := next(execCtx)
		resultChan <- struct {
			result *processor.NodeResult
			err    error
		}{result, err}
	}()

	select {
	case res := <-resultChan:
		if res.err != nil {
			return nil, res.err
		}
		return res.result.Output, nil

	case <-execCtx.Done():
		// Timeout occurred, call handler if set
		if m.onTimeout != nil {
			m.onTimeout(ctx, rctx, node)
		}

		// Give grace period for cleanup
		graceCtx, graceCancel := context.WithTimeout(context.Background(), m.gracePeriod)
		defer graceCancel()

		select {
		case res := <-resultChan:
			// Completed during grace period
			if res.err != nil {
				return nil, res.err
			}
			return res.result.Output, nil
		case <-graceCtx.Done():
			return nil, fmt.Errorf("node execution timed out after %v (plus %v grace period)", timeout, m.gracePeriod)
		}
	}
}

// CircuitBreakerTimeoutMiddleware combines timeout with circuit breaker
type CircuitBreakerTimeoutMiddleware struct {
	timeout        time.Duration
	failureCount   map[string]int
	lastFailure    map[string]time.Time
	threshold      int
	resetTimeout   time.Duration
}

// NewCircuitBreakerTimeoutMiddleware creates a circuit breaker timeout middleware
func NewCircuitBreakerTimeoutMiddleware(timeout time.Duration, threshold int, resetTimeout time.Duration) *CircuitBreakerTimeoutMiddleware {
	return &CircuitBreakerTimeoutMiddleware{
		timeout:      timeout,
		failureCount: make(map[string]int),
		lastFailure:  make(map[string]time.Time),
		threshold:    threshold,
		resetTimeout: resetTimeout,
	}
}

// Execute implements Middleware
func (m *CircuitBreakerTimeoutMiddleware) Execute(ctx context.Context, rctx *processor.RuntimeContext, node *processor.NodeDefinition, next NextFunc) (map[string]interface{}, error) {
	key := node.Type

	// Check circuit breaker
	if m.failureCount[key] >= m.threshold {
		if time.Since(m.lastFailure[key]) < m.resetTimeout {
			return nil, fmt.Errorf("circuit breaker open for node type %s", node.Type)
		}
		// Reset after timeout
		m.failureCount[key] = 0
	}

	// Apply timeout
	ctx, cancel := context.WithTimeout(ctx, m.timeout)
	defer cancel()

	result, err := next(ctx)
	if err != nil {
		m.failureCount[key]++
		m.lastFailure[key] = time.Now()
		return nil, err
	}

	// Success resets failure count
	m.failureCount[key] = 0
	return result.Output, nil
}
