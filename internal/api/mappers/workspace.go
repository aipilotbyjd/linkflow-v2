package mappers

import (
	"github.com/linkflow-ai/linkflow/internal/api/dto"
	"github.com/linkflow-ai/linkflow/internal/domain/models"
)

// WorkspaceToResponse converts a Workspace model to WorkspaceResponse DTO
func WorkspaceToResponse(ws *models.Workspace) dto.WorkspaceResponse {
	var settings interface{}
	if ws.Settings != nil {
		settings = ws.Settings
	}

	return dto.WorkspaceResponse{
		ID:           ws.ID.String(),
		OwnerID:      ws.OwnerID.String(),
		Name:         ws.Name,
		Slug:         ws.Slug,
		Description:  ws.Description,
		LogoURL:      ws.LogoURL,
		Website:      ws.Website,
		Timezone:     ws.Timezone,
		Language:     ws.Language,
		Currency:     ws.Currency,
		Country:      ws.Country,
		Industry:     ws.Industry,
		CompanySize:  ws.CompanySize,
		BillingEmail: ws.BillingEmail,
		Settings:     settings,
		PlanID:       ws.PlanID,
		CreatedAt:    ws.CreatedAt.Unix(),
		UpdatedAt:    ws.UpdatedAt.Unix(),
	}
}

// WorkspacesToResponse converts a slice of Workspace models to WorkspaceResponse DTOs
func WorkspacesToResponse(workspaces []models.Workspace) []dto.WorkspaceResponse {
	result := make([]dto.WorkspaceResponse, len(workspaces))
	for i := range workspaces {
		result[i] = WorkspaceToResponse(&workspaces[i])
	}
	return result
}

// WorkspaceMemberToResponse converts a WorkspaceMember model to WorkspaceMemberResponse DTO
func WorkspaceMemberToResponse(m *models.WorkspaceMember) dto.WorkspaceMemberResponse {
	var joinedAt, invitedAt *int64
	if m.JoinedAt != nil {
		ts := m.JoinedAt.Unix()
		joinedAt = &ts
	}
	if m.InvitedAt != nil {
		ts := m.InvitedAt.Unix()
		invitedAt = &ts
	}

	return dto.WorkspaceMemberResponse{
		ID:        m.ID.String(),
		User:      UserToResponse(&m.User),
		Role:      m.Role,
		JoinedAt:  joinedAt,
		InvitedAt: invitedAt,
	}
}

// WorkspaceMembersToResponse converts a slice of WorkspaceMember models to DTOs
func WorkspaceMembersToResponse(members []models.WorkspaceMember) []dto.WorkspaceMemberResponse {
	result := make([]dto.WorkspaceMemberResponse, len(members))
	for i := range members {
		result[i] = WorkspaceMemberToResponse(&members[i])
	}
	return result
}
