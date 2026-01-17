package folder

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/middleware"
	"github.com/linkflow-ai/linkflow/internal/core/domain/folder"
)

type DeleteFolderHandler struct {
	folderRepo folder.Repository
}

func NewDeleteFolderHandler(folderRepo folder.Repository) *DeleteFolderHandler {
	return &DeleteFolderHandler{folderRepo: folderRepo}
}

func (h *DeleteFolderHandler) Handle(w http.ResponseWriter, r *http.Request) {
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

	f, err := h.folderRepo.FindByID(r.Context(), folderID)
	if err != nil {
		common.HandleError(w, err)
		return
	}

	if f.WorkspaceID != wsCtx.WorkspaceID {
		common.NotFound(w, "folder")
		return
	}

	// Check if folder has children
	hasChildren, err := h.folderRepo.HasChildren(r.Context(), folderID)
	if err != nil {
		common.HandleError(w, err)
		return
	}
	if hasChildren {
		common.BadRequest(w, "folder has sub-folders, delete them first")
		return
	}

	// Check if folder has workflows
	hasWorkflows, err := h.folderRepo.HasWorkflows(r.Context(), folderID)
	if err != nil {
		common.HandleError(w, err)
		return
	}
	if hasWorkflows {
		common.BadRequest(w, "folder contains workflows, move or delete them first")
		return
	}

	if err := h.folderRepo.Delete(r.Context(), folderID); err != nil {
		common.HandleError(w, err)
		return
	}

	common.NoContent(w)
}
