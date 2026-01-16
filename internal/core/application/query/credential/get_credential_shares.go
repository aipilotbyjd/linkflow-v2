package credential

import (
	"context"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/core/domain/credential"
)

type GetCredentialSharesQuery struct {
	CredentialID uuid.UUID
}

type GetCredentialSharesHandler struct {
	shareRepo credential.ShareRepository
}

func NewGetCredentialSharesHandler(shareRepo credential.ShareRepository) *GetCredentialSharesHandler {
	return &GetCredentialSharesHandler{shareRepo: shareRepo}
}

func (h *GetCredentialSharesHandler) Handle(ctx context.Context, q GetCredentialSharesQuery) ([]credential.Share, error) {
	return h.shareRepo.FindByCredentialID(ctx, q.CredentialID)
}
