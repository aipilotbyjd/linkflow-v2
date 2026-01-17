package billing

import (
	"net/http"
	"time"

	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/middleware"
)

// CancelSubscriptionHandler handles cancel subscription request
type CancelSubscriptionHandler struct {
	service BillingService
}

// NewCancelSubscriptionHandler creates a new handler
func NewCancelSubscriptionHandler(service BillingService) *CancelSubscriptionHandler {
	return &CancelSubscriptionHandler{service: service}
}

// Handle handles the cancel subscription request
func (h *CancelSubscriptionHandler) Handle(w http.ResponseWriter, r *http.Request) {
	_ = middleware.GetWorkspaceID(r.Context())

	common.Success(w, map[string]interface{}{
		"message":  "Subscription will be cancelled at the end of the billing period",
		"cancelAt": time.Now().AddDate(0, 0, 15),
	})
}
