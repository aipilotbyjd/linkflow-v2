package pinneddata

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/middleware"
)

// GetAllHandler handles get all pinned data request
type GetAllHandler struct {
	repo PinnedDataRepository
}

// NewGetAllHandler creates a new handler
func NewGetAllHandler(repo PinnedDataRepository) *GetAllHandler {
	return &GetAllHandler{repo: repo}
}

// Handle handles the get all pinned data request
func (h *GetAllHandler) Handle(w http.ResponseWriter, r *http.Request) {
	workflowID := chi.URLParam(r, "workflowId")
	_ = middleware.GetWorkspaceID(r.Context())

	pinnedData := []PinnedData{
		{
			ID:         uuid.New().String(),
			WorkflowID: workflowID,
			NodeID:     "node-1",
			Data:       json.RawMessage(`{"key": "sample data"}`),
			CreatedAt:  time.Now().AddDate(0, 0, -1),
			UpdatedAt:  time.Now(),
		},
	}

	common.Success(w, map[string]interface{}{
		"pinnedData": pinnedData,
		"workflowId": workflowID,
	})
}
