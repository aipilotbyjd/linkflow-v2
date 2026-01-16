package billing

import (
	"net/http"

	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/middleware"
	"github.com/linkflow-ai/linkflow/internal/core/domain/billing"
)

type SubscriptionResponse struct {
	ID              string  `json:"id"`
	WorkspaceID     string  `json:"workspace_id"`
	PlanID          string  `json:"plan_id"`
	Status          string  `json:"status"`
	CurrentPeriodStart string `json:"current_period_start"`
	CurrentPeriodEnd   string `json:"current_period_end"`
	CancelAtPeriodEnd  bool   `json:"cancel_at_period_end"`
	CreatedAt       string  `json:"created_at"`
}

type GetSubscriptionHandler struct {
	subscriptionRepo billing.SubscriptionRepository
}

func NewGetSubscriptionHandler(subscriptionRepo billing.SubscriptionRepository) *GetSubscriptionHandler {
	return &GetSubscriptionHandler{subscriptionRepo: subscriptionRepo}
}

func (h *GetSubscriptionHandler) Handle(w http.ResponseWriter, r *http.Request) {
	wsCtx := middleware.GetWorkspaceFromContext(r.Context())
	if wsCtx == nil {
		common.BadRequest(w, "workspace context required")
		return
	}

	sub, err := h.subscriptionRepo.FindByWorkspaceID(r.Context(), wsCtx.WorkspaceID)
	if err != nil {
		common.HandleError(w, err)
		return
	}

	common.Success(w, toSubscriptionResponse(sub))
}

func toSubscriptionResponse(s *billing.Subscription) SubscriptionResponse {
	return SubscriptionResponse{
		ID:                 s.ID.String(),
		WorkspaceID:        s.WorkspaceID.String(),
		PlanID:             s.PlanID,
		Status:             string(s.Status),
		CurrentPeriodStart: s.CurrentPeriodStart.Format("2006-01-02T15:04:05Z"),
		CurrentPeriodEnd:   s.CurrentPeriodEnd.Format("2006-01-02T15:04:05Z"),
		CancelAtPeriodEnd:  s.CancelAtPeriodEnd,
		CreatedAt:          s.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
}
