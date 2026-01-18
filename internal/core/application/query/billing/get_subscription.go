package billing

import (
	"context"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/core/domain/billing"
)

type GetSubscriptionQuery struct {
	WorkspaceID uuid.UUID
}

type GetSubscriptionHandler struct {
	repo billing.SubscriptionRepository
}

func NewGetSubscriptionHandler(repo billing.SubscriptionRepository) *GetSubscriptionHandler {
	return &GetSubscriptionHandler{repo: repo}
}

func (h *GetSubscriptionHandler) Handle(ctx context.Context, q GetSubscriptionQuery) (*billing.Subscription, error) {
	return h.repo.FindByWorkspaceID(ctx, q.WorkspaceID)
}
