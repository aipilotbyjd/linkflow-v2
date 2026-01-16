package credential

import (
	"context"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/shared/types"
)

// Repository defines the interface for credential persistence
type Repository interface {
	Create(ctx context.Context, credential *Credential) error
	Update(ctx context.Context, credential *Credential) error
	Delete(ctx context.Context, id uuid.UUID) error
	FindByID(ctx context.Context, id uuid.UUID) (*Credential, error)
	FindByIDWithShares(ctx context.Context, id uuid.UUID) (*Credential, error)
	FindByWorkspaceID(ctx context.Context, workspaceID uuid.UUID, opts *ListOptions) ([]Credential, int64, error)
	FindByType(ctx context.Context, workspaceID uuid.UUID, credType Type) ([]Credential, error)
	FindByProvider(ctx context.Context, workspaceID uuid.UUID, provider string) ([]Credential, error)
	ExistsByName(ctx context.Context, workspaceID uuid.UUID, name string) (bool, error)
	UpdateLastUsed(ctx context.Context, id uuid.UUID) error
	CountByWorkspaceID(ctx context.Context, workspaceID uuid.UUID) (int64, error)
	FindExpiring(ctx context.Context, withinDuration string) ([]Credential, error)
}

// ShareRepository defines the interface for credential share persistence
type ShareRepository interface {
	Create(ctx context.Context, share *Share) error
	Delete(ctx context.Context, id uuid.UUID) error
	FindByID(ctx context.Context, id uuid.UUID) (*Share, error)
	FindByCredentialID(ctx context.Context, credentialID uuid.UUID) ([]Share, error)
	FindByUserID(ctx context.Context, userID uuid.UUID) ([]Share, error)
	FindByCredentialAndUser(ctx context.Context, credentialID, userID uuid.UUID) (*Share, error)
	DeleteByCredentialID(ctx context.Context, credentialID uuid.UUID) error
}

// ListOptions for credential queries
type ListOptions struct {
	*types.ListOptions
	Type         *Type
	Provider     *string
	SharingScope *SharingScope
	Search       string
	CreatedBy    *uuid.UUID
}

// NewListOptions creates default list options
func NewListOptions(page, perPage int) *ListOptions {
	return &ListOptions{
		ListOptions: types.NewListOptions(page, perPage),
	}
}

// EncryptionService defines the interface for credential encryption
type EncryptionService interface {
	Encrypt(data []byte) (string, error)
	Decrypt(encrypted string) ([]byte, error)
}
