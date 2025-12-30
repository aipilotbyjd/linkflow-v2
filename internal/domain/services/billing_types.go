package services

import (
	"errors"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/domain/models"
)

// Billing errors
var (
	ErrPlanNotFound            = errors.New("plan not found")
	ErrSubscriptionNotFound    = errors.New("subscription not found")
	ErrSubscriptionExists      = errors.New("subscription already exists for this workspace")
	ErrSamePlan                = errors.New("subscription is already on this plan")
	ErrCreditsExceeded         = errors.New("credit limit exceeded")
	ErrFeatureNotAvailable     = errors.New("feature not available on current plan")
	ErrWorkflowLimitExceeded   = errors.New("workflow limit exceeded")
	ErrMemberLimitExceeded     = errors.New("member limit exceeded")
	ErrExecutionLimitExceeded  = errors.New("execution limit exceeded")
	ErrScheduleLimitExceeded   = errors.New("schedule limit exceeded")
	ErrWebhookLimitExceeded    = errors.New("webhook limit exceeded")
	ErrCredentialLimitExceeded = errors.New("credential limit exceeded")
	ErrInvalidBillingCycle     = errors.New("invalid billing cycle: must be 'monthly' or 'yearly'")
	ErrInvalidCredits          = errors.New("credits must be positive")
)

// Subscription status constants
const (
	SubscriptionStatusActive   = "active"
	SubscriptionStatusCanceled = "canceled"
	SubscriptionStatusPastDue  = "past_due"
)

// Billing cycle constants
const (
	BillingCycleMonthly = "monthly"
	BillingCycleYearly  = "yearly"
)

// Feature name constants for type-safe feature checking
const (
	FeatureWebhooks          = "webhooks"
	FeatureSchedules         = "schedules"
	FeatureManualTrigger     = "manual_trigger"
	FeatureBasicNodes        = "basic_nodes"
	FeatureAdvancedNodes     = "advanced_nodes"
	FeatureSubWorkflows      = "sub_workflows"
	FeatureErrorWorkflow     = "error_workflow"
	FeatureAPIAccess         = "api_access"
	FeatureCustomFunctions   = "custom_functions"
	FeatureCustomAI          = "custom_ai"
	FeaturePriorityExecution = "priority_execution"
	FeatureParallelExecution = "parallel_execution"
	FeatureRetryOnFailure    = "retry_on_failure"
	FeatureTeamRoles         = "team_roles"
	FeatureSharedTemplates   = "shared_templates"
	FeatureWorkflowComments  = "workflow_comments"
	FeatureSSO               = "sso"
	FeatureAuditLogs         = "audit_logs"
	FeatureIPWhitelist       = "ip_whitelist"
	FeatureDataEncryption    = "data_encryption"
	FeaturePrioritySupport   = "priority_support"
	FeatureDedicatedSupport  = "dedicated_support"
	FeatureSLAGuarantee      = "sla_guarantee"
	FeatureCustomBranding    = "custom_branding"
	FeatureWhiteLabel        = "white_label"
)

// CreditBalance represents the credit balance for a workspace
type CreditBalance struct {
	Included   int `json:"included"`
	Purchased  int `json:"purchased"`
	Used       int `json:"used"`
	Remaining  int `json:"remaining"`
	MaxAllowed int `json:"max_allowed"`
	Overage    int `json:"overage"`
}

// PlanLimits represents all limits for a plan
type PlanLimits struct {
	PlanID              string `json:"plan_id"`
	PlanName            string `json:"plan_name"`
	CreditsIncluded     int    `json:"credits_included"`
	CreditsMax          int    `json:"credits_max"`
	ExecutionsLimit     int    `json:"executions_limit"`
	WorkflowsLimit      int    `json:"workflows_limit"`
	MembersLimit        int    `json:"members_limit"`
	CredentialsLimit    int    `json:"credentials_limit"`
	SchedulesLimit      int    `json:"schedules_limit"`
	WebhooksLimit       int    `json:"webhooks_limit"`
	ExecutionTimeout    int    `json:"execution_timeout"`
	MaxNodesPerWorkflow int    `json:"max_nodes_per_workflow"`
	RetentionDays       int    `json:"retention_days"`
}

// RecordOperationInput is the input for recording an operation
type RecordOperationInput struct {
	WorkspaceID uuid.UUID
	ExecutionID uuid.UUID
	WorkflowID  uuid.UUID
	NodeID      string
	NodeType    string
	Success     bool
	DurationMs  int
}

// CreateSubscriptionInput holds input for creating a subscription
type CreateSubscriptionInput struct {
	WorkspaceID          uuid.UUID
	PlanID               string
	BillingCycle         string
	StripeSubscriptionID *string
	StripeCustomerID     *string
}

// featureCheckMap maps feature names to their getter functions
var featureCheckMap = map[string]func(models.PlanFeatures) bool{
	FeatureWebhooks:          func(f models.PlanFeatures) bool { return f.Webhooks },
	FeatureSchedules:         func(f models.PlanFeatures) bool { return f.Schedules },
	FeatureManualTrigger:     func(f models.PlanFeatures) bool { return f.ManualTrigger },
	FeatureBasicNodes:        func(f models.PlanFeatures) bool { return f.BasicNodes },
	FeatureAdvancedNodes:     func(f models.PlanFeatures) bool { return f.AdvancedNodes },
	FeatureSubWorkflows:      func(f models.PlanFeatures) bool { return f.SubWorkflows },
	FeatureErrorWorkflow:     func(f models.PlanFeatures) bool { return f.ErrorWorkflow },
	FeatureAPIAccess:         func(f models.PlanFeatures) bool { return f.APIAccess },
	FeatureCustomFunctions:   func(f models.PlanFeatures) bool { return f.CustomFunctions },
	FeatureCustomAI:          func(f models.PlanFeatures) bool { return f.CustomAI },
	FeaturePriorityExecution: func(f models.PlanFeatures) bool { return f.PriorityExecution },
	FeatureParallelExecution: func(f models.PlanFeatures) bool { return f.ParallelExecution },
	FeatureRetryOnFailure:    func(f models.PlanFeatures) bool { return f.RetryOnFailure },
	FeatureTeamRoles:         func(f models.PlanFeatures) bool { return f.TeamRoles },
	FeatureSharedTemplates:   func(f models.PlanFeatures) bool { return f.SharedTemplates },
	FeatureWorkflowComments:  func(f models.PlanFeatures) bool { return f.WorkflowComments },
	FeatureSSO:               func(f models.PlanFeatures) bool { return f.SSO },
	FeatureAuditLogs:         func(f models.PlanFeatures) bool { return f.AuditLogs },
	FeatureIPWhitelist:       func(f models.PlanFeatures) bool { return f.IPWhitelist },
	FeatureDataEncryption:    func(f models.PlanFeatures) bool { return f.DataEncryption },
	FeaturePrioritySupport:   func(f models.PlanFeatures) bool { return f.PrioritySupport },
	FeatureDedicatedSupport:  func(f models.PlanFeatures) bool { return f.DedicatedSupport },
	FeatureSLAGuarantee:      func(f models.PlanFeatures) bool { return f.SLAGuarantee },
	FeatureCustomBranding:    func(f models.PlanFeatures) bool { return f.CustomBranding },
	FeatureWhiteLabel:        func(f models.PlanFeatures) bool { return f.WhiteLabel },
}
