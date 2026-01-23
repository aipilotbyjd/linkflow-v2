package workflow

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/middleware"
	"github.com/linkflow-ai/linkflow/internal/core/domain/workflow"
	"github.com/linkflow-ai/linkflow/internal/infrastructure/validation"
)

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
	// Body is optional
	if r.ContentLength > 0 {
		if err := json.NewDecoder(r.Body).Decode(&req); err == nil {
			if errors := validation.Validate(req); len(errors) > 0 {
				details := make([]common.ValidationDetail, len(errors))
				for i, e := range errors {
					details[i] = common.ValidationDetail{Field: e.Field, Message: e.Message}
				}
				common.ValidationErrors(w, details)
				return
			}
		}
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

	duplicated, err := workflow.NewWorkflow(original.WorkspaceID, userID, name)
	if err != nil {
		common.HandleError(w, err)
		return
	}

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
