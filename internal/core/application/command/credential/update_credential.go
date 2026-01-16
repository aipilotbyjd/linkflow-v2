package credential

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/core/domain/credential"
	"github.com/linkflow-ai/linkflow/internal/shared/types"
)

type UpdateCredentialCommand struct {
	CredentialID uuid.UUID
	Name         *string
	Description  *string
	Data         types.JSON
	SharingScope *credential.SharingScope
}

type UpdateCredentialHandler struct {
	credentialRepo credential.Repository
}

func NewUpdateCredentialHandler(credentialRepo credential.Repository) *UpdateCredentialHandler {
	return &UpdateCredentialHandler{credentialRepo: credentialRepo}
}

func (h *UpdateCredentialHandler) Handle(ctx context.Context, cmd UpdateCredentialCommand) (*credential.Credential, error) {
	cred, err := h.credentialRepo.FindByID(ctx, cmd.CredentialID)
	if err != nil {
		return nil, credential.ErrCredentialNotFound
	}

	if cmd.Name != nil {
		cred.Name = *cmd.Name
	}
	if cmd.Description != nil {
		cred.Description = cmd.Description
	}
	if cmd.SharingScope != nil {
		cred.SharingScope = *cmd.SharingScope
	}

	if err := h.credentialRepo.Update(ctx, cred); err != nil {
		return nil, fmt.Errorf("failed to update credential: %w", err)
	}

	return cred, nil
}
