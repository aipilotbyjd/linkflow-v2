package google

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
	defaultBaseURL = "https://generativelanguage.googleapis.com/v1beta"
	defaultTimeout = 120 * time.Second
)

// Adapter implements the ProviderAdapter interface for Google AI
type Adapter struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
}

// NewAdapter creates a new Google adapter
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
	return ai.ProviderGoogle
}

// Capabilities returns supported capabilities
func (a *Adapter) Capabilities() []ai.Capability {
	return []ai.Capability{
		ai.CapabilityChat,
		ai.CapabilityVision,
		ai.CapabilityEmbedding,
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
	// Convert messages to Google format
	contents := make([]map[string]interface{}, 0)
	var systemInstruction string

	for _, msg := range req.Messages {
		if msg.Role == ai.RoleSystem {
			systemInstruction = msg.GetText()
			continue
		}
		contents = append(contents, a.convertMessage(msg))
	}

	payload := map[string]interface{}{
		"contents": contents,
	}

	if systemInstruction != "" {
		payload["systemInstruction"] = map[string]interface{}{
			"parts": []map[string]string{
				{"text": systemInstruction},
			},
		}
	}

	// Generation config
	genConfig := map[string]interface{}{}
	if req.MaxTokens > 0 {
		genConfig["maxOutputTokens"] = req.MaxTokens
	}
	if req.Temperature != nil {
		genConfig["temperature"] = *req.Temperature
	}
	if req.TopP != nil {
		genConfig["topP"] = *req.TopP
	}
	if req.TopK != nil {
		genConfig["topK"] = *req.TopK
	}
	if len(req.Stop) > 0 {
		genConfig["stopSequences"] = req.Stop
	}

	if len(genConfig) > 0 {
		payload["generationConfig"] = genConfig
	}

	// Add tools if provided
	if len(req.Tools) > 0 {
		tools := make([]map[string]interface{}, 0)
		functionDeclarations := make([]map[string]interface{}, len(req.Tools))
		for i, t := range req.Tools {
			functionDeclarations[i] = map[string]interface{}{
				"name":        t.Function.Name,
				"description": t.Function.Description,
				"parameters":  t.Function.Parameters,
			}
		}
		tools = append(tools, map[string]interface{}{
			"functionDeclarations": functionDeclarations,
		})
		payload["tools"] = tools
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/models/%s:generateContent?key=%s", a.baseURL, req.Model, a.apiKey)

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

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

	var result generateResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return a.convertChatResponse(&result, req.Model), nil
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
		Provider:     ai.ProviderGoogle,
		FinishReason: chatResp.FinishReason,
		CostUSD:      chatResp.CostUSD,
	}, nil
}

// Embed generates embeddings
func (a *Adapter) Embed(ctx context.Context, req *ai.EmbeddingRequest) (*ai.EmbeddingResponse, error) {
	model := req.Model
	if model == "" {
		model = "text-embedding-004"
	}

	embeddings := make([][]float64, len(req.Input))

	for i, text := range req.Input {
		payload := map[string]interface{}{
			"model": "models/" + model,
			"content": map[string]interface{}{
				"parts": []map[string]string{
					{"text": text},
				},
			},
		}

		body, err := json.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request: %w", err)
		}

		url := fmt.Sprintf("%s/models/%s:embedContent?key=%s", a.baseURL, model, a.apiKey)

		httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		httpReq.Header.Set("Content-Type", "application/json")

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

		var result embeddingResponse
		if err := json.Unmarshal(respBody, &result); err != nil {
			return nil, fmt.Errorf("failed to parse response: %w", err)
		}

		embeddings[i] = result.Embedding.Values
	}

	dimensions := 0
	if len(embeddings) > 0 && len(embeddings[0]) > 0 {
		dimensions = len(embeddings[0])
	}

	return &ai.EmbeddingResponse{
		ID:         uuid.New().String(),
		Model:      model,
		Embeddings: embeddings,
		Provider:   ai.ProviderGoogle,
		Dimensions: dimensions,
	}, nil
}

// GenerateImage generates images (not supported by Gemini API directly)
func (a *Adapter) GenerateImage(ctx context.Context, req *ai.ImageRequest) (*ai.ImageResponse, error) {
	return nil, fmt.Errorf("image generation not supported by Google Gemini API")
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
		Provider: ai.ProviderGoogle,
		Created:  chatResp.Created,
		CostUSD:  chatResp.CostUSD,
	}, nil
}

// TextToSpeech converts text to speech (not supported)
func (a *Adapter) TextToSpeech(ctx context.Context, req *ai.TTSRequest) (*ai.TTSResponse, error) {
	return nil, fmt.Errorf("text-to-speech not supported by Google Gemini API")
}

// SpeechToText converts speech to text (not supported in this adapter)
func (a *Adapter) SpeechToText(ctx context.Context, req *ai.STTRequest) (*ai.STTResponse, error) {
	return nil, fmt.Errorf("speech-to-text not supported by Google Gemini API")
}

// ListModels returns available models
func (a *Adapter) ListModels(ctx context.Context) ([]ai.Model, error) {
	return ai.GetModelsByProvider(ai.ProviderGoogle), nil
}

func (a *Adapter) convertMessage(msg ai.Message) map[string]interface{} {
	role := "user"
	if msg.Role == ai.RoleAssistant {
		role = "model"
	}

	parts := make([]map[string]interface{}, 0)

	for _, c := range msg.Content {
		switch c.Type {
		case ai.ContentTypeText:
			parts = append(parts, map[string]interface{}{
				"text": c.Text,
			})
		case ai.ContentTypeImage:
			if c.ImageBase64 != "" {
				mimeType := "image/jpeg"
				if c.ImageType != "" {
					mimeType = "image/" + c.ImageType
				}
				parts = append(parts, map[string]interface{}{
					"inlineData": map[string]string{
						"mimeType": mimeType,
						"data":     c.ImageBase64,
					},
				})
			}
		}
	}

	// Handle tool responses
	if msg.Role == ai.RoleTool && msg.ToolCallID != "" {
		parts = []map[string]interface{}{
			{
				"functionResponse": map[string]interface{}{
					"name":     msg.ToolCallID,
					"response": map[string]string{"result": msg.GetText()},
				},
			},
		}
	}

	return map[string]interface{}{
		"role":  role,
		"parts": parts,
	}
}

func (a *Adapter) convertChatResponse(resp *generateResponse, model string) *ai.ChatResponse {
	if len(resp.Candidates) == 0 {
		return &ai.ChatResponse{
			ID:       uuid.New().String(),
			Model:    model,
			Provider: ai.ProviderGoogle,
			Created:  time.Now(),
		}
	}

	candidate := resp.Candidates[0]

	var text string
	var toolCalls []ai.ToolCall

	for _, part := range candidate.Content.Parts {
		if part.Text != "" {
			text = part.Text
		}
		if part.FunctionCall != nil {
			argsJSON, _ := json.Marshal(part.FunctionCall.Args)
			toolCalls = append(toolCalls, ai.ToolCall{
				ID:   part.FunctionCall.Name,
				Type: "function",
				Function: struct {
					Name      string `json:"name"`
					Arguments string `json:"arguments"`
				}{
					Name:      part.FunctionCall.Name,
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

	usage := ai.Usage{}
	if resp.UsageMetadata != nil {
		usage.InputTokens = resp.UsageMetadata.PromptTokenCount
		usage.OutputTokens = resp.UsageMetadata.CandidatesTokenCount
		usage.TotalTokens = resp.UsageMetadata.TotalTokenCount
	}

	// Calculate cost
	modelInfo, ok := ai.GetModel(model)
	costUSD := 0.0
	if ok {
		costUSD = float64(usage.InputTokens)*modelInfo.InputPricePer1M/1_000_000 +
			float64(usage.OutputTokens)*modelInfo.OutputPricePer1M/1_000_000
	}

	finishReason := ""
	if candidate.FinishReason != "" {
		finishReason = candidate.FinishReason
	}

	return &ai.ChatResponse{
		ID:           uuid.New().String(),
		Model:        model,
		Message:      message,
		Usage:        usage,
		Created:      time.Now(),
		Provider:     ai.ProviderGoogle,
		FinishReason: finishReason,
		CostUSD:      costUSD,
	}
}

func (a *Adapter) parseError(statusCode int, body []byte) error {
	var errResp struct {
		Error struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
			Status  string `json:"status"`
		} `json:"error"`
	}

	if err := json.Unmarshal(body, &errResp); err != nil {
		return ai.NewProviderError(ai.ProviderGoogle, statusCode, "unknown", string(body), false)
	}

	retryable := statusCode == 429 || statusCode >= 500
	return ai.NewProviderError(ai.ProviderGoogle, statusCode, errResp.Error.Status, errResp.Error.Message, retryable)
}
