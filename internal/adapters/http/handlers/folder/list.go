package folder

import (
	"net/http"

	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/middleware"
	"github.com/linkflow-ai/linkflow/internal/core/domain/folder"
)

type ListFoldersHandler struct {
	folderRepo folder.Repository
}

func NewListFoldersHandler(folderRepo folder.Repository) *ListFoldersHandler {
	return &ListFoldersHandler{folderRepo: folderRepo}
}

func (h *ListFoldersHandler) Handle(w http.ResponseWriter, r *http.Request) {
	wsCtx := middleware.GetWorkspaceFromContext(r.Context())
	if wsCtx == nil {
		common.BadRequest(w, "workspace context required")
		return
	}

	folders, _, err := h.folderRepo.FindByWorkspace(r.Context(), wsCtx.WorkspaceID, nil)
	if err != nil {
		common.HandleError(w, err)
		return
	}

	common.Success(w, folders)
}
