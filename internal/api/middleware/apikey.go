package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/linkflow-ai/linkflow/internal/api/dto"
	"github.com/linkflow-ai/linkflow/internal/domain/models"
	"github.com/linkflow-ai/linkflow/internal/domain/services"
	"github.com/linkflow-ai/linkflow/internal/pkg/crypto"
)

const APIKeyContextKey contextKey = "apikey"

type APIKeyMiddleware struct {
	apiKeySvc *services.APIKeyService
}

func NewAPIKeyMiddleware(apiKeySvc *services.APIKeyService) *APIKeyMiddleware {
	return &APIKeyMiddleware{apiKeySvc: apiKeySvc}
}

func (m *APIKeyMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiKey := extractAPIKey(r)
		if apiKey == "" {
			next.ServeHTTP(w, r)
			return
		}

		key, err := m.apiKeySvc.Validate(r.Context(), apiKey)
		if err != nil {
			dto.ErrorResponse(w, http.StatusUnauthorized, "invalid API key")
			return
		}

		ctx := context.WithValue(r.Context(), APIKeyContextKey, key)
		ctx = context.WithValue(ctx, UserContextKey, &crypto.Claims{
			UserID: key.UserID,
		})

		if key.WorkspaceID != nil {
			ctx = context.WithValue(ctx, WorkspaceContextKey, &WorkspaceContext{
				WorkspaceID: *key.WorkspaceID,
			})
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (m *APIKeyMiddleware) RequireAPIKey(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiKey := extractAPIKey(r)
		if apiKey == "" {
			dto.ErrorResponse(w, http.StatusUnauthorized, "API key required")
			return
		}

		key, err := m.apiKeySvc.Validate(r.Context(), apiKey)
		if err != nil {
			dto.ErrorResponse(w, http.StatusUnauthorized, "invalid API key")
			return
		}

		ctx := context.WithValue(r.Context(), APIKeyContextKey, key)
		ctx = context.WithValue(ctx, UserContextKey, &crypto.Claims{
			UserID: key.UserID,
		})

		if key.WorkspaceID != nil {
			ctx = context.WithValue(ctx, WorkspaceContextKey, &WorkspaceContext{
				WorkspaceID: *key.WorkspaceID,
			})
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (m *APIKeyMiddleware) RequireScope(scope string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			apiKey := GetAPIKeyFromContext(r.Context())
			if apiKey == nil {
				dto.ErrorResponse(w, http.StatusUnauthorized, "API key required")
				return
			}

			if !m.apiKeySvc.HasScope(apiKey, scope) {
				dto.ErrorResponse(w, http.StatusForbidden, "insufficient scope: "+scope)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func extractAPIKey(r *http.Request) string {
	auth := r.Header.Get("Authorization")
	if strings.HasPrefix(auth, "Bearer lf_") {
		return strings.TrimPrefix(auth, "Bearer ")
	}

	if key := r.Header.Get("X-API-Key"); key != "" {
		return key
	}

	if key := r.URL.Query().Get("api_key"); key != "" {
		return key
	}

	return ""
}

func GetAPIKeyFromContext(ctx context.Context) *models.APIKey {
	if key, ok := ctx.Value(APIKeyContextKey).(*models.APIKey); ok {
		return key
	}
	return nil
}
