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
	templateRepo template.Repository
}

func NewGetTemplateHandler(templateRepo template.Repository) *GetTemplateHandler {
	return &GetTemplateHandler{templateRepo: templateRepo}
}

func (h *GetTemplateHandler) Handle(ctx context.Context, query GetTemplateQuery) (*template.Template, error) {
	return h.templateRepo.FindByID(ctx, query.TemplateID)
}
