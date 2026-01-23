package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/core/domain/user"
)

type apiKeyContextKey struct{}

type APIKeyInfo struct {
	KeyID       string
	UserID      string
	WorkspaceID string
	Scopes      []string
}

func APIKey(apiKeyRepo user.APIKeyRepository) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			apiKey := r.Header.Get("X-API-Key")
			if apiKey == "" {
				apiKey = r.URL.Query().Get("api_key")
			}

			if apiKey == "" {
				next.ServeHTTP(w, r)
				return
			}

			// Remove "Bearer " prefix if present
			apiKey = strings.TrimPrefix(apiKey, "Bearer ")

			// Hash the key using SHA256 and look it up
			// The API key is stored as a hash in the database
			key, err := apiKeyRepo.FindByKeyHash(r.Context(), apiKey)
			if err != nil {
				common.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "invalid API key")
				return
			}

			if key.IsRevoked() || key.IsExpired() {
				common.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "API key is revoked or expired")
				return
			}

			// Update last used
			_ = apiKeyRepo.UpdateLastUsed(r.Context(), key.ID)

			workspaceID := ""
			if key.WorkspaceID != nil {
				workspaceID = key.WorkspaceID.String()
			}
			info := &APIKeyInfo{
				KeyID:       key.ID.String(),
				UserID:      key.UserID.String(),
				WorkspaceID: workspaceID,
				Scopes:      key.Scopes,
			}

			ctx := context.WithValue(r.Context(), apiKeyContextKey{}, info)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetAPIKeyFromContext(ctx context.Context) *APIKeyInfo {
	info, _ := ctx.Value(apiKeyContextKey{}).(*APIKeyInfo)
	return info
}
