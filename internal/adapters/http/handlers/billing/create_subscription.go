package billing

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/middleware"
)

// CreateSubscriptionRequest represents subscription creation request
type CreateSubscriptionRequest struct {
	PlanID string `json:"planId"`
}

// CreateSubscriptionHandler handles create subscription request
type CreateSubscriptionHandler struct {
	service BillingService
}

// NewCreateSubscriptionHandler creates a new handler
func NewCreateSubscriptionHandler(service BillingService) *CreateSubscriptionHandler {
	return &CreateSubscriptionHandler{service: service}
}

// Handle handles the create subscription request
func (h *CreateSubscriptionHandler) Handle(w http.ResponseWriter, r *http.Request) {
	workspaceID := middleware.GetWorkspaceID(r.Context())

	var req CreateSubscriptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common.BadRequest(w, "invalid request body")
		return
	}

	if req.PlanID == "" {
		common.BadRequest(w, "plan ID is required")
		return
	}

	subscription := Subscription{
		ID:     "sub-" + workspaceID.String()[:8],
		PlanID: req.PlanID,
		Status: "active",
		CurrentPeriod: Period{
			Start: time.Now(),
			End:   time.Now().AddDate(0, 1, 0),
		},
		CreatedAt: time.Now(),
	}

	common.Created(w, subscription)
}
