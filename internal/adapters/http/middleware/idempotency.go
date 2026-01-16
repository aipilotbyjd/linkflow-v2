package middleware

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"time"

	"github.com/linkflow-ai/linkflow/internal/infrastructure/cache"
)

const (
	IdempotencyKeyHeader = "Idempotency-Key"
	idempotencyTTL       = 24 * time.Hour
)

type cachedResponse struct {
	StatusCode int               `json:"status_code"`
	Headers    map[string]string `json:"headers"`
	Body       []byte            `json:"body"`
}

type idempotencyResponseWriter struct {
	http.ResponseWriter
	statusCode int
	body       bytes.Buffer
}

func (w *idempotencyResponseWriter) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *idempotencyResponseWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func Idempotency(cacheStore cache.Cache) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Only apply to mutating methods
			if r.Method != "POST" && r.Method != "PUT" && r.Method != "PATCH" {
				next.ServeHTTP(w, r)
				return
			}

			idempotencyKey := r.Header.Get(IdempotencyKeyHeader)
			if idempotencyKey == "" {
				next.ServeHTTP(w, r)
				return
			}

			// Generate cache key
			cacheKey := generateIdempotencyCacheKey(r, idempotencyKey)

			// Check for cached response
			cached, err := cacheStore.Get(r.Context(), cacheKey)
			if err == nil && cached != nil {
				// Return cached response
				w.Header().Set("X-Idempotency-Replay", "true")
				w.Write(cached)
				return
			}

			// Wrap response writer to capture response
			wrapped := &idempotencyResponseWriter{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
			}

			next.ServeHTTP(wrapped, r)

			// Cache successful responses
			if wrapped.statusCode >= 200 && wrapped.statusCode < 300 {
				_ = cacheStore.Set(r.Context(), cacheKey, wrapped.body.Bytes(), idempotencyTTL)
			}
		})
	}
}

func generateIdempotencyCacheKey(r *http.Request, key string) string {
	// Include user context if available
	userID := ""
	if claims := GetUserFromContext(r.Context()); claims != nil {
		userID = claims.UserID.String()
	}

	hash := sha256.New()
	hash.Write([]byte(key))
	hash.Write([]byte(userID))
	hash.Write([]byte(r.Method))
	hash.Write([]byte(r.URL.Path))

	return "idempotency:" + hex.EncodeToString(hash.Sum(nil))
}

// IdempotencyKeyFromRequest reads and resets the request body to generate a key
func IdempotencyKeyFromRequest(r *http.Request) string {
	if r.Body == nil {
		return ""
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		return ""
	}
	r.Body = io.NopCloser(bytes.NewBuffer(body))

	hash := sha256.Sum256(body)
	return hex.EncodeToString(hash[:])
}

type idempotencyContextKey struct{}

func SetIdempotencyKey(ctx context.Context, key string) context.Context {
	return context.WithValue(ctx, idempotencyContextKey{}, key)
}

func GetIdempotencyKey(ctx context.Context) string {
	key, _ := ctx.Value(idempotencyContextKey{}).(string)
	return key
}
