package services

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/domain/models"
	"github.com/linkflow-ai/linkflow/internal/domain/repositories"
)

// Billing errors
var (
	ErrPlanNotFound           = errors.New("plan not found")
	ErrSubscriptionNotFound   = errors.New("subscription not found")
	ErrLimitExceeded          = errors.New("plan limit exceeded")
	ErrCreditsExceeded        = errors.New("credit limit exceeded")
	ErrFeatureNotAvailable    = errors.New("feature not available on current plan")
	ErrWorkflowLimitExceeded  = errors.New("workflow limit exceeded")
	ErrMemberLimitExceeded    = errors.New("member limit exceeded")
	ErrExecutionLimitExceeded = errors.New("execution limit exceeded")
	ErrScheduleLimitExceeded  = errors.New("schedule limit exceeded")
	ErrWebhookLimitExceeded   = errors.New("webhook limit exceeded")
)

type BillingService struct {
	planRepo         *repositories.PlanRepository
	subscriptionRepo *repositories.SubscriptionRepository
	usageRepo        *repositories.UsageRepository
	invoiceRepo      *repositories.InvoiceRepository
	workspaceRepo    *repositories.WorkspaceRepository
}

func NewBillingService(
	planRepo *repositories.PlanRepository,
	subscriptionRepo *repositories.SubscriptionRepository,
	usageRepo *repositories.UsageRepository,
	invoiceRepo *repositories.InvoiceRepository,
	workspaceRepo *repositories.WorkspaceRepository,
) *BillingService {
	return &BillingService{
		planRepo:         planRepo,
		subscriptionRepo: subscriptionRepo,
		usageRepo:        usageRepo,
		invoiceRepo:      invoiceRepo,
		workspaceRepo:    workspaceRepo,
	}
}

func (s *BillingService) GetPlans(ctx context.Context) ([]models.Plan, error) {
	return s.planRepo.FindActive(ctx)
}

func (s *BillingService) GetPlan(ctx context.Context, planID string) (*models.Plan, error) {
	return s.planRepo.FindByID(ctx, planID)
}

func (s *BillingService) GetSubscription(ctx context.Context, workspaceID uuid.UUID) (*models.Subscription, error) {
	return s.subscriptionRepo.FindByWorkspaceID(ctx, workspaceID)
}

type CreateSubscriptionInput struct {
	WorkspaceID          uuid.UUID
	PlanID               string
	BillingCycle         string
	StripeSubscriptionID *string
	StripeCustomerID     *string
}

func (s *BillingService) CreateSubscription(ctx context.Context, input CreateSubscriptionInput) (*models.Subscription, error) {
	now := time.Now()
	periodEnd := now.AddDate(0, 1, 0)
	if input.BillingCycle == "yearly" {
		periodEnd = now.AddDate(1, 0, 0)
	}

	subscription := &models.Subscription{
		WorkspaceID:          input.WorkspaceID,
		PlanID:               input.PlanID,
		Status:               "active",
		BillingCycle:         input.BillingCycle,
		StripeSubscriptionID: input.StripeSubscriptionID,
		StripeCustomerID:     input.StripeCustomerID,
		CurrentPeriodStart:   now,
		CurrentPeriodEnd:     periodEnd,
	}

	if err := s.subscriptionRepo.Create(ctx, subscription); err != nil {
		return nil, err
	}

	_ = s.workspaceRepo.UpdatePlan(ctx, input.WorkspaceID, input.PlanID)

	return subscription, nil
}

func (s *BillingService) UpdateSubscription(ctx context.Context, workspaceID uuid.UUID, planID string) error {
	subscription, err := s.subscriptionRepo.FindByWorkspaceID(ctx, workspaceID)
	if err != nil {
		return err
	}

	subscription.PlanID = planID
	if err := s.subscriptionRepo.Update(ctx, subscription); err != nil {
		return err
	}

	return s.workspaceRepo.UpdatePlan(ctx, workspaceID, planID)
}

func (s *BillingService) CancelSubscription(ctx context.Context, workspaceID uuid.UUID, cancelAtPeriodEnd bool) error {
	subscription, err := s.subscriptionRepo.FindByWorkspaceID(ctx, workspaceID)
	if err != nil {
		return err
	}

	if cancelAtPeriodEnd {
		return s.subscriptionRepo.SetCancelAt(ctx, subscription.ID, &subscription.CurrentPeriodEnd)
	}

	now := time.Now()
	if err := s.subscriptionRepo.SetCancelAt(ctx, subscription.ID, &now); err != nil {
		return err
	}
	return s.subscriptionRepo.UpdateStatus(ctx, subscription.ID, "canceled")
}

func (s *BillingService) GetUsage(ctx context.Context, workspaceID uuid.UUID) (*models.Usage, error) {
	return s.usageRepo.GetOrCreateCurrentPeriod(ctx, workspaceID)
}

func (s *BillingService) IncrementExecutions(ctx context.Context, workspaceID uuid.UUID) error {
	usage, err := s.usageRepo.GetOrCreateCurrentPeriod(ctx, workspaceID)
	if err != nil {
		return err
	}

	return s.usageRepo.IncrementExecutions(ctx, workspaceID, usage.PeriodStart, usage.PeriodEnd)
}

func (s *BillingService) CheckExecutionLimit(ctx context.Context, workspaceID uuid.UUID) (bool, error) {
	workspace, err := s.workspaceRepo.FindByID(ctx, workspaceID)
	if err != nil {
		return false, err
	}

	plan, err := s.planRepo.FindByID(ctx, workspace.PlanID)
	if err != nil {
		return false, err
	}

	if plan.ExecutionsLimit == -1 {
		return true, nil
	}

	usage, err := s.usageRepo.GetOrCreateCurrentPeriod(ctx, workspaceID)
	if err != nil {
		return false, err
	}

	return usage.Executions < plan.ExecutionsLimit, nil
}

func (s *BillingService) CheckWorkflowLimit(ctx context.Context, workspaceID uuid.UUID, currentCount int) (bool, error) {
	workspace, err := s.workspaceRepo.FindByID(ctx, workspaceID)
	if err != nil {
		return false, err
	}

	plan, err := s.planRepo.FindByID(ctx, workspace.PlanID)
	if err != nil {
		return false, err
	}

	if plan.WorkflowsLimit == -1 {
		return true, nil
	}

	return currentCount < plan.WorkflowsLimit, nil
}

func (s *BillingService) CheckMemberLimit(ctx context.Context, workspaceID uuid.UUID, currentCount int) (bool, error) {
	workspace, err := s.workspaceRepo.FindByID(ctx, workspaceID)
	if err != nil {
		return false, err
	}

	plan, err := s.planRepo.FindByID(ctx, workspace.PlanID)
	if err != nil {
		return false, err
	}

	if plan.MembersLimit == -1 {
		return true, nil
	}

	return currentCount < plan.MembersLimit, nil
}

func (s *BillingService) GetInvoices(ctx context.Context, workspaceID uuid.UUID, opts *repositories.ListOptions) ([]models.Invoice, int64, error) {
	return s.invoiceRepo.FindByWorkspaceID(ctx, workspaceID, opts)
}

func (s *BillingService) HandleStripeWebhook(ctx context.Context, eventType string, data map[string]interface{}) error {
	switch eventType {
	case "customer.subscription.updated":
		// Handle subscription update
	case "customer.subscription.deleted":
		// Handle subscription cancellation
	case "invoice.paid":
		// Handle successful payment
	case "invoice.payment_failed":
		// Handle failed payment
	}
	return nil
}

// Credit-based billing methods (Make.com style)

// ConsumeCredits consumes credits for an operation
func (s *BillingService) ConsumeCredits(ctx context.Context, workspaceID uuid.UUID, credits int) error {
	usage, err := s.usageRepo.GetOrCreateCurrentPeriod(ctx, workspaceID)
	if err != nil {
		return err
	}

	workspace, err := s.workspaceRepo.FindByID(ctx, workspaceID)
	if err != nil {
		return err
	}

	plan, err := s.planRepo.FindByID(ctx, workspace.PlanID)
	if err != nil {
		return err
	}

	// Check if credits would exceed max
	if plan.CreditsMax != -1 && usage.CreditsUsed+credits > plan.CreditsMax {
		return ErrCreditsExceeded
	}

	// Update usage
	return s.usageRepo.IncrementCredits(ctx, workspaceID, usage.PeriodStart, usage.PeriodEnd, credits)
}

// CheckCredits checks if workspace has enough credits
func (s *BillingService) CheckCredits(ctx context.Context, workspaceID uuid.UUID, requiredCredits int) (bool, error) {
	usage, err := s.usageRepo.GetOrCreateCurrentPeriod(ctx, workspaceID)
	if err != nil {
		return false, err
	}

	workspace, err := s.workspaceRepo.FindByID(ctx, workspaceID)
	if err != nil {
		return false, err
	}

	plan, err := s.planRepo.FindByID(ctx, workspace.PlanID)
	if err != nil {
		return false, err
	}

	// Unlimited credits
	if plan.CreditsMax == -1 {
		return true, nil
	}

	totalCredits := usage.CreditsIncluded + usage.CreditsPurchased
	return usage.CreditsUsed+requiredCredits <= totalCredits, nil
}

// GetCreditBalance returns current credit balance
func (s *BillingService) GetCreditBalance(ctx context.Context, workspaceID uuid.UUID) (*CreditBalance, error) {
	usage, err := s.usageRepo.GetOrCreateCurrentPeriod(ctx, workspaceID)
	if err != nil {
		return nil, err
	}

	workspace, err := s.workspaceRepo.FindByID(ctx, workspaceID)
	if err != nil {
		return nil, err
	}

	plan, err := s.planRepo.FindByID(ctx, workspace.PlanID)
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

// CreditBalance represents the credit balance for a workspace
type CreditBalance struct {
	Included   int `json:"included"`
	Purchased  int `json:"purchased"`
	Used       int `json:"used"`
	Remaining  int `json:"remaining"`
	MaxAllowed int `json:"max_allowed"`
	Overage    int `json:"overage"`
}

// CheckFeature checks if a feature is available on the current plan
func (s *BillingService) CheckFeature(ctx context.Context, workspaceID uuid.UUID, feature string) (bool, error) {
	workspace, err := s.workspaceRepo.FindByID(ctx, workspaceID)
	if err != nil {
		return false, err
	}

	plan, err := s.planRepo.FindByID(ctx, workspace.PlanID)
	if err != nil {
		return false, err
	}

	var features models.PlanFeatures
	featuresBytes, err := json.Marshal(plan.Features)
	if err != nil {
		return false, err
	}
	if err := json.Unmarshal(featuresBytes, &features); err != nil {
		return false, err
	}

	return isFeatureEnabled(features, feature), nil
}

// isFeatureEnabled checks if a specific feature is enabled
func isFeatureEnabled(features models.PlanFeatures, feature string) bool {
	switch feature {
	case "webhooks":
		return features.Webhooks
	case "schedules":
		return features.Schedules
	case "manual_trigger":
		return features.ManualTrigger
	case "basic_nodes":
		return features.BasicNodes
	case "advanced_nodes":
		return features.AdvancedNodes
	case "sub_workflows":
		return features.SubWorkflows
	case "error_workflow":
		return features.ErrorWorkflow
	case "api_access":
		return features.APIAccess
	case "custom_functions":
		return features.CustomFunctions
	case "custom_ai":
		return features.CustomAI
	case "priority_execution":
		return features.PriorityExecution
	case "parallel_execution":
		return features.ParallelExecution
	case "retry_on_failure":
		return features.RetryOnFailure
	case "team_roles":
		return features.TeamRoles
	case "shared_templates":
		return features.SharedTemplates
	case "workflow_comments":
		return features.WorkflowComments
	case "sso":
		return features.SSO
	case "audit_logs":
		return features.AuditLogs
	case "ip_whitelist":
		return features.IPWhitelist
	case "data_encryption":
		return features.DataEncryption
	case "priority_support":
		return features.PrioritySupport
	case "dedicated_support":
		return features.DedicatedSupport
	case "sla_guarantee":
		return features.SLAGuarantee
	case "custom_branding":
		return features.CustomBranding
	case "white_label":
		return features.WhiteLabel
	default:
		return false
	}
}

// CheckScheduleLimit checks if workspace can create more schedules
func (s *BillingService) CheckScheduleLimit(ctx context.Context, workspaceID uuid.UUID, currentCount int) (bool, error) {
	workspace, err := s.workspaceRepo.FindByID(ctx, workspaceID)
	if err != nil {
		return false, err
	}

	plan, err := s.planRepo.FindByID(ctx, workspace.PlanID)
	if err != nil {
		return false, err
	}

	if plan.SchedulesLimit == -1 {
		return true, nil
	}

	return currentCount < plan.SchedulesLimit, nil
}

// CheckWebhookLimit checks if workspace can create more webhooks
func (s *BillingService) CheckWebhookLimit(ctx context.Context, workspaceID uuid.UUID, currentCount int) (bool, error) {
	workspace, err := s.workspaceRepo.FindByID(ctx, workspaceID)
	if err != nil {
		return false, err
	}

	plan, err := s.planRepo.FindByID(ctx, workspace.PlanID)
	if err != nil {
		return false, err
	}

	if plan.WebhooksLimit == -1 {
		return true, nil
	}

	return currentCount < plan.WebhooksLimit, nil
}

// GetPlanLimits returns all limits for a workspace's plan
func (s *BillingService) GetPlanLimits(ctx context.Context, workspaceID uuid.UUID) (*PlanLimits, error) {
	workspace, err := s.workspaceRepo.FindByID(ctx, workspaceID)
	if err != nil {
		return nil, err
	}

	plan, err := s.planRepo.FindByID(ctx, workspace.PlanID)
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

// RecordOperation records an operation for credit tracking
func (s *BillingService) RecordOperation(ctx context.Context, input RecordOperationInput) error {
	credits := models.GetCreditCost(input.NodeType)

	// Log the operation (optional, for detailed tracking)
	// This could be async/batched for performance

	// Consume credits
	return s.ConsumeCredits(ctx, input.WorkspaceID, credits)
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

// IncrementOperations increments the operation counter
func (s *BillingService) IncrementOperations(ctx context.Context, workspaceID uuid.UUID) error {
	usage, err := s.usageRepo.GetOrCreateCurrentPeriod(ctx, workspaceID)
	if err != nil {
		return err
	}

	return s.usageRepo.IncrementOperations(ctx, workspaceID, usage.PeriodStart, usage.PeriodEnd)
}
