package models

import (
	"time"

	"github.com/google/uuid"
)

// Plan IDs
const (
	PlanFree       = "free"
	PlanStarter    = "starter"
	PlanPro        = "pro"
	PlanBusiness   = "business"
	PlanEnterprise = "enterprise"
)

// Subscription statuses
const (
	SubscriptionStatusActive    = "active"
	SubscriptionStatusCanceled  = "canceled"
	SubscriptionStatusPastDue   = "past_due"
	SubscriptionStatusTrialing  = "trialing"
	SubscriptionStatusPaused    = "paused"
	SubscriptionStatusExpired   = "expired"
)

// Billing cycles
const (
	BillingCycleMonthly = "monthly"
	BillingCycleYearly  = "yearly"
)

// Plan represents a subscription plan with limits and features
type Plan struct {
	ID   string  `gorm:"size:50;primaryKey" json:"id"`
	Name string  `gorm:"size:100;not null" json:"name"`
	Tier string  `gorm:"size:20;not null;default:starter" json:"tier"` // free, starter, pro, business, enterprise

	// Pricing (in cents)
	PriceMonthly int `gorm:"not null" json:"price_monthly"`
	PriceYearly  int `gorm:"not null" json:"price_yearly"`

	// Credits/Operations (Make.com style)
	CreditsIncluded   int `gorm:"not null;default:1000" json:"credits_included"`     // Base credits per month
	CreditsMax        int `gorm:"not null;default:1000" json:"credits_max"`          // Max credits allowed (with overage)
	CreditOverageCost int `gorm:"not null;default:0" json:"credit_overage_cost"`     // Cost per 1000 extra credits (cents)

	// Resource Limits (-1 = unlimited)
	ExecutionsLimit  int `gorm:"not null" json:"executions_limit"`  // Per month
	WorkflowsLimit   int `gorm:"not null" json:"workflows_limit"`   // Active workflows
	MembersLimit     int `gorm:"not null" json:"members_limit"`     // Team members
	CredentialsLimit int `gorm:"not null" json:"credentials_limit"` // Stored credentials
	SchedulesLimit   int `gorm:"not null;default:5" json:"schedules_limit"`
	WebhooksLimit    int `gorm:"not null;default:5" json:"webhooks_limit"`

	// Execution Limits
	ExecutionTimeout   int `gorm:"not null;default:30" json:"execution_timeout"`       // Seconds
	MaxNodesPerWorkflow int `gorm:"not null;default:50" json:"max_nodes_per_workflow"`

	// Data Retention
	RetentionDays    int `gorm:"not null" json:"retention_days"`
	LogRetentionDays int `gorm:"not null;default:7" json:"log_retention_days"`

	// Features (JSON flags)
	Features JSON `gorm:"type:jsonb;default:'{}'" json:"features"`

	// Stripe Integration
	StripePriceID       *string `gorm:"size:255" json:"-"`
	StripeYearlyPriceID *string `gorm:"size:255" json:"-"`
	StripeProductID     *string `gorm:"size:255" json:"-"`

	// Metadata
	Description *string   `gorm:"type:text" json:"description,omitempty"`
	IsActive    bool      `gorm:"default:true" json:"is_active"`
	IsPublic    bool      `gorm:"default:true" json:"is_public"` // Show on pricing page
	SortOrder   int       `gorm:"default:0" json:"sort_order"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (Plan) TableName() string {
	return "plans"
}

// PlanFeatures represents the feature flags for a plan
type PlanFeatures struct {
	// Core Features
	Webhooks       bool `json:"webhooks"`
	Schedules      bool `json:"schedules"`
	ManualTrigger  bool `json:"manual_trigger"`
	BasicNodes     bool `json:"basic_nodes"`
	AdvancedNodes  bool `json:"advanced_nodes"`
	SubWorkflows   bool `json:"sub_workflows"`
	ErrorWorkflow  bool `json:"error_workflow"`

	// API & Integration
	APIAccess       bool `json:"api_access"`
	CustomFunctions bool `json:"custom_functions"`
	CustomAI        bool `json:"custom_ai"` // Connect own AI providers

	// Execution
	PriorityExecution bool `json:"priority_execution"`
	ParallelExecution bool `json:"parallel_execution"`
	RetryOnFailure    bool `json:"retry_on_failure"`

	// Team & Collaboration
	TeamRoles         bool `json:"team_roles"`
	SharedTemplates   bool `json:"shared_templates"`
	WorkflowComments  bool `json:"workflow_comments"`

	// Security & Compliance
	SSO             bool `json:"sso"`
	AuditLogs       bool `json:"audit_logs"`
	IPWhitelist     bool `json:"ip_whitelist"`
	DataEncryption  bool `json:"data_encryption"`

	// Support
	PrioritySupport bool `json:"priority_support"`
	DedicatedSupport bool `json:"dedicated_support"`
	SLAGuarantee    bool `json:"sla_guarantee"`

	// Customization
	CustomBranding  bool `json:"custom_branding"`
	WhiteLabel      bool `json:"white_label"`
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

// Usage tracks resource consumption for a billing period
type Usage struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	WorkspaceID uuid.UUID `gorm:"type:uuid;index;not null" json:"workspace_id"`
	PeriodStart time.Time `gorm:"not null;index" json:"period_start"`
	PeriodEnd   time.Time `gorm:"not null" json:"period_end"`

	// Credits/Operations (Make.com style)
	CreditsUsed      int `gorm:"default:0" json:"credits_used"`       // Total credits consumed
	CreditsIncluded  int `gorm:"default:0" json:"credits_included"`   // From plan
	CreditsPurchased int `gorm:"default:0" json:"credits_purchased"`  // Extra purchased
	CreditsRemaining int `gorm:"default:0" json:"credits_remaining"`  // Calculated field

	// Execution Metrics
	Executions         int `gorm:"default:0" json:"executions"`           // Total workflow executions
	ExecutionsSuccess  int `gorm:"default:0" json:"executions_success"`   // Successful
	ExecutionsFailed   int `gorm:"default:0" json:"executions_failed"`    // Failed
	Operations         int `gorm:"default:0" json:"operations"`           // Total node operations
	WebhooksCalled     int `gorm:"default:0" json:"webhooks_called"`      // Webhook triggers
	SchedulesTriggered int `gorm:"default:0" json:"schedules_triggered"`  // Schedule triggers

	// Resource Counts (point-in-time snapshots)
	Workflows   int `gorm:"default:0" json:"workflows"`    // Active workflows
	Members     int `gorm:"default:0" json:"members"`      // Team members
	Credentials int `gorm:"default:0" json:"credentials"`  // Stored credentials
	Schedules   int `gorm:"default:0" json:"schedules"`    // Active schedules
	Webhooks    int `gorm:"default:0" json:"webhooks"`     // Active webhooks

	// Data Usage
	StorageBytes     int64 `gorm:"default:0" json:"storage_bytes"`      // Total storage
	DataTransferIn   int64 `gorm:"default:0" json:"data_transfer_in"`   // Bytes in
	DataTransferOut  int64 `gorm:"default:0" json:"data_transfer_out"`  // Bytes out

	// Billing
	OverageCredits int `gorm:"default:0" json:"overage_credits"`  // Credits over limit
	OverageCost    int `gorm:"default:0" json:"overage_cost"`     // Cost in cents

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Workspace Workspace `gorm:"foreignKey:WorkspaceID" json:"-"`
}

func (Usage) TableName() string {
	return "usage"
}

// OperationLog tracks individual operations for credit consumption
type OperationLog struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	WorkspaceID  uuid.UUID `gorm:"type:uuid;index;not null" json:"workspace_id"`
	ExecutionID  uuid.UUID `gorm:"type:uuid;index;not null" json:"execution_id"`
	WorkflowID   uuid.UUID `gorm:"type:uuid;index;not null" json:"workflow_id"`
	NodeID       string    `gorm:"size:100;not null" json:"node_id"`
	NodeType     string    `gorm:"size:100;not null" json:"node_type"`
	Credits      int       `gorm:"not null;default:1" json:"credits"`        // Credits consumed
	DataSizeIn   int64     `gorm:"default:0" json:"data_size_in"`            // Bytes processed in
	DataSizeOut  int64     `gorm:"default:0" json:"data_size_out"`           // Bytes processed out
	DurationMs   int       `gorm:"default:0" json:"duration_ms"`             // Execution time
	Success      bool      `gorm:"default:true" json:"success"`
	ErrorMessage *string   `gorm:"type:text" json:"error_message,omitempty"`
	CreatedAt    time.Time `gorm:"index" json:"created_at"`
}

func (OperationLog) TableName() string {
	return "operation_logs"
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
