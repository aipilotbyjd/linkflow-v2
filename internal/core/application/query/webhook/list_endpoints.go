package webhook

import (
	"context"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/core/domain/webhook"
	"github.com/linkflow-ai/linkflow/internal/shared/types"
)

type ListEndpointsQuery struct {
	WorkspaceID uuid.UUID
	WorkflowID  *uuid.UUID
	Page        int
	PageSize    int
}

type ListEndpointsResult struct {
	Endpoints []webhook.Endpoint
	Total     int64
}

type ListEndpointsHandler struct {
	webhookRepo webhook.Repository
}

func NewListEndpointsHandler(webhookRepo webhook.Repository) *ListEndpointsHandler {
	return &ListEndpointsHandler{webhookRepo: webhookRepo}
}

func (h *ListEndpointsHandler) Handle(ctx context.Context, query ListEndpointsQuery) (*ListEndpointsResult, error) {
	opts := types.NewListOptions(query.Page, query.PageSize)

	endpoints, total, err := h.webhookRepo.FindByWorkspaceID(ctx, query.WorkspaceID, opts)
	if err != nil {
		return nil, err
	}

	return &ListEndpointsResult{
		Endpoints: endpoints,
		Total:     total,
	}, nil
}
