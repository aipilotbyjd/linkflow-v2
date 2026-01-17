package folder

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/middleware"
	"github.com/linkflow-ai/linkflow/internal/core/domain/folder"
)

type UpdateFolderHandler struct {
	folderRepo folder.Repository
}

func NewUpdateFolderHandler(folderRepo folder.Repository) *UpdateFolderHandler {
	return &UpdateFolderHandler{folderRepo: folderRepo}
}

type UpdateFolderRequest struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	Color       *string `json:"color,omitempty"`
}

func (h *UpdateFolderHandler) Handle(w http.ResponseWriter, r *http.Request) {
	wsCtx := middleware.GetWorkspaceFromContext(r.Context())
	if wsCtx == nil {
		common.BadRequest(w, "workspace context required")
		return
	}

	folderIDStr := chi.URLParam(r, "folderId")
	folderID, err := uuid.Parse(folderIDStr)
	if err != nil {
		common.BadRequest(w, "invalid folder ID")
		return
	}

	var req UpdateFolderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common.BadRequest(w, "invalid request body")
		return
	}

	f, err := h.folderRepo.FindByID(r.Context(), folderID)
	if err != nil {
		common.HandleError(w, err)
		return
	}

	if f.WorkspaceID != wsCtx.WorkspaceID {
		common.NotFound(w, "folder")
		return
	}

	name := f.Name
	if req.Name != nil {
		name = *req.Name
	}

	f.Update(name, req.Description, req.Color)

	if err := h.folderRepo.Update(r.Context(), f); err != nil {
		common.HandleError(w, err)
		return
	}

	common.Success(w, f)
}
