package services

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/domain/models"
	"github.com/rs/zerolog/log"
)

// GetSubscription returns the subscription for a workspace.
func (s *BillingService) GetSubscription(ctx context.Context, workspaceID uuid.UUID) (*models.Subscription, error) {
	subscription, err := s.subscriptionRepo.FindByWorkspaceID(ctx, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("%w for workspace %s", ErrSubscriptionNotFound, workspaceID)
	}
	return subscription, nil
}

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

// CancelSubscription cancels a subscription
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

// HandleStripeWebhook processes Stripe webhook events
func (s *BillingService) HandleStripeWebhook(ctx context.Context, eventType string, data map[string]interface{}) error {
	log.Info().
		Str("event_type", eventType).
		Msg("Processing Stripe webhook")

	switch eventType {
	case "customer.subscription.updated":
		log.Debug().Msg("Stripe webhook: subscription updated - not yet implemented")

	case "customer.subscription.deleted":
		log.Debug().Msg("Stripe webhook: subscription deleted - not yet implemented")

	case "invoice.paid":
		log.Debug().Msg("Stripe webhook: invoice paid - not yet implemented")

	case "invoice.payment_failed":
		log.Debug().Msg("Stripe webhook: payment failed - not yet implemented")

	default:
		log.Debug().Str("event_type", eventType).Msg("Unhandled Stripe webhook event")
	}

	return nil
}
