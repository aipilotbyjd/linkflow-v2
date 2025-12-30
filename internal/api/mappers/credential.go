package mappers

import (
	"github.com/linkflow-ai/linkflow/internal/api/dto"
	"github.com/linkflow-ai/linkflow/internal/domain/models"
)

// CredentialToResponse converts a Credential model to CredentialResponse DTO
func CredentialToResponse(c *models.Credential) dto.CredentialResponse {
	var lastUsedAt *int64
	if c.LastUsedAt != nil {
		ts := c.LastUsedAt.Unix()
		lastUsedAt = &ts
	}

	return dto.CredentialResponse{
		ID:          c.ID.String(),
		Name:        c.Name,
		Type:        c.Type,
		Description: c.Description,
		LastUsedAt:  lastUsedAt,
		CreatedAt:   c.CreatedAt.Unix(),
	}
}

// CredentialsToResponse converts a slice of Credential models to CredentialResponse DTOs
func CredentialsToResponse(credentials []models.Credential) []dto.CredentialResponse {
	result := make([]dto.CredentialResponse, len(credentials))
	for i := range credentials {
		result[i] = CredentialToResponse(&credentials[i])
	}
	return result
}
