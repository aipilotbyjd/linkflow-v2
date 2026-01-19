package anthropic

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/core/domain/ai"
)

const (
	defaultBaseURL   = "https://api.anthropic.com/v1"
	defaultTimeout   = 120 * time.Second
	anthropicVersion = "2023-06-01"
)

// Adapter implements the ProviderAdapter interface for Anthropic
type Adapter struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
}

// NewAdapter creates a new Anthropic adapter
func NewAdapter(config *ai.ProviderConfig) *Adapter {
	baseURL := defaultBaseURL
	if config.BaseURL != "" {
		baseURL = config.BaseURL
	}

	timeout := defaultTimeout
	if config.Timeout > 0 {
		timeout = time.Duration(config.Timeout) * time.Second
	}

	return &Adapter{
		apiKey:  config.APIKey,
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

// Provider returns the provider type
func (a *Adapter) Provider() ai.Provider {
	return ai.ProviderAnthropic
}

// Capabilities returns supported capabilities
func (a *Adapter) Capabilities() []ai.Capability {
	return []ai.Capability{
		ai.CapabilityChat,
		ai.CapabilityVision,
		ai.CapabilityTools,
	}
}

// SupportsCapability checks if a capability is supported
func (a *Adapter) SupportsCapability(cap ai.Capability) bool {
	for _, c := range a.Capabilities() {
		if c == cap {
			return true
		}
	}
	return false
}

// Chat sends a chat completion request
func (a *Adapter) Chat(ctx context.Context, req *ai.ChatRequest) (*ai.ChatResponse, error) {
	// Extract system message and convert other messages
	var system string
	messages := make([]map[string]interface{}, 0)

	for _, msg := range req.Messages {
		if msg.Role == ai.RoleSystem {
			system = msg.GetText()
			continue
		}
		messages = append(messages, a.convertMessage(msg))
	}

	payload := map[string]interface{}{
		"model":    req.Model,
		"messages": messages,
	}

	if system != "" {
		payload["system"] = system
	}

	maxTokens := req.MaxTokens
	if maxTokens == 0 {
		maxTokens = 4096
	}
	payload["max_tokens"] = maxTokens

	if req.Temperature != nil {
		payload["temperature"] = *req.Temperature
	}
	if req.TopP != nil {
		payload["top_p"] = *req.TopP
	}
	if req.TopK != nil {
		payload["top_k"] = *req.TopK
	}
	if len(req.Stop) > 0 {
		payload["stop_sequences"] = req.Stop
	}

	// Add tools if provided
	if len(req.Tools) > 0 {
		tools := make([]map[string]interface{}, len(req.Tools))
		for i, t := range req.Tools {
			tools[i] = map[string]interface{}{
				"name":         t.Function.Name,
				"description":  t.Function.Description,
				"input_schema": t.Function.Parameters,
			}
		}
		payload["tools"] = tools
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", a.baseURL+"/messages", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	a.setHeaders(httpReq)

	resp, err := a.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, a.parseError(resp.StatusCode, respBody)
	}

	var result messageResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return a.convertChatResponse(&result), nil
}

// Complete sends a text completion request
func (a *Adapter) Complete(ctx context.Context, req *ai.CompletionRequest) (*ai.CompletionResponse, error) {
	chatReq := &ai.ChatRequest{
		Messages: []ai.Message{
			ai.NewUserMessage(req.Prompt),
		},
		Model:       req.Model,
		MaxTokens:   req.MaxTokens,
		Temperature: req.Temperature,
		TopP:        req.TopP,
		Stop:        req.Stop,
	}

	chatResp, err := a.Chat(ctx, chatReq)
	if err != nil {
		return nil, err
	}

	return &ai.CompletionResponse{
		ID:           chatResp.ID,
		Model:        chatResp.Model,
		Text:         chatResp.GetText(),
		Usage:        chatResp.Usage,
		Created:      chatResp.Created,
		Provider:     ai.ProviderAnthropic,
		FinishReason: chatResp.FinishReason,
		CostUSD:      chatResp.CostUSD,
	}, nil
}

// Embed generates embeddings (not supported by Anthropic)
func (a *Adapter) Embed(ctx context.Context, req *ai.EmbeddingRequest) (*ai.EmbeddingResponse, error) {
	return nil, fmt.Errorf("embeddings not supported by Anthropic")
}

// GenerateImage generates images (not supported by Anthropic)
func (a *Adapter) GenerateImage(ctx context.Context, req *ai.ImageRequest) (*ai.ImageResponse, error) {
	return nil, fmt.Errorf("image generation not supported by Anthropic")
}

// AnalyzeImage analyzes images using vision
func (a *Adapter) AnalyzeImage(ctx context.Context, req *ai.VisionRequest) (*ai.VisionResponse, error) {
	chatReq := &ai.ChatRequest{
		Messages:    req.Messages,
		Model:       req.Model,
		MaxTokens:   req.MaxTokens,
		Temperature: req.Temperature,
	}

	chatResp, err := a.Chat(ctx, chatReq)
	if err != nil {
		return nil, err
	}

	return &ai.VisionResponse{
		ID:       chatResp.ID,
		Model:    chatResp.Model,
		Message:  chatResp.Message,
		Usage:    chatResp.Usage,
		Provider: ai.ProviderAnthropic,
		Created:  chatResp.Created,
		CostUSD:  chatResp.CostUSD,
	}, nil
}

// TextToSpeech converts text to speech (not supported by Anthropic)
func (a *Adapter) TextToSpeech(ctx context.Context, req *ai.TTSRequest) (*ai.TTSResponse, error) {
	return nil, fmt.Errorf("text-to-speech not supported by Anthropic")
}

// SpeechToText converts speech to text (not supported by Anthropic)
func (a *Adapter) SpeechToText(ctx context.Context, req *ai.STTRequest) (*ai.STTResponse, error) {
	return nil, fmt.Errorf("speech-to-text not supported by Anthropic")
}

// ListModels returns available models
func (a *Adapter) ListModels(ctx context.Context) ([]ai.Model, error) {
	return ai.GetModelsByProvider(ai.ProviderAnthropic), nil
}

func (a *Adapter) setHeaders(req *http.Request) {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", a.apiKey)
	req.Header.Set("anthropic-version", anthropicVersion)
}

func (a *Adapter) convertMessage(msg ai.Message) map[string]interface{} {
	role := msg.Role.String()
	if role == "tool" {
		role = "user" // Anthropic uses user role for tool results
	}

	result := map[string]interface{}{
		"role": role,
	}

	// Handle tool results
	if msg.Role == ai.RoleTool && msg.ToolCallID != "" {
		result["content"] = []map[string]interface{}{
			{
				"type":        "tool_result",
				"tool_use_id": msg.ToolCallID,
				"content":     msg.GetText(),
			},
		}
		return result
	}

	// Handle simple text content
	if len(msg.Content) == 1 && msg.Content[0].Type == ai.ContentTypeText {
		result["content"] = msg.Content[0].Text
		return result
	}

	// Handle multi-modal content
	content := make([]map[string]interface{}, 0, len(msg.Content))
	for _, c := range msg.Content {
		switch c.Type {
		case ai.ContentTypeText:
			content = append(content, map[string]interface{}{
				"type": "text",
				"text": c.Text,
			})
		case ai.ContentTypeImage:
			imgContent := map[string]interface{}{
				"type": "image",
			}
			if c.ImageBase64 != "" {
				mediaType := c.ImageType
				if mediaType == "" {
					mediaType = "image/jpeg"
				} else {
					mediaType = "image/" + mediaType
				}
				imgContent["source"] = map[string]string{
					"type":         "base64",
					"media_type":   mediaType,
					"data":         c.ImageBase64,
				}
			} else if c.ImageURL != "" {
				imgContent["source"] = map[string]string{
					"type": "url",
					"url":  c.ImageURL,
				}
			}
			content = append(content, imgContent)
		}
	}

	if len(content) > 0 {
		result["content"] = content
	}

	return result
}

func (a *Adapter) convertChatResponse(resp *messageResponse) *ai.ChatResponse {
	// Extract text and tool calls from content
	var text string
	var toolCalls []ai.ToolCall

	for _, block := range resp.Content {
		switch block.Type {
		case "text":
			text = block.Text
		case "tool_use":
			argsJSON, _ := json.Marshal(block.Input)
			toolCalls = append(toolCalls, ai.ToolCall{
				ID:   block.ID,
				Type: "function",
				Function: struct {
					Name      string `json:"name"`
					Arguments string `json:"arguments"`
				}{
					Name:      block.Name,
					Arguments: string(argsJSON),
				},
			})
		}
	}

	message := ai.Message{
		Role: ai.RoleAssistant,
		Content: []ai.Content{
			{Type: ai.ContentTypeText, Text: text},
		},
		ToolCalls: toolCalls,
	}

	usage := ai.Usage{
		InputTokens:  resp.Usage.InputTokens,
		OutputTokens: resp.Usage.OutputTokens,
		TotalTokens:  resp.Usage.InputTokens + resp.Usage.OutputTokens,
	}

	// Calculate cost
	model, ok := ai.GetModel(resp.Model)
	costUSD := 0.0
	if ok {
		costUSD = float64(usage.InputTokens)*model.InputPricePer1M/1_000_000 +
			float64(usage.OutputTokens)*model.OutputPricePer1M/1_000_000
	}

	return &ai.ChatResponse{
		ID:           resp.ID,
		Model:        resp.Model,
		Message:      message,
		Usage:        usage,
		Created:      time.Now(),
		Provider:     ai.ProviderAnthropic,
		FinishReason: resp.StopReason,
		CostUSD:      costUSD,
	}
}

func (a *Adapter) parseError(statusCode int, body []byte) error {
	var errResp struct {
		Error struct {
			Type    string `json:"type"`
			Message string `json:"message"`
		} `json:"error"`
	}

	if err := json.Unmarshal(body, &errResp); err != nil {
		return ai.NewProviderError(ai.ProviderAnthropic, statusCode, "unknown", string(body), false)
	}

	retryable := statusCode == 429 || statusCode >= 500
	return ai.NewProviderError(ai.ProviderAnthropic, statusCode, errResp.Error.Type, errResp.Error.Message, retryable)
}

// Helper to generate unique ID
func generateID() string {
	return uuid.New().String()
}
