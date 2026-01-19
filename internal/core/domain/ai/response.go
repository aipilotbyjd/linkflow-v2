package ai

import (
	"time"
)

// ChatResponse represents a chat completion response
type ChatResponse struct {
	ID      string    `json:"id"`
	Model   string    `json:"model"`
	Message Message   `json:"message"`
	Usage   Usage     `json:"usage"`
	Created time.Time `json:"created"`

	// Additional info
	Provider    Provider `json:"provider"`
	FinishReason string  `json:"finish_reason,omitempty"` // stop, length, tool_calls, content_filter

	// For streaming
	Streaming bool `json:"streaming,omitempty"`

	// Cost calculation
	CostUSD float64 `json:"cost_usd,omitempty"`

	// Was this response cached?
	Cached bool `json:"cached,omitempty"`

	// Latency
	LatencyMS int64 `json:"latency_ms,omitempty"`
}

// GetText returns the text content from the response message
func (r *ChatResponse) GetText() string {
	return r.Message.GetText()
}

// HasToolCalls returns true if the response contains tool calls
func (r *ChatResponse) HasToolCalls() bool {
	return len(r.Message.ToolCalls) > 0
}

// CompletionResponse represents a text completion response
type CompletionResponse struct {
	ID       string    `json:"id"`
	Model    string    `json:"model"`
	Text     string    `json:"text"`
	Usage    Usage     `json:"usage"`
	Created  time.Time `json:"created"`
	Provider Provider  `json:"provider"`

	FinishReason string  `json:"finish_reason,omitempty"`
	CostUSD      float64 `json:"cost_usd,omitempty"`
	LatencyMS    int64   `json:"latency_ms,omitempty"`
}

// EmbeddingResponse represents an embedding response
type EmbeddingResponse struct {
	ID         string      `json:"id"`
	Model      string      `json:"model"`
	Embeddings [][]float64 `json:"embeddings"`
	Usage      Usage       `json:"usage"`
	Provider   Provider    `json:"provider"`

	Dimensions int     `json:"dimensions"`
	CostUSD    float64 `json:"cost_usd,omitempty"`
	LatencyMS  int64   `json:"latency_ms,omitempty"`
}

// ImageResponse represents an image generation response
type ImageResponse struct {
	ID       string       `json:"id"`
	Model    string       `json:"model"`
	Images   []ImageData  `json:"images"`
	Provider Provider     `json:"provider"`
	Created  time.Time    `json:"created"`

	CostUSD   float64 `json:"cost_usd,omitempty"`
	LatencyMS int64   `json:"latency_ms,omitempty"`
}

// ImageData represents a generated image
type ImageData struct {
	URL           string `json:"url,omitempty"`
	Base64        string `json:"b64_json,omitempty"`
	RevisedPrompt string `json:"revised_prompt,omitempty"`
}

// VisionResponse represents an image analysis response
type VisionResponse struct {
	ID       string    `json:"id"`
	Model    string    `json:"model"`
	Message  Message   `json:"message"`
	Usage    Usage     `json:"usage"`
	Provider Provider  `json:"provider"`
	Created  time.Time `json:"created"`

	CostUSD   float64 `json:"cost_usd,omitempty"`
	LatencyMS int64   `json:"latency_ms,omitempty"`
}

// GetText returns the text content from the vision response
func (r *VisionResponse) GetText() string {
	return r.Message.GetText()
}

// TTSResponse represents a text-to-speech response
type TTSResponse struct {
	Audio     []byte   `json:"audio"`
	AudioURL  string   `json:"audio_url,omitempty"`
	Format    string   `json:"format"`
	Provider  Provider `json:"provider"`

	CostUSD   float64 `json:"cost_usd,omitempty"`
	LatencyMS int64   `json:"latency_ms,omitempty"`
}

// STTResponse represents a speech-to-text response
type STTResponse struct {
	Text      string   `json:"text"`
	Language  string   `json:"language,omitempty"`
	Duration  float64  `json:"duration,omitempty"`
	Provider  Provider `json:"provider"`

	// Detailed segments with timestamps
	Segments []TranscriptSegment `json:"segments,omitempty"`

	CostUSD   float64 `json:"cost_usd,omitempty"`
	LatencyMS int64   `json:"latency_ms,omitempty"`
}

// TranscriptSegment represents a segment of transcribed audio
type TranscriptSegment struct {
	ID    int     `json:"id"`
	Start float64 `json:"start"`
	End   float64 `json:"end"`
	Text  string  `json:"text"`
}

// StructuredOutputResponse represents a structured JSON output response
type StructuredOutputResponse struct {
	ID       string                 `json:"id"`
	Model    string                 `json:"model"`
	Data     map[string]interface{} `json:"data"`
	Usage    Usage                  `json:"usage"`
	Provider Provider               `json:"provider"`

	CostUSD   float64 `json:"cost_usd,omitempty"`
	LatencyMS int64   `json:"latency_ms,omitempty"`
}

// Usage represents token usage
type Usage struct {
	InputTokens   int `json:"input_tokens"`
	OutputTokens  int `json:"output_tokens"`
	TotalTokens   int `json:"total_tokens"`
	CachedTokens  int `json:"cached_tokens,omitempty"`
}

// StreamEvent represents a streaming event
type StreamEvent struct {
	Type    string `json:"type"` // content, tool_call, done, error
	Content string `json:"content,omitempty"`

	// For tool calls
	ToolCall *ToolCall `json:"tool_call,omitempty"`

	// For completion
	Message      *Message `json:"message,omitempty"`
	Usage        *Usage   `json:"usage,omitempty"`
	FinishReason string   `json:"finish_reason,omitempty"`

	// For errors
	Error string `json:"error,omitempty"`
}

// RouterResponse represents a routing decision response
type RouterResponse struct {
	SelectedModel    Model    `json:"selected_model"`
	SelectedProvider Provider `json:"selected_provider"`
	Reason           string   `json:"reason"`

	// Alternative options
	Alternatives []Model `json:"alternatives,omitempty"`
}
