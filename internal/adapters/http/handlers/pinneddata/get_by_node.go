package pinneddata

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/core/domain/pinneddata"
	"gorm.io/gorm"
)

// GetByNodeHandler handles get pinned data by node request
type GetByNodeHandler struct {
	repo pinneddata.Repository
}

// NewGetByNodeHandler creates a new handler
func NewGetByNodeHandler(repo pinneddata.Repository) *GetByNodeHandler {
	return &GetByNodeHandler{repo: repo}
}

// Handle handles the get pinned data by node request
func (h *GetByNodeHandler) Handle(w http.ResponseWriter, r *http.Request) {
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

	data, err := h.repo.GetByNode(r.Context(), workflowID, nodeID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			common.NotFound(w, "Pinned data")
			return
		}
		common.HandleError(w, err)
		return
	}

	common.Success(w, ToPinnedDataResponse(*data))
}
