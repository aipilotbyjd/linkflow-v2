package share

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/middleware"
	"github.com/linkflow-ai/linkflow/internal/core/domain/share"
	"gorm.io/gorm"
)

// RevokeHandler handles revoke share request
type RevokeHandler struct {
	repo share.Repository
}

// NewRevokeHandler creates a new handler
func NewRevokeHandler(repo share.Repository) *RevokeHandler {
	return &RevokeHandler{repo: repo}
}

// Handle handles the revoke share request
func (h *RevokeHandler) Handle(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	shareIDStr := chi.URLParam(r, "shareId")
	shareID, err := uuid.Parse(shareIDStr)
	if err != nil {
		common.BadRequest(w, "Invalid share ID")
		return
	}

	shareObj, err := h.repo.FindByID(r.Context(), shareID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			common.NotFound(w, "Share")
			return
		}
		common.HandleError(w, err)
		return
	}

	// Verify user is the owner
	if shareObj.SharedByID != userID {
		common.Forbidden(w, "Not authorized to revoke this share")
		return
	}

	if err := h.repo.Delete(r.Context(), shareID); err != nil {
		common.HandleError(w, err)
		return
	}

	common.Success(w, map[string]interface{}{
		"message": "Share revoked successfully",
	})
}
