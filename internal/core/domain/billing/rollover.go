package billing

import (
	"time"

	"github.com/google/uuid"
)

// CreditRollover tracks rolled over credits from previous periods
type CreditRollover struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	WorkspaceID uuid.UUID `gorm:"type:uuid;index;not null" json:"workspace_id"`

	// Rolled over amounts
	RolledCredits   int `gorm:"default:0" json:"rolled_credits"`
	RolledAICredits int `gorm:"default:0" json:"rolled_ai_credits"`

	// Source period
	FromPeriodStart time.Time `json:"from_period_start"`
	FromPeriodEnd   time.Time `json:"from_period_end"`

	// Expiry
	ExpiresAt time.Time `json:"expires_at"`
	Expired   bool      `gorm:"default:false" json:"expired"`

	// Consumption tracking
	UsedCredits   int `gorm:"default:0" json:"used_credits"`
	UsedAICredits int `gorm:"default:0" json:"used_ai_credits"`

	CreatedAt time.Time `json:"created_at"`
}

// RolloverSettings configures credit rollover behavior
type RolloverSettings struct {
	// Whether rollover is enabled for this plan
	Enabled bool `json:"enabled"`

	// Maximum percentage of unused credits that roll over
	MaxRolloverPercent int `json:"max_rollover_percent"` // e.g., 20 = 20%

	// Maximum absolute credits that can roll over
	MaxRolloverCredits int `json:"max_rollover_credits"`
	MaxRolloverAI      int `json:"max_rollover_ai_credits"`

	// How long rolled credits last (months)
	ExpiryMonths int `json:"expiry_months"` // e.g., 3 = expires after 3 months
}

// Plan rollover settings
var PlanRolloverSettings = map[string]RolloverSettings{
	"free": {
		Enabled: false,
	},
	"core": {
		Enabled:            true,
		MaxRolloverPercent: 10,
		MaxRolloverCredits: 1000,
		MaxRolloverAI:      50,
		ExpiryMonths:       1,
	},
	"pro": {
		Enabled:            true,
		MaxRolloverPercent: 20,
		MaxRolloverCredits: 2000,
		MaxRolloverAI:      200,
		ExpiryMonths:       2,
	},
	"teams": {
		Enabled:            true,
		MaxRolloverPercent: 25,
		MaxRolloverCredits: 5000,
		MaxRolloverAI:      500,
		ExpiryMonths:       3,
	},
	"enterprise": {
		Enabled:            true,
		MaxRolloverPercent: 50,
		MaxRolloverCredits: -1, // Unlimited
		MaxRolloverAI:      -1,
		ExpiryMonths:       6,
	},
}

// CalculateRollover calculates credits to roll over
func CalculateRollover(planID string, unusedCredits, unusedAI, planCredits, planAI int) (rollCredits, rollAI int) {
	settings, ok := PlanRolloverSettings[planID]
	if !ok || !settings.Enabled {
		return 0, 0
	}

	// Calculate percentage-based limit
	maxByPercent := planCredits * settings.MaxRolloverPercent / 100
	maxAIByPercent := planAI * settings.MaxRolloverPercent / 100

	// Take minValueimum of unused, percentage limit, and absolute limit
	rollCredits = minValue(unusedCredits, maxByPercent)
	if settings.MaxRolloverCredits > 0 {
		rollCredits = minValue(rollCredits, settings.MaxRolloverCredits)
	}

	rollAI = minValue(unusedAI, maxAIByPercent)
	if settings.MaxRolloverAI > 0 {
		rollAI = minValue(rollAI, settings.MaxRolloverAI)
	}

	return rollCredits, rollAI
}

// NewCreditRollover creates a new rollover record
func NewCreditRollover(workspaceID uuid.UUID, planID string, unusedCredits, unusedAI, planCredits, planAI int, periodStart, periodEnd time.Time) *CreditRollover {
	rollCredits, rollAI := CalculateRollover(planID, unusedCredits, unusedAI, planCredits, planAI)

	settings := PlanRolloverSettings[planID]
	expiresAt := periodEnd.AddDate(0, settings.ExpiryMonths, 0)

	return &CreditRollover{
		ID:              uuid.New(),
		WorkspaceID:     workspaceID,
		RolledCredits:   rollCredits,
		RolledAICredits: rollAI,
		FromPeriodStart: periodStart,
		FromPeriodEnd:   periodEnd,
		ExpiresAt:       expiresAt,
		CreatedAt:       time.Now(),
	}
}

// RemainingCredits returns unused rolled credits
func (r *CreditRollover) RemainingCredits() int {
	return r.RolledCredits - r.UsedCredits
}

// RemainingAICredits returns unused rolled AI credits
func (r *CreditRollover) RemainingAICredits() int {
	return r.RolledAICredits - r.UsedAICredits
}

// IsExpired checks if rollover has expired
func (r *CreditRollover) IsExpired() bool {
	return time.Now().After(r.ExpiresAt)
}

// ConsumeCredits consumes from rolled credits, returns amount consumed
func (r *CreditRollover) ConsumeCredits(amount int) int {
	if r.IsExpired() || r.Expired {
		return 0
	}

	available := r.RemainingCredits()
	consume := minValue(amount, available)
	r.UsedCredits += consume
	return consume
}

// ConsumeAICredits consumes from rolled AI credits
func (r *CreditRollover) ConsumeAICredits(amount int) int {
	if r.IsExpired() || r.Expired {
		return 0
	}

	available := r.RemainingAICredits()
	consume := minValue(amount, available)
	r.UsedAICredits += consume
	return consume
}

func minValue(a, b int) int {
	if a < b {
		return a
	}
	return b
}
