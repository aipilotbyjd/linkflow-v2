package ai

// Model represents an AI model
type Model struct {
	ID           string       `json:"id"`
	Name         string       `json:"name"`
	Provider     Provider     `json:"provider"`
	Capabilities []Capability `json:"capabilities"`

	// Pricing (per 1M tokens)
	InputPricePer1M  float64 `json:"input_price_per_1m"`
	OutputPricePer1M float64 `json:"output_price_per_1m"`

	// Limits
	MaxInputTokens  int `json:"max_input_tokens"`
	MaxOutputTokens int `json:"max_output_tokens"`
	ContextWindow   int `json:"context_window"`

	// Features
	SupportsVision    bool `json:"supports_vision"`
	SupportsTools     bool `json:"supports_tools"`
	SupportsStreaming bool `json:"supports_streaming"`
	SupportsJSON      bool `json:"supports_json"`

	// Metadata
	Description string `json:"description,omitempty"`
	Deprecated  bool   `json:"deprecated,omitempty"`
}

// Common model definitions
var (
	// OpenAI Models
	ModelGPT4o = Model{
		ID:                "gpt-4o",
		Name:              "GPT-4o",
		Provider:          ProviderOpenAI,
		Capabilities:      []Capability{CapabilityChat, CapabilityVision, CapabilityTools},
		InputPricePer1M:   2.50,
		OutputPricePer1M:  10.00,
		ContextWindow:     128000,
		MaxOutputTokens:   16384,
		SupportsVision:    true,
		SupportsTools:     true,
		SupportsStreaming: true,
		SupportsJSON:      true,
	}

	ModelGPT4oMini = Model{
		ID:                "gpt-4o-mini",
		Name:              "GPT-4o Mini",
		Provider:          ProviderOpenAI,
		Capabilities:      []Capability{CapabilityChat, CapabilityVision, CapabilityTools},
		InputPricePer1M:   0.15,
		OutputPricePer1M:  0.60,
		ContextWindow:     128000,
		MaxOutputTokens:   16384,
		SupportsVision:    true,
		SupportsTools:     true,
		SupportsStreaming: true,
		SupportsJSON:      true,
	}

	ModelO1 = Model{
		ID:                "o1",
		Name:              "O1",
		Provider:          ProviderOpenAI,
		Capabilities:      []Capability{CapabilityChat},
		InputPricePer1M:   15.00,
		OutputPricePer1M:  60.00,
		ContextWindow:     200000,
		MaxOutputTokens:   100000,
		SupportsVision:    true,
		SupportsTools:     true,
		SupportsStreaming: true,
		SupportsJSON:      true,
	}

	ModelO1Mini = Model{
		ID:                "o1-mini",
		Name:              "O1 Mini",
		Provider:          ProviderOpenAI,
		Capabilities:      []Capability{CapabilityChat},
		InputPricePer1M:   3.00,
		OutputPricePer1M:  12.00,
		ContextWindow:     128000,
		MaxOutputTokens:   65536,
		SupportsVision:    false,
		SupportsTools:     false,
		SupportsStreaming: true,
		SupportsJSON:      true,
	}

	// Anthropic Models
	ModelClaude35Sonnet = Model{
		ID:                "claude-3-5-sonnet-20241022",
		Name:              "Claude 3.5 Sonnet",
		Provider:          ProviderAnthropic,
		Capabilities:      []Capability{CapabilityChat, CapabilityVision, CapabilityTools},
		InputPricePer1M:   3.00,
		OutputPricePer1M:  15.00,
		ContextWindow:     200000,
		MaxOutputTokens:   8192,
		SupportsVision:    true,
		SupportsTools:     true,
		SupportsStreaming: true,
		SupportsJSON:      true,
	}

	ModelClaude35Haiku = Model{
		ID:                "claude-3-5-haiku-20241022",
		Name:              "Claude 3.5 Haiku",
		Provider:          ProviderAnthropic,
		Capabilities:      []Capability{CapabilityChat, CapabilityVision, CapabilityTools},
		InputPricePer1M:   0.80,
		OutputPricePer1M:  4.00,
		ContextWindow:     200000,
		MaxOutputTokens:   8192,
		SupportsVision:    true,
		SupportsTools:     true,
		SupportsStreaming: true,
		SupportsJSON:      true,
	}

	ModelClaude3Opus = Model{
		ID:                "claude-3-opus-20240229",
		Name:              "Claude 3 Opus",
		Provider:          ProviderAnthropic,
		Capabilities:      []Capability{CapabilityChat, CapabilityVision, CapabilityTools},
		InputPricePer1M:   15.00,
		OutputPricePer1M:  75.00,
		ContextWindow:     200000,
		MaxOutputTokens:   4096,
		SupportsVision:    true,
		SupportsTools:     true,
		SupportsStreaming: true,
		SupportsJSON:      true,
	}

	// Google Models
	ModelGemini15Pro = Model{
		ID:                "gemini-1.5-pro",
		Name:              "Gemini 1.5 Pro",
		Provider:          ProviderGoogle,
		Capabilities:      []Capability{CapabilityChat, CapabilityVision, CapabilityTools},
		InputPricePer1M:   1.25,
		OutputPricePer1M:  5.00,
		ContextWindow:     2097152,
		MaxOutputTokens:   8192,
		SupportsVision:    true,
		SupportsTools:     true,
		SupportsStreaming: true,
		SupportsJSON:      true,
	}

	ModelGemini15Flash = Model{
		ID:                "gemini-1.5-flash",
		Name:              "Gemini 1.5 Flash",
		Provider:          ProviderGoogle,
		Capabilities:      []Capability{CapabilityChat, CapabilityVision, CapabilityTools},
		InputPricePer1M:   0.075,
		OutputPricePer1M:  0.30,
		ContextWindow:     1048576,
		MaxOutputTokens:   8192,
		SupportsVision:    true,
		SupportsTools:     true,
		SupportsStreaming: true,
		SupportsJSON:      true,
	}

	ModelGemini20Flash = Model{
		ID:                "gemini-2.0-flash-exp",
		Name:              "Gemini 2.0 Flash",
		Provider:          ProviderGoogle,
		Capabilities:      []Capability{CapabilityChat, CapabilityVision, CapabilityTools},
		InputPricePer1M:   0.10,
		OutputPricePer1M:  0.40,
		ContextWindow:     1048576,
		MaxOutputTokens:   8192,
		SupportsVision:    true,
		SupportsTools:     true,
		SupportsStreaming: true,
		SupportsJSON:      true,
	}
)

// ModelRegistry holds all available models
var ModelRegistry = map[string]Model{
	// OpenAI
	"gpt-4o":          ModelGPT4o,
	"gpt-4o-mini":     ModelGPT4oMini,
	"o1":              ModelO1,
	"o1-mini":         ModelO1Mini,

	// Anthropic
	"claude-3-5-sonnet-20241022": ModelClaude35Sonnet,
	"claude-3-5-haiku-20241022":  ModelClaude35Haiku,
	"claude-3-opus-20240229":     ModelClaude3Opus,

	// Google
	"gemini-1.5-pro":      ModelGemini15Pro,
	"gemini-1.5-flash":    ModelGemini15Flash,
	"gemini-2.0-flash-exp": ModelGemini20Flash,
}

// GetModel returns a model by ID
func GetModel(id string) (Model, bool) {
	m, ok := ModelRegistry[id]
	return m, ok
}

// GetModelsByProvider returns all models for a provider
func GetModelsByProvider(provider Provider) []Model {
	var models []Model
	for _, m := range ModelRegistry {
		if m.Provider == provider {
			models = append(models, m)
		}
	}
	return models
}

// GetModelsByCapability returns all models with a specific capability
func GetModelsByCapability(cap Capability) []Model {
	var models []Model
	for _, m := range ModelRegistry {
		for _, c := range m.Capabilities {
			if c == cap {
				models = append(models, m)
				break
			}
		}
	}
	return models
}
