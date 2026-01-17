package marketplace

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

// DeleteRatingHandler handles delete rating request
type DeleteRatingHandler struct{}

// NewDeleteRatingHandler creates a new handler
func NewDeleteRatingHandler() *DeleteRatingHandler {
	return &DeleteRatingHandler{}
}

// Handle handles the delete rating request
func (h *DeleteRatingHandler) Handle(w http.ResponseWriter, r *http.Request) {
	templateID := chi.URLParam(r, "templateId")
	_ = templateID

	w.WriteHeader(http.StatusNoContent)
}
