package topup

import (
	"time"

	"github.com/linkflow-ai/linkflow/internal/core/domain/billing"
)

// UpdateAutoTopUpRequest for configuring auto top-up
type UpdateAutoTopUpRequest struct {
	Enabled              *bool   `json:"enabled,omitempty"`
	TriggerThreshold     *int    `json:"trigger_threshold,omitempty"`
	PurchaseType         *string `json:"purchase_type,omitempty"`
	CreditAmount         *int    `json:"credit_amount,omitempty"`
	AICreditsAmount      *int    `json:"ai_credits_amount,omitempty"`
	MaxPurchasesPerMonth *int    `json:"max_purchases_per_month,omitempty"`
	MaxSpendPerMonth     *int64  `json:"max_spend_per_month_cents,omitempty"`
}

// PurchaseCreditsRequest for manual credit purchase
type PurchaseCreditsRequest struct {
	PackID string `json:"pack_id" validate:"required"`
}

// AutoTopUpResponse for API response
type AutoTopUpResponse struct {
	ID                   string     `json:"id"`
	Enabled              bool       `json:"enabled"`
	TriggerThreshold     int        `json:"trigger_threshold"`
	PurchaseType         string     `json:"purchase_type"`
	CreditAmount         int        `json:"credit_amount"`
	AICreditsAmount      int        `json:"ai_credits_amount"`
	MaxPurchasesPerMonth int        `json:"max_purchases_per_month"`
	MaxSpendPerMonth     int64      `json:"max_spend_per_month_cents"`
	PurchasesThisMonth   int        `json:"purchases_this_month"`
	SpentThisMonth       int64      `json:"spent_this_month_cents"`
	LastPurchaseAt       *time.Time `json:"last_purchase_at,omitempty"`
	CreatedAt            time.Time  `json:"created_at"`
}

// CreditPackResponse for available packs
type CreditPackResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Credits     int    `json:"credits,omitempty"`
	AICredits   int    `json:"ai_credits,omitempty"`
	PriceCents  int64  `json:"price_cents"`
	Description string `json:"description"`
	Popular     bool   `json:"popular"`
	SavePercent int    `json:"save_percent,omitempty"`
}

// PurchaseHistoryResponse for purchase logs
type PurchaseHistoryResponse struct {
	ID             string    `json:"id"`
	PurchaseType   string    `json:"purchase_type"`
	CreditsAdded   int       `json:"credits_added"`
	AICreditsAdded int       `json:"ai_credits_added"`
	AmountCharged  int64     `json:"amount_charged_cents"`
	TriggerReason  string    `json:"trigger_reason"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"created_at"`
}

// ToAutoTopUpResponse converts domain to response
func ToAutoTopUpResponse(a *billing.AutoTopUp) AutoTopUpResponse {
	return AutoTopUpResponse{
		ID:                   a.ID.String(),
		Enabled:              a.Enabled,
		TriggerThreshold:     a.TriggerThreshold,
		PurchaseType:         string(a.PurchaseType),
		CreditAmount:         a.CreditAmount,
		AICreditsAmount:      a.AICreditsAmount,
		MaxPurchasesPerMonth: a.MaxPurchasesPerMonth,
		MaxSpendPerMonth:     a.MaxSpendPerMonth,
		PurchasesThisMonth:   a.PurchasesThisMonth,
		SpentThisMonth:       a.SpentThisMonth,
		LastPurchaseAt:       a.LastPurchaseAt,
		CreatedAt:            a.CreatedAt,
	}
}

// ToCreditPackResponse converts domain to response
func ToCreditPackResponse(p billing.CreditPack) CreditPackResponse {
	return CreditPackResponse{
		ID:          p.ID,
		Name:        p.Name,
		Credits:     p.Credits,
		AICredits:   p.AICredits,
		PriceCents:  p.PriceCents,
		Description: p.Description,
		Popular:     p.Popular,
		SavePercent: p.SavePercent,
	}
}
