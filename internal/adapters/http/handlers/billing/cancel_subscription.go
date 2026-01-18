package billing

import (
	"net/http"

	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/middleware"
	billingCmd "github.com/linkflow-ai/linkflow/internal/core/application/command/billing"
)

type CancelSubscriptionHandler struct {
	handler *billingCmd.CancelSubscriptionHandler
}

func NewCancelSubscriptionHandler(handler *billingCmd.CancelSubscriptionHandler) *CancelSubscriptionHandler {
	return &CancelSubscriptionHandler{handler: handler}
}

func (h *CancelSubscriptionHandler) Handle(w http.ResponseWriter, r *http.Request) {
	workspaceID := middleware.GetWorkspaceID(r.Context())

	err := h.handler.Handle(r.Context(), billingCmd.CancelSubscriptionCommand{
		WorkspaceID: workspaceID,
	})
	if err != nil {
		common.HandleError(w, err)
		return
	}

	common.Success(w, map[string]string{
		"message": "subscription cancelled",
	})
}
