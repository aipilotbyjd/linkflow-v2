package pinneddata

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
)

// SetPinnedDataRequest represents set pinned data request
type SetPinnedDataRequest struct {
	NodeID string          `json:"nodeId"`
	Data   json.RawMessage `json:"data"`
}

// SetHandler handles set pinned data request
type SetHandler struct {
	repo PinnedDataRepository
}

// NewSetHandler creates a new handler
func NewSetHandler(repo PinnedDataRepository) *SetHandler {
	return &SetHandler{repo: repo}
}

// Handle handles the set pinned data request
func (h *SetHandler) Handle(w http.ResponseWriter, r *http.Request) {
	workflowID := chi.URLParam(r, "workflowId")

	var req SetPinnedDataRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common.BadRequest(w, "invalid request body")
		return
	}

	if req.NodeID == "" {
		common.BadRequest(w, "node ID is required")
		return
	}

	if len(req.Data) == 0 {
		common.BadRequest(w, "data is required")
		return
	}

	pinnedData := PinnedData{
		ID:         uuid.New().String(),
		WorkflowID: workflowID,
		NodeID:     req.NodeID,
		Data:       req.Data,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	common.Created(w, pinnedData)
}
