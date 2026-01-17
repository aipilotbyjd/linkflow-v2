package billing

import (
	"net/http"
	"time"

	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/middleware"
)

// GetSubscriptionHandler handles get subscription request
type GetSubscriptionHandler struct {
	service BillingService
}

// NewGetSubscriptionHandler creates a new handler
func NewGetSubscriptionHandler(service BillingService) *GetSubscriptionHandler {
	return &GetSubscriptionHandler{service: service}
}

// Handle handles the get subscription request
func (h *GetSubscriptionHandler) Handle(w http.ResponseWriter, r *http.Request) {
	workspaceID := middleware.GetWorkspaceID(r.Context())

	subscription := Subscription{
		ID:     "sub-" + workspaceID.String()[:8],
		PlanID: "free",
		Status: "active",
		CurrentPeriod: Period{
			Start: time.Now().AddDate(0, 0, -15),
			End:   time.Now().AddDate(0, 0, 15),
		},
		CreatedAt: time.Now().AddDate(0, -1, 0),
	}

	common.Success(w, subscription)
}
