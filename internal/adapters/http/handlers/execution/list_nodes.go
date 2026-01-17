package execution

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/middleware"
	"github.com/linkflow-ai/linkflow/internal/core/domain/execution"
)

type GetExecutionNodesHandler struct {
	executionRepo     execution.Repository
	nodeExecutionRepo execution.NodeExecutionRepository
}

func NewGetExecutionNodesHandler(
	executionRepo execution.Repository,
	nodeExecutionRepo execution.NodeExecutionRepository,
) *GetExecutionNodesHandler {
	return &GetExecutionNodesHandler{
		executionRepo:     executionRepo,
		nodeExecutionRepo: nodeExecutionRepo,
	}
}

func (h *GetExecutionNodesHandler) Handle(w http.ResponseWriter, r *http.Request) {
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

	// Verify execution belongs to workspace
	exec, err := h.executionRepo.FindByID(r.Context(), execID)
	if err != nil {
		common.HandleError(w, err)
		return
	}

	if exec.WorkspaceID != wsCtx.WorkspaceID {
		common.NotFound(w, "execution")
		return
	}

	// Get node executions
	nodeExecutions, err := h.nodeExecutionRepo.FindByExecutionID(r.Context(), execID)
	if err != nil {
		common.HandleError(w, err)
		return
	}

	common.Success(w, nodeExecutions)
}
