package marketplace

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

// UnpublishHandler handles unpublish from marketplace request
type UnpublishHandler struct{}

// NewUnpublishHandler creates a new handler
func NewUnpublishHandler() *UnpublishHandler {
	return &UnpublishHandler{}
}

// Handle handles the unpublish request
func (h *UnpublishHandler) Handle(w http.ResponseWriter, r *http.Request) {
	templateID := chi.URLParam(r, "templateId")
	_ = templateID

	w.WriteHeader(http.StatusNoContent)
}
