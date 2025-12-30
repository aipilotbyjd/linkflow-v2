package mappers

import (
	"github.com/linkflow-ai/linkflow/internal/api/dto"
	"github.com/linkflow-ai/linkflow/internal/domain/models"
)

// PlanToResponse converts a Plan model to PlanResponse DTO
func PlanToResponse(p *models.Plan) dto.PlanResponse {
	return dto.PlanResponse{
		ID:               p.ID,
		Name:             p.Name,
		Description:      p.Description,
		PriceMonthly:     p.PriceMonthly,
		PriceYearly:      p.PriceYearly,
		ExecutionsLimit:  p.ExecutionsLimit,
		WorkflowsLimit:   p.WorkflowsLimit,
		MembersLimit:     p.MembersLimit,
		CredentialsLimit: p.CredentialsLimit,
		Features:         p.Features,
	}
}

// PlansToResponse converts a slice of Plan models to PlanResponse DTOs
func PlansToResponse(plans []models.Plan) []dto.PlanResponse {
	result := make([]dto.PlanResponse, len(plans))
	for i := range plans {
		result[i] = PlanToResponse(&plans[i])
	}
	return result
}

// SubscriptionToResponse converts a Subscription model to SubscriptionResponse DTO
func SubscriptionToResponse(s *models.Subscription) dto.SubscriptionResponse {
	var cancelAt *int64
	if s.CancelAt != nil {
		ts := s.CancelAt.Unix()
		cancelAt = &ts
	}

	return dto.SubscriptionResponse{
		ID:                 s.ID.String(),
		PlanID:             s.PlanID,
		Status:             s.Status,
		BillingCycle:       s.BillingCycle,
		CurrentPeriodStart: s.CurrentPeriodStart.Unix(),
		CurrentPeriodEnd:   s.CurrentPeriodEnd.Unix(),
		CancelAt:           cancelAt,
	}
}
