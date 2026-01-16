package credential

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/core/domain/credential"
	"github.com/linkflow-ai/linkflow/internal/shared/events"
	"github.com/linkflow-ai/linkflow/internal/shared/types"
)

type CreateCredentialCommand struct {
	WorkspaceID  uuid.UUID
	CreatedBy    uuid.UUID
	Name         string
	Description  *string
	Type         credential.Type
	Provider     string
	Data         types.JSON
	SharingScope credential.SharingScope
}

type CreateCredentialHandler struct {
	credentialRepo credential.Repository
	eventBus       events.Bus
}

func NewCreateCredentialHandler(credentialRepo credential.Repository, eventBus events.Bus) *CreateCredentialHandler {
	return &CreateCredentialHandler{credentialRepo: credentialRepo, eventBus: eventBus}
}

func (h *CreateCredentialHandler) Handle(ctx context.Context, cmd CreateCredentialCommand) (*credential.Credential, error) {
	if cmd.Name == "" {
		return nil, fmt.Errorf("credential name is required")
	}

	exists, err := h.credentialRepo.ExistsByName(ctx, cmd.WorkspaceID, cmd.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to check credential name: %w", err)
	}
	if exists {
		return nil, credential.ErrCredentialNameExists
	}

	cred := credential.NewCredential(cmd.WorkspaceID, cmd.CreatedBy, cmd.Name, cmd.Type, cmd.Provider)
	cred.Description = cmd.Description
	cred.SharingScope = cmd.SharingScope

	if err := h.credentialRepo.Create(ctx, cred); err != nil {
		return nil, fmt.Errorf("failed to create credential: %w", err)
	}

	if h.eventBus != nil {
		_ = h.eventBus.Publish(ctx, events.CredentialCreated{
			BaseEvent:    events.NewBaseEvent("credential.created", cred.ID, "credential"),
			CredentialID: cred.ID,
			WorkspaceID:  cred.WorkspaceID,
			Type:         string(cred.Type),
			CreatedBy:    cred.CreatedBy,
		})
	}

	return cred, nil
}
