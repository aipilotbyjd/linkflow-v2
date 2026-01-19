package ai

// ChatRequest represents a chat completion request
type ChatRequest struct {
	// Required
	Messages []Message `json:"messages"`
	Model    string    `json:"model"`

	// Optional - Generation params
	MaxTokens        int      `json:"max_tokens,omitempty"`
	Temperature      *float64 `json:"temperature,omitempty"`
	TopP             *float64 `json:"top_p,omitempty"`
	TopK             *int     `json:"top_k,omitempty"`
	FrequencyPenalty *float64 `json:"frequency_penalty,omitempty"`
	PresencePenalty  *float64 `json:"presence_penalty,omitempty"`
	Stop             []string `json:"stop,omitempty"`

	// Tools/Functions
	Tools      []Tool `json:"tools,omitempty"`
	ToolChoice string `json:"tool_choice,omitempty"` // auto, none, required, or specific tool name

	// Output format
	ResponseFormat *ResponseFormat `json:"response_format,omitempty"`

	// Streaming
	Stream bool `json:"stream,omitempty"`

	// Provider-specific
	ProviderConfig *ProviderConfig `json:"-"`

	// Metadata
	User       string            `json:"user,omitempty"`
	Metadata   map[string]string `json:"metadata,omitempty"`
	WorkspaceID string           `json:"-"`
	ExecutionID string           `json:"-"`
}

// ResponseFormat specifies the output format
type ResponseFormat struct {
	Type       string                 `json:"type"` // text, json_object, json_schema
	JSONSchema map[string]interface{} `json:"json_schema,omitempty"`
}

// CompletionRequest represents a text completion request
type CompletionRequest struct {
	Prompt string `json:"prompt"`
	Model  string `json:"model"`

	MaxTokens   int      `json:"max_tokens,omitempty"`
	Temperature *float64 `json:"temperature,omitempty"`
	TopP        *float64 `json:"top_p,omitempty"`
	Stop        []string `json:"stop,omitempty"`

	Stream bool `json:"stream,omitempty"`

	ProviderConfig *ProviderConfig `json:"-"`
	WorkspaceID    string          `json:"-"`
}

// EmbeddingRequest represents an embedding request
type EmbeddingRequest struct {
	Input          []string `json:"input"`
	Model          string   `json:"model"`
	Dimensions     int      `json:"dimensions,omitempty"`
	EncodingFormat string   `json:"encoding_format,omitempty"` // float, base64

	ProviderConfig *ProviderConfig `json:"-"`
	WorkspaceID    string          `json:"-"`
}

// ImageRequest represents an image generation request
type ImageRequest struct {
	Prompt         string `json:"prompt"`
	Model          string `json:"model,omitempty"`
	N              int    `json:"n,omitempty"`
	Size           string `json:"size,omitempty"`   // 256x256, 512x512, 1024x1024, etc.
	Quality        string `json:"quality,omitempty"` // standard, hd
	Style          string `json:"style,omitempty"`   // vivid, natural
	ResponseFormat string `json:"response_format,omitempty"` // url, b64_json

	ProviderConfig *ProviderConfig `json:"-"`
	WorkspaceID    string          `json:"-"`
}

// VisionRequest represents an image analysis request
type VisionRequest struct {
	Messages []Message `json:"messages"`
	Model    string    `json:"model"`

	MaxTokens   int      `json:"max_tokens,omitempty"`
	Temperature *float64 `json:"temperature,omitempty"`

	ProviderConfig *ProviderConfig `json:"-"`
	WorkspaceID    string          `json:"-"`
}

// TTSRequest represents a text-to-speech request
type TTSRequest struct {
	Input          string `json:"input"`
	Model          string `json:"model"`
	Voice          string `json:"voice"`
	ResponseFormat string `json:"response_format,omitempty"` // mp3, opus, aac, flac
	Speed          float64 `json:"speed,omitempty"`

	ProviderConfig *ProviderConfig `json:"-"`
	WorkspaceID    string          `json:"-"`
}

// STTRequest represents a speech-to-text request
type STTRequest struct {
	Audio          []byte  `json:"audio"`
	AudioURL       string  `json:"audio_url,omitempty"`
	Model          string  `json:"model"`
	Language       string  `json:"language,omitempty"`
	Prompt         string  `json:"prompt,omitempty"`
	ResponseFormat string  `json:"response_format,omitempty"` // json, text, srt, vtt
	Temperature    float64 `json:"temperature,omitempty"`
	Timestamps     bool    `json:"timestamps,omitempty"`

	ProviderConfig *ProviderConfig `json:"-"`
	WorkspaceID    string          `json:"-"`
}

// StructuredOutputRequest represents a request for structured JSON output
type StructuredOutputRequest struct {
	Messages   []Message              `json:"messages"`
	Model      string                 `json:"model"`
	Schema     map[string]interface{} `json:"schema"`
	SchemaName string                 `json:"schema_name,omitempty"`

	MaxTokens   int      `json:"max_tokens,omitempty"`
	Temperature *float64 `json:"temperature,omitempty"`

	ProviderConfig *ProviderConfig `json:"-"`
	WorkspaceID    string          `json:"-"`
}

// RouterRequest represents a request to route to the best model
type RouterRequest struct {
	Messages []Message `json:"messages"`

	// Routing preferences
	PreferredProviders []Provider `json:"preferred_providers,omitempty"`
	MaxCostPer1M       float64    `json:"max_cost_per_1m,omitempty"`
	RequireVision      bool       `json:"require_vision,omitempty"`
	RequireTools       bool       `json:"require_tools,omitempty"`
	PreferSpeed        bool       `json:"prefer_speed,omitempty"`
	PreferQuality      bool       `json:"prefer_quality,omitempty"`

	ProviderConfig *ProviderConfig `json:"-"`
	WorkspaceID    string          `json:"-"`
}
