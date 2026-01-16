package credential

import (
	"context"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/core/domain/credential"
)

type ShareCredentialCommand struct {
	CredentialID uuid.UUID
	UserID       uuid.UUID
	Permission   credential.Permission
	SharedBy     uuid.UUID
}

type ShareCredentialHandler struct {
	credentialRepo credential.Repository
	shareRepo      credential.ShareRepository
}

func NewShareCredentialHandler(credentialRepo credential.Repository, shareRepo credential.ShareRepository) *ShareCredentialHandler {
	return &ShareCredentialHandler{credentialRepo: credentialRepo, shareRepo: shareRepo}
}

func (h *ShareCredentialHandler) Handle(ctx context.Context, cmd ShareCredentialCommand) (*credential.Share, error) {
	cred, err := h.credentialRepo.FindByID(ctx, cmd.CredentialID)
	if err != nil {
		return nil, credential.ErrCredentialNotFound
	}

	share := credential.NewShare(cred.ID, cmd.UserID, cmd.SharedBy, cmd.Permission)

	if err := h.shareRepo.Create(ctx, share); err != nil {
		return nil, err
	}

	return share, nil
}
