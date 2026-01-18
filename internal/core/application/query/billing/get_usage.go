package billing

import (
	"context"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/core/domain/billing"
)

type GetUsageQuery struct {
	WorkspaceID uuid.UUID
}

type GetUsageHandler struct {
	repo billing.UsageRepository
}

func NewGetUsageHandler(repo billing.UsageRepository) *GetUsageHandler {
	return &GetUsageHandler{repo: repo}
}

func (h *GetUsageHandler) Handle(ctx context.Context, q GetUsageQuery) (*billing.Usage, error) {
	return h.repo.GetCurrentPeriodUsage(ctx, q.WorkspaceID)
}
