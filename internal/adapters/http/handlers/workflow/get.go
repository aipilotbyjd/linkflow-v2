package workflow

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	workflowQuery "github.com/linkflow-ai/linkflow/internal/core/application/query/workflow"
)

// GetHandler handles getting a single workflow
type GetHandler struct {
	handler *workflowQuery.GetWorkflowHandler
}

// NewGetHandler creates a new get handler
func NewGetHandler(handler *workflowQuery.GetWorkflowHandler) *GetHandler {
	return &GetHandler{handler: handler}
}

// Handle handles the get workflow request
func (h *GetHandler) Handle(w http.ResponseWriter, r *http.Request) {
	workflowIDStr := chi.URLParam(r, "workflowId")
	workflowID, err := uuid.Parse(workflowIDStr)
	if err != nil {
		common.BadRequest(w, "invalid workflow ID")
		return
	}

	wf, err := h.handler.Handle(r.Context(), workflowQuery.GetWorkflowQuery{
		WorkflowID: workflowID,
	})
	if err != nil {
		common.HandleError(w, err)
		return
	}

	common.Success(w, ToWorkflowResponse(wf))
}
