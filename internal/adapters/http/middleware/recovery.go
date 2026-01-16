package middleware

import (
	"net/http"
	"runtime/debug"

	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/infrastructure/observability/logger"
)

// Recovery creates a panic recovery middleware
func Recovery(log logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					// Log the panic
					log.Error().
						Interface("error", err).
						Str("stack", string(debug.Stack())).
						Str("method", r.Method).
						Str("path", r.URL.Path).
						Msg("panic recovered")

					// Return internal server error
					common.InternalError(w, "internal server error")
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}
