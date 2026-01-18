package template

import (
	"context"

	"github.com/linkflow-ai/linkflow/internal/core/domain/template"
	"github.com/linkflow-ai/linkflow/internal/shared/types"
)

type SearchTemplatesQuery struct {
	Query  string
	Limit  int
	Offset int
}

type SearchTemplatesHandler struct {
	repo template.Repository
}

func NewSearchTemplatesHandler(repo template.Repository) *SearchTemplatesHandler {
	return &SearchTemplatesHandler{repo: repo}
}

func (h *SearchTemplatesHandler) Handle(ctx context.Context, q SearchTemplatesQuery) ([]template.Template, int64, error) {
	opts := &template.ListOptions{
		ListOptions: &types.ListOptions{
			Limit:  q.Limit,
			Offset: q.Offset,
		},
	}
	return h.repo.Search(ctx, q.Query, opts)
}
