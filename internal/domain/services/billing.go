package services

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/domain/models"
	"github.com/linkflow-ai/linkflow/internal/domain/repositories"
)

var (
	ErrPlanNotFound         = errors.New("plan not found")
	ErrSubscriptionNotFound = errors.New("subscription not found")
	ErrLimitExceeded        = errors.New("plan limit exceeded")
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

	s.workspaceRepo.UpdatePlan(ctx, input.WorkspaceID, input.PlanID)

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
