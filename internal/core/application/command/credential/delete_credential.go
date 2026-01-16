package credential

import (
	"context"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/core/domain/credential"
)

type DeleteCredentialCommand struct {
	CredentialID uuid.UUID
}

type DeleteCredentialHandler struct {
	credentialRepo credential.Repository
}

func NewDeleteCredentialHandler(credentialRepo credential.Repository) *DeleteCredentialHandler {
	return &DeleteCredentialHandler{credentialRepo: credentialRepo}
}

func (h *DeleteCredentialHandler) Handle(ctx context.Context, cmd DeleteCredentialCommand) error {
	_, err := h.credentialRepo.FindByID(ctx, cmd.CredentialID)
	if err != nil {
		return credential.ErrCredentialNotFound
	}

	return h.credentialRepo.Delete(ctx, cmd.CredentialID)
}
