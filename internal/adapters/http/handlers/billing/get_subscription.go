package billing

import (
	"net/http"

	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/middleware"
	billingQuery "github.com/linkflow-ai/linkflow/internal/core/application/query/billing"
)

type GetSubscriptionHandler struct {
	handler *billingQuery.GetSubscriptionHandler
}

func NewGetSubscriptionHandler(handler *billingQuery.GetSubscriptionHandler) *GetSubscriptionHandler {
	return &GetSubscriptionHandler{handler: handler}
}

func (h *GetSubscriptionHandler) Handle(w http.ResponseWriter, r *http.Request) {
	workspaceID := middleware.GetWorkspaceID(r.Context())

	sub, err := h.handler.Handle(r.Context(), billingQuery.GetSubscriptionQuery{
		WorkspaceID: workspaceID,
	})
	if err != nil {
		common.HandleError(w, err)
		return
	}

	common.Success(w, ToSubscriptionResponse(sub))
}
