package webhook

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/core/domain/webhook"
	"github.com/linkflow-ai/linkflow/internal/shared/events"
)

type CreateEndpointCommand struct {
	WorkflowID  uuid.UUID
	WorkspaceID uuid.UUID
	NodeID      string
	Path        string
	Method      string
}

type CreateEndpointHandler struct {
	webhookRepo webhook.Repository
	eventBus    events.Bus
}

func NewCreateEndpointHandler(webhookRepo webhook.Repository, eventBus events.Bus) *CreateEndpointHandler {
	return &CreateEndpointHandler{webhookRepo: webhookRepo, eventBus: eventBus}
}

func (h *CreateEndpointHandler) Handle(ctx context.Context, cmd CreateEndpointCommand) (*webhook.Endpoint, error) {
	exists, err := h.webhookRepo.ExistsByPath(ctx, cmd.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to check path: %w", err)
	}
	if exists {
		return nil, webhook.ErrPathAlreadyExists
	}

	endpoint, err := webhook.NewEndpoint(cmd.WorkflowID, cmd.WorkspaceID, cmd.NodeID, cmd.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to create endpoint entity: %w", err)
	}
	endpoint.Method = cmd.Method

	if err := h.webhookRepo.Create(ctx, endpoint); err != nil {
		return nil, fmt.Errorf("failed to create webhook endpoint: %w", err)
	}

	if h.eventBus != nil {
		_ = h.eventBus.Publish(ctx, events.WebhookTriggered{
			BaseEvent:  events.NewBaseEvent("webhook.created", endpoint.ID, "webhook"),
			EndpointID: endpoint.ID,
			WorkflowID: endpoint.WorkflowID,
			Method:     endpoint.Method,
		})
	}

	return endpoint, nil
}
