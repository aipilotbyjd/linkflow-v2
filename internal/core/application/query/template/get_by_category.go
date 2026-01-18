package template

import (
	"context"

	"github.com/linkflow-ai/linkflow/internal/core/domain/template"
	"github.com/linkflow-ai/linkflow/internal/shared/types"
)

type GetByCategoryQuery struct {
	Category string
	Limit    int
	Offset   int
}

type GetByCategoryHandler struct {
	repo template.Repository
}

func NewGetByCategoryHandler(repo template.Repository) *GetByCategoryHandler {
	return &GetByCategoryHandler{repo: repo}
}

func (h *GetByCategoryHandler) Handle(ctx context.Context, q GetByCategoryQuery) ([]template.Template, int64, error) {
	opts := &template.ListOptions{
		ListOptions: &types.ListOptions{
			Limit:  q.Limit,
			Offset: q.Offset,
		},
	}
	return h.repo.FindByCategory(ctx, q.Category, opts)
}
