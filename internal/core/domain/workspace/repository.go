package workspace

import (
	"context"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/shared/types"
)

// Repository defines the interface for workspace persistence
type Repository interface {
	Create(ctx context.Context, workspace *Workspace) error
	Update(ctx context.Context, workspace *Workspace) error
	Delete(ctx context.Context, id uuid.UUID) error
	FindByID(ctx context.Context, id uuid.UUID) (*Workspace, error)
	FindBySlug(ctx context.Context, slug string) (*Workspace, error)
	FindByOwnerID(ctx context.Context, ownerID uuid.UUID) ([]Workspace, error)
	ExistsBySlug(ctx context.Context, slug string) (bool, error)
	FindAll(ctx context.Context, opts *types.ListOptions) ([]Workspace, int64, error)
}

// MemberRepository defines the interface for member persistence
type MemberRepository interface {
	Create(ctx context.Context, member *Member) error
	Update(ctx context.Context, member *Member) error
	Delete(ctx context.Context, id uuid.UUID) error
	FindByID(ctx context.Context, id uuid.UUID) (*Member, error)
	FindByWorkspaceAndUser(ctx context.Context, workspaceID, userID uuid.UUID) (*Member, error)
	FindByWorkspaceID(ctx context.Context, workspaceID uuid.UUID, opts *types.ListOptions) ([]Member, int64, error)
	FindByUserID(ctx context.Context, userID uuid.UUID) ([]Member, error)
	FindWorkspacesByUserID(ctx context.Context, userID uuid.UUID) ([]Workspace, error)
	CountByWorkspaceID(ctx context.Context, workspaceID uuid.UUID) (int64, error)
	IsMember(ctx context.Context, workspaceID, userID uuid.UUID) (bool, error)
	DeleteByWorkspaceAndUser(ctx context.Context, workspaceID, userID uuid.UUID) error
}

// InvitationRepository defines the interface for invitation persistence
type InvitationRepository interface {
	Create(ctx context.Context, invitation *Invitation) error
	Update(ctx context.Context, invitation *Invitation) error
	Delete(ctx context.Context, id uuid.UUID) error
	FindByID(ctx context.Context, id uuid.UUID) (*Invitation, error)
	FindByToken(ctx context.Context, token string) (*Invitation, error)
	FindByWorkspaceID(ctx context.Context, workspaceID uuid.UUID) ([]Invitation, error)
	FindByEmail(ctx context.Context, email string) ([]Invitation, error)
	FindPendingByWorkspaceAndEmail(ctx context.Context, workspaceID uuid.UUID, email string) (*Invitation, error)
	CleanupExpired(ctx context.Context) (int64, error)
}

// ListOptions for workspace queries
type ListOptions struct {
	*types.ListOptions
	OwnerID *uuid.UUID
	PlanID  *string
	Search  string
}
