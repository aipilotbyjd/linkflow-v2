package workflow

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/core/domain/workflow"
)

type DeleteHandler struct {
	workflowRepo workflow.Repository
}

func NewDeleteHandler(workflowRepo workflow.Repository) *DeleteHandler {
	return &DeleteHandler{workflowRepo: workflowRepo}
}

func (h *DeleteHandler) Handle(w http.ResponseWriter, r *http.Request) {
	workflowIDStr := chi.URLParam(r, "workflowId")
	workflowID, err := uuid.Parse(workflowIDStr)
	if err != nil {
		common.BadRequest(w, "invalid workflow ID")
		return
	}

	if err := h.workflowRepo.Delete(r.Context(), workflowID); err != nil {
		common.HandleError(w, err)
		return
	}

	common.NoContent(w)
}
