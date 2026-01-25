package note

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/middleware"
	"github.com/linkflow-ai/linkflow/internal/core/domain/note"
)

type UpdateNoteHandler struct {
	repo note.Repository
}

func NewUpdateNoteHandler(repo note.Repository) *UpdateNoteHandler {
	return &UpdateNoteHandler{repo: repo}
}

func (h *UpdateNoteHandler) Handle(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	noteID, err := uuid.Parse(chi.URLParam(r, "noteID"))
	if err != nil {
		common.BadRequest(w, "invalid note ID")
		return
	}

	var req UpdateNoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common.BadRequest(w, "invalid request body")
		return
	}

	// Validate
	if req.Content == "" {
		common.BadRequest(w, "content is required")
		return
	}
	if len(req.Content) > 10000 {
		common.BadRequest(w, "content too long (max 10000 characters)")
		return
	}

	// Find note
	c, err := h.repo.FindByID(r.Context(), noteID)
	if err != nil {
		common.HandleError(w, err)
		return
	}

	// Check ownership
	if c.UserID != userID {
		common.Forbidden(w, "you can only update your own notes")
		return
	}

	// Update
	if err := c.Update(req.Content); err != nil {
		common.HandleError(w, err)
		return
	}

	if err := h.repo.Update(r.Context(), c); err != nil {
		common.HandleError(w, err)
		return
	}

	common.Success(w, ToResponse(c))
}
