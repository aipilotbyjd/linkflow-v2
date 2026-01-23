package billing

import (
	"time"

	"github.com/google/uuid"
)

// AutoTopUp configures automatic credit purchases
type AutoTopUp struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	WorkspaceID uuid.UUID `gorm:"type:uuid;uniqueIndex;not null" json:"workspace_id"`
	Enabled     bool      `gorm:"default:false" json:"enabled"`

	// Trigger settings
	TriggerThreshold int `gorm:"default:10" json:"trigger_threshold"` // Buy when below X% remaining

	// Purchase settings
	PurchaseType    TopUpType `gorm:"size:20;default:credits" json:"purchase_type"`
	CreditAmount    int       `gorm:"default:10000" json:"credit_amount"`    // Credits to buy
	AICreditsAmount int       `gorm:"default:1000" json:"ai_credits_amount"` // AI credits to buy

	// Spending limits
	MaxPurchasesPerMonth int   `gorm:"default:5" json:"max_purchases_per_month"`
	MaxSpendPerMonth     int64 `gorm:"default:10000" json:"max_spend_per_month_cents"` // $100 default cap

	// Payment
	PaymentMethodID string `gorm:"size:100" json:"payment_method_id,omitempty"`

	// Tracking
	PurchasesThisMonth int        `gorm:"default:0" json:"purchases_this_month"`
	SpentThisMonth     int64      `gorm:"default:0" json:"spent_this_month_cents"`
	LastPurchaseAt     *time.Time `json:"last_purchase_at,omitempty"`
	LastResetAt        time.Time  `json:"last_reset_at"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type TopUpType string

const (
	TopUpTypeCredits   TopUpType = "credits"
	TopUpTypeAICredits TopUpType = "ai_credits"
	TopUpTypeBoth      TopUpType = "both"
	TopUpTypeUpgrade   TopUpType = "upgrade" // Auto-upgrade to next plan
)

// TopUpPurchase records auto top-up transactions
type TopUpPurchase struct {
	ID             uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	WorkspaceID    uuid.UUID `gorm:"type:uuid;index;not null" json:"workspace_id"`
	AutoTopUpID    uuid.UUID `gorm:"type:uuid;index" json:"auto_topup_id"`
	PurchaseType   TopUpType `gorm:"size:20" json:"purchase_type"`
	CreditsAdded   int       `json:"credits_added"`
	AICreditsAdded int       `json:"ai_credits_added"`
	AmountCharged  int64     `json:"amount_charged_cents"`
	StripeChargeID string    `gorm:"size:100" json:"stripe_charge_id,omitempty"`
	TriggerReason  string    `gorm:"size:200" json:"trigger_reason"`
	UsageAtTrigger int64     `json:"usage_at_trigger"`
	LimitAtTrigger int64     `json:"limit_at_trigger"`
	Status         string    `gorm:"size:20" json:"status"` // pending, completed, failed
	FailureReason  string    `gorm:"size:500" json:"failure_reason,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
}

// CreditPack represents purchasable credit packages
type CreditPack struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Credits     int    `json:"credits"`
	AICredits   int    `json:"ai_credits"`
	PriceCents  int64  `json:"price_cents"`
	Description string `json:"description"`
	Popular     bool   `json:"popular"`
	SavePercent int    `json:"save_percent,omitempty"`
}

// Available credit packs for purchase
var CreditPacks = []CreditPack{
	{
		ID:          "pack_5k",
		Name:        "5,000 Credits",
		Credits:     5000,
		PriceCents:  1500, // $15
		Description: "Good for small projects",
	},
	{
		ID:          "pack_10k",
		Name:        "10,000 Credits",
		Credits:     10000,
		PriceCents:  2500, // $25
		Description: "Most popular choice",
		Popular:     true,
		SavePercent: 17,
	},
	{
		ID:          "pack_50k",
		Name:        "50,000 Credits",
		Credits:     50000,
		PriceCents:  10000, // $100
		Description: "Best value for teams",
		SavePercent: 33,
	},
	{
		ID:          "pack_ai_500",
		Name:        "500 AI Credits",
		AICredits:   500,
		PriceCents:  1000, // $10
		Description: "For AI-powered workflows",
	},
	{
		ID:          "pack_ai_2000",
		Name:        "2,000 AI Credits",
		AICredits:   2000,
		PriceCents:  3000, // $30
		Description: "Heavy AI usage",
		Popular:     true,
		SavePercent: 25,
	},
}

// NewAutoTopUp creates default auto top-up settings
func NewAutoTopUp(workspaceID uuid.UUID) *AutoTopUp {
	return &AutoTopUp{
		ID:                   uuid.New(),
		WorkspaceID:          workspaceID,
		Enabled:              false,
		TriggerThreshold:     10,
		PurchaseType:         TopUpTypeCredits,
		CreditAmount:         10000,
		AICreditsAmount:      1000,
		MaxPurchasesPerMonth: 5,
		MaxSpendPerMonth:     10000, // $100
		LastResetAt:          time.Now(),
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
	}
}

// CanPurchase checks if auto top-up can make another purchase
func (a *AutoTopUp) CanPurchase(purchaseAmount int64) (bool, string) {
	if !a.Enabled {
		return false, "auto top-up is disabled"
	}

	if a.PurchasesThisMonth >= a.MaxPurchasesPerMonth {
		return false, "monthly purchase limit reached"
	}

	if a.SpentThisMonth+purchaseAmount > a.MaxSpendPerMonth {
		return false, "monthly spending limit would be exceeded"
	}

	return true, ""
}

// RecordPurchase records a successful purchase
func (a *AutoTopUp) RecordPurchase(amount int64) {
	now := time.Now()
	a.PurchasesThisMonth++
	a.SpentThisMonth += amount
	a.LastPurchaseAt = &now
	a.UpdatedAt = now
}

// ResetMonthlyLimits resets monthly counters
func (a *AutoTopUp) ResetMonthlyLimits() {
	a.PurchasesThisMonth = 0
	a.SpentThisMonth = 0
	a.LastResetAt = time.Now()
	a.UpdatedAt = time.Now()
}

// GetCreditPack returns a credit pack by ID
func GetCreditPack(id string) *CreditPack {
	for _, pack := range CreditPacks {
		if pack.ID == id {
			return &pack
		}
	}
	return nil
}
