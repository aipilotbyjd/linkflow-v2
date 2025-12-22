package dispatcher

import (
	"context"
	"fmt"
	"sync"
	"time"

	pkgredis "github.com/linkflow-ai/linkflow/internal/pkg/redis"
	"github.com/redis/go-redis/v9"
)

type RateLimiter interface {
	Allow(ctx context.Context, key string) bool
	AllowN(ctx context.Context, key string, n int) bool
}

// SlidingWindowLimiter implements sliding window rate limiting
type SlidingWindowLimiter struct {
	redis      *pkgredis.Client
	keyPrefix  string
	limit      int
	windowSize time.Duration
}

func NewSlidingWindowLimiter(redis *pkgredis.Client, keyPrefix string, limit int, windowSize time.Duration) *SlidingWindowLimiter {
	return &SlidingWindowLimiter{
		redis:      redis,
		keyPrefix:  keyPrefix,
		limit:      limit,
		windowSize: windowSize,
	}
}

func (l *SlidingWindowLimiter) Allow(ctx context.Context, key string) bool {
	return l.AllowN(ctx, key, 1)
}

func (l *SlidingWindowLimiter) AllowN(ctx context.Context, key string, n int) bool {
	fullKey := l.keyPrefix + ":" + key
	now := time.Now()
	windowStart := now.Add(-l.windowSize)

	pipe := l.redis.Pipeline()
	pipe.ZRemRangeByScore(ctx, fullKey, "-inf", fmt.Sprintf("%d", windowStart.UnixNano()))
	countCmd := pipe.ZCard(ctx, fullKey)
	_, err := pipe.Exec(ctx)

	if err != nil {
		return true // Allow on error
	}

	if int(countCmd.Val())+n > l.limit {
		return false
	}

	// Add entries
	members := make([]redis.Z, n)
	for i := 0; i < n; i++ {
		members[i] = redis.Z{
			Score:  float64(now.UnixNano() + int64(i)),
			Member: fmt.Sprintf("%d-%d", now.UnixNano(), i),
		}
	}

	pipe = l.redis.Pipeline()
	pipe.ZAdd(ctx, fullKey, members...)
	pipe.Expire(ctx, fullKey, l.windowSize*2)
	_, _ = pipe.Exec(ctx)

	return true
}

// LocalLimiter is an in-memory rate limiter
type LocalLimiter struct {
	limit      int
	windowSize time.Duration
	windows    map[string]*window
	mu         sync.Mutex
}

type window struct {
	count     int
	startTime time.Time
}

func NewLocalLimiter(limit int, windowSize time.Duration) *LocalLimiter {
	l := &LocalLimiter{
		limit:      limit,
		windowSize: windowSize,
		windows:    make(map[string]*window),
	}
	go l.cleanup()
	return l
}

func (l *LocalLimiter) Allow(ctx context.Context, key string) bool {
	return l.AllowN(ctx, key, 1)
}

func (l *LocalLimiter) AllowN(ctx context.Context, key string, n int) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	w, exists := l.windows[key]

	if !exists || now.Sub(w.startTime) > l.windowSize {
		l.windows[key] = &window{count: n, startTime: now}
		return n <= l.limit
	}

	if w.count+n > l.limit {
		return false
	}

	w.count += n
	return true
}

func (l *LocalLimiter) cleanup() {
	ticker := time.NewTicker(time.Minute)
	for range ticker.C {
		l.mu.Lock()
		now := time.Now()
		for key, w := range l.windows {
			if now.Sub(w.startTime) > l.windowSize*2 {
				delete(l.windows, key)
			}
		}
		l.mu.Unlock()
	}
}

// CompositeLimiter combines multiple limiters
type CompositeLimiter struct {
	limiters []RateLimiter
}

func NewCompositeLimiter(limiters ...RateLimiter) *CompositeLimiter {
	return &CompositeLimiter{limiters: limiters}
}

func (l *CompositeLimiter) Allow(ctx context.Context, key string) bool {
	for _, limiter := range l.limiters {
		if !limiter.Allow(ctx, key) {
			return false
		}
	}
	return true
}

func (l *CompositeLimiter) AllowN(ctx context.Context, key string, n int) bool {
	for _, limiter := range l.limiters {
		if !limiter.AllowN(ctx, key, n) {
			return false
		}
	}
	return true
}
