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

type DuplicateRequest struct {
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	FolderID    *string `json:"folderId,omitempty"`
}

type DuplicateHandler struct {
	workflowRepo workflow.Repository
}

func NewDuplicateHandler(workflowRepo workflow.Repository) *DuplicateHandler {
	return &DuplicateHandler{workflowRepo: workflowRepo}
}

func (h *DuplicateHandler) Handle(w http.ResponseWriter, r *http.Request) {
	workflowIDStr := chi.URLParam(r, "workflowId")
	workflowID, err := uuid.Parse(workflowIDStr)
	if err != nil {
		common.BadRequest(w, "invalid workflow ID")
		return
	}

	userID := middleware.GetUserID(r.Context())
	workspaceID := middleware.GetWorkspaceID(r.Context())

	var req DuplicateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common.BadRequest(w, "invalid request body")
		return
	}

	original, err := h.workflowRepo.FindByID(r.Context(), workflowID)
	if err != nil {
		common.HandleError(w, err)
		return
	}

	if original.WorkspaceID.String() != workspaceID.String() {
		common.Forbidden(w, "access denied")
		return
	}

	name := req.Name
	if name == "" {
		name = original.Name + " (Copy)"
	}

	duplicated := workflow.NewWorkflow(original.WorkspaceID, userID, name)

	duplicated.Description = original.Description
	if req.Description != nil {
		duplicated.Description = req.Description
	}

	duplicated.Nodes = original.Nodes
	duplicated.Connections = original.Connections
	duplicated.Settings = original.Settings

	if req.FolderID != nil {
		folderID, err := uuid.Parse(*req.FolderID)
		if err == nil {
			duplicated.FolderID = &folderID
		}
	} else if original.FolderID != nil {
		duplicated.FolderID = original.FolderID
	}

	if err := h.workflowRepo.Create(r.Context(), duplicated); err != nil {
		common.HandleError(w, err)
		return
	}

	common.Created(w, map[string]interface{}{
		"id":        duplicated.ID.String(),
		"name":      duplicated.Name,
		"sourceId":  original.ID.String(),
		"createdAt": duplicated.CreatedAt,
	})
}
