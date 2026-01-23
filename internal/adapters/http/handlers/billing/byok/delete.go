package byok

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/middleware"
)

// DeleteHandler handles deleting BYOK configurations
type DeleteHandler struct {
	repo BYOKRepository
}

func NewDeleteHandler(repo BYOKRepository) *DeleteHandler {
	return &DeleteHandler{repo: repo}
}

func (h *DeleteHandler) Handle(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	workspaceID := middleware.GetWorkspaceID(ctx)

	configID, err := uuid.Parse(chi.URLParam(r, "configId"))
	if err != nil {
		common.BadRequest(w, "Invalid configuration ID")
		return
	}

	// Verify ownership before deletion
	config, err := h.repo.FindByID(ctx, configID)
	if err != nil {
		common.HandleError(w, err)
		return
	}
	if config == nil {
		common.NotFound(w, "BYOK configuration")
		return
	}
	if config.WorkspaceID != workspaceID {
		common.Forbidden(w, "You don't have permission to delete this configuration")
		return
	}

	if err := h.repo.Delete(ctx, configID); err != nil {
		common.HandleError(w, err)
		return
	}

	common.NoContent(w)
}
