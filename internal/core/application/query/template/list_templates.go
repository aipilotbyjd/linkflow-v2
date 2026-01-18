package template

import (
	"context"

	"github.com/linkflow-ai/linkflow/internal/core/domain/template"
	"github.com/linkflow-ai/linkflow/internal/shared/types"
)

type ListTemplatesQuery struct {
	Limit  int
	Offset int
}

type ListTemplatesHandler struct {
	repo template.Repository
}

func NewListTemplatesHandler(repo template.Repository) *ListTemplatesHandler {
	return &ListTemplatesHandler{repo: repo}
}

func (h *ListTemplatesHandler) Handle(ctx context.Context, q ListTemplatesQuery) ([]template.Template, int64, error) {
	opts := &template.ListOptions{
		ListOptions: &types.ListOptions{
			Limit:  q.Limit,
			Offset: q.Offset,
		},
	}
	return h.repo.FindAll(ctx, opts)
}
