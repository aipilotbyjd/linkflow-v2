package alerts

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/middleware"
	"github.com/linkflow-ai/linkflow/internal/core/domain/billing"
)

// AcknowledgeHandler handles acknowledging alerts
type AcknowledgeHandler struct {
	repo billing.UsageAlertLogRepository
}

func NewAcknowledgeHandler(repo billing.UsageAlertLogRepository) *AcknowledgeHandler {
	return &AcknowledgeHandler{repo: repo}
}

func (h *AcknowledgeHandler) Handle(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := middleware.GetUserID(ctx)

	alertLogID, err := uuid.Parse(chi.URLParam(r, "logId"))
	if err != nil {
		common.BadRequest(w, "Invalid alert log ID")
		return
	}

	if err := h.repo.Acknowledge(ctx, alertLogID, userID); err != nil {
		common.HandleError(w, err)
		return
	}

	common.Success(w, map[string]interface{}{
		"acknowledged": true,
	})
}
