package billing

import (
	"net/http"
	"strconv"

	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/middleware"
	"github.com/linkflow-ai/linkflow/internal/core/domain/billing"
	"github.com/linkflow-ai/linkflow/internal/shared/types"
)

type InvoiceResponse struct {
	ID            string  `json:"id"`
	WorkspaceID   string  `json:"workspace_id"`
	Number        string  `json:"number"`
	Status        string  `json:"status"`
	AmountDue     int64   `json:"amount_due"`
	AmountPaid    int64   `json:"amount_paid"`
	Currency      string  `json:"currency"`
	PeriodStart   string  `json:"period_start"`
	PeriodEnd     string  `json:"period_end"`
	DueDate       *string `json:"due_date,omitempty"`
	PaidAt        *string `json:"paid_at,omitempty"`
	InvoicePDFURL *string `json:"invoice_pdf_url,omitempty"`
	CreatedAt     string  `json:"created_at"`
}

type GetInvoicesHandler struct {
	invoiceRepo billing.InvoiceRepository
}

func NewGetInvoicesHandler(invoiceRepo billing.InvoiceRepository) *GetInvoicesHandler {
	return &GetInvoicesHandler{invoiceRepo: invoiceRepo}
}

func (h *GetInvoicesHandler) Handle(w http.ResponseWriter, r *http.Request) {
	wsCtx := middleware.GetWorkspaceFromContext(r.Context())
	if wsCtx == nil {
		common.BadRequest(w, "workspace context required")
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize
	invoices, total, err := h.invoiceRepo.FindByWorkspaceID(r.Context(), wsCtx.WorkspaceID, pageSize, offset)
	if err != nil {
		common.HandleError(w, err)
		return
	}

	var responses []InvoiceResponse
	for _, inv := range invoices {
		responses = append(responses, toInvoiceResponse(&inv))
	}

	common.List(w, responses, types.PageResponse{
		Page:       page,
		PageSize:   pageSize,
		TotalItems: total,
		TotalPages: int((total + int64(pageSize) - 1) / int64(pageSize)),
		HasMore:    int64(page*pageSize) < total,
	})
}

func toInvoiceResponse(inv *billing.Invoice) InvoiceResponse {
	resp := InvoiceResponse{
		ID:            inv.ID.String(),
		WorkspaceID:   inv.WorkspaceID.String(),
		Number:        inv.Number,
		Status:        string(inv.Status),
		AmountDue:     inv.AmountDue,
		AmountPaid:    inv.AmountPaid,
		Currency:      inv.Currency,
		PeriodStart:   inv.PeriodStart.Format("2006-01-02T15:04:05Z"),
		PeriodEnd:     inv.PeriodEnd.Format("2006-01-02T15:04:05Z"),
		InvoicePDFURL: inv.InvoicePDF,
		CreatedAt:     inv.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}

	if inv.DueDate != nil {
		t := inv.DueDate.Format("2006-01-02T15:04:05Z")
		resp.DueDate = &t
	}
	if inv.PaidAt != nil {
		t := inv.PaidAt.Format("2006-01-02T15:04:05Z")
		resp.PaidAt = &t
	}

	return resp
}
