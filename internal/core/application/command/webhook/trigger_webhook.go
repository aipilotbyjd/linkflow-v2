package webhook

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/core/domain/webhook"
	"github.com/linkflow-ai/linkflow/internal/shared/types"
)

type TriggerWebhookCommand struct {
	Path       string
	Method     string
	Headers    map[string]string
	Body       types.JSON
	QueryParams map[string]string
	IPAddress  string
}

type TriggerWebhookResult struct {
	EndpointID  uuid.UUID
	WorkflowID  uuid.UUID
	ExecutionID uuid.UUID
}

type TriggerWebhookHandler struct {
	webhookRepo webhook.Repository
	eventRepo   webhook.EventRepository
}

func NewTriggerWebhookHandler(webhookRepo webhook.Repository, eventRepo webhook.EventRepository) *TriggerWebhookHandler {
	return &TriggerWebhookHandler{webhookRepo: webhookRepo, eventRepo: eventRepo}
}

func (h *TriggerWebhookHandler) Handle(ctx context.Context, cmd TriggerWebhookCommand) (*TriggerWebhookResult, error) {
	endpoint, err := h.webhookRepo.FindByPath(ctx, cmd.Path)
	if err != nil {
		return nil, webhook.ErrEndpointNotFound
	}

	if !endpoint.IsActive {
		return nil, webhook.ErrEndpointInactive
	}

	if endpoint.Method != "" && endpoint.Method != cmd.Method {
		return nil, fmt.Errorf("method not allowed")
	}

	// Log the webhook event
	event := webhook.NewEvent(endpoint.ID, cmd.Method, endpoint.Path)
	event.WithIPAddress(cmd.IPAddress)
	if err := h.eventRepo.Create(ctx, event); err != nil {
		// Non-fatal, continue
	}

	// Record the call
	if err := h.webhookRepo.RecordCall(ctx, endpoint.ID); err != nil {
		// Non-fatal
	}

	// Workflow execution is triggered asynchronously via the event bus
	// The actual execution is handled by the worker service

	return &TriggerWebhookResult{
		EndpointID: endpoint.ID,
		WorkflowID: endpoint.WorkflowID,
	}, nil
}
