package billing

import (
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/linkflow-ai/linkflow/internal/adapters/http/dto/common"
	"github.com/linkflow-ai/linkflow/internal/adapters/http/middleware"
)

type Plan struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Price       float64  `json:"price"`
	Currency    string   `json:"currency"`
	Interval    string   `json:"interval"`
	Features    []string `json:"features"`
	Limits      Limits   `json:"limits"`
}

type Limits struct {
	Workflows     int `json:"workflows"`
	Executions    int `json:"executionsPerMonth"`
	TeamMembers   int `json:"teamMembers"`
	DataRetention int `json:"dataRetentionDays"`
}

type Subscription struct {
	ID            string    `json:"id"`
	PlanID        string    `json:"planId"`
	Status        string    `json:"status"`
	CurrentPeriod Period    `json:"currentPeriod"`
	CancelAt      *time.Time `json:"cancelAt,omitempty"`
	CreatedAt     time.Time `json:"createdAt"`
}

type Period struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

type Usage struct {
	Period          Period        `json:"period"`
	Workflows       UsageItem     `json:"workflows"`
	Executions      UsageItem     `json:"executions"`
	Storage         StorageUsage  `json:"storage"`
	ExecutionsByDay []DailyUsage  `json:"executionsByDay"`
}

type UsageItem struct {
	Used  int `json:"used"`
	Limit int `json:"limit"`
}

type StorageUsage struct {
	UsedBytes  int64 `json:"usedBytes"`
	LimitBytes int64 `json:"limitBytes"`
}

type DailyUsage struct {
	Date       string `json:"date"`
	Executions int    `json:"executions"`
}

type Invoice struct {
	ID          string    `json:"id"`
	Number      string    `json:"number"`
	Amount      float64   `json:"amount"`
	Currency    string    `json:"currency"`
	Status      string    `json:"status"`
	PeriodStart time.Time `json:"periodStart"`
	PeriodEnd   time.Time `json:"periodEnd"`
	PaidAt      *time.Time `json:"paidAt,omitempty"`
	PDFURL      string    `json:"pdfUrl"`
}

type BillingService interface {
	GetPlans() ([]Plan, error)
	GetSubscription(workspaceID string) (*Subscription, error)
	CreateSubscription(workspaceID, planID string) (*Subscription, error)
	CancelSubscription(workspaceID string) error
	GetUsage(workspaceID string) (*Usage, error)
	GetInvoices(workspaceID string) ([]Invoice, error)
}

type Handler struct {
	service BillingService
	webhookSecret string
}

func NewHandler(service BillingService, webhookSecret string) *Handler {
	return &Handler{
		service:       service,
		webhookSecret: webhookSecret,
	}
}

func (h *Handler) GetPlans(w http.ResponseWriter, r *http.Request) {
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

func (h *Handler) GetSubscription(w http.ResponseWriter, r *http.Request) {
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

type CreateSubscriptionRequest struct {
	PlanID string `json:"planId"`
}

func (h *Handler) CreateSubscription(w http.ResponseWriter, r *http.Request) {
	workspaceID := middleware.GetWorkspaceID(r.Context())

	var req CreateSubscriptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common.Error(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	if req.PlanID == "" {
		common.Error(w, http.StatusBadRequest, "MISSING_PLAN", "Plan ID is required")
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

	common.JSON(w, http.StatusCreated, subscription)
}

func (h *Handler) CancelSubscription(w http.ResponseWriter, r *http.Request) {
	_ = middleware.GetWorkspaceID(r.Context())

	common.Success(w, map[string]interface{}{
		"message":   "Subscription will be cancelled at the end of the billing period",
		"cancelAt":  time.Now().AddDate(0, 0, 15),
	})
}

func (h *Handler) GetUsage(w http.ResponseWriter, r *http.Request) {
	_ = middleware.GetWorkspaceID(r.Context())

	now := time.Now()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	
	dailyUsage := make([]DailyUsage, 0, 30)
	for d := 0; d < now.Day(); d++ {
		date := startOfMonth.AddDate(0, 0, d)
		dailyUsage = append(dailyUsage, DailyUsage{
			Date:       date.Format("2006-01-02"),
			Executions: 50 + d*10,
		})
	}

	usage := Usage{
		Period: Period{
			Start: startOfMonth,
			End:   startOfMonth.AddDate(0, 1, 0).Add(-time.Second),
		},
		Workflows:  UsageItem{Used: 3, Limit: 5},
		Executions: UsageItem{Used: 450, Limit: 1000},
		Storage: StorageUsage{
			UsedBytes:  52428800,
			LimitBytes: 1073741824,
		},
		ExecutionsByDay: dailyUsage,
	}

	common.Success(w, usage)
}

func (h *Handler) GetInvoices(w http.ResponseWriter, r *http.Request) {
	_ = middleware.GetWorkspaceID(r.Context())

	invoices := []Invoice{
		{
			ID:          "inv-001",
			Number:      "INV-2024-001",
			Amount:      0,
			Currency:    "USD",
			Status:      "paid",
			PeriodStart: time.Now().AddDate(0, -1, 0),
			PeriodEnd:   time.Now(),
			PaidAt:      func() *time.Time { t := time.Now().AddDate(0, -1, 0); return &t }(),
			PDFURL:      "/api/v1/billing/invoices/inv-001/pdf",
		},
	}

	common.Success(w, map[string]interface{}{
		"invoices": invoices,
	})
}

func (h *Handler) StripeWebhook(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		common.Error(w, http.StatusBadRequest, "INVALID_BODY", "Could not read request body")
		return
	}

	_ = r.Header.Get("Stripe-Signature")
	_ = body

	common.Success(w, map[string]string{
		"received": "true",
	})
}
