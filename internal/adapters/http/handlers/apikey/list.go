package apikey

import (
	"net/http"

	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/middleware"
	"github.com/linkflow-ai/linkflow/internal/core/domain/user"
)

type ListAPIKeysHandler struct {
	apiKeyRepo user.APIKeyRepository
}

func NewListAPIKeysHandler(apiKeyRepo user.APIKeyRepository) *ListAPIKeysHandler {
	return &ListAPIKeysHandler{apiKeyRepo: apiKeyRepo}
}

type APIKeyResponse struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	KeyPrefix  string   `json:"key_prefix"`
	Scopes     []string `json:"scopes"`
	LastUsedAt *string  `json:"last_used_at,omitempty"`
	ExpiresAt  *string  `json:"expires_at,omitempty"`
	CreatedAt  string   `json:"created_at"`
}

func (h *ListAPIKeysHandler) Handle(w http.ResponseWriter, r *http.Request) {
	userClaims := middleware.GetUserFromContext(r.Context())
	if userClaims == nil {
		common.Unauthorized(w, "authentication required")
		return
	}

	keys, err := h.apiKeyRepo.FindByUserID(r.Context(), userClaims.UserID)
	if err != nil {
		common.HandleError(w, err)
		return
	}

	response := make([]APIKeyResponse, 0, len(keys))
	for _, key := range keys {
		if key.IsRevoked() {
			continue
		}

		resp := APIKeyResponse{
			ID:        key.ID.String(),
			Name:      key.Name,
			KeyPrefix: key.KeyPrefix,
			Scopes:    key.Scopes,
			CreatedAt: key.CreatedAt.Format("2006-01-02T15:04:05Z"),
		}
		if key.LastUsedAt != nil {
			s := key.LastUsedAt.Format("2006-01-02T15:04:05Z")
			resp.LastUsedAt = &s
		}
		if key.ExpiresAt != nil {
			s := key.ExpiresAt.Format("2006-01-02T15:04:05Z")
			resp.ExpiresAt = &s
		}
		response = append(response, resp)
	}

	common.Success(w, response)
}
