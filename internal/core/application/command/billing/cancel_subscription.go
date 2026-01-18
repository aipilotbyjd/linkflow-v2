package billing

import (
	"context"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/core/domain/billing"
)

type CancelSubscriptionCommand struct {
	WorkspaceID uuid.UUID
}

type CancelSubscriptionHandler struct {
	repo billing.SubscriptionRepository
}

func NewCancelSubscriptionHandler(repo billing.SubscriptionRepository) *CancelSubscriptionHandler {
	return &CancelSubscriptionHandler{repo: repo}
}

func (h *CancelSubscriptionHandler) Handle(ctx context.Context, cmd CancelSubscriptionCommand) error {
	sub, err := h.repo.FindByWorkspaceID(ctx, cmd.WorkspaceID)
	if err != nil {
		return err
	}

	sub.Cancel()

	return h.repo.Update(ctx, sub)
}
