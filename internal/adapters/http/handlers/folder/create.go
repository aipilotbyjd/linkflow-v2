package folder

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/middleware"
	"github.com/linkflow-ai/linkflow/internal/core/domain/folder"
	"github.com/linkflow-ai/linkflow/internal/infrastructure/validation"
)

type CreateFolderHandler struct {
	folderRepo folder.Repository
}

func NewCreateFolderHandler(folderRepo folder.Repository) *CreateFolderHandler {
	return &CreateFolderHandler{folderRepo: folderRepo}
}

type CreateFolderRequest struct {
	Name        string     `json:"name" validate:"required"`
	Description *string    `json:"description,omitempty"`
	Color       *string    `json:"color,omitempty"`
	ParentID    *uuid.UUID `json:"parent_id,omitempty"`
}

func (h *CreateFolderHandler) Handle(w http.ResponseWriter, r *http.Request) {
	wsCtx := middleware.GetWorkspaceFromContext(r.Context())
	if wsCtx == nil {
		common.BadRequest(w, "Workspace context required")
		return
	}

	userClaims := middleware.GetUserFromContext(r.Context())
	if userClaims == nil {
		common.Unauthorized(w, "Authentication required")
		return
	}

	var req CreateFolderRequest
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

	f := folder.NewFolder(wsCtx.WorkspaceID, req.Name, userClaims.UserID)

	if req.ParentID != nil {
		f = f.WithParent(*req.ParentID)
	}
	if req.Description != nil {
		f = f.WithDescription(*req.Description)
	}
	if req.Color != nil {
		f = f.WithColor(*req.Color)
	}

	if err := h.folderRepo.Create(r.Context(), f); err != nil {
		common.HandleError(w, err)
		return
	}

	common.Created(w, f)
}
