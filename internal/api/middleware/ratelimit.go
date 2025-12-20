package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/linkflow-ai/linkflow/internal/api/dto"
	"github.com/linkflow-ai/linkflow/internal/pkg/crypto"
	pkgredis "github.com/linkflow-ai/linkflow/internal/pkg/redis"
)

type RateLimiter struct {
	redis *pkgredis.Client
}

func NewRateLimiter(redis *pkgredis.Client) *RateLimiter {
	return &RateLimiter{redis: redis}
}

func (rl *RateLimiter) Limit(limit int, window time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := rl.getKey(r)
			
			allowed, remaining, err := rl.redis.RateLimit(r.Context(), key, limit, window)
			if err != nil {
				// If Redis fails, allow the request
				next.ServeHTTP(w, r)
				return
			}

			w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", limit))
			w.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
			w.Header().Set("X-RateLimit-Reset", fmt.Sprintf("%d", time.Now().Add(window).Unix()))

			if !allowed {
				dto.ErrorResponse(w, http.StatusTooManyRequests, "rate limit exceeded")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func (rl *RateLimiter) LimitByUser(limit int, window time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims := GetUserFromContext(r.Context())
			if claims == nil {
				next.ServeHTTP(w, r)
				return
			}

			key := fmt.Sprintf("ratelimit:user:%s", claims.UserID.String())
			
			allowed, remaining, err := rl.redis.RateLimit(r.Context(), key, limit, window)
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}

			w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", limit))
			w.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))

			if !allowed {
				dto.ErrorResponse(w, http.StatusTooManyRequests, "rate limit exceeded")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func (rl *RateLimiter) getKey(r *http.Request) string {
	// Try to get user ID first
	claims := GetUserFromContext(r.Context())
	if claims != nil {
		return fmt.Sprintf("ratelimit:user:%s", claims.UserID.String())
	}

	// Fall back to IP address
	ip := r.RemoteAddr
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		ip = forwarded
	}
	return fmt.Sprintf("ratelimit:ip:%s", ip)
}

func (rl *RateLimiter) LimitByEndpoint(limit int, window time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims := GetUserFromContext(r.Context())
			var identifier string
			if claims != nil {
				identifier = claims.UserID.String()
			} else {
				identifier = r.RemoteAddr
			}

			key := fmt.Sprintf("ratelimit:endpoint:%s:%s:%s", r.Method, r.URL.Path, identifier)
			
			allowed, remaining, err := rl.redis.RateLimit(r.Context(), key, limit, window)
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}

			w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", limit))
			w.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))

			if !allowed {
				dto.ErrorResponse(w, http.StatusTooManyRequests, "rate limit exceeded")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

var _ = crypto.GenerateRandomString // silence import
