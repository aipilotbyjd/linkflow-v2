package folder

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/middleware"
	"github.com/linkflow-ai/linkflow/internal/core/domain/folder"
)

type GetFolderHandler struct {
	folderRepo folder.Repository
}

func NewGetFolderHandler(folderRepo folder.Repository) *GetFolderHandler {
	return &GetFolderHandler{folderRepo: folderRepo}
}

func (h *GetFolderHandler) Handle(w http.ResponseWriter, r *http.Request) {
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

	common.Success(w, f)
}
