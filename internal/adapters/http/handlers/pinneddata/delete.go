package pinneddata

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/core/domain/pinneddata"
)

// DeleteHandler handles delete pinned data request
type DeleteHandler struct {
	repo pinneddata.Repository
}

// NewDeleteHandler creates a new handler
func NewDeleteHandler(repo pinneddata.Repository) *DeleteHandler {
	return &DeleteHandler{repo: repo}
}

// Handle handles the delete pinned data request
func (h *DeleteHandler) Handle(w http.ResponseWriter, r *http.Request) {
	workflowIDStr := chi.URLParam(r, "workflowId")
	workflowID, err := uuid.Parse(workflowIDStr)
	if err != nil {
		common.BadRequest(w, "Invalid workflow ID")
		return
	}

	nodeID := chi.URLParam(r, "nodeId")
	if nodeID == "" {
		common.BadRequest(w, "Node ID is required")
		return
	}

	if err := h.repo.Delete(r.Context(), workflowID, nodeID); err != nil {
		common.HandleError(w, err)
		return
	}

	common.Success(w, map[string]interface{}{
		"message": "Pinned data deleted successfully",
	})
}
