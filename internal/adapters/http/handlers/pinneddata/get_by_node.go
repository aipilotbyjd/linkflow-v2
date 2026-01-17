package pinneddata

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
)

// GetByNodeHandler handles get pinned data by node request
type GetByNodeHandler struct {
	repo PinnedDataRepository
}

// NewGetByNodeHandler creates a new handler
func NewGetByNodeHandler(repo PinnedDataRepository) *GetByNodeHandler {
	return &GetByNodeHandler{repo: repo}
}

// Handle handles the get pinned data by node request
func (h *GetByNodeHandler) Handle(w http.ResponseWriter, r *http.Request) {
	workflowID := chi.URLParam(r, "workflowId")
	nodeID := chi.URLParam(r, "nodeId")

	pinnedData := PinnedData{
		ID:         uuid.New().String(),
		WorkflowID: workflowID,
		NodeID:     nodeID,
		Data:       json.RawMessage(`{"key": "sample data for node"}`),
		CreatedAt:  time.Now().AddDate(0, 0, -1),
		UpdatedAt:  time.Now(),
	}

	common.Success(w, pinnedData)
}
