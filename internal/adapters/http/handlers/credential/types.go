package credential

import (
	"github.com/linkflow-ai/linkflow/internal/core/domain/credential"
	"github.com/linkflow-ai/linkflow/internal/shared/types"
)

// CreateRequest represents credential creation request
type CreateRequest struct {
	Name        string                  `json:"name" validate:"required,min=1,max=100"`
	Description *string                 `json:"description,omitempty" validate:"omitempty,max=500"`
	Type        credential.Type         `json:"type" validate:"required,credential_type"`
	Provider    string                  `json:"provider" validate:"omitempty,max=50"`
	Data        types.JSON              `json:"data" validate:"required"`
	Scope       credential.SharingScope `json:"scope" validate:"omitempty,sharing_scope"`
}

// UpdateRequest represents credential update request
type UpdateRequest struct {
	Name        *string                  `json:"name,omitempty"`
	Description *string                  `json:"description,omitempty"`
	Data        types.JSON               `json:"data,omitempty"`
	Scope       *credential.SharingScope `json:"scope,omitempty"`
}

// CredentialResponse represents credential in responses
type CredentialResponse struct {
	ID          string  `json:"id"`
	WorkspaceID string  `json:"workspace_id"`
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	Type        string  `json:"type"`
	Provider    string  `json:"provider"`
	Scope       string  `json:"scope"`
	CreatedBy   string  `json:"created_by"`
	LastUsedAt  *string `json:"last_used_at,omitempty"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}

// ToCredentialResponse converts domain credential to response
func ToCredentialResponse(c *credential.Credential) CredentialResponse {
	provider := ""
	if c.Provider != nil {
		provider = *c.Provider
	}
	resp := CredentialResponse{
		ID:          c.ID.String(),
		WorkspaceID: c.WorkspaceID.String(),
		Name:        c.Name,
		Description: c.Description,
		Type:        string(c.Type),
		Provider:    provider,
		Scope:       string(c.SharingScope),
		CreatedBy:   c.CreatedBy.String(),
		CreatedAt:   c.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:   c.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}

	if c.LastUsedAt != nil {
		s := c.LastUsedAt.Format("2006-01-02T15:04:05Z")
		resp.LastUsedAt = &s
	}

	return resp
}
