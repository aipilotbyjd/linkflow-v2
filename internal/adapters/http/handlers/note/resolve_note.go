package note

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/middleware"
	"github.com/linkflow-ai/linkflow/internal/core/domain/note"
)

type ResolveNoteHandler struct {
	repo note.Repository
}

func NewResolveNoteHandler(repo note.Repository) *ResolveNoteHandler {
	return &ResolveNoteHandler{repo: repo}
}

func (h *ResolveNoteHandler) Handle(w http.ResponseWriter, r *http.Request) {
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

	// Resolve
	if err := c.Resolve(userID); err != nil {
		common.HandleError(w, err)
		return
	}

	if err := h.repo.Update(r.Context(), c); err != nil {
		common.HandleError(w, err)
		return
	}

	common.Success(w, ToResponse(c))
}
