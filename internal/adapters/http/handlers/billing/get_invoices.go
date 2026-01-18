package billing

import (
	"net/http"

	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/middleware"
	billingQuery "github.com/linkflow-ai/linkflow/internal/core/application/query/billing"
)

type GetInvoicesHandler struct {
	handler *billingQuery.GetInvoicesHandler
}

func NewGetInvoicesHandler(handler *billingQuery.GetInvoicesHandler) *GetInvoicesHandler {
	return &GetInvoicesHandler{handler: handler}
}

func (h *GetInvoicesHandler) Handle(w http.ResponseWriter, r *http.Request) {
	workspaceID := middleware.GetWorkspaceID(r.Context())

	invoices, err := h.handler.Handle(r.Context(), billingQuery.GetInvoicesQuery{
		WorkspaceID: workspaceID,
	})
	if err != nil {
		common.HandleError(w, err)
		return
	}

	common.Success(w, map[string]interface{}{
		"invoices": ToInvoiceResponses(invoices),
	})
}
