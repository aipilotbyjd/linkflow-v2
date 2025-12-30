package services

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/domain/models"
	"github.com/linkflow-ai/linkflow/internal/domain/repositories"
)

// BillingService handles all billing-related operations
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
