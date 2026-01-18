package billing

import (
	"github.com/linkflow-ai/linkflow/internal/core/domain/billing"
)

type GetPlansHandler struct{}

func NewGetPlansHandler() *GetPlansHandler {
	return &GetPlansHandler{}
}

func (h *GetPlansHandler) Handle() []billing.Plan {
	return billing.GetAllPlans()
}
