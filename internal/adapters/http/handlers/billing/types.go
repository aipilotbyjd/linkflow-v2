package billing

import (
	"time"

	"github.com/linkflow-ai/linkflow/internal/core/domain/billing"
)

// Request DTOs

type CreateSubscriptionRequest struct {
	PlanID string `json:"planId" validate:"required"`
}

// Response DTOs

type PlanResponse struct {
	ID          string        `json:"id"`
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Price       float64       `json:"price"`
	Currency    string        `json:"currency"`
	Interval    string        `json:"interval"`
	Features    []string      `json:"features"`
	Limits      LimitsResponse `json:"limits"`
}

type LimitsResponse struct {
	Workflows     int `json:"workflows"`
	Executions    int `json:"executionsPerMonth"`
	TeamMembers   int `json:"teamMembers"`
	DataRetention int `json:"dataRetentionDays"`
}

type SubscriptionResponse struct {
	ID            string          `json:"id"`
	PlanID        string          `json:"planId"`
	Status        string          `json:"status"`
	CurrentPeriod PeriodResponse  `json:"currentPeriod"`
	CancelAt      *time.Time      `json:"cancelAt,omitempty"`
	CreatedAt     time.Time       `json:"createdAt"`
}

type PeriodResponse struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

type UsageResponse struct {
	Period          PeriodResponse       `json:"period"`
	Workflows       UsageItemResponse    `json:"workflows"`
	Executions      UsageItemResponse    `json:"executions"`
	Storage         StorageUsageResponse `json:"storage"`
	ExecutionsByDay []DailyUsageResponse `json:"executionsByDay"`
}

type UsageItemResponse struct {
	Used  int `json:"used"`
	Limit int `json:"limit"`
}

type StorageUsageResponse struct {
	UsedBytes  int64 `json:"usedBytes"`
	LimitBytes int64 `json:"limitBytes"`
}

type DailyUsageResponse struct {
	Date       string `json:"date"`
	Executions int    `json:"executions"`
}

type InvoiceResponse struct {
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

// Mappers - Domain to Response

func ToPlanResponse(p billing.Plan) PlanResponse {
	return PlanResponse{
		ID:          p.ID,
		Name:        p.Name,
		Description: p.Description,
		Price:       float64(p.PriceMonthly) / 100,
		Currency:    p.Currency,
		Interval:    "month",
		Features:    p.Features,
		Limits: LimitsResponse{
			Workflows:     p.Limits.Workflows,
			Executions:    p.Limits.ExecutionsPerMonth,
			TeamMembers:   p.Limits.TeamMembers,
			DataRetention: p.Limits.RetentionDays,
		},
	}
}

func ToPlanResponses(plans []billing.Plan) []PlanResponse {
	responses := make([]PlanResponse, len(plans))
	for i, p := range plans {
		responses[i] = ToPlanResponse(p)
	}
	return responses
}

func ToSubscriptionResponse(s *billing.Subscription) SubscriptionResponse {
	return SubscriptionResponse{
		ID:     s.ID.String(),
		PlanID: s.PlanID,
		Status: string(s.Status),
		CurrentPeriod: PeriodResponse{
			Start: s.CurrentPeriodStart,
			End:   s.CurrentPeriodEnd,
		},
		CancelAt:  s.CanceledAt,
		CreatedAt: s.CreatedAt,
	}
}

func ToUsageResponse(u *billing.Usage) UsageResponse {
	return UsageResponse{
		Period: PeriodResponse{
			Start: u.PeriodStart,
			End:   u.PeriodEnd,
		},
		Executions: UsageItemResponse{
			Used:  int(u.Executions),
			Limit: -1,
		},
		Storage: StorageUsageResponse{
			UsedBytes:  u.StorageBytes,
			LimitBytes: -1,
		},
		ExecutionsByDay: []DailyUsageResponse{},
	}
}

func ToInvoiceResponse(i billing.Invoice) InvoiceResponse {
	pdfURL := ""
	if i.InvoicePDF != nil {
		pdfURL = *i.InvoicePDF
	}
	return InvoiceResponse{
		ID:          i.ID.String(),
		Number:      i.Number,
		Amount:      float64(i.Total) / 100,
		Currency:    i.Currency,
		Status:      string(i.Status),
		PeriodStart: i.PeriodStart,
		PeriodEnd:   i.PeriodEnd,
		PaidAt:      i.PaidAt,
		PDFURL:      pdfURL,
	}
}

func ToInvoiceResponses(invoices []billing.Invoice) []InvoiceResponse {
	responses := make([]InvoiceResponse, len(invoices))
	for i, inv := range invoices {
		responses[i] = ToInvoiceResponse(inv)
	}
	return responses
}
