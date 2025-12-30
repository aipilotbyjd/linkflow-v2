package services

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/domain/models"
)

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
