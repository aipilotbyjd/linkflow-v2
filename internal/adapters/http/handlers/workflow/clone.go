package workflow

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/middleware"
	"github.com/linkflow-ai/linkflow/internal/core/domain/workflow"
)

type CloneRequest struct {
	Name string `json:"name" validate:"required"`
}

type CloneHandler struct {
	workflowRepo workflow.Repository
}

func NewCloneHandler(workflowRepo workflow.Repository) *CloneHandler {
	return &CloneHandler{workflowRepo: workflowRepo}
}

func (h *CloneHandler) Handle(w http.ResponseWriter, r *http.Request) {
	workflowIDStr := chi.URLParam(r, "workflowId")
	workflowID, err := uuid.Parse(workflowIDStr)
	if err != nil {
		common.BadRequest(w, "invalid workflow ID")
		return
	}

	var req CloneRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common.BadRequest(w, "invalid request body")
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

	original, err := h.workflowRepo.FindByID(r.Context(), workflowID)
	if err != nil {
		common.HandleError(w, err)
		return
	}

	cloned := workflow.NewWorkflow(wsCtx.WorkspaceID, userClaims.UserID, req.Name)
	cloned.Description = original.Description
	cloned.Nodes = original.Nodes
	cloned.Connections = original.Connections
	cloned.Settings = original.Settings
	cloned.Tags = original.Tags

	if err := h.workflowRepo.Create(r.Context(), cloned); err != nil {
		common.HandleError(w, err)
		return
	}

	common.Created(w, toWorkflowResponse(cloned))
}
