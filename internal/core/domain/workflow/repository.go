package workflow

import (
	"context"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/shared/types"
)

// Repository defines the interface for workflow persistence
type Repository interface {
	Create(ctx context.Context, workflow *Workflow) error
	Update(ctx context.Context, workflow *Workflow) error
	Delete(ctx context.Context, id uuid.UUID) error
	FindByID(ctx context.Context, id uuid.UUID) (*Workflow, error)
	FindByWorkspaceID(ctx context.Context, workspaceID uuid.UUID, opts *ListOptions) ([]Workflow, int64, error)
	FindByFolderID(ctx context.Context, folderID uuid.UUID, opts *types.ListOptions) ([]Workflow, int64, error)
	ExistsByName(ctx context.Context, workspaceID uuid.UUID, name string) (bool, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status Status) error
	IncrementExecutionCount(ctx context.Context, id uuid.UUID) error
	CountByWorkspaceID(ctx context.Context, workspaceID uuid.UUID) (int64, error)
	CountActiveByWorkspaceID(ctx context.Context, workspaceID uuid.UUID) (int64, error)
}

// VersionRepository defines the interface for version persistence
type VersionRepository interface {
	Create(ctx context.Context, version *Version) error
	FindByID(ctx context.Context, id uuid.UUID) (*Version, error)
	FindByWorkflowID(ctx context.Context, workflowID uuid.UUID, opts *types.ListOptions) ([]Version, int64, error)
	FindByWorkflowAndVersion(ctx context.Context, workflowID uuid.UUID, version int) (*Version, error)
	FindLatestByWorkflowID(ctx context.Context, workflowID uuid.UUID) (*Version, error)
	Delete(ctx context.Context, id uuid.UUID) error
	DeleteByWorkflowID(ctx context.Context, workflowID uuid.UUID) error
}

// FolderRepository defines the interface for folder persistence
type FolderRepository interface {
	Create(ctx context.Context, folder *Folder) error
	Update(ctx context.Context, folder *Folder) error
	Delete(ctx context.Context, id uuid.UUID) error
	FindByID(ctx context.Context, id uuid.UUID) (*Folder, error)
	FindByWorkspaceID(ctx context.Context, workspaceID uuid.UUID) ([]Folder, error)
	FindByParentID(ctx context.Context, workspaceID uuid.UUID, parentID *uuid.UUID) ([]Folder, error)
	ExistsByName(ctx context.Context, workspaceID uuid.UUID, parentID *uuid.UUID, name string) (bool, error)
}

// ListOptions for workflow queries
type ListOptions struct {
	*types.ListOptions
	Status    *Status
	Tags      []string
	FolderID  *uuid.UUID
	Search    string
	IsFavorite *bool
	CreatedBy *uuid.UUID
}

// NewListOptions creates default list options
func NewListOptions(page, perPage int) *ListOptions {
	return &ListOptions{
		ListOptions: types.NewListOptions(page, perPage),
	}
}
