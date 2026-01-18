package billing

import (
	"encoding/json"
	"net/http"

	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/middleware"
	billingCmd "github.com/linkflow-ai/linkflow/internal/core/application/command/billing"
)

type CreateSubscriptionHandler struct {
	handler *billingCmd.CreateSubscriptionHandler
}

func NewCreateSubscriptionHandler(handler *billingCmd.CreateSubscriptionHandler) *CreateSubscriptionHandler {
	return &CreateSubscriptionHandler{handler: handler}
}

func (h *CreateSubscriptionHandler) Handle(w http.ResponseWriter, r *http.Request) {
	var req CreateSubscriptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common.BadRequest(w, "invalid request body")
		return
	}

	workspaceID := middleware.GetWorkspaceID(r.Context())
	userID := middleware.GetUserID(r.Context())

	sub, err := h.handler.Handle(r.Context(), billingCmd.CreateSubscriptionCommand{
		WorkspaceID: workspaceID,
		PlanID:      req.PlanID,
		CreatedBy:   userID,
	})
	if err != nil {
		common.HandleError(w, err)
		return
	}

	common.Created(w, ToSubscriptionResponse(sub))
}
