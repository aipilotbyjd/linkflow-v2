package pinneddata

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

// DeleteHandler handles delete pinned data request
type DeleteHandler struct {
	repo PinnedDataRepository
}

// NewDeleteHandler creates a new handler
func NewDeleteHandler(repo PinnedDataRepository) *DeleteHandler {
	return &DeleteHandler{repo: repo}
}

// Handle handles the delete pinned data request
func (h *DeleteHandler) Handle(w http.ResponseWriter, r *http.Request) {
	workflowID := chi.URLParam(r, "workflowId")
	nodeID := chi.URLParam(r, "nodeId")

	_ = workflowID
	_ = nodeID

	w.WriteHeader(http.StatusNoContent)
}
