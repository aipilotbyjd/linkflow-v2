package apikey

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/middleware"
	"github.com/linkflow-ai/linkflow/internal/core/domain/user"
)

type RevokeAPIKeyHandler struct {
	apiKeyRepo user.APIKeyRepository
}

func NewRevokeAPIKeyHandler(apiKeyRepo user.APIKeyRepository) *RevokeAPIKeyHandler {
	return &RevokeAPIKeyHandler{apiKeyRepo: apiKeyRepo}
}

func (h *RevokeAPIKeyHandler) Handle(w http.ResponseWriter, r *http.Request) {
	userClaims := middleware.GetUserFromContext(r.Context())
	if userClaims == nil {
		common.Unauthorized(w, "authentication required")
		return
	}

	keyIDStr := chi.URLParam(r, "keyId")
	keyID, err := uuid.Parse(keyIDStr)
	if err != nil {
		common.BadRequest(w, "invalid API key ID")
		return
	}

	// Verify key belongs to user
	key, err := h.apiKeyRepo.FindByID(r.Context(), keyID)
	if err != nil {
		common.NotFound(w, "API key")
		return
	}

	if key.UserID != userClaims.UserID {
		common.NotFound(w, "API key")
		return
	}

	if err := h.apiKeyRepo.Revoke(r.Context(), keyID); err != nil {
		common.HandleError(w, err)
		return
	}

	common.NoContent(w)
}
