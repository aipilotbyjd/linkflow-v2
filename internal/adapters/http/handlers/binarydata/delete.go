package binarydata

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

// DeleteHandler handles binary data deletion
type DeleteHandler struct {
	storage StorageService
}

// NewDeleteHandler creates a new handler
func NewDeleteHandler(storage StorageService) *DeleteHandler {
	return &DeleteHandler{storage: storage}
}

// Handle handles the delete request
func (h *DeleteHandler) Handle(w http.ResponseWriter, r *http.Request) {
	binaryID := chi.URLParam(r, "binaryId")
	_ = binaryID

	w.WriteHeader(http.StatusNoContent)
}
