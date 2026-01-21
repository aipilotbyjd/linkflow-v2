package alerts

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/middleware"
	"github.com/linkflow-ai/linkflow/internal/core/domain/billing"
)

// DeleteHandler handles deleting usage alerts
type DeleteHandler struct {
	repo billing.UsageAlertRepository
}

func NewDeleteHandler(repo billing.UsageAlertRepository) *DeleteHandler {
	return &DeleteHandler{repo: repo}
}

func (h *DeleteHandler) Handle(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	workspaceID := middleware.GetWorkspaceID(ctx)
	
	alertID, err := uuid.Parse(chi.URLParam(r, "alertId"))
	if err != nil {
		common.BadRequest(w, "Invalid alert ID")
		return
	}

	// Verify ownership before deletion
	alert, err := h.repo.FindByID(ctx, alertID)
	if err != nil {
		common.HandleError(w, err)
		return
	}
	if alert == nil {
		common.NotFound(w, "Alert")
		return
	}
	if alert.WorkspaceID != workspaceID {
		common.Forbidden(w, "You don't have permission to delete this alert")
		return
	}

	if err := h.repo.Delete(ctx, alertID); err != nil {
		common.HandleError(w, err)
		return
	}

	common.NoContent(w)
}
