package template

import (
	"context"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/shared/types"
)

type Repository interface {
	Create(ctx context.Context, template *Template) error
	Update(ctx context.Context, template *Template) error
	Delete(ctx context.Context, id uuid.UUID) error
	FindByID(ctx context.Context, id uuid.UUID) (*Template, error)
	FindAll(ctx context.Context, opts *ListOptions) ([]Template, int64, error)
	FindByCategory(ctx context.Context, category string, opts *ListOptions) ([]Template, int64, error)
	FindFeatured(ctx context.Context, limit int) ([]Template, error)
	Search(ctx context.Context, query string, opts *ListOptions) ([]Template, int64, error)
	IncrementUsage(ctx context.Context, id uuid.UUID) error
}

type ListOptions struct {
	*types.ListOptions
	Category   string
	Tags       []string
	IsFeatured *bool
	IsPublic   *bool
}
