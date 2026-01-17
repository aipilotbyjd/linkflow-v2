package execution

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/middleware"
	"github.com/linkflow-ai/linkflow/internal/core/domain/execution"
)

type GetNodeHandler struct {
	executionRepo     execution.Repository
	nodeExecutionRepo execution.NodeExecutionRepository
}

func NewGetNodeHandler(executionRepo execution.Repository, nodeExecutionRepo execution.NodeExecutionRepository) *GetNodeHandler {
	return &GetNodeHandler{
		executionRepo:     executionRepo,
		nodeExecutionRepo: nodeExecutionRepo,
	}
}

func (h *GetNodeHandler) Handle(w http.ResponseWriter, r *http.Request) {
	executionIDStr := chi.URLParam(r, "executionId")
	executionID, err := uuid.Parse(executionIDStr)
	if err != nil {
		common.BadRequest(w, "invalid execution ID")
		return
	}

	nodeID := chi.URLParam(r, "nodeId")
	if nodeID == "" {
		common.BadRequest(w, "node ID is required")
		return
	}

	workspaceID := middleware.GetWorkspaceID(r.Context())

	exec, err := h.executionRepo.FindByID(r.Context(), executionID)
	if err != nil {
		common.HandleError(w, err)
		return
	}

	if exec.WorkspaceID.String() != workspaceID.String() {
		common.Forbidden(w, "access denied")
		return
	}

	nodeExec, err := h.nodeExecutionRepo.FindByExecutionAndNodeID(r.Context(), executionID, nodeID)
	if err != nil {
		common.HandleError(w, err)
		return
	}

	var duration *int64
	if nodeExec.StartedAt != nil && nodeExec.CompletedAt != nil {
		d := nodeExec.CompletedAt.Sub(*nodeExec.StartedAt).Milliseconds()
		duration = &d
	}

	var nodeName string
	if nodeExec.NodeName != nil {
		nodeName = *nodeExec.NodeName
	}

	response := NodeExecutionResponse{
		ID:          nodeExec.ID.String(),
		ExecutionID: nodeExec.ExecutionID.String(),
		NodeID:      nodeExec.NodeID,
		NodeType:    nodeExec.NodeType,
		NodeName:    nodeName,
		Status:      string(nodeExec.Status),
		StartedAt:   nodeExec.StartedAt,
		CompletedAt: nodeExec.CompletedAt,
		Duration:    duration,
		InputData:   nodeExec.InputData,
		OutputData:  nodeExec.OutputData,
		Error:       nodeExec.ErrorMessage,
		RetryCount:  nodeExec.RetryCount,
	}

	common.Success(w, response)
}
