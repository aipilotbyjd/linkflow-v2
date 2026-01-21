package credential

import (
	"context"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/core/domain/credential"
	"github.com/linkflow-ai/linkflow/internal/shared/errors"
)

type GetCredentialQuery struct {
	CredentialID uuid.UUID
	WorkspaceID  uuid.UUID // Required for access control
}

type GetCredentialHandler struct {
	credentialRepo credential.Repository
}

func NewGetCredentialHandler(credentialRepo credential.Repository) *GetCredentialHandler {
	return &GetCredentialHandler{credentialRepo: credentialRepo}
}

func (h *GetCredentialHandler) Handle(ctx context.Context, q GetCredentialQuery) (*credential.Credential, error) {
	cred, err := h.credentialRepo.FindByID(ctx, q.CredentialID)
	if err != nil {
		return nil, err
	}

	// Verify workspace ownership
	if cred.WorkspaceID != q.WorkspaceID {
		return nil, errors.NewForbiddenError("credential does not belong to this workspace")
	}

	return cred, nil
}
