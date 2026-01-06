package dto

import "github.com/linkflow-ai/linkflow/internal/domain/models"

// Credential requests

type CreateCredentialRequest struct {
	Name         string                `json:"name" validate:"required,min=1,max=100"`
	Type         string                `json:"type" validate:"required,oneof=api_key oauth2 basic bearer custom"`
	Data         models.CredentialData `json:"data" validate:"required"`
	Description  *string               `json:"description,omitempty" validate:"omitempty,max=500"`
	SharingScope *string               `json:"sharing_scope,omitempty" validate:"omitempty,oneof=private workspace specific"`
}

type UpdateCredentialRequest struct {
	Name        *string                `json:"name,omitempty" validate:"omitempty,min=1,max=100"`
	Data        *models.CredentialData `json:"data,omitempty"`
	Description *string                `json:"description,omitempty" validate:"omitempty,max=500"`
}

type UpdateSharingScopeRequest struct {
	SharingScope string `json:"sharing_scope" validate:"required,oneof=private workspace specific"`
}

type ShareCredentialRequest struct {
	UserIDs []string `json:"user_ids" validate:"required,min=1,dive,uuid"`
}

// Credential responses

type CredentialResponse struct {
	ID                string                    `json:"id"`
	WorkspaceID       string                    `json:"workspace_id"`
	CreatedBy         string                    `json:"created_by"`
	Name              string                    `json:"name"`
	Type              string                    `json:"type"`
	Description       *string                   `json:"description,omitempty"`
	Provider          *string                   `json:"provider,omitempty"`
	ProviderAccountID *string                   `json:"provider_account_id,omitempty"`
	TokenExpiresAt    *int64                    `json:"token_expires_at,omitempty"`
	SharingScope      string                    `json:"sharing_scope"`
	IsOwner           bool                      `json:"is_owner"`
	CanEdit           bool                      `json:"can_edit"`
	CanShare          bool                      `json:"can_share"`
	Shares            []CredentialShareResponse `json:"shares,omitempty"`
	LastUsedAt        *int64                    `json:"last_used_at,omitempty"`
	CreatedAt         int64                     `json:"created_at"`
	UpdatedAt         int64                     `json:"updated_at"`
}

type CredentialShareResponse struct {
	ID         string       `json:"id"`
	UserID     string       `json:"user_id"`
	User       *UserSummary `json:"user,omitempty"`
	Permission string       `json:"permission"`
	SharedBy   string       `json:"shared_by"`
	CreatedAt  int64        `json:"created_at"`
}

type UserSummary struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	FirstName string `json:"first_name,omitempty"`
	LastName  string `json:"last_name,omitempty"`
}
