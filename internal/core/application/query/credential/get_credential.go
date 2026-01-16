package credential

import (
	"context"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/core/domain/credential"
)

type GetCredentialQuery struct {
	CredentialID uuid.UUID
}

type GetCredentialHandler struct {
	credentialRepo credential.Repository
}

func NewGetCredentialHandler(credentialRepo credential.Repository) *GetCredentialHandler {
	return &GetCredentialHandler{credentialRepo: credentialRepo}
}

func (h *GetCredentialHandler) Handle(ctx context.Context, q GetCredentialQuery) (*credential.Credential, error) {
	return h.credentialRepo.FindByID(ctx, q.CredentialID)
}
