package workflow

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	workflowCmd "github.com/linkflow-ai/linkflow/internal/core/application/command/workflow"
	"github.com/linkflow-ai/linkflow/internal/shared/types"
)

type UpdateRequest struct {
	Name        *string         `json:"name,omitempty"`
	Description *string         `json:"description,omitempty"`
	Nodes       types.JSONArray `json:"nodes,omitempty"`
	Connections types.JSONArray `json:"connections,omitempty"`
	Settings    types.JSON      `json:"settings,omitempty"`
}

type UpdateHandler struct {
	handler *workflowCmd.UpdateWorkflowHandler
}

func NewUpdateHandler(handler *workflowCmd.UpdateWorkflowHandler) *UpdateHandler {
	return &UpdateHandler{handler: handler}
}

func (h *UpdateHandler) Handle(w http.ResponseWriter, r *http.Request) {
	workflowIDStr := chi.URLParam(r, "workflowId")
	workflowID, err := uuid.Parse(workflowIDStr)
	if err != nil {
		common.BadRequest(w, "invalid workflow ID")
		return
	}

	var req UpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common.BadRequest(w, "invalid request body")
		return
	}

	wf, err := h.handler.Handle(r.Context(), workflowCmd.UpdateWorkflowCommand{
		WorkflowID:  workflowID,
		Name:        req.Name,
		Description: req.Description,
		Nodes:       req.Nodes,
		Connections: req.Connections,
		Settings:    req.Settings,
	})
	if err != nil {
		common.HandleError(w, err)
		return
	}

	common.Success(w, toWorkflowResponse(wf))
}
