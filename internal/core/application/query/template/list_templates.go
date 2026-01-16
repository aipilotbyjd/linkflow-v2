package template

import (
	"context"

	"github.com/linkflow-ai/linkflow/internal/core/domain/template"
	"github.com/linkflow-ai/linkflow/internal/shared/types"
)

type ListTemplatesQuery struct {
	Category string
	Search   string
	Featured *bool
	Page     int
	PageSize int
}

type ListTemplatesResult struct {
	Templates []template.Template
	Total     int64
}

type ListTemplatesHandler struct {
	templateRepo template.Repository
}

func NewListTemplatesHandler(templateRepo template.Repository) *ListTemplatesHandler {
	return &ListTemplatesHandler{templateRepo: templateRepo}
}

func (h *ListTemplatesHandler) Handle(ctx context.Context, query ListTemplatesQuery) (*ListTemplatesResult, error) {
	opts := &template.ListOptions{
		ListOptions: types.NewListOptions(query.Page, query.PageSize),
		Category:    query.Category,
		IsFeatured:  query.Featured,
	}

	templates, total, err := h.templateRepo.FindAll(ctx, opts)
	if err != nil {
		return nil, err
	}

	return &ListTemplatesResult{
		Templates: templates,
		Total:     total,
	}, nil
}
