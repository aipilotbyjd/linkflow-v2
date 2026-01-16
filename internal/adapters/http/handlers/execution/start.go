package execution

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/middleware"
	executionCmd "github.com/linkflow-ai/linkflow/internal/core/application/command/execution"
	"github.com/linkflow-ai/linkflow/internal/shared/types"
)

// StartRequest represents execution start request
type StartRequest struct {
	InputData types.JSON `json:"input_data,omitempty"`
}

// StartHandler handles starting workflow execution
type StartHandler struct {
	handler *executionCmd.StartExecutionHandler
}

// NewStartHandler creates a new handler
func NewStartHandler(handler *executionCmd.StartExecutionHandler) *StartHandler {
	return &StartHandler{handler: handler}
}

// Handle handles the start execution request
func (h *StartHandler) Handle(w http.ResponseWriter, r *http.Request) {
	workflowIDStr := chi.URLParam(r, "workflowId")
	workflowID, err := uuid.Parse(workflowIDStr)
	if err != nil {
		common.BadRequest(w, "invalid workflow ID")
		return
	}

	wsCtx := middleware.GetWorkspaceFromContext(r.Context())
	if wsCtx == nil {
		common.BadRequest(w, "workspace context required")
		return
	}

	userClaims := middleware.GetUserFromContext(r.Context())
	if userClaims == nil {
		common.Unauthorized(w, "")
		return
	}

	var req StartRequest
	if r.ContentLength > 0 {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			common.BadRequest(w, "invalid request body")
			return
		}
	}

	exec, err := h.handler.Handle(r.Context(), executionCmd.StartExecutionCommand{
		WorkflowID:  workflowID,
		WorkspaceID: wsCtx.WorkspaceID,
		TriggeredBy: &userClaims.UserID,
		TriggerType: "manual",
		InputData:   req.InputData,
	})
	if err != nil {
		common.HandleError(w, err)
		return
	}

	common.Created(w, toExecutionResponse(exec))
}
