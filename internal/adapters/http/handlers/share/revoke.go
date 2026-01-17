package share

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

// RevokeHandler handles revoke share request
type RevokeHandler struct {
	repo ShareRepository
}

// NewRevokeHandler creates a new handler
func NewRevokeHandler(repo ShareRepository) *RevokeHandler {
	return &RevokeHandler{repo: repo}
}

// Handle handles the revoke share request
func (h *RevokeHandler) Handle(w http.ResponseWriter, r *http.Request) {
	shareID := chi.URLParam(r, "shareId")
	_ = shareID

	w.WriteHeader(http.StatusNoContent)
}
