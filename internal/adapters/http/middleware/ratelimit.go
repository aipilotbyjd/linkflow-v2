package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"golang.org/x/time/rate"
)

// RateLimiter provides rate limiting functionality
type RateLimiter struct {
	visitors map[string]*visitor
	mu       sync.RWMutex
	rate     rate.Limit
	burst    int
	cleanup  time.Duration
}

type visitor struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(requestsPerMinute int, burst int) *RateLimiter {
	rl := &RateLimiter{
		visitors: make(map[string]*visitor),
		rate:     rate.Limit(float64(requestsPerMinute) / 60.0),
		burst:    burst,
		cleanup:  5 * time.Minute,
	}
	go rl.cleanupVisitors()
	return rl
}

func (rl *RateLimiter) getVisitor(ip string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	v, exists := rl.visitors[ip]
	if !exists {
		limiter := rate.NewLimiter(rl.rate, rl.burst)
		rl.visitors[ip] = &visitor{limiter: limiter, lastSeen: time.Now()}
		return limiter
	}

	v.lastSeen = time.Now()
	return v.limiter
}

func (rl *RateLimiter) cleanupVisitors() {
	for {
		time.Sleep(rl.cleanup)
		rl.mu.Lock()
		for ip, v := range rl.visitors {
			if time.Since(v.lastSeen) > rl.cleanup {
				delete(rl.visitors, ip)
			}
		}
		rl.mu.Unlock()
	}
}

// RateLimit creates a rate limiting middleware
func RateLimit(requestsPerMinute, burst int) func(http.Handler) http.Handler {
	rl := NewRateLimiter(requestsPerMinute, burst)
	
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := getClientIP(r)
			limiter := rl.getVisitor(ip)
			
			if !limiter.Allow() {
				common.RateLimited(w)
				return
			}
			
			next.ServeHTTP(w, r)
		})
	}
}

// RateLimitByUser creates a rate limiting middleware keyed by user ID
func RateLimitByUser(requestsPerMinute, burst int) func(http.Handler) http.Handler {
	rl := NewRateLimiter(requestsPerMinute, burst)
	
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user := GetUserFromContext(r.Context())
			key := getClientIP(r)
			if user != nil {
				key = user.UserID.String()
			}
			
			limiter := rl.getVisitor(key)
			
			if !limiter.Allow() {
				common.RateLimited(w)
				return
			}
			
			next.ServeHTTP(w, r)
		})
	}
}

func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return xff
	}
	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}
	// Fall back to RemoteAddr
	return r.RemoteAddr
}
