package template

import (
	"context"

	"github.com/linkflow-ai/linkflow/internal/core/domain/template"
)

type GetFeaturedQuery struct {
	Limit int
}

type GetFeaturedHandler struct {
	repo template.Repository
}

func NewGetFeaturedHandler(repo template.Repository) *GetFeaturedHandler {
	return &GetFeaturedHandler{repo: repo}
}

func (h *GetFeaturedHandler) Handle(ctx context.Context, q GetFeaturedQuery) ([]template.Template, error) {
	limit := q.Limit
	if limit <= 0 {
		limit = 10
	}
	return h.repo.FindFeatured(ctx, limit)
}
