package folder

import (
	"context"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/shared/types"
)

// Repository defines the interface for folder persistence
type Repository interface {
	Create(ctx context.Context, folder *Folder) error
	Update(ctx context.Context, folder *Folder) error
	Delete(ctx context.Context, id uuid.UUID) error
	FindByID(ctx context.Context, id uuid.UUID) (*Folder, error)
	FindByWorkspace(ctx context.Context, workspaceID uuid.UUID, opts *types.ListOptions) ([]*Folder, int64, error)
	FindByParent(ctx context.Context, workspaceID uuid.UUID, parentID *uuid.UUID) ([]*Folder, error)
	FindRootFolders(ctx context.Context, workspaceID uuid.UUID) ([]*Folder, error)
	HasChildren(ctx context.Context, folderID uuid.UUID) (bool, error)
	HasWorkflows(ctx context.Context, folderID uuid.UUID) (bool, error)
}

// ListOptions for folder queries
type ListOptions struct {
	*types.ListOptions
	ParentID *uuid.UUID
}
