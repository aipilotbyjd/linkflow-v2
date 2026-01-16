package workflow

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	workflowCmd "github.com/linkflow-ai/linkflow/internal/core/application/command/workflow"
)

type ActivateHandler struct {
	handler *workflowCmd.ActivateWorkflowHandler
}

func NewActivateHandler(handler *workflowCmd.ActivateWorkflowHandler) *ActivateHandler {
	return &ActivateHandler{handler: handler}
}

func (h *ActivateHandler) Handle(w http.ResponseWriter, r *http.Request) {
	workflowIDStr := chi.URLParam(r, "workflowId")
	workflowID, err := uuid.Parse(workflowIDStr)
	if err != nil {
		common.BadRequest(w, "invalid workflow ID")
		return
	}

	if err := h.handler.Handle(r.Context(), workflowCmd.ActivateWorkflowCommand{
		WorkflowID: workflowID,
	}); err != nil {
		common.HandleError(w, err)
		return
	}

	common.Success(w, map[string]string{"status": "activated"})
}
