package pinneddata

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/core/domain/pinneddata"
)

// SetHandler handles set pinned data request
type SetHandler struct {
	repo pinneddata.Repository
}

// NewSetHandler creates a new handler
func NewSetHandler(repo pinneddata.Repository) *SetHandler {
	return &SetHandler{repo: repo}
}

// Handle handles the set pinned data request
func (h *SetHandler) Handle(w http.ResponseWriter, r *http.Request) {
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

	var req SetPinnedDataRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common.BadRequest(w, "Invalid request body")
		return
	}

	data, err := h.repo.Set(r.Context(), workflowID, nodeID, req.Data)
	if err != nil {
		common.HandleError(w, err)
		return
	}

	common.Success(w, ToPinnedDataResponse(*data))
}
