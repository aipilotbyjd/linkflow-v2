package ai

import (
	"context"
)

// Provider represents an AI provider (OpenAI, Anthropic, Google, etc.)
type Provider string

const (
	ProviderOpenAI    Provider = "openai"
	ProviderAnthropic Provider = "anthropic"
	ProviderGoogle    Provider = "google"
	ProviderAzure     Provider = "azure_openai"
	ProviderOllama    Provider = "ollama"
	ProviderCustom    Provider = "custom"
)

func (p Provider) String() string {
	return string(p)
}

func (p Provider) IsValid() bool {
	switch p {
	case ProviderOpenAI, ProviderAnthropic, ProviderGoogle, ProviderAzure, ProviderOllama, ProviderCustom:
		return true
	default:
		return false
	}
}

func ParseProvider(s string) (Provider, bool) {
	p := Provider(s)
	return p, p.IsValid()
}

// Capability represents AI capabilities
type Capability string

const (
	CapabilityChat       Capability = "chat"
	CapabilityCompletion Capability = "completion"
	CapabilityEmbedding  Capability = "embedding"
	CapabilityVision     Capability = "vision"
	CapabilityImage      Capability = "image_generation"
	CapabilitySpeech     Capability = "speech"
	CapabilityTranscribe Capability = "transcribe"
	CapabilityTools      Capability = "tools"
)

// ProviderAdapter is the interface that all AI provider adapters must implement
type ProviderAdapter interface {
	// Provider returns the provider type
	Provider() Provider

	// Capabilities returns the capabilities supported by this provider
	Capabilities() []Capability

	// Chat sends a chat completion request
	Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error)

	// Complete sends a text completion request
	Complete(ctx context.Context, req *CompletionRequest) (*CompletionResponse, error)

	// Embed generates embeddings for text
	Embed(ctx context.Context, req *EmbeddingRequest) (*EmbeddingResponse, error)

	// GenerateImage generates images from text
	GenerateImage(ctx context.Context, req *ImageRequest) (*ImageResponse, error)

	// AnalyzeImage analyzes an image (vision)
	AnalyzeImage(ctx context.Context, req *VisionRequest) (*VisionResponse, error)

	// TextToSpeech converts text to speech
	TextToSpeech(ctx context.Context, req *TTSRequest) (*TTSResponse, error)

	// SpeechToText converts speech to text
	SpeechToText(ctx context.Context, req *STTRequest) (*STTResponse, error)

	// SupportsCapability checks if the provider supports a capability
	SupportsCapability(cap Capability) bool

	// ListModels returns available models
	ListModels(ctx context.Context) ([]Model, error)
}

// ProviderConfig holds configuration for a provider
type ProviderConfig struct {
	Provider    Provider          `json:"provider"`
	APIKey      string            `json:"api_key"`
	BaseURL     string            `json:"base_url,omitempty"`
	OrgID       string            `json:"org_id,omitempty"`
	ProjectID   string            `json:"project_id,omitempty"`
	Region      string            `json:"region,omitempty"`
	Timeout     int               `json:"timeout,omitempty"`
	MaxRetries  int               `json:"max_retries,omitempty"`
	ExtraConfig map[string]string `json:"extra_config,omitempty"`
}
