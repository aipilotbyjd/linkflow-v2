package middleware

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/linkflow-ai/linkflow/internal/worker/processor"
	"golang.org/x/time/rate"
)

// RateLimitMiddleware enforces rate limits on node execution
type RateLimitMiddleware struct {
	limiters sync.Map // key -> *rate.Limiter
	config   RateLimitConfig
}

// RateLimitConfig configures the rate limiter
type RateLimitConfig struct {
	// Global rate limit (requests per second)
	GlobalRPS float64

	// Per-workspace limits
	WorkspaceRPS float64
	WorkspaceBurst int

	// Per-node-type limits
	NodeTypeRPS map[string]float64
	NodeTypeBurst map[string]int

	// Enabled categories (e.g., integration nodes)
	EnabledCategories []string

	// Whether to wait or fail on limit
	WaitOnLimit bool
	MaxWaitTime time.Duration
}

// DefaultRateLimitConfig returns default configuration
func DefaultRateLimitConfig() RateLimitConfig {
	return RateLimitConfig{
		GlobalRPS:      1000,
		WorkspaceRPS:   100,
		WorkspaceBurst: 50,
		NodeTypeRPS: map[string]float64{
			"integration.openai":    10,
			"integration.anthropic": 10,
			"integration.slack":     20,
			"integration.email":     10,
			"action.http":           50,
		},
		NodeTypeBurst: map[string]int{
			"integration.openai":    5,
			"integration.anthropic": 5,
			"integration.slack":     10,
			"integration.email":     5,
			"action.http":           25,
		},
		EnabledCategories: []string{"integration"},
		WaitOnLimit:       true,
		MaxWaitTime:       30 * time.Second,
	}
}

// NewRateLimitMiddleware creates a new rate limit middleware
func NewRateLimitMiddleware(cfg RateLimitConfig) *RateLimitMiddleware {
	return &RateLimitMiddleware{
		config: cfg,
	}
}

// Execute implements Middleware
func (m *RateLimitMiddleware) Execute(ctx context.Context, rctx *processor.RuntimeContext, node *processor.NodeDefinition, next NextFunc) (map[string]interface{}, error) {
	// Check if rate limiting is enabled for this node type
	if !m.shouldRateLimit(node) {
		result, err := next(ctx)
		if err != nil {
			return nil, err
		}
		return result.Output, nil
	}

	// Get workspace limiter
	workspaceLimiter := m.getWorkspaceLimiter(rctx.WorkspaceID.String())

	// Get node type limiter
	nodeTypeLimiter := m.getNodeTypeLimiter(node.Type)

	// Check/wait for both limiters
	if m.config.WaitOnLimit {
		// Create context with max wait time
		waitCtx := ctx
		if m.config.MaxWaitTime > 0 {
			var cancel context.CancelFunc
			waitCtx, cancel = context.WithTimeout(ctx, m.config.MaxWaitTime)
			defer cancel()
		}

		// Wait for workspace limiter
		if err := workspaceLimiter.Wait(waitCtx); err != nil {
			return nil, fmt.Errorf("workspace rate limit exceeded: %w", err)
		}

		// Wait for node type limiter
		if nodeTypeLimiter != nil {
			if err := nodeTypeLimiter.Wait(waitCtx); err != nil {
				return nil, fmt.Errorf("node type rate limit exceeded for %s: %w", node.Type, err)
			}
		}
	} else {
		// Fail immediately if limit exceeded
		if !workspaceLimiter.Allow() {
			return nil, fmt.Errorf("workspace rate limit exceeded")
		}
		if nodeTypeLimiter != nil && !nodeTypeLimiter.Allow() {
			return nil, fmt.Errorf("node type rate limit exceeded for %s", node.Type)
		}
	}

	// Execute
	result, err := next(ctx)
	if err != nil {
		return nil, err
	}
	return result.Output, nil
}

func (m *RateLimitMiddleware) shouldRateLimit(node *processor.NodeDefinition) bool {
	// Check if node type has specific limit
	if _, ok := m.config.NodeTypeRPS[node.Type]; ok {
		return true
	}

	// Check if category is enabled
	for _, cat := range m.config.EnabledCategories {
		if len(node.Type) >= len(cat) && node.Type[:len(cat)] == cat {
			return true
		}
	}

	return false
}

func (m *RateLimitMiddleware) getWorkspaceLimiter(workspaceID string) *rate.Limiter {
	key := "workspace:" + workspaceID

	if limiter, ok := m.limiters.Load(key); ok {
		return limiter.(*rate.Limiter)
	}

	limiter := rate.NewLimiter(rate.Limit(m.config.WorkspaceRPS), m.config.WorkspaceBurst)
	m.limiters.Store(key, limiter)
	return limiter
}

func (m *RateLimitMiddleware) getNodeTypeLimiter(nodeType string) *rate.Limiter {
	rps, ok := m.config.NodeTypeRPS[nodeType]
	if !ok {
		return nil
	}

	key := "nodetype:" + nodeType

	if limiter, ok := m.limiters.Load(key); ok {
		return limiter.(*rate.Limiter)
	}

	burst := 10
	if b, ok := m.config.NodeTypeBurst[nodeType]; ok {
		burst = b
	}

	limiter := rate.NewLimiter(rate.Limit(rps), burst)
	m.limiters.Store(key, limiter)
	return limiter
}

// TokenBucketLimiter provides a token bucket rate limiter
type TokenBucketLimiter struct {
	tokens     float64
	maxTokens  float64
	refillRate float64 // tokens per second
	lastRefill time.Time
	mu         sync.Mutex
}

// NewTokenBucketLimiter creates a new token bucket limiter
func NewTokenBucketLimiter(maxTokens, refillRate float64) *TokenBucketLimiter {
	return &TokenBucketLimiter{
		tokens:     maxTokens,
		maxTokens:  maxTokens,
		refillRate: refillRate,
		lastRefill: time.Now(),
	}
}

// Allow checks if a request can proceed
func (l *TokenBucketLimiter) Allow() bool {
	return l.AllowN(1)
}

// AllowN checks if n requests can proceed
func (l *TokenBucketLimiter) AllowN(n float64) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.refill()

	if l.tokens >= n {
		l.tokens -= n
		return true
	}
	return false
}

// Wait waits until a request can proceed
func (l *TokenBucketLimiter) Wait(ctx context.Context) error {
	return l.WaitN(ctx, 1)
}

// WaitN waits until n requests can proceed
func (l *TokenBucketLimiter) WaitN(ctx context.Context, n float64) error {
	for {
		l.mu.Lock()
		l.refill()

		if l.tokens >= n {
			l.tokens -= n
			l.mu.Unlock()
			return nil
		}

		// Calculate wait time
		needed := n - l.tokens
		waitTime := time.Duration(needed/l.refillRate*1000) * time.Millisecond
		l.mu.Unlock()

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(waitTime):
			// Try again
		}
	}
}

func (l *TokenBucketLimiter) refill() {
	now := time.Now()
	elapsed := now.Sub(l.lastRefill).Seconds()
	l.tokens += elapsed * l.refillRate
	if l.tokens > l.maxTokens {
		l.tokens = l.maxTokens
	}
	l.lastRefill = now
}

// SlidingWindowLimiter provides a sliding window rate limiter
type SlidingWindowLimiter struct {
	windowSize time.Duration
	maxReqs    int
	requests   []time.Time
	mu         sync.Mutex
}

// NewSlidingWindowLimiter creates a new sliding window limiter
func NewSlidingWindowLimiter(windowSize time.Duration, maxReqs int) *SlidingWindowLimiter {
	return &SlidingWindowLimiter{
		windowSize: windowSize,
		maxReqs:    maxReqs,
		requests:   make([]time.Time, 0),
	}
}

// Allow checks if a request can proceed
func (l *SlidingWindowLimiter) Allow() bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	windowStart := now.Add(-l.windowSize)

	// Remove old requests
	valid := make([]time.Time, 0)
	for _, t := range l.requests {
		if t.After(windowStart) {
			valid = append(valid, t)
		}
	}
	l.requests = valid

	// Check limit
	if len(l.requests) >= l.maxReqs {
		return false
	}

	// Add new request
	l.requests = append(l.requests, now)
	return true
}

// ConcurrencyLimiter limits concurrent executions
type ConcurrencyLimiter struct {
	semaphore chan struct{}
	timeout   time.Duration
}

// NewConcurrencyLimiter creates a concurrency limiter
func NewConcurrencyLimiter(maxConcurrent int, timeout time.Duration) *ConcurrencyLimiter {
	return &ConcurrencyLimiter{
		semaphore: make(chan struct{}, maxConcurrent),
		timeout:   timeout,
	}
}

// Acquire acquires a slot
func (l *ConcurrencyLimiter) Acquire(ctx context.Context) error {
	if l.timeout > 0 {
		ctx, cancel := context.WithTimeout(ctx, l.timeout)
		defer cancel()

		select {
		case l.semaphore <- struct{}{}:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	select {
	case l.semaphore <- struct{}{}:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Release releases a slot
func (l *ConcurrencyLimiter) Release() {
	<-l.semaphore
}

// ConcurrencyMiddleware limits concurrent node executions
type ConcurrencyMiddleware struct {
	limiters sync.Map
	config   ConcurrencyConfig
}

// ConcurrencyConfig configures concurrency limits
type ConcurrencyConfig struct {
	MaxGlobal            int
	MaxPerWorkspace      int
	MaxPerNodeType       map[string]int
	AcquireTimeout       time.Duration
}

// NewConcurrencyMiddleware creates a concurrency middleware
func NewConcurrencyMiddleware(cfg ConcurrencyConfig) *ConcurrencyMiddleware {
	return &ConcurrencyMiddleware{config: cfg}
}

// Execute implements Middleware
func (m *ConcurrencyMiddleware) Execute(ctx context.Context, rctx *processor.RuntimeContext, node *processor.NodeDefinition, next NextFunc) (map[string]interface{}, error) {
	// Get workspace limiter
	limiter := m.getLimiter(rctx.WorkspaceID.String())

	// Acquire
	if err := limiter.Acquire(ctx); err != nil {
		return nil, fmt.Errorf("concurrency limit exceeded: %w", err)
	}
	defer limiter.Release()

	// Execute
	result, err := next(ctx)
	if err != nil {
		return nil, err
	}
	return result.Output, nil
}

func (m *ConcurrencyMiddleware) getLimiter(workspaceID string) *ConcurrencyLimiter {
	if limiter, ok := m.limiters.Load(workspaceID); ok {
		return limiter.(*ConcurrencyLimiter)
	}

	limiter := NewConcurrencyLimiter(m.config.MaxPerWorkspace, m.config.AcquireTimeout)
	m.limiters.Store(workspaceID, limiter)
	return limiter
}
