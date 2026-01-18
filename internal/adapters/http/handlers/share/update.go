package share

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/middleware"
	"github.com/linkflow-ai/linkflow/internal/core/domain/share"
	"github.com/linkflow-ai/linkflow/internal/infrastructure/validation"
	"gorm.io/gorm"
)

// UpdateHandler handles update share request
type UpdateHandler struct {
	repo share.Repository
}

// NewUpdateHandler creates a new handler
func NewUpdateHandler(repo share.Repository) *UpdateHandler {
	return &UpdateHandler{repo: repo}
}

// Handle handles the update share request
func (h *UpdateHandler) Handle(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	shareIDStr := chi.URLParam(r, "shareId")
	shareID, err := uuid.Parse(shareIDStr)
	if err != nil {
		common.BadRequest(w, "Invalid share ID")
		return
	}

	var req UpdateShareRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common.BadRequest(w, "Invalid request body")
		return
	}

	if errors := validation.Validate(req); len(errors) > 0 {
		details := make([]common.ValidationDetail, len(errors))
		for i, e := range errors {
			details[i] = common.ValidationDetail{Field: e.Field, Message: e.Message}
		}
		common.ValidationErrors(w, details)
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
		common.Forbidden(w, "Not authorized to update this share")
		return
	}

	shareObj.Permission = share.SharePermission(req.Permission)
	if err := h.repo.Update(r.Context(), shareObj); err != nil {
		common.HandleError(w, err)
		return
	}

	common.Success(w, ToShareResponse(*shareObj))
}
