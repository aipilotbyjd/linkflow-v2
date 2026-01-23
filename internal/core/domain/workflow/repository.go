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
	Status     *Status
	Tags       []string
	FolderID   *uuid.UUID
	Search     string
	IsFavorite *bool
	CreatedBy  *uuid.UUID
}

// NewListOptions creates default list options
func NewListOptions(page, perPage int) *ListOptions {
	return &ListOptions{
		ListOptions: types.NewListOptions(page, perPage),
	}
}

// AdvancedSearchOptions for advanced workflow search
type AdvancedSearchOptions struct {
	*types.ListOptions

	// Text search
	Query    string   // Search in name, description
	SearchIn []string // Fields to search: name, description, nodes, tags

	// Filters
	Status     []Status // Multiple statuses
	Tags       []string // Match any tag
	TagsAll    []string // Match all tags
	NodeTypes  []string // Workflows containing these node types
	Category   string
	IsFavorite *bool
	FolderID   *uuid.UUID
	CreatedBy  *uuid.UUID

	// Date filters
	CreatedAfter   *int64 // Unix timestamp
	CreatedBefore  *int64
	UpdatedAfter   *int64
	UpdatedBefore  *int64
	ExecutedAfter  *int64
	ExecutedBefore *int64

	// Execution filters
	MinExecutions *int
	MaxExecutions *int
	HasErrors     *bool // Has error workflow configured

	// Sorting
	SortBy    string // name, created_at, updated_at, execution_count, last_executed_at
	SortOrder string // asc, desc
}

// NewAdvancedSearchOptions creates default advanced search options
func NewAdvancedSearchOptions(page, perPage int) *AdvancedSearchOptions {
	return &AdvancedSearchOptions{
		ListOptions: types.NewListOptions(page, perPage),
		SearchIn:    []string{"name", "description"},
		SortBy:      "updated_at",
		SortOrder:   "desc",
	}
}

// SearchRepository extends Repository with advanced search
type SearchRepository interface {
	Repository
	AdvancedSearch(ctx context.Context, workspaceID uuid.UUID, opts *AdvancedSearchOptions) ([]Workflow, int64, error)
	GetNodeTypesInWorkspace(ctx context.Context, workspaceID uuid.UUID) ([]string, error)
	GetTagsInWorkspace(ctx context.Context, workspaceID uuid.UUID) ([]string, error)
	GetCategoriesInWorkspace(ctx context.Context, workspaceID uuid.UUID) ([]string, error)
}
