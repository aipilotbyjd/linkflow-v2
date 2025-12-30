package dto

import "github.com/linkflow-ai/linkflow/internal/domain/models"

// Credential requests

type CreateCredentialRequest struct {
	Name        string                `json:"name" validate:"required,min=1,max=100"`
	Type        string                `json:"type" validate:"required,oneof=api_key oauth2 basic bearer custom"`
	Data        models.CredentialData `json:"data" validate:"required"`
	Description *string               `json:"description,omitempty" validate:"omitempty,max=500"`
}

type UpdateCredentialRequest struct {
	Name        *string                `json:"name,omitempty" validate:"omitempty,min=1,max=100"`
	Data        *models.CredentialData `json:"data,omitempty"`
	Description *string                `json:"description,omitempty" validate:"omitempty,max=500"`
}

// Credential responses

type CredentialResponse struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Type        string  `json:"type"`
	Description *string `json:"description,omitempty"`
	LastUsedAt  *int64  `json:"last_used_at,omitempty"`
	CreatedAt   int64   `json:"created_at"`
}
