package billing

import (
	"context"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/core/domain/billing"
)

type GetInvoicesQuery struct {
	WorkspaceID uuid.UUID
}

type GetInvoicesHandler struct {
	repo billing.InvoiceRepository
}

func NewGetInvoicesHandler(repo billing.InvoiceRepository) *GetInvoicesHandler {
	return &GetInvoicesHandler{repo: repo}
}

func (h *GetInvoicesHandler) Handle(ctx context.Context, q GetInvoicesQuery) ([]billing.Invoice, error) {
	invoices, _, err := h.repo.FindByWorkspaceID(ctx, q.WorkspaceID, 100, 0)
	return invoices, err
}
