package execution

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/middleware"
	execCmd "github.com/linkflow-ai/linkflow/internal/core/application/command/execution"
	"github.com/linkflow-ai/linkflow/internal/core/domain/execution"
	"github.com/linkflow-ai/linkflow/internal/core/domain/workflow"
	"github.com/linkflow-ai/linkflow/internal/infrastructure/validation"
)

type ReplayExecutionHandler struct {
	executionRepo    execution.Repository
	workflowRepo     workflow.Repository
	startExecHandler *execCmd.StartExecutionHandler
}

func NewReplayExecutionHandler(
	executionRepo execution.Repository,
	workflowRepo workflow.Repository,
	startExecHandler *execCmd.StartExecutionHandler,
) *ReplayExecutionHandler {
	return &ReplayExecutionHandler{
		executionRepo:    executionRepo,
		workflowRepo:     workflowRepo,
		startExecHandler: startExecHandler,
	}
}

func (h *ReplayExecutionHandler) Handle(w http.ResponseWriter, r *http.Request) {
	wsCtx := middleware.GetWorkspaceFromContext(r.Context())
	if wsCtx == nil {
		common.BadRequest(w, "workspace context required")
		return
	}

	userClaims := middleware.GetUserFromContext(r.Context())
	if userClaims == nil {
		common.Unauthorized(w, "authentication required")
		return
	}

	execIDStr := chi.URLParam(r, "executionId")
	execID, err := uuid.Parse(execIDStr)
	if err != nil {
		common.BadRequest(w, "invalid execution ID")
		return
	}

	var req ReplayRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		req.UseOriginalInput = true
	} else if errors := validation.Validate(req); len(errors) > 0 {
		details := make([]common.ValidationDetail, len(errors))
		for i, e := range errors {
			details[i] = common.ValidationDetail{Field: e.Field, Message: e.Message}
		}
		common.ValidationErrors(w, details)
		return
	}

	// Get original execution
	originalExec, err := h.executionRepo.FindByID(r.Context(), execID)
	if err != nil {
		common.HandleError(w, err)
		return
	}

	if originalExec.WorkspaceID != wsCtx.WorkspaceID {
		common.NotFound(w, "execution")
		return
	}

	// Determine input data
	inputData := req.InputData
	if req.UseOriginalInput && originalExec.InputData != nil {
		inputData = originalExec.InputData
	}

	// Start new execution
	result, err := h.startExecHandler.Handle(r.Context(), execCmd.StartExecutionCommand{
		WorkflowID:  originalExec.WorkflowID,
		WorkspaceID: wsCtx.WorkspaceID,
		TriggerType: "manual",
		TriggeredBy: &userClaims.UserID,
		InputData:   inputData,
	})
	if err != nil {
		common.HandleError(w, err)
		return
	}

	common.Success(w, map[string]interface{}{
		"execution_id":          result.ID,
		"original_execution_id": execID,
		"status":                result.Status,
		"message":               "Execution replay started",
	})
}
