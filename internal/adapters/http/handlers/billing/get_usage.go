package billing

import (
	"net/http"

	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/middleware"
	billingQuery "github.com/linkflow-ai/linkflow/internal/core/application/query/billing"
)

type GetUsageHandler struct {
	handler *billingQuery.GetUsageHandler
}

func NewGetUsageHandler(handler *billingQuery.GetUsageHandler) *GetUsageHandler {
	return &GetUsageHandler{handler: handler}
}

func (h *GetUsageHandler) Handle(w http.ResponseWriter, r *http.Request) {
	workspaceID := middleware.GetWorkspaceID(r.Context())

	usage, err := h.handler.Handle(r.Context(), billingQuery.GetUsageQuery{
		WorkspaceID: workspaceID,
	})
	if err != nil {
		common.HandleError(w, err)
		return
	}

	common.Success(w, ToUsageResponse(usage))
}
