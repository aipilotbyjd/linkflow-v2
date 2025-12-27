package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/domain/models"
	"github.com/linkflow-ai/linkflow/internal/domain/repositories"
	"github.com/rs/zerolog/log"
)

// Billing errors
var (
	ErrPlanNotFound            = errors.New("plan not found")
	ErrSubscriptionNotFound    = errors.New("subscription not found")
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

type BillingService struct {
	planRepo         *repositories.PlanRepository
	subscriptionRepo *repositories.SubscriptionRepository
	usageRepo        *repositories.UsageRepository
	invoiceRepo      *repositories.InvoiceRepository
	workspaceRepo    *repositories.WorkspaceRepository
	// Optional repos for live counting
	workflowRepo   *repositories.WorkflowRepository
	memberRepo     *repositories.WorkspaceMemberRepository
	credentialRepo *repositories.CredentialRepository
}

// NewBillingService creates a new BillingService with required repositories.
// All repository parameters are required and must not be nil.
func NewBillingService(
	planRepo *repositories.PlanRepository,
	subscriptionRepo *repositories.SubscriptionRepository,
	usageRepo *repositories.UsageRepository,
	invoiceRepo *repositories.InvoiceRepository,
	workspaceRepo *repositories.WorkspaceRepository,
) *BillingService {
	if planRepo == nil || subscriptionRepo == nil || usageRepo == nil || invoiceRepo == nil || workspaceRepo == nil {
		panic("billing service: all repositories are required")
	}
	return &BillingService{
		planRepo:         planRepo,
		subscriptionRepo: subscriptionRepo,
		usageRepo:        usageRepo,
		invoiceRepo:      invoiceRepo,
		workspaceRepo:    workspaceRepo,
	}
}

// SetCountingRepos sets optional repositories for live resource counting
func (s *BillingService) SetCountingRepos(
	workflowRepo *repositories.WorkflowRepository,
	memberRepo *repositories.WorkspaceMemberRepository,
	credentialRepo *repositories.CredentialRepository,
) {
	s.workflowRepo = workflowRepo
	s.memberRepo = memberRepo
	s.credentialRepo = credentialRepo
}

// GetPlans returns all active billing plans.
func (s *BillingService) GetPlans(ctx context.Context) ([]models.Plan, error) {
	plans, err := s.planRepo.FindActive(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get plans: %w", err)
	}
	return plans, nil
}

// GetPlan returns a specific plan by ID.
func (s *BillingService) GetPlan(ctx context.Context, planID string) (*models.Plan, error) {
	plan, err := s.planRepo.FindByID(ctx, planID)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrPlanNotFound, planID)
	}
	return plan, nil
}

// GetSubscription returns the subscription for a workspace.
func (s *BillingService) GetSubscription(ctx context.Context, workspaceID uuid.UUID) (*models.Subscription, error) {
	subscription, err := s.subscriptionRepo.FindByWorkspaceID(ctx, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("%w for workspace %s", ErrSubscriptionNotFound, workspaceID)
	}
	return subscription, nil
}

// ErrSubscriptionExists is returned when trying to create a subscription for a workspace that already has one.
var ErrSubscriptionExists = errors.New("subscription already exists for this workspace")

// CreateSubscription creates a new subscription for a workspace.
func (s *BillingService) CreateSubscription(ctx context.Context, input CreateSubscriptionInput) (*models.Subscription, error) {
	// Validate billing cycle
	if input.BillingCycle != BillingCycleMonthly && input.BillingCycle != BillingCycleYearly {
		return nil, ErrInvalidBillingCycle
	}

	// Validate plan exists
	if _, err := s.planRepo.FindByID(ctx, input.PlanID); err != nil {
		return nil, fmt.Errorf("%w: %s", ErrPlanNotFound, input.PlanID)
	}

	// Check if subscription already exists
	existing, err := s.subscriptionRepo.FindByWorkspaceID(ctx, input.WorkspaceID)
	if err == nil && existing != nil {
		return nil, ErrSubscriptionExists
	}

	now := time.Now()
	periodEnd := now.AddDate(0, 1, 0)
	if input.BillingCycle == BillingCycleYearly {
		periodEnd = now.AddDate(1, 0, 0)
	}

	subscription := &models.Subscription{
		WorkspaceID:          input.WorkspaceID,
		PlanID:               input.PlanID,
		Status:               SubscriptionStatusActive,
		BillingCycle:         input.BillingCycle,
		StripeSubscriptionID: input.StripeSubscriptionID,
		StripeCustomerID:     input.StripeCustomerID,
		CurrentPeriodStart:   now,
		CurrentPeriodEnd:     periodEnd,
	}

	if err := s.subscriptionRepo.Create(ctx, subscription); err != nil {
		return nil, fmt.Errorf("failed to create subscription: %w", err)
	}

	if err := s.workspaceRepo.UpdatePlan(ctx, input.WorkspaceID, input.PlanID); err != nil {
		log.Error().
			Err(err).
			Str("workspace_id", input.WorkspaceID.String()).
			Str("plan_id", input.PlanID).
			Msg("Failed to update workspace plan after subscription creation")
	}

	log.Info().
		Str("workspace_id", input.WorkspaceID.String()).
		Str("plan_id", input.PlanID).
		Str("billing_cycle", input.BillingCycle).
		Msg("Subscription created")

	return subscription, nil
}

// ErrSamePlan is returned when trying to update to the same plan.
var ErrSamePlan = errors.New("subscription is already on this plan")

// UpdateSubscription updates the plan for an existing subscription.
func (s *BillingService) UpdateSubscription(ctx context.Context, workspaceID uuid.UUID, planID string) error {
	// Validate plan exists
	if _, err := s.planRepo.FindByID(ctx, planID); err != nil {
		return fmt.Errorf("%w: %s", ErrPlanNotFound, planID)
	}

	subscription, err := s.subscriptionRepo.FindByWorkspaceID(ctx, workspaceID)
	if err != nil {
		return fmt.Errorf("failed to find subscription: %w", err)
	}

	// Check if trying to update to the same plan
	if subscription.PlanID == planID {
		return ErrSamePlan
	}

	oldPlanID := subscription.PlanID
	subscription.PlanID = planID
	if err := s.subscriptionRepo.Update(ctx, subscription); err != nil {
		return fmt.Errorf("failed to update subscription: %w", err)
	}

	if err := s.workspaceRepo.UpdatePlan(ctx, workspaceID, planID); err != nil {
		return fmt.Errorf("failed to update workspace plan: %w", err)
	}

	log.Info().
		Str("workspace_id", workspaceID.String()).
		Str("old_plan_id", oldPlanID).
		Str("new_plan_id", planID).
		Msg("Subscription updated")

	return nil
}

func (s *BillingService) CancelSubscription(ctx context.Context, workspaceID uuid.UUID, cancelAtPeriodEnd bool) error {
	subscription, err := s.subscriptionRepo.FindByWorkspaceID(ctx, workspaceID)
	if err != nil {
		return fmt.Errorf("failed to find subscription: %w", err)
	}

	if cancelAtPeriodEnd {
		if err := s.subscriptionRepo.SetCancelAt(ctx, subscription.ID, &subscription.CurrentPeriodEnd); err != nil {
			return fmt.Errorf("failed to set cancel at period end: %w", err)
		}
		log.Info().
			Str("workspace_id", workspaceID.String()).
			Time("cancel_at", subscription.CurrentPeriodEnd).
			Msg("Subscription scheduled for cancellation at period end")
		return nil
	}

	now := time.Now()
	if err := s.subscriptionRepo.SetCancelAt(ctx, subscription.ID, &now); err != nil {
		return fmt.Errorf("failed to set cancel at: %w", err)
	}
	if err := s.subscriptionRepo.UpdateStatus(ctx, subscription.ID, SubscriptionStatusCanceled); err != nil {
		return fmt.Errorf("failed to update subscription status: %w", err)
	}

	log.Info().
		Str("workspace_id", workspaceID.String()).
		Msg("Subscription canceled immediately")

	return nil
}

func (s *BillingService) GetUsage(ctx context.Context, workspaceID uuid.UUID) (*models.Usage, error) {
	usage, err := s.usageRepo.GetOrCreateCurrentPeriod(ctx, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get usage: %w", err)
	}

	// Update live counts if repositories are available
	if s.workflowRepo != nil {
		if count, err := s.workflowRepo.CountByWorkspace(ctx, workspaceID); err != nil {
			log.Warn().Err(err).Str("workspace_id", workspaceID.String()).Msg("Failed to count workflows for usage")
		} else {
			usage.Workflows = int(count)
		}
	}
	if s.memberRepo != nil {
		if count, err := s.memberRepo.CountMembers(ctx, workspaceID); err != nil {
			log.Warn().Err(err).Str("workspace_id", workspaceID.String()).Msg("Failed to count members for usage")
		} else {
			usage.Members = int(count)
		}
	}
	if s.credentialRepo != nil {
		if count, err := s.credentialRepo.CountByWorkspace(ctx, workspaceID); err != nil {
			log.Warn().Err(err).Str("workspace_id", workspaceID.String()).Msg("Failed to count credentials for usage")
		} else {
			usage.Credentials = int(count)
		}
	}

	return usage, nil
}

func (s *BillingService) IncrementExecutions(ctx context.Context, workspaceID uuid.UUID) error {
	usage, err := s.usageRepo.GetOrCreateCurrentPeriod(ctx, workspaceID)
	if err != nil {
		return fmt.Errorf("failed to get usage: %w", err)
	}

	if err := s.usageRepo.IncrementExecutions(ctx, workspaceID, usage.PeriodStart, usage.PeriodEnd); err != nil {
		return fmt.Errorf("failed to increment executions: %w", err)
	}
	return nil
}

func (s *BillingService) CheckExecutionLimit(ctx context.Context, workspaceID uuid.UUID) (bool, error) {
	_, plan, err := s.getWorkspacePlan(ctx, workspaceID)
	if err != nil {
		return false, err
	}

	if plan.ExecutionsLimit == -1 {
		return true, nil
	}

	usage, err := s.usageRepo.GetOrCreateCurrentPeriod(ctx, workspaceID)
	if err != nil {
		return false, fmt.Errorf("failed to get usage: %w", err)
	}

	return usage.Executions < plan.ExecutionsLimit, nil
}

// getWorkspacePlan is a helper that fetches workspace and its associated plan
func (s *BillingService) getWorkspacePlan(ctx context.Context, workspaceID uuid.UUID) (*models.Workspace, *models.Plan, error) {
	workspace, err := s.workspaceRepo.FindByID(ctx, workspaceID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to find workspace: %w", err)
	}

	plan, err := s.planRepo.FindByID(ctx, workspace.PlanID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to find plan: %w", err)
	}

	return workspace, plan, nil
}

// checkLimit is a generic helper for checking resource limits
func checkLimit(limit, currentCount int) bool {
	if limit == -1 {
		return true
	}
	return currentCount < limit
}

func (s *BillingService) CheckWorkflowLimit(ctx context.Context, workspaceID uuid.UUID, currentCount int) (bool, error) {
	_, plan, err := s.getWorkspacePlan(ctx, workspaceID)
	if err != nil {
		return false, err
	}
	return checkLimit(plan.WorkflowsLimit, currentCount), nil
}

func (s *BillingService) CheckMemberLimit(ctx context.Context, workspaceID uuid.UUID, currentCount int) (bool, error) {
	_, plan, err := s.getWorkspacePlan(ctx, workspaceID)
	if err != nil {
		return false, err
	}
	return checkLimit(plan.MembersLimit, currentCount), nil
}

// CheckCredentialsLimit checks if workspace can create more credentials
func (s *BillingService) CheckCredentialsLimit(ctx context.Context, workspaceID uuid.UUID, currentCount int) (bool, error) {
	_, plan, err := s.getWorkspacePlan(ctx, workspaceID)
	if err != nil {
		return false, err
	}
	return checkLimit(plan.CredentialsLimit, currentCount), nil
}

// GetInvoices returns paginated invoices for a workspace.
func (s *BillingService) GetInvoices(ctx context.Context, workspaceID uuid.UUID, opts *repositories.ListOptions) ([]models.Invoice, int64, error) {
	invoices, total, err := s.invoiceRepo.FindByWorkspaceID(ctx, workspaceID, opts)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get invoices: %w", err)
	}
	return invoices, total, nil
}

func (s *BillingService) HandleStripeWebhook(ctx context.Context, eventType string, data map[string]interface{}) error {
	log.Info().
		Str("event_type", eventType).
		Msg("Processing Stripe webhook")

	switch eventType {
	case "customer.subscription.updated":
		// TODO: Implement subscription update handling
		// - Extract subscription ID from data
		// - Update local subscription record
		log.Debug().Msg("Stripe webhook: subscription updated - not yet implemented")

	case "customer.subscription.deleted":
		// TODO: Implement subscription deletion handling
		// - Extract subscription ID from data
		// - Mark subscription as canceled
		// - Downgrade workspace to free plan
		log.Debug().Msg("Stripe webhook: subscription deleted - not yet implemented")

	case "invoice.paid":
		// TODO: Implement successful payment handling
		// - Extract invoice details from data
		// - Create invoice record
		// - Update credits if applicable
		log.Debug().Msg("Stripe webhook: invoice paid - not yet implemented")

	case "invoice.payment_failed":
		// TODO: Implement failed payment handling
		// - Extract invoice details from data
		// - Mark subscription as past_due
		// - Send notification to workspace owner
		log.Debug().Msg("Stripe webhook: payment failed - not yet implemented")

	default:
		log.Debug().Str("event_type", eventType).Msg("Unhandled Stripe webhook event")
	}

	return nil
}

// Credit-based billing methods (Make.com style)

// ConsumeCredits consumes credits for an operation
func (s *BillingService) ConsumeCredits(ctx context.Context, workspaceID uuid.UUID, credits int) error {
	if credits <= 0 {
		return ErrInvalidCredits
	}

	usage, err := s.usageRepo.GetOrCreateCurrentPeriod(ctx, workspaceID)
	if err != nil {
		return fmt.Errorf("failed to get usage: %w", err)
	}

	_, plan, err := s.getWorkspacePlan(ctx, workspaceID)
	if err != nil {
		return err
	}

	// Check if credits would exceed max allowed
	if plan.CreditsMax != -1 && usage.CreditsUsed+credits > plan.CreditsMax {
		log.Warn().
			Str("workspace_id", workspaceID.String()).
			Int("requested", credits).
			Int("used", usage.CreditsUsed).
			Int("max", plan.CreditsMax).
			Msg("Credit limit exceeded")
		return ErrCreditsExceeded
	}

	if err := s.usageRepo.IncrementCredits(ctx, workspaceID, usage.PeriodStart, usage.PeriodEnd, credits); err != nil {
		return fmt.Errorf("failed to increment credits: %w", err)
	}

	return nil
}

// CheckCredits checks if workspace has enough credits
func (s *BillingService) CheckCredits(ctx context.Context, workspaceID uuid.UUID, requiredCredits int) (bool, error) {
	usage, err := s.usageRepo.GetOrCreateCurrentPeriod(ctx, workspaceID)
	if err != nil {
		return false, fmt.Errorf("failed to get usage: %w", err)
	}

	_, plan, err := s.getWorkspacePlan(ctx, workspaceID)
	if err != nil {
		return false, err
	}

	// Unlimited credits
	if plan.CreditsMax == -1 {
		return true, nil
	}

	// Check against max allowed (consistent with ConsumeCredits)
	return usage.CreditsUsed+requiredCredits <= plan.CreditsMax, nil
}

// GetCreditBalance returns current credit balance
func (s *BillingService) GetCreditBalance(ctx context.Context, workspaceID uuid.UUID) (*CreditBalance, error) {
	usage, err := s.usageRepo.GetOrCreateCurrentPeriod(ctx, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get usage: %w", err)
	}

	_, plan, err := s.getWorkspacePlan(ctx, workspaceID)
	if err != nil {
		return nil, err
	}

	totalCredits := usage.CreditsIncluded + usage.CreditsPurchased
	remaining := totalCredits - usage.CreditsUsed
	if remaining < 0 {
		remaining = 0
	}

	return &CreditBalance{
		Included:   usage.CreditsIncluded,
		Purchased:  usage.CreditsPurchased,
		Used:       usage.CreditsUsed,
		Remaining:  remaining,
		MaxAllowed: plan.CreditsMax,
		Overage:    usage.OverageCredits,
	}, nil
}

// CheckFeature checks if a feature is available on the current plan
func (s *BillingService) CheckFeature(ctx context.Context, workspaceID uuid.UUID, feature string) (bool, error) {
	_, plan, err := s.getWorkspacePlan(ctx, workspaceID)
	if err != nil {
		return false, err
	}

	features, err := parsePlanFeatures(plan.Features)
	if err != nil {
		return false, fmt.Errorf("failed to parse plan features: %w", err)
	}

	return isFeatureEnabled(features, feature), nil
}

// parsePlanFeatures converts JSON features to PlanFeatures struct
func parsePlanFeatures(featuresJSON models.JSON) (models.PlanFeatures, error) {
	var features models.PlanFeatures
	featuresBytes, err := json.Marshal(featuresJSON)
	if err != nil {
		return features, err
	}
	if err := json.Unmarshal(featuresBytes, &features); err != nil {
		return features, err
	}
	return features, nil
}

// isFeatureEnabled checks if a specific feature is enabled using the feature map
func isFeatureEnabled(features models.PlanFeatures, feature string) bool {
	if checker, ok := featureCheckMap[feature]; ok {
		return checker(features)
	}
	return false
}

// CheckScheduleLimit checks if workspace can create more schedules
func (s *BillingService) CheckScheduleLimit(ctx context.Context, workspaceID uuid.UUID, currentCount int) (bool, error) {
	_, plan, err := s.getWorkspacePlan(ctx, workspaceID)
	if err != nil {
		return false, err
	}
	return checkLimit(plan.SchedulesLimit, currentCount), nil
}

// CheckWebhookLimit checks if workspace can create more webhooks
func (s *BillingService) CheckWebhookLimit(ctx context.Context, workspaceID uuid.UUID, currentCount int) (bool, error) {
	_, plan, err := s.getWorkspacePlan(ctx, workspaceID)
	if err != nil {
		return false, err
	}
	return checkLimit(plan.WebhooksLimit, currentCount), nil
}

// GetPlanLimits returns all limits for a workspace's plan
func (s *BillingService) GetPlanLimits(ctx context.Context, workspaceID uuid.UUID) (*PlanLimits, error) {
	_, plan, err := s.getWorkspacePlan(ctx, workspaceID)
	if err != nil {
		return nil, err
	}

	return &PlanLimits{
		PlanID:              plan.ID,
		PlanName:            plan.Name,
		CreditsIncluded:     plan.CreditsIncluded,
		CreditsMax:          plan.CreditsMax,
		ExecutionsLimit:     plan.ExecutionsLimit,
		WorkflowsLimit:      plan.WorkflowsLimit,
		MembersLimit:        plan.MembersLimit,
		CredentialsLimit:    plan.CredentialsLimit,
		SchedulesLimit:      plan.SchedulesLimit,
		WebhooksLimit:       plan.WebhooksLimit,
		ExecutionTimeout:    plan.ExecutionTimeout,
		MaxNodesPerWorkflow: plan.MaxNodesPerWorkflow,
		RetentionDays:       plan.RetentionDays,
	}, nil
}

// RecordOperation records an operation for credit tracking
func (s *BillingService) RecordOperation(ctx context.Context, input RecordOperationInput) error {
	credits := models.GetCreditCost(input.NodeType)

	// Log the operation (optional, for detailed tracking)
	// This could be async/batched for performance

	// Consume credits
	return s.ConsumeCredits(ctx, input.WorkspaceID, credits)
}

// IncrementOperations increments the operation counter by a given count
func (s *BillingService) IncrementOperations(ctx context.Context, workspaceID uuid.UUID, count int) error {
	if count <= 0 {
		return nil // No-op for zero or negative counts
	}

	usage, err := s.usageRepo.GetOrCreateCurrentPeriod(ctx, workspaceID)
	if err != nil {
		return fmt.Errorf("failed to get usage: %w", err)
	}

	if err := s.usageRepo.IncrementOperationsBy(ctx, workspaceID, usage.PeriodStart, usage.PeriodEnd, count); err != nil {
		return fmt.Errorf("failed to increment operations: %w", err)
	}
	return nil
}

// IncrementExecutionSuccess increments successful execution counter
func (s *BillingService) IncrementExecutionSuccess(ctx context.Context, workspaceID uuid.UUID) error {
	usage, err := s.usageRepo.GetOrCreateCurrentPeriod(ctx, workspaceID)
	if err != nil {
		return fmt.Errorf("failed to get usage: %w", err)
	}

	if err := s.usageRepo.IncrementExecutionSuccess(ctx, workspaceID, usage.PeriodStart, usage.PeriodEnd); err != nil {
		return fmt.Errorf("failed to increment execution success: %w", err)
	}
	return nil
}

// IncrementExecutionFailed increments failed execution counter
func (s *BillingService) IncrementExecutionFailed(ctx context.Context, workspaceID uuid.UUID) error {
	usage, err := s.usageRepo.GetOrCreateCurrentPeriod(ctx, workspaceID)
	if err != nil {
		return fmt.Errorf("failed to get usage: %w", err)
	}

	if err := s.usageRepo.IncrementExecutionFailed(ctx, workspaceID, usage.PeriodStart, usage.PeriodEnd); err != nil {
		return fmt.Errorf("failed to increment execution failed: %w", err)
	}
	return nil
}
