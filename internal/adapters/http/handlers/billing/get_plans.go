package billing

import (
	"net/http"

	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
)

// GetPlansHandler handles get billing plans request
type GetPlansHandler struct {
	service BillingService
}

// NewGetPlansHandler creates a new handler
func NewGetPlansHandler(service BillingService) *GetPlansHandler {
	return &GetPlansHandler{service: service}
}

// Handle handles the get plans request
func (h *GetPlansHandler) Handle(w http.ResponseWriter, r *http.Request) {
	plans := []Plan{
		{
			ID:          "free",
			Name:        "Free",
			Description: "For personal projects",
			Price:       0,
			Currency:    "USD",
			Interval:    "month",
			Features:    []string{"5 workflows", "1,000 executions/month", "7 day retention"},
			Limits:      Limits{Workflows: 5, Executions: 1000, TeamMembers: 1, DataRetention: 7},
		},
		{
			ID:          "pro",
			Name:        "Pro",
			Description: "For growing teams",
			Price:       29,
			Currency:    "USD",
			Interval:    "month",
			Features:    []string{"Unlimited workflows", "50,000 executions/month", "30 day retention", "5 team members"},
			Limits:      Limits{Workflows: -1, Executions: 50000, TeamMembers: 5, DataRetention: 30},
		},
		{
			ID:          "enterprise",
			Name:        "Enterprise",
			Description: "For large organizations",
			Price:       199,
			Currency:    "USD",
			Interval:    "month",
			Features:    []string{"Everything in Pro", "Unlimited executions", "90 day retention", "Unlimited team members", "SSO", "Priority support"},
			Limits:      Limits{Workflows: -1, Executions: -1, TeamMembers: -1, DataRetention: 90},
		},
	}

	common.Success(w, map[string]interface{}{
		"plans": plans,
	})
}
