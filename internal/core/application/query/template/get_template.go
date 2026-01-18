package template

import (
	"context"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/core/domain/template"
)

type GetTemplateQuery struct {
	TemplateID uuid.UUID
}

type GetTemplateHandler struct {
	repo template.Repository
}

func NewGetTemplateHandler(repo template.Repository) *GetTemplateHandler {
	return &GetTemplateHandler{repo: repo}
}

func (h *GetTemplateHandler) Handle(ctx context.Context, q GetTemplateQuery) (*template.Template, error) {
	return h.repo.FindByID(ctx, q.TemplateID)
}
