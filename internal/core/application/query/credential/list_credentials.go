package credential

import (
	"context"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/core/domain/credential"
	"github.com/linkflow-ai/linkflow/internal/shared/types"
)

type ListCredentialsQuery struct {
	WorkspaceID uuid.UUID
	Type        *credential.Type
	Provider    *string
	Search      string
	Page        int
	PageSize    int
}

type ListCredentialsResult struct {
	Credentials []credential.Credential
	Total       int64
	Page        int
	PageSize    int
	TotalPages  int
}

type ListCredentialsHandler struct {
	credentialRepo credential.Repository
}

func NewListCredentialsHandler(credentialRepo credential.Repository) *ListCredentialsHandler {
	return &ListCredentialsHandler{credentialRepo: credentialRepo}
}

func (h *ListCredentialsHandler) Handle(ctx context.Context, q ListCredentialsQuery) (*ListCredentialsResult, error) {
	opts := &credential.ListOptions{
		ListOptions: types.NewListOptions(q.Page, q.PageSize),
		Type:        q.Type,
		Provider:    q.Provider,
		Search:      q.Search,
	}

	credentials, total, err := h.credentialRepo.FindByWorkspaceID(ctx, q.WorkspaceID, opts)
	if err != nil {
		return nil, err
	}

	pageSize := q.PageSize
	if pageSize <= 0 {
		pageSize = types.DefaultPageSize
	}
	totalPages := int(total) / pageSize
	if int(total)%pageSize != 0 {
		totalPages++
	}

	return &ListCredentialsResult{
		Credentials: credentials,
		Total:       total,
		Page:        q.Page,
		PageSize:    pageSize,
		TotalPages:  totalPages,
	}, nil
}
