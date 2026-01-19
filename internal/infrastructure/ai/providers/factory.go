package providers

import (
	"fmt"

	"github.com/linkflow-ai/linkflow/internal/core/domain/ai"
	"github.com/linkflow-ai/linkflow/internal/infrastructure/ai/providers/anthropic"
	"github.com/linkflow-ai/linkflow/internal/infrastructure/ai/providers/google"
	"github.com/linkflow-ai/linkflow/internal/infrastructure/ai/providers/openai"
)

// Factory creates provider adapters
type Factory struct{}

// NewFactory creates a new provider factory
func NewFactory() *Factory {
	return &Factory{}
}

// CreateAdapter creates a provider adapter based on config
func (f *Factory) CreateAdapter(config *ai.ProviderConfig) (ai.ProviderAdapter, error) {
	if config == nil {
		return nil, fmt.Errorf("provider config is required")
	}

	if config.APIKey == "" {
		return nil, fmt.Errorf("API key is required")
	}

	switch config.Provider {
	case ai.ProviderOpenAI:
		return openai.NewAdapter(config), nil

	case ai.ProviderAnthropic:
		return anthropic.NewAdapter(config), nil

	case ai.ProviderGoogle:
		return google.NewAdapter(config), nil

	case ai.ProviderAzure:
		// Azure uses OpenAI-compatible API
		return openai.NewAdapter(config), nil

	case ai.ProviderOllama:
		// Ollama uses OpenAI-compatible API
		if config.BaseURL == "" {
			config.BaseURL = "http://localhost:11434/v1"
		}
		return openai.NewAdapter(config), nil

	default:
		return nil, fmt.Errorf("unsupported provider: %s", config.Provider)
	}
}

// CreateAdapterFromCredential creates an adapter from credential data
func (f *Factory) CreateAdapterFromCredential(provider ai.Provider, credData map[string]interface{}) (ai.ProviderAdapter, error) {
	config := &ai.ProviderConfig{
		Provider: provider,
	}

	// Extract API key
	if apiKey, ok := credData["api_key"].(string); ok {
		config.APIKey = apiKey
	}

	// Extract organization ID (OpenAI)
	if orgID, ok := credData["org_id"].(string); ok {
		config.OrgID = orgID
	}

	// Extract base URL
	if baseURL, ok := credData["base_url"].(string); ok {
		config.BaseURL = baseURL
	}

	// Extract project ID (Google)
	if projectID, ok := credData["project_id"].(string); ok {
		config.ProjectID = projectID
	}

	// Extract region (Google/Azure)
	if region, ok := credData["region"].(string); ok {
		config.Region = region
	}

	return f.CreateAdapter(config)
}

// GetSupportedProviders returns all supported providers
func (f *Factory) GetSupportedProviders() []ai.Provider {
	return []ai.Provider{
		ai.ProviderOpenAI,
		ai.ProviderAnthropic,
		ai.ProviderGoogle,
		ai.ProviderAzure,
		ai.ProviderOllama,
	}
}

// GetProviderCapabilities returns capabilities for a provider
func (f *Factory) GetProviderCapabilities(provider ai.Provider) []ai.Capability {
	switch provider {
	case ai.ProviderOpenAI:
		return []ai.Capability{
			ai.CapabilityChat,
			ai.CapabilityCompletion,
			ai.CapabilityEmbedding,
			ai.CapabilityVision,
			ai.CapabilityImage,
			ai.CapabilitySpeech,
			ai.CapabilityTranscribe,
			ai.CapabilityTools,
		}
	case ai.ProviderAnthropic:
		return []ai.Capability{
			ai.CapabilityChat,
			ai.CapabilityVision,
			ai.CapabilityTools,
		}
	case ai.ProviderGoogle:
		return []ai.Capability{
			ai.CapabilityChat,
			ai.CapabilityVision,
			ai.CapabilityEmbedding,
			ai.CapabilityTools,
		}
	case ai.ProviderAzure:
		return []ai.Capability{
			ai.CapabilityChat,
			ai.CapabilityCompletion,
			ai.CapabilityEmbedding,
			ai.CapabilityVision,
			ai.CapabilityImage,
			ai.CapabilityTools,
		}
	case ai.ProviderOllama:
		return []ai.Capability{
			ai.CapabilityChat,
			ai.CapabilityEmbedding,
		}
	default:
		return []ai.Capability{}
	}
}

// GetDefaultModel returns the default model for a provider
func (f *Factory) GetDefaultModel(provider ai.Provider) string {
	switch provider {
	case ai.ProviderOpenAI:
		return "gpt-4o-mini"
	case ai.ProviderAnthropic:
		return "claude-3-5-haiku-20241022"
	case ai.ProviderGoogle:
		return "gemini-1.5-flash"
	case ai.ProviderAzure:
		return "gpt-4o-mini"
	case ai.ProviderOllama:
		return "llama3.2"
	default:
		return "gpt-4o-mini"
	}
}
