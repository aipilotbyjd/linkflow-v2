package billing

import (
	"time"

	"github.com/google/uuid"
)

// BYOKConfig - Bring Your Own Key configuration for AI providers
type BYOKConfig struct {
	ID          uuid.UUID    `gorm:"type:uuid;primaryKey" json:"id"`
	WorkspaceID uuid.UUID    `gorm:"type:uuid;index;not null" json:"workspace_id"`
	Provider    AIProvider   `gorm:"size:50;not null" json:"provider"`
	
	// Encrypted API key (stored encrypted in DB)
	APIKeyEncrypted string `gorm:"size:500" json:"-"`
	APIKeyMasked    string `gorm:"-" json:"api_key_masked"` // For display: sk-...xxxx
	
	// Optional settings
	OrganizationID string `gorm:"size:100" json:"organization_id,omitempty"`
	BaseURL        string `gorm:"size:255" json:"base_url,omitempty"` // For custom endpoints
	
	// Status
	IsActive     bool       `gorm:"default:true" json:"is_active"`
	IsValid      bool       `gorm:"default:false" json:"is_valid"` // Set after validation
	LastValidated *time.Time `json:"last_validated,omitempty"`
	LastUsed     *time.Time `json:"last_used,omitempty"`
	ErrorMessage string     `gorm:"size:500" json:"error_message,omitempty"`
	
	// Usage tracking (even with BYOK, track for analytics)
	TotalRequests   int64 `gorm:"default:0" json:"total_requests"`
	TotalTokensUsed int64 `gorm:"default:0" json:"total_tokens_used"`
	
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type AIProvider string

const (
	ProviderOpenAI    AIProvider = "openai"
	ProviderAnthropic AIProvider = "anthropic"
	ProviderGoogle    AIProvider = "google"
	ProviderAzure     AIProvider = "azure_openai"
	ProviderCohere    AIProvider = "cohere"
	ProviderMistral   AIProvider = "mistral"
	ProviderGroq      AIProvider = "groq"
	ProviderTogether  AIProvider = "together"
	ProviderPerplexity AIProvider = "perplexity"
)

// ProviderInfo contains info about AI providers
type ProviderInfo struct {
	ID          AIProvider `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	DocsURL     string     `json:"docs_url"`
	Models      []string   `json:"models"`
	SupportsOrg bool       `json:"supports_organization_id"`
}

// SupportedProviders lists all supported BYOK providers
var SupportedProviders = []ProviderInfo{
	{
		ID:          ProviderOpenAI,
		Name:        "OpenAI",
		Description: "GPT-4, GPT-3.5, DALL-E, Whisper",
		DocsURL:     "https://platform.openai.com/api-keys",
		Models:      []string{"gpt-4", "gpt-4-turbo", "gpt-4o", "gpt-3.5-turbo", "dall-e-3", "whisper-1"},
		SupportsOrg: true,
	},
	{
		ID:          ProviderAnthropic,
		Name:        "Anthropic",
		Description: "Claude 3 Opus, Sonnet, Haiku",
		DocsURL:     "https://console.anthropic.com/settings/keys",
		Models:      []string{"claude-3-opus", "claude-3-sonnet", "claude-3-haiku", "claude-3.5-sonnet"},
		SupportsOrg: false,
	},
	{
		ID:          ProviderGoogle,
		Name:        "Google AI",
		Description: "Gemini Pro, Gemini Ultra",
		DocsURL:     "https://aistudio.google.com/app/apikey",
		Models:      []string{"gemini-pro", "gemini-ultra", "gemini-1.5-pro"},
		SupportsOrg: false,
	},
	{
		ID:          ProviderAzure,
		Name:        "Azure OpenAI",
		Description: "OpenAI models on Azure",
		DocsURL:     "https://portal.azure.com",
		Models:      []string{"gpt-4", "gpt-35-turbo"},
		SupportsOrg: false,
	},
	{
		ID:          ProviderGroq,
		Name:        "Groq",
		Description: "Ultra-fast LLM inference",
		DocsURL:     "https://console.groq.com/keys",
		Models:      []string{"llama-3.1-70b", "mixtral-8x7b"},
		SupportsOrg: false,
	},
}

// NewBYOKConfig creates a new BYOK configuration
func NewBYOKConfig(workspaceID uuid.UUID, provider AIProvider) *BYOKConfig {
	return &BYOKConfig{
		ID:          uuid.New(),
		WorkspaceID: workspaceID,
		Provider:    provider,
		IsActive:    true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

// MaskAPIKey creates a masked version of the API key for display
func MaskAPIKey(key string) string {
	if len(key) < 8 {
		return "****"
	}
	return key[:4] + "..." + key[len(key)-4:]
}

// SetAPIKey sets and masks the API key
func (b *BYOKConfig) SetAPIKey(encryptedKey, originalKey string) {
	b.APIKeyEncrypted = encryptedKey
	b.APIKeyMasked = MaskAPIKey(originalKey)
	b.UpdatedAt = time.Now()
}

// MarkValid marks the key as validated
func (b *BYOKConfig) MarkValid() {
	now := time.Now()
	b.IsValid = true
	b.LastValidated = &now
	b.ErrorMessage = ""
	b.UpdatedAt = now
}

// MarkInvalid marks the key as invalid with error
func (b *BYOKConfig) MarkInvalid(err string) {
	b.IsValid = false
	b.ErrorMessage = err
	b.UpdatedAt = time.Now()
}

// RecordUsage records API usage
func (b *BYOKConfig) RecordUsage(tokens int64) {
	now := time.Now()
	b.TotalRequests++
	b.TotalTokensUsed += tokens
	b.LastUsed = &now
	b.UpdatedAt = now
}

// IsBYOKEnabled checks if workspace has valid BYOK for a provider
func IsBYOKEnabled(configs []*BYOKConfig, provider AIProvider) bool {
	for _, c := range configs {
		if c.Provider == provider && c.IsActive && c.IsValid {
			return true
		}
	}
	return false
}

// GetProviderInfo returns info for a provider
func GetProviderInfo(provider AIProvider) *ProviderInfo {
	for _, p := range SupportedProviders {
		if p.ID == provider {
			return &p
		}
	}
	return nil
}
