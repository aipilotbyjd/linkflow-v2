package billing

import (
	"net/http"
	"time"

	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/middleware"
)

// GetInvoicesHandler handles get invoices request
type GetInvoicesHandler struct {
	service BillingService
}

// NewGetInvoicesHandler creates a new handler
func NewGetInvoicesHandler(service BillingService) *GetInvoicesHandler {
	return &GetInvoicesHandler{service: service}
}

// Handle handles the get invoices request
func (h *GetInvoicesHandler) Handle(w http.ResponseWriter, r *http.Request) {
	_ = middleware.GetWorkspaceID(r.Context())

	invoices := []Invoice{
		{
			ID:          "inv-001",
			Number:      "INV-2024-001",
			Amount:      0,
			Currency:    "USD",
			Status:      "paid",
			PeriodStart: time.Now().AddDate(0, -1, 0),
			PeriodEnd:   time.Now(),
			PaidAt:      func() *time.Time { t := time.Now().AddDate(0, -1, 0); return &t }(),
			PDFURL:      "/api/v1/billing/invoices/inv-001/pdf",
		},
	}

	common.Success(w, map[string]interface{}{
		"invoices": invoices,
	})
}
