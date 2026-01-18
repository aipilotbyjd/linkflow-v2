package billing

import (
	"context"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/core/domain/billing"
	"github.com/linkflow-ai/linkflow/internal/shared/events"
)

type CreateSubscriptionCommand struct {
	WorkspaceID uuid.UUID
	PlanID      string
	CreatedBy   uuid.UUID
}

type CreateSubscriptionHandler struct {
	repo     billing.SubscriptionRepository
	eventBus events.Bus
}

func NewCreateSubscriptionHandler(repo billing.SubscriptionRepository, eventBus events.Bus) *CreateSubscriptionHandler {
	return &CreateSubscriptionHandler{repo: repo, eventBus: eventBus}
}

func (h *CreateSubscriptionHandler) Handle(ctx context.Context, cmd CreateSubscriptionCommand) (*billing.Subscription, error) {
	plan := billing.GetPlan(cmd.PlanID)
	if plan == nil {
		return nil, billing.ErrPlanNotFound
	}

	sub := billing.NewSubscription(cmd.WorkspaceID, cmd.PlanID)

	if err := h.repo.Create(ctx, sub); err != nil {
		return nil, err
	}

	if h.eventBus != nil {
		_ = h.eventBus.Publish(ctx, events.NewBaseEvent("subscription.created", sub.ID, "subscription"))
	}

	return sub, nil
}
