package gateway

import (
	"context"
	"sync"
	"time"

	"github.com/linkflow-ai/linkflow/internal/core/domain/ai"
)

// RateLimiter handles rate limiting for AI providers
type RateLimiter struct {
	limits   map[ai.Provider]*providerLimit
	buckets  map[string]*tokenBucket
	mu       sync.RWMutex
}

type providerLimit struct {
	requestsPerMinute int
	tokensPerMinute   int
	requestsPerDay    int
}

type tokenBucket struct {
	tokens     float64
	maxTokens  float64
	refillRate float64 // tokens per second
	lastRefill time.Time
	mu         sync.Mutex
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter() *RateLimiter {
	rl := &RateLimiter{
		limits:  make(map[ai.Provider]*providerLimit),
		buckets: make(map[string]*tokenBucket),
	}

	// Set default limits per provider
	rl.limits[ai.ProviderOpenAI] = &providerLimit{
		requestsPerMinute: 500,
		tokensPerMinute:   150000,
		requestsPerDay:    10000,
	}
	rl.limits[ai.ProviderAnthropic] = &providerLimit{
		requestsPerMinute: 60,
		tokensPerMinute:   100000,
		requestsPerDay:    5000,
	}
	rl.limits[ai.ProviderGoogle] = &providerLimit{
		requestsPerMinute: 60,
		tokensPerMinute:   120000,
		requestsPerDay:    10000,
	}

	return rl
}

// Allow checks if a request is allowed under rate limits
func (rl *RateLimiter) Allow(ctx context.Context, workspaceID string, provider ai.Provider) error {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	key := workspaceID + ":" + string(provider)

	bucket, ok := rl.buckets[key]
	if !ok {
		limit, hasLimit := rl.limits[provider]
		if !hasLimit {
			// No limit configured, allow
			return nil
		}

		bucket = &tokenBucket{
			tokens:     float64(limit.requestsPerMinute),
			maxTokens:  float64(limit.requestsPerMinute),
			refillRate: float64(limit.requestsPerMinute) / 60.0,
			lastRefill: time.Now(),
		}
		rl.buckets[key] = bucket
	}

	return bucket.take(1)
}

// AllowTokens checks if token usage is allowed
func (rl *RateLimiter) AllowTokens(ctx context.Context, workspaceID string, provider ai.Provider, tokens int) error {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	key := workspaceID + ":" + string(provider) + ":tokens"

	bucket, ok := rl.buckets[key]
	if !ok {
		limit, hasLimit := rl.limits[provider]
		if !hasLimit {
			return nil
		}

		bucket = &tokenBucket{
			tokens:     float64(limit.tokensPerMinute),
			maxTokens:  float64(limit.tokensPerMinute),
			refillRate: float64(limit.tokensPerMinute) / 60.0,
			lastRefill: time.Now(),
		}
		rl.buckets[key] = bucket
	}

	return bucket.take(float64(tokens))
}

// SetLimit sets custom rate limits for a provider
func (rl *RateLimiter) SetLimit(provider ai.Provider, requestsPerMinute, tokensPerMinute, requestsPerDay int) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	rl.limits[provider] = &providerLimit{
		requestsPerMinute: requestsPerMinute,
		tokensPerMinute:   tokensPerMinute,
		requestsPerDay:    requestsPerDay,
	}
}

// SetWorkspaceLimit sets custom limits for a specific workspace
func (rl *RateLimiter) SetWorkspaceLimit(workspaceID string, provider ai.Provider, requestsPerMinute int) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	key := workspaceID + ":" + string(provider)

	rl.buckets[key] = &tokenBucket{
		tokens:     float64(requestsPerMinute),
		maxTokens:  float64(requestsPerMinute),
		refillRate: float64(requestsPerMinute) / 60.0,
		lastRefill: time.Now(),
	}
}

// GetUsage returns current usage stats for a workspace
func (rl *RateLimiter) GetUsage(workspaceID string, provider ai.Provider) (remaining int, resetAt time.Time) {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	key := workspaceID + ":" + string(provider)
	bucket, ok := rl.buckets[key]
	if !ok {
		limit, hasLimit := rl.limits[provider]
		if hasLimit {
			return limit.requestsPerMinute, time.Now().Add(time.Minute)
		}
		return -1, time.Time{} // Unlimited
	}

	bucket.refill()
	return int(bucket.tokens), time.Now().Add(time.Duration(bucket.maxTokens/bucket.refillRate) * time.Second)
}

func (b *tokenBucket) take(tokens float64) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.refill()

	if b.tokens < tokens {
		return ai.ErrRateLimited
	}

	b.tokens -= tokens
	return nil
}

func (b *tokenBucket) refill() {
	now := time.Now()
	elapsed := now.Sub(b.lastRefill).Seconds()
	b.tokens += elapsed * b.refillRate

	if b.tokens > b.maxTokens {
		b.tokens = b.maxTokens
	}

	b.lastRefill = now
}

// Reset resets rate limits for a workspace
func (rl *RateLimiter) Reset(workspaceID string) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	for key := range rl.buckets {
		if len(key) > len(workspaceID) && key[:len(workspaceID)] == workspaceID {
			delete(rl.buckets, key)
		}
	}
}

// Cleanup removes stale buckets
func (rl *RateLimiter) Cleanup(maxAge time.Duration) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	cutoff := time.Now().Add(-maxAge)
	for key, bucket := range rl.buckets {
		bucket.mu.Lock()
		if bucket.lastRefill.Before(cutoff) {
			delete(rl.buckets, key)
		}
		bucket.mu.Unlock()
	}
}
