package workspace

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/core/domain/workspace"
)

type DeleteHandler struct {
	workspaceRepo workspace.Repository
}

func NewDeleteHandler(workspaceRepo workspace.Repository) *DeleteHandler {
	return &DeleteHandler{workspaceRepo: workspaceRepo}
}

func (h *DeleteHandler) Handle(w http.ResponseWriter, r *http.Request) {
	workspaceIDStr := chi.URLParam(r, "workspaceId")
	workspaceID, err := uuid.Parse(workspaceIDStr)
	if err != nil {
		common.BadRequest(w, "invalid workspace ID")
		return
	}

	if err := h.workspaceRepo.Delete(r.Context(), workspaceID); err != nil {
		common.HandleError(w, err)
		return
	}

	common.NoContent(w)
}
