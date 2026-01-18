package execution

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/middleware"
	"github.com/linkflow-ai/linkflow/internal/core/domain/execution"
	"github.com/linkflow-ai/linkflow/internal/core/domain/workflow"
	"github.com/linkflow-ai/linkflow/internal/infrastructure/validation"
)

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
		common.BadRequest(w, "Invalid request body")
		return
	}

	if errors := validation.Validate(req); len(errors) > 0 {
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

	// For now, return not implemented
	common.Success(w, map[string]interface{}{
		"message":      "Replay from node feature is in development",
		"execution_id": execID,
		"node_id":      req.NodeID,
	})
}
