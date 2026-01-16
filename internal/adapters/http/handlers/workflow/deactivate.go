package workflow

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/core/domain/workflow"
)

type DeactivateHandler struct {
	workflowRepo workflow.Repository
}

func NewDeactivateHandler(workflowRepo workflow.Repository) *DeactivateHandler {
	return &DeactivateHandler{workflowRepo: workflowRepo}
}

func (h *DeactivateHandler) Handle(w http.ResponseWriter, r *http.Request) {
	workflowIDStr := chi.URLParam(r, "workflowId")
	workflowID, err := uuid.Parse(workflowIDStr)
	if err != nil {
		common.BadRequest(w, "invalid workflow ID")
		return
	}

	wf, err := h.workflowRepo.FindByID(r.Context(), workflowID)
	if err != nil {
		common.HandleError(w, err)
		return
	}

	wf.Deactivate()

	if err := h.workflowRepo.Update(r.Context(), wf); err != nil {
		common.HandleError(w, err)
		return
	}

	common.Success(w, toWorkflowResponse(wf))
}
