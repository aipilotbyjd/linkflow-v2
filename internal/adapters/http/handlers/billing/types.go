package billing

import "time"

// Plan represents a billing plan
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

// Limits represents plan limits
type Limits struct {
	Workflows     int `json:"workflows"`
	Executions    int `json:"executionsPerMonth"`
	TeamMembers   int `json:"teamMembers"`
	DataRetention int `json:"dataRetentionDays"`
}

// Subscription represents a workspace subscription
type Subscription struct {
	ID            string     `json:"id"`
	PlanID        string     `json:"planId"`
	Status        string     `json:"status"`
	CurrentPeriod Period     `json:"currentPeriod"`
	CancelAt      *time.Time `json:"cancelAt,omitempty"`
	CreatedAt     time.Time  `json:"createdAt"`
}

// Period represents a billing period
type Period struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// Usage represents workspace usage
type Usage struct {
	Period          Period       `json:"period"`
	Workflows       UsageItem    `json:"workflows"`
	Executions      UsageItem    `json:"executions"`
	Storage         StorageUsage `json:"storage"`
	ExecutionsByDay []DailyUsage `json:"executionsByDay"`
}

// UsageItem represents usage of a specific resource
type UsageItem struct {
	Used  int `json:"used"`
	Limit int `json:"limit"`
}

// StorageUsage represents storage usage
type StorageUsage struct {
	UsedBytes  int64 `json:"usedBytes"`
	LimitBytes int64 `json:"limitBytes"`
}

// DailyUsage represents daily execution usage
type DailyUsage struct {
	Date       string `json:"date"`
	Executions int    `json:"executions"`
}

// Invoice represents a billing invoice
type Invoice struct {
	ID          string     `json:"id"`
	Number      string     `json:"number"`
	Amount      float64    `json:"amount"`
	Currency    string     `json:"currency"`
	Status      string     `json:"status"`
	PeriodStart time.Time  `json:"periodStart"`
	PeriodEnd   time.Time  `json:"periodEnd"`
	PaidAt      *time.Time `json:"paidAt,omitempty"`
	PDFURL      string     `json:"pdfUrl"`
}

// BillingService defines the billing service interface
type BillingService interface {
	GetPlans() ([]Plan, error)
	GetSubscription(workspaceID string) (*Subscription, error)
	CreateSubscription(workspaceID, planID string) (*Subscription, error)
	CancelSubscription(workspaceID string) error
	GetUsage(workspaceID string) (*Usage, error)
	GetInvoices(workspaceID string) ([]Invoice, error)
}
