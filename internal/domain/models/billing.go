package models

import (
	"time"

	"github.com/google/uuid"
)

type Plan struct {
	ID                string    `gorm:"size:50;primaryKey" json:"id"`
	Name              string    `gorm:"size:100;not null" json:"name"`
	Description       *string   `gorm:"type:text" json:"description,omitempty"`
	PriceMonthly      int       `gorm:"not null" json:"price_monthly"`      // cents
	PriceYearly       int       `gorm:"not null" json:"price_yearly"`       // cents
	ExecutionsLimit   int       `gorm:"not null" json:"executions_limit"`   // per month
	WorkflowsLimit    int       `gorm:"not null" json:"workflows_limit"`
	MembersLimit      int       `gorm:"not null" json:"members_limit"`
	CredentialsLimit  int       `gorm:"not null" json:"credentials_limit"`
	RetentionDays     int       `gorm:"not null" json:"retention_days"`
	Features          JSON      `gorm:"type:jsonb;default:'{}'" json:"features"`
	StripePriceID     *string   `gorm:"size:255" json:"-"`
	StripeYearlyPriceID *string `gorm:"size:255" json:"-"`
	IsActive          bool      `gorm:"default:true" json:"is_active"`
	SortOrder         int       `gorm:"default:0" json:"sort_order"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

func (Plan) TableName() string {
	return "plans"
}

type Subscription struct {
	ID                   uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	WorkspaceID          uuid.UUID  `gorm:"type:uuid;uniqueIndex;not null" json:"workspace_id"`
	PlanID               string     `gorm:"size:50;not null" json:"plan_id"`
	Status               string     `gorm:"size:20;not null;default:active" json:"status"`
	BillingCycle         string     `gorm:"size:10;not null;default:monthly" json:"billing_cycle"`
	StripeSubscriptionID *string    `gorm:"size:255" json:"-"`
	StripeCustomerID     *string    `gorm:"size:255" json:"-"`
	CurrentPeriodStart   time.Time  `gorm:"not null" json:"current_period_start"`
	CurrentPeriodEnd     time.Time  `gorm:"not null" json:"current_period_end"`
	CancelAt             *time.Time `json:"cancel_at,omitempty"`
	CanceledAt           *time.Time `json:"canceled_at,omitempty"`
	TrialStart           *time.Time `json:"trial_start,omitempty"`
	TrialEnd             *time.Time `json:"trial_end,omitempty"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`

	Workspace Workspace `gorm:"foreignKey:WorkspaceID" json:"-"`
	Plan      Plan      `gorm:"foreignKey:PlanID" json:"-"`
}

func (Subscription) TableName() string {
	return "subscriptions"
}

type Usage struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	WorkspaceID  uuid.UUID `gorm:"type:uuid;index;not null" json:"workspace_id"`
	PeriodStart  time.Time `gorm:"not null;index" json:"period_start"`
	PeriodEnd    time.Time `gorm:"not null" json:"period_end"`
	Executions   int       `gorm:"default:0" json:"executions"`
	Workflows    int       `gorm:"default:0" json:"workflows"`
	Members      int       `gorm:"default:0" json:"members"`
	Credentials  int       `gorm:"default:0" json:"credentials"`
	StorageBytes int64     `gorm:"default:0" json:"storage_bytes"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`

	Workspace Workspace `gorm:"foreignKey:WorkspaceID" json:"-"`
}

func (Usage) TableName() string {
	return "usage"
}

type Invoice struct {
	ID              uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	WorkspaceID     uuid.UUID  `gorm:"type:uuid;index;not null" json:"workspace_id"`
	StripeInvoiceID *string    `gorm:"size:255" json:"-"`
	Number          string     `gorm:"size:50;not null" json:"number"`
	Status          string     `gorm:"size:20;not null" json:"status"`
	AmountDue       int        `gorm:"not null" json:"amount_due"`
	AmountPaid      int        `gorm:"not null" json:"amount_paid"`
	Currency        string     `gorm:"size:3;default:usd" json:"currency"`
	PeriodStart     time.Time  `gorm:"not null" json:"period_start"`
	PeriodEnd       time.Time  `gorm:"not null" json:"period_end"`
	DueDate         *time.Time `json:"due_date,omitempty"`
	PaidAt          *time.Time `json:"paid_at,omitempty"`
	InvoiceURL      *string    `gorm:"size:500" json:"invoice_url,omitempty"`
	InvoicePDF      *string    `gorm:"size:500" json:"invoice_pdf,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`

	Workspace Workspace `gorm:"foreignKey:WorkspaceID" json:"-"`
}

func (Invoice) TableName() string {
	return "invoices"
}
