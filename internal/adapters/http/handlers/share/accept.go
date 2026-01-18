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

// AcceptHandler handles accept share request
type AcceptHandler struct {
	repo share.Repository
}

// NewAcceptHandler creates a new handler
func NewAcceptHandler(repo share.Repository) *AcceptHandler {
	return &AcceptHandler{repo: repo}
}

// Handle handles the accept share request
func (h *AcceptHandler) Handle(w http.ResponseWriter, r *http.Request) {
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

	// Verify user is the recipient
	if shareObj.SharedWithID != userID {
		common.Forbidden(w, "Not authorized to accept this share")
		return
	}

	shareObj.Accept()
	if err := h.repo.Update(r.Context(), shareObj); err != nil {
		common.HandleError(w, err)
		return
	}

	common.Success(w, ToShareResponse(*shareObj))
}
