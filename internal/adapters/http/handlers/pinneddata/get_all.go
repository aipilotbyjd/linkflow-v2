package pinneddata

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/core/domain/pinneddata"
)

// GetAllHandler handles get all pinned data request
type GetAllHandler struct {
	repo pinneddata.Repository
}

// NewGetAllHandler creates a new handler
func NewGetAllHandler(repo pinneddata.Repository) *GetAllHandler {
	return &GetAllHandler{repo: repo}
}

// Handle handles the get all pinned data request
func (h *GetAllHandler) Handle(w http.ResponseWriter, r *http.Request) {
	workflowIDStr := chi.URLParam(r, "workflowId")
	workflowID, err := uuid.Parse(workflowIDStr)
	if err != nil {
		common.BadRequest(w, "Invalid workflow ID")
		return
	}

	pinnedDataList, err := h.repo.GetByWorkflow(r.Context(), workflowID)
	if err != nil {
		common.HandleError(w, err)
		return
	}

	common.Success(w, map[string]interface{}{
		"pinnedData": ToPinnedDataResponseList(pinnedDataList),
	})
}
