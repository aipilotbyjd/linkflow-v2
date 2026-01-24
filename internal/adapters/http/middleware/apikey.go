package middleware

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strings"

	"github.com/google/uuid"
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

			// Remove "lf_" prefix if present
			apiKey = strings.TrimPrefix(apiKey, "lf_")

			// Hash the key using SHA256 and look it up
			hash := sha256.Sum256([]byte(apiKey))
			hashHex := hex.EncodeToString(hash[:])

			// The API key is stored as a hash in the database
			key, err := apiKeyRepo.FindByKeyHash(r.Context(), hashHex)
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

			// Set workspace context
			workspaceID := uuid.Nil
			var workspaceIDPtr *uuid.UUID
			if key.WorkspaceID != nil {
				workspaceID = *key.WorkspaceID
				workspaceIDPtr = key.WorkspaceID
			}

			// Create compatible UserClaims for other middlewares
			userClaims := &UserClaims{
				UserID:      key.UserID,
				Email:       "api-key@programmatic", // Placeholder for programmatic access
				WorkspaceID: workspaceIDPtr,
			}

			info := &APIKeyInfo{
				KeyID:       key.ID.String(),
				UserID:      key.UserID.String(),
				WorkspaceID: workspaceID.String(),
				Scopes:      key.Scopes,
			}

			ctx := context.WithValue(r.Context(), apiKeyContextKey{}, info)
			ctx = context.WithValue(ctx, UserContextKey, userClaims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetAPIKeyFromContext(ctx context.Context) *APIKeyInfo {
	info, _ := ctx.Value(apiKeyContextKey{}).(*APIKeyInfo)
	return info
}
