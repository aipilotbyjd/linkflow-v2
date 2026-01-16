package mappers

import (
	"github.com/linkflow-ai/linkflow/internal/adapters/persistence/postgres/models"
	"github.com/linkflow-ai/linkflow/internal/core/domain/workspace"
)

func WorkspaceToModel(ws *workspace.Workspace) *models.Workspace {
	return &models.Workspace{
		ID:               ws.ID,
		OwnerID:          ws.OwnerID,
		Name:             ws.Name,
		Slug:             ws.Slug,
		Description:      ws.Description,
		LogoURL:          ws.LogoURL,
		Website:          ws.Website,
		Timezone:         ws.Timezone,
		Language:         ws.Language,
		Currency:         ws.Currency,
		Country:          ws.Country,
		Industry:         ws.Industry,
		CompanySize:      ws.CompanySize,
		BillingEmail:     ws.BillingEmail,
		Settings:         ws.Settings,
		PlanID:           ws.PlanID,
		StripeCustomerID: ws.StripeCustomerID,
		CreatedAt:        ws.CreatedAt,
		UpdatedAt:        ws.UpdatedAt,
	}
}

func WorkspaceToDomain(m *models.Workspace) *workspace.Workspace {
	return &workspace.Workspace{
		ID:               m.ID,
		OwnerID:          m.OwnerID,
		Name:             m.Name,
		Slug:             m.Slug,
		Description:      m.Description,
		LogoURL:          m.LogoURL,
		Website:          m.Website,
		Timezone:         m.Timezone,
		Language:         m.Language,
		Currency:         m.Currency,
		Country:          m.Country,
		Industry:         m.Industry,
		CompanySize:      m.CompanySize,
		BillingEmail:     m.BillingEmail,
		Settings:         m.Settings,
		PlanID:           m.PlanID,
		StripeCustomerID: m.StripeCustomerID,
		CreatedAt:        m.CreatedAt,
		UpdatedAt:        m.UpdatedAt,
	}
}
