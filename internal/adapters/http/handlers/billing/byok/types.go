package byok

import (
	"time"

	"github.com/linkflow-ai/linkflow/internal/core/domain/billing"
)

// CreateBYOKRequest for adding a new API key
type CreateBYOKRequest struct {
	Provider       string `json:"provider" validate:"required"`
	Name           string `json:"name" validate:"required,max=100"`
	APIKey         string `json:"api_key" validate:"required"`
	OrganizationID string `json:"organization_id,omitempty"`
	BaseURL        string `json:"base_url,omitempty"`
}

// UpdateBYOKRequest for updating BYOK config
type UpdateBYOKRequest struct {
	Name           *string `json:"name,omitempty"`
	APIKey         *string `json:"api_key,omitempty"`
	OrganizationID *string `json:"organization_id,omitempty"`
	BaseURL        *string `json:"base_url,omitempty"`
	IsActive       *bool   `json:"is_active,omitempty"`
}

// BYOKResponse for API responses
type BYOKResponse struct {
	ID              string     `json:"id"`
	Provider        string     `json:"provider"`
	ProviderName    string     `json:"provider_name"`
	Name            string     `json:"name"`
	APIKeyMasked    string     `json:"api_key_masked"`
	OrganizationID  string     `json:"organization_id,omitempty"`
	BaseURL         string     `json:"base_url,omitempty"`
	IsActive        bool       `json:"is_active"`
	IsValid         bool       `json:"is_valid"`
	LastValidated   *time.Time `json:"last_validated,omitempty"`
	LastUsed        *time.Time `json:"last_used,omitempty"`
	ErrorMessage    string     `json:"error_message,omitempty"`
	TotalRequests   int64      `json:"total_requests"`
	TotalTokensUsed int64      `json:"total_tokens_used"`
	CreatedAt       time.Time  `json:"created_at"`
}

// ProviderResponse for listing supported providers
type ProviderResponse struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	DocsURL     string   `json:"docs_url"`
	Models      []string `json:"models"`
	SupportsOrg bool     `json:"supports_organization_id"`
}

// ToBYOKResponse converts domain to response
func ToBYOKResponse(b *billing.BYOKConfig) BYOKResponse {
	providerName := string(b.Provider)
	if info := billing.GetProviderInfo(b.Provider); info != nil {
		providerName = info.Name
	}

	return BYOKResponse{
		ID:              b.ID.String(),
		Provider:        string(b.Provider),
		ProviderName:    providerName,
		Name:            b.Name,
		APIKeyMasked:    b.APIKeyMasked,
		OrganizationID:  b.OrganizationID,
		BaseURL:         b.BaseURL,
		IsActive:        b.IsActive,
		IsValid:         b.IsValid,
		LastValidated:   b.LastValidated,
		LastUsed:        b.LastUsed,
		ErrorMessage:    b.ErrorMessage,
		TotalRequests:   b.TotalRequests,
		TotalTokensUsed: b.TotalTokensUsed,
		CreatedAt:       b.CreatedAt,
	}
}

// ToProviderResponse converts provider info to response
func ToProviderResponse(p billing.ProviderInfo) ProviderResponse {
	return ProviderResponse{
		ID:          string(p.ID),
		Name:        p.Name,
		Description: p.Description,
		DocsURL:     p.DocsURL,
		Models:      p.Models,
		SupportsOrg: p.SupportsOrg,
	}
}
