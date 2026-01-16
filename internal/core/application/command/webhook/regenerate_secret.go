package webhook

import (
	"context"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/core/domain/webhook"
)

type RegenerateSecretCommand struct {
	EndpointID uuid.UUID
}

type RegenerateSecretHandler struct {
	webhookRepo webhook.Repository
}

func NewRegenerateSecretHandler(webhookRepo webhook.Repository) *RegenerateSecretHandler {
	return &RegenerateSecretHandler{webhookRepo: webhookRepo}
}

func (h *RegenerateSecretHandler) Handle(ctx context.Context, cmd RegenerateSecretCommand) (*webhook.Endpoint, error) {
	endpoint, err := h.webhookRepo.FindByID(ctx, cmd.EndpointID)
	if err != nil {
		return nil, webhook.ErrEndpointNotFound
	}

	// Generate new secret (in practice, use crypto/rand)
	newSecret := uuid.New().String()
	endpoint.RegenerateSecret(newSecret)

	if err := h.webhookRepo.Update(ctx, endpoint); err != nil {
		return nil, err
	}

	return endpoint, nil
}
