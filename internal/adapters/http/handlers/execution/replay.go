package execution

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/middleware"
	execCmd "github.com/linkflow-ai/linkflow/internal/core/application/command/execution"
	"github.com/linkflow-ai/linkflow/internal/core/domain/execution"
	"github.com/linkflow-ai/linkflow/internal/core/domain/workflow"
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

type ReplayRequest struct {
	UseOriginalInput bool                   `json:"useOriginalInput"`
	InputData        map[string]interface{} `json:"input_data,omitempty"`
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

// ReplayFromNodeHandler handles replaying from a specific node
type ReplayFromNodeHandler struct {
	executionRepo execution.Repository
	workflowRepo  workflow.Repository
	asynqClient   *asynq.Client
}

func NewReplayFromNodeHandler(
	executionRepo execution.Repository,
	workflowRepo workflow.Repository,
	asynqClient *asynq.Client,
) *ReplayFromNodeHandler {
	return &ReplayFromNodeHandler{
		executionRepo: executionRepo,
		workflowRepo:  workflowRepo,
		asynqClient:   asynqClient,
	}
}

type ReplayFromNodeRequest struct {
	NodeID string `json:"node_id"`
}

func (h *ReplayFromNodeHandler) Handle(w http.ResponseWriter, r *http.Request) {
	wsCtx := middleware.GetWorkspaceFromContext(r.Context())
	if wsCtx == nil {
		common.BadRequest(w, "workspace context required")
		return
	}

	execIDStr := chi.URLParam(r, "executionId")
	execID, err := uuid.Parse(execIDStr)
	if err != nil {
		common.BadRequest(w, "invalid execution ID")
		return
	}

	var req ReplayFromNodeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common.BadRequest(w, "invalid request body")
		return
	}

	if req.NodeID == "" {
		common.BadRequest(w, "node_id is required")
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

	// For now, return not implemented
	common.Success(w, map[string]interface{}{
		"message":      "Replay from node feature is in development",
		"execution_id": execID,
		"node_id":      req.NodeID,
	})
}
