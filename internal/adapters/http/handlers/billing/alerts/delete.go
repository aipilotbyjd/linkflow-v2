package alerts

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
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
	alertID, err := uuid.Parse(chi.URLParam(r, "alertId"))
	if err != nil {
		common.BadRequest(w, "Invalid alert ID")
		return
	}

	if err := h.repo.Delete(ctx, alertID); err != nil {
		common.HandleError(w, err)
		return
	}

	common.NoContent(w)
}
