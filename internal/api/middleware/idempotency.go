package middleware

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"time"

	pkgredis "github.com/linkflow-ai/linkflow/internal/pkg/redis"
	"github.com/rs/zerolog/log"
)

const (
	IdempotencyHeader = "X-Idempotency-Key"
	idempotencyPrefix = "idempotency:"
	idempotencyTTL    = 24 * time.Hour
)

type cachedResponse struct {
	Status  int               `json:"status"`
	Headers map[string]string `json:"headers"`
	Body    []byte            `json:"body"`
}

type idempotencyResponseWriter struct {
	http.ResponseWriter
	status int
	body   *bytes.Buffer
}

func (w *idempotencyResponseWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func (w *idempotencyResponseWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// IdempotencyMiddleware prevents duplicate operations for POST/PUT/DELETE requests.
// Clients send X-Idempotency-Key header; server caches and returns same response on retry.
func IdempotencyMiddleware(redis *pkgredis.Client) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Only apply to mutating methods
			if r.Method != http.MethodPost && r.Method != http.MethodPut && r.Method != http.MethodDelete {
				next.ServeHTTP(w, r)
				return
			}

			idempotencyKey := r.Header.Get(IdempotencyHeader)
			if idempotencyKey == "" {
				next.ServeHTTP(w, r)
				return
			}

			// Include user ID and path in cache key for isolation
			claims := GetUserFromContext(r.Context())
			userID := "anonymous"
			if claims != nil {
				userID = claims.UserID.String()
			}

			cacheKey := buildIdempotencyKey(userID, r.URL.Path, r.Method, idempotencyKey)

			// Check for cached response
			ctx := r.Context()
			cached, err := getCachedResponse(ctx, redis, cacheKey)
			if err == nil && cached != nil {
				log.Debug().
					Str("idempotency_key", idempotencyKey).
					Str("path", r.URL.Path).
					Msg("returning cached idempotent response")

				w.Header().Set("X-Idempotent-Replayed", "true")
				for k, v := range cached.Headers {
					w.Header().Set(k, v)
				}
				w.WriteHeader(cached.Status)
				w.Write(cached.Body)
				return
			}

			// Wrap response writer to capture response
			recorder := &idempotencyResponseWriter{
				ResponseWriter: w,
				status:         http.StatusOK,
				body:           &bytes.Buffer{},
			}

			next.ServeHTTP(recorder, r)

			// Cache successful responses (2xx status codes)
			if recorder.status >= 200 && recorder.status < 300 {
				headers := make(map[string]string)
				for k, v := range recorder.Header() {
					if len(v) > 0 {
						headers[k] = v[0]
					}
				}

				toCache := &cachedResponse{
					Status:  recorder.status,
					Headers: headers,
					Body:    recorder.body.Bytes(),
				}

				if err := setCachedResponse(ctx, redis, cacheKey, toCache); err != nil {
					log.Warn().Err(err).Str("key", idempotencyKey).Msg("failed to cache idempotent response")
				}
			}
		})
	}
}

func buildIdempotencyKey(userID, path, method, key string) string {
	h := sha256.New()
	h.Write([]byte(userID + ":" + path + ":" + method + ":" + key))
	return idempotencyPrefix + hex.EncodeToString(h.Sum(nil))[:32]
}

func getCachedResponse(ctx context.Context, redis *pkgredis.Client, key string) (*cachedResponse, error) {
	data, err := redis.Get(ctx, key).Result()
	if err != nil || data == "" {
		return nil, err
	}

	var cached cachedResponse
	if err := json.Unmarshal([]byte(data), &cached); err != nil {
		return nil, err
	}
	return &cached, nil
}

func setCachedResponse(ctx context.Context, redis *pkgredis.Client, key string, resp *cachedResponse) error {
	data, err := json.Marshal(resp)
	if err != nil {
		return err
	}
	return redis.Set(ctx, key, string(data), idempotencyTTL).Err()
}

// ReadBodyWithReset reads the request body and resets it for further use.
func ReadBodyWithReset(r *http.Request) ([]byte, error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	r.Body = io.NopCloser(bytes.NewBuffer(body))
	return body, nil
}
