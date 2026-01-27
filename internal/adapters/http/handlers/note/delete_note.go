package note

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/middleware"
	"github.com/linkflow-ai/linkflow/internal/core/domain/note"
)

type DeleteNoteHandler struct {
	repo note.Repository
}

func NewDeleteNoteHandler(repo note.Repository) *DeleteNoteHandler {
	return &DeleteNoteHandler{repo: repo}
}

func (h *DeleteNoteHandler) Handle(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	noteID, err := uuid.Parse(chi.URLParam(r, "noteId"))
	if err != nil {
		common.BadRequest(w, "invalid note ID")
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
		common.Forbidden(w, "you can only delete your own notes")
		return
	}

	// Delete
	if err := h.repo.Delete(r.Context(), noteID); err != nil {
		common.HandleError(w, err)
		return
	}

	common.NoContent(w)
}
