package execution

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/middleware"
	executionCmd "github.com/linkflow-ai/linkflow/internal/core/application/command/execution"
	executionDomain "github.com/linkflow-ai/linkflow/internal/core/domain/execution"
)

type RetryHandler struct {
	handler       *executionCmd.StartExecutionHandler
	executionRepo executionDomain.Repository
}

func NewRetryHandler(handler *executionCmd.StartExecutionHandler, executionRepo executionDomain.Repository) *RetryHandler {
	return &RetryHandler{
		handler:       handler,
		executionRepo: executionRepo,
	}
}

func (h *RetryHandler) Handle(w http.ResponseWriter, r *http.Request) {
	executionIDStr := chi.URLParam(r, "executionId")
	executionID, err := uuid.Parse(executionIDStr)
	if err != nil {
		common.BadRequest(w, "invalid execution ID")
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

	// Get original execution
	originalExec, err := h.executionRepo.FindByID(r.Context(), executionID)
	if err != nil {
		common.NotFound(w, "Execution not found")
		return
	}

	// Verify workspace ownership
	if originalExec.WorkspaceID != wsCtx.WorkspaceID {
		common.Forbidden(w, "Execution does not belong to this workspace")
		return
	}

	// Create new execution with same input
	cmd := executionCmd.StartExecutionCommand{
		WorkflowID:  originalExec.WorkflowID,
		WorkspaceID: originalExec.WorkspaceID,
		TriggeredBy: &userClaims.UserID,
		TriggerType: "retry",
		InputData:   originalExec.InputData,
	}

	result, err := h.handler.Handle(r.Context(), cmd)
	if err != nil {
		common.HandleError(w, err)
		return
	}

	common.Success(w, map[string]interface{}{
		"status":           "retry queued",
		"execution_id":     result.ID.String(),
		"original_exec_id": executionID.String(),
	})
}
