package webhook

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/core/domain/webhook"
	"github.com/linkflow-ai/linkflow/internal/infrastructure/security"
	"github.com/linkflow-ai/linkflow/internal/shared/types"
)

type TriggerWebhookCommand struct {
	Path        string
	Method      string
	Headers     map[string]string
	Body        types.JSON
	RawBody     string
	QueryParams map[string]string
	IPAddress   string
}

type TriggerWebhookResult struct {
	EndpointID  uuid.UUID
	WorkflowID  uuid.UUID
	ExecutionID uuid.UUID
}

type TriggerWebhookHandler struct {
	webhookRepo       webhook.Repository
	eventRepo         webhook.EventRepository
	securityValidator *security.WebhookValidator
}

func NewTriggerWebhookHandler(webhookRepo webhook.Repository, eventRepo webhook.EventRepository) *TriggerWebhookHandler {
	return &TriggerWebhookHandler{
		webhookRepo:       webhookRepo,
		eventRepo:         eventRepo,
		securityValidator: security.NewWebhookValidator(),
	}
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

	// Security validation
	if endpoint.IsSecured() {
		secConfig := &security.WebhookSecurityConfig{
			AllowedIPs:         endpoint.GetAllowedIPsList(),
			RequireTimestamp:   endpoint.RequireTimestamp,
			TimestampHeader:    endpoint.TimestampHeader,
			TimestampMaxAgeSec: int64(endpoint.TimestampMaxAgeSec),
			RequireNonce:       endpoint.RequireNonce,
			NonceHeader:        endpoint.NonceHeader,
		}

		// Add signature config if secret is set
		if endpoint.RequireSignature && endpoint.HasSecret() {
			secConfig.Secret = *endpoint.Secret
			secConfig.SignatureHeader = endpoint.SignatureHeader
		}

		secReq := &security.WebhookRequest{
			Headers:   cmd.Headers,
			RawBody:   cmd.RawBody,
			IPAddress: cmd.IPAddress,
		}

		result := h.securityValidator.Validate(secConfig, secReq)
		if !result.Valid {
			// Log the security failure
			if h.eventRepo != nil {
				event := webhook.NewEvent(endpoint.ID, cmd.Method, endpoint.Path)
				event.WithIPAddress(cmd.IPAddress)
				event.MarkFailed(fmt.Sprintf("security validation failed: %s", result.ErrorCode))
				if err := h.eventRepo.Create(ctx, event); err != nil {
					// Non-fatal, continue
				}
			}
			return nil, fmt.Errorf("webhook security validation failed: %s", result.Error)
		}
	}

	// Log the webhook event (if event repo is available)
	if h.eventRepo != nil {
		event := webhook.NewEvent(endpoint.ID, cmd.Method, endpoint.Path)
		event.WithIPAddress(cmd.IPAddress)
		if err := h.eventRepo.Create(ctx, event); err != nil {
			// Non-fatal, continue
		}
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
