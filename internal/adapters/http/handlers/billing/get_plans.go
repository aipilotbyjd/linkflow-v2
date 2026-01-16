package billing

import (
	"net/http"

	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
)

type Plan struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Price       int64    `json:"price"`
	Currency    string   `json:"currency"`
	Interval    string   `json:"interval"`
	Features    []string `json:"features"`
	Limits      Limits   `json:"limits"`
}

type Limits struct {
	Workflows          int `json:"workflows"`
	ExecutionsPerMonth int `json:"executions_per_month"`
	TeamMembers        int `json:"team_members"`
	RetentionDays      int `json:"retention_days"`
}

type GetPlansHandler struct{}

func NewGetPlansHandler() *GetPlansHandler {
	return &GetPlansHandler{}
}

func (h *GetPlansHandler) Handle(w http.ResponseWriter, r *http.Request) {
	plans := []Plan{
		{
			ID:          "free",
			Name:        "Free",
			Description: "For individuals getting started",
			Price:       0,
			Currency:    "USD",
			Interval:    "month",
			Features:    []string{"5 workflows", "1,000 executions/month", "Community support"},
			Limits:      Limits{Workflows: 5, ExecutionsPerMonth: 1000, TeamMembers: 1, RetentionDays: 7},
		},
		{
			ID:          "pro",
			Name:        "Pro",
			Description: "For professionals and small teams",
			Price:       2900,
			Currency:    "USD",
			Interval:    "month",
			Features:    []string{"Unlimited workflows", "50,000 executions/month", "Priority support", "Team collaboration"},
			Limits:      Limits{Workflows: -1, ExecutionsPerMonth: 50000, TeamMembers: 10, RetentionDays: 30},
		},
		{
			ID:          "enterprise",
			Name:        "Enterprise",
			Description: "For large organizations",
			Price:       0,
			Currency:    "USD",
			Interval:    "month",
			Features:    []string{"Unlimited everything", "SSO", "Dedicated support", "SLA"},
			Limits:      Limits{Workflows: -1, ExecutionsPerMonth: -1, TeamMembers: -1, RetentionDays: 365},
		},
	}

	common.Success(w, plans)
}
