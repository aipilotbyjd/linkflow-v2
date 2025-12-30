package services

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/domain/models"
	"github.com/linkflow-ai/linkflow/internal/domain/repositories"
	"github.com/rs/zerolog/log"
)

// GetUsage returns usage stats for a workspace
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

// GetInvoices returns paginated invoices for a workspace.
func (s *BillingService) GetInvoices(ctx context.Context, workspaceID uuid.UUID, opts *repositories.ListOptions) ([]models.Invoice, int64, error) {
	invoices, total, err := s.invoiceRepo.FindByWorkspaceID(ctx, workspaceID, opts)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get invoices: %w", err)
	}
	return invoices, total, nil
}

// IncrementExecutions increments the execution counter
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

// IncrementOperations increments the operation counter by a given count
func (s *BillingService) IncrementOperations(ctx context.Context, workspaceID uuid.UUID, count int) error {
	if count <= 0 {
		return nil
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

// RecordOperation records an operation for credit tracking
func (s *BillingService) RecordOperation(ctx context.Context, input RecordOperationInput) error {
	credits := models.GetCreditCost(input.NodeType)
	return s.ConsumeCredits(ctx, input.WorkspaceID, credits)
}

// CheckExecutionLimit checks if workspace can execute more workflows
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

// CheckWorkflowLimit checks if workspace can create more workflows
func (s *BillingService) CheckWorkflowLimit(ctx context.Context, workspaceID uuid.UUID, currentCount int) (bool, error) {
	_, plan, err := s.getWorkspacePlan(ctx, workspaceID)
	if err != nil {
		return false, err
	}
	return checkLimit(plan.WorkflowsLimit, currentCount), nil
}

// CheckMemberLimit checks if workspace can add more members
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

// checkLimit is a generic helper for checking resource limits
func checkLimit(limit, currentCount int) bool {
	if limit == -1 {
		return true
	}
	return currentCount < limit
}
