package openai

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/core/domain/ai"
)

const (
	defaultBaseURL = "https://api.openai.com/v1"
	defaultTimeout = 120 * time.Second
)

// Adapter implements the ProviderAdapter interface for OpenAI
type Adapter struct {
	apiKey     string
	orgID      string
	baseURL    string
	httpClient *http.Client
}

// NewAdapter creates a new OpenAI adapter
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
		orgID:   config.OrgID,
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

// Provider returns the provider type
func (a *Adapter) Provider() ai.Provider {
	return ai.ProviderOpenAI
}

// Capabilities returns supported capabilities
func (a *Adapter) Capabilities() []ai.Capability {
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
	// Convert messages to OpenAI format
	messages := make([]map[string]interface{}, len(req.Messages))
	for i, msg := range req.Messages {
		messages[i] = a.convertMessage(msg)
	}

	payload := map[string]interface{}{
		"model":    req.Model,
		"messages": messages,
	}

	if req.MaxTokens > 0 {
		payload["max_tokens"] = req.MaxTokens
	}
	if req.Temperature != nil {
		payload["temperature"] = *req.Temperature
	}
	if req.TopP != nil {
		payload["top_p"] = *req.TopP
	}
	if len(req.Stop) > 0 {
		payload["stop"] = req.Stop
	}
	if req.FrequencyPenalty != nil {
		payload["frequency_penalty"] = *req.FrequencyPenalty
	}
	if req.PresencePenalty != nil {
		payload["presence_penalty"] = *req.PresencePenalty
	}

	// Add tools if provided
	if len(req.Tools) > 0 {
		payload["tools"] = req.Tools
		if req.ToolChoice != "" {
			payload["tool_choice"] = req.ToolChoice
		}
	}

	// Response format
	if req.ResponseFormat != nil {
		payload["response_format"] = req.ResponseFormat
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", a.baseURL+"/chat/completions", bytes.NewReader(body))
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

	var result chatCompletionResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return a.convertChatResponse(&result), nil
}

// Complete sends a text completion request (using chat API)
func (a *Adapter) Complete(ctx context.Context, req *ai.CompletionRequest) (*ai.CompletionResponse, error) {
	// OpenAI deprecated the completions endpoint, use chat
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
		Provider:     ai.ProviderOpenAI,
		FinishReason: chatResp.FinishReason,
		CostUSD:      chatResp.CostUSD,
	}, nil
}

// Embed generates embeddings
func (a *Adapter) Embed(ctx context.Context, req *ai.EmbeddingRequest) (*ai.EmbeddingResponse, error) {
	model := req.Model
	if model == "" {
		model = "text-embedding-3-small"
	}

	payload := map[string]interface{}{
		"model": model,
		"input": req.Input,
	}

	if req.Dimensions > 0 {
		payload["dimensions"] = req.Dimensions
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", a.baseURL+"/embeddings", bytes.NewReader(body))
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

	var result embeddingResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	embeddings := make([][]float64, len(result.Data))
	for i, d := range result.Data {
		embeddings[i] = d.Embedding
	}

	dimensions := 0
	if len(embeddings) > 0 && len(embeddings[0]) > 0 {
		dimensions = len(embeddings[0])
	}

	return &ai.EmbeddingResponse{
		ID:         uuid.New().String(),
		Model:      model,
		Embeddings: embeddings,
		Usage: ai.Usage{
			TotalTokens: result.Usage.TotalTokens,
		},
		Provider:   ai.ProviderOpenAI,
		Dimensions: dimensions,
	}, nil
}

// GenerateImage generates images
func (a *Adapter) GenerateImage(ctx context.Context, req *ai.ImageRequest) (*ai.ImageResponse, error) {
	model := req.Model
	if model == "" {
		model = "dall-e-3"
	}

	payload := map[string]interface{}{
		"model":  model,
		"prompt": req.Prompt,
	}

	if req.N > 0 {
		payload["n"] = req.N
	}
	if req.Size != "" {
		payload["size"] = req.Size
	} else {
		payload["size"] = "1024x1024"
	}
	if req.Quality != "" {
		payload["quality"] = req.Quality
	}
	if req.Style != "" {
		payload["style"] = req.Style
	}
	if req.ResponseFormat != "" {
		payload["response_format"] = req.ResponseFormat
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", a.baseURL+"/images/generations", bytes.NewReader(body))
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

	var result imageResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	images := make([]ai.ImageData, len(result.Data))
	for i, img := range result.Data {
		images[i] = ai.ImageData{
			URL:           img.URL,
			Base64:        img.B64JSON,
			RevisedPrompt: img.RevisedPrompt,
		}
	}

	return &ai.ImageResponse{
		ID:       uuid.New().String(),
		Model:    model,
		Images:   images,
		Provider: ai.ProviderOpenAI,
		Created:  time.Unix(result.Created, 0),
	}, nil
}

// AnalyzeImage analyzes images using vision
func (a *Adapter) AnalyzeImage(ctx context.Context, req *ai.VisionRequest) (*ai.VisionResponse, error) {
	// Vision is handled through regular chat with image content
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
		Provider: ai.ProviderOpenAI,
		Created:  chatResp.Created,
		CostUSD:  chatResp.CostUSD,
	}, nil
}

// TextToSpeech converts text to speech
func (a *Adapter) TextToSpeech(ctx context.Context, req *ai.TTSRequest) (*ai.TTSResponse, error) {
	model := req.Model
	if model == "" {
		model = "tts-1"
	}

	voice := req.Voice
	if voice == "" {
		voice = "alloy"
	}

	payload := map[string]interface{}{
		"model": model,
		"input": req.Input,
		"voice": voice,
	}

	if req.ResponseFormat != "" {
		payload["response_format"] = req.ResponseFormat
	}
	if req.Speed > 0 {
		payload["speed"] = req.Speed
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", a.baseURL+"/audio/speech", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	a.setHeaders(httpReq)

	resp, err := a.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, a.parseError(resp.StatusCode, respBody)
	}

	audio, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read audio: %w", err)
	}

	format := req.ResponseFormat
	if format == "" {
		format = "mp3"
	}

	return &ai.TTSResponse{
		Audio:    audio,
		Format:   format,
		Provider: ai.ProviderOpenAI,
	}, nil
}

// SpeechToText converts speech to text
func (a *Adapter) SpeechToText(ctx context.Context, req *ai.STTRequest) (*ai.STTResponse, error) {
	model := req.Model
	if model == "" {
		model = "whisper-1"
	}

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// Add audio file
	part, err := writer.CreateFormFile("file", "audio.mp3")
	if err != nil {
		return nil, fmt.Errorf("failed to create form file: %w", err)
	}
	if _, err := part.Write(req.Audio); err != nil {
		return nil, fmt.Errorf("failed to write audio: %w", err)
	}

	// Add model
	if err := writer.WriteField("model", model); err != nil {
		return nil, fmt.Errorf("failed to write model field: %w", err)
	}

	// Add optional fields
	if req.Language != "" {
		if err := writer.WriteField("language", req.Language); err != nil {
			return nil, fmt.Errorf("failed to write language field: %w", err)
		}
	}
	if req.Prompt != "" {
		if err := writer.WriteField("prompt", req.Prompt); err != nil {
			return nil, fmt.Errorf("failed to write prompt field: %w", err)
		}
	}
	if req.ResponseFormat != "" {
		if err := writer.WriteField("response_format", req.ResponseFormat); err != nil {
			return nil, fmt.Errorf("failed to write response_format field: %w", err)
		}
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("failed to close writer: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", a.baseURL+"/audio/transcriptions", &buf)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", writer.FormDataContentType())
	httpReq.Header.Set("Authorization", "Bearer "+a.apiKey)
	if a.orgID != "" {
		httpReq.Header.Set("OpenAI-Organization", a.orgID)
	}

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

	var result transcriptionResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		// May be plain text response
		return &ai.STTResponse{
			Text:     string(respBody),
			Provider: ai.ProviderOpenAI,
		}, nil
	}

	return &ai.STTResponse{
		Text:     result.Text,
		Language: result.Language,
		Duration: result.Duration,
		Provider: ai.ProviderOpenAI,
	}, nil
}

// ListModels returns available models
func (a *Adapter) ListModels(ctx context.Context) ([]ai.Model, error) {
	return ai.GetModelsByProvider(ai.ProviderOpenAI), nil
}

func (a *Adapter) setHeaders(req *http.Request) {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+a.apiKey)
	if a.orgID != "" {
		req.Header.Set("OpenAI-Organization", a.orgID)
	}
}

func (a *Adapter) convertMessage(msg ai.Message) map[string]interface{} {
	result := map[string]interface{}{
		"role": msg.Role.String(),
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
				"type": "image_url",
			}
			if c.ImageURL != "" {
				imgContent["image_url"] = map[string]string{
					"url": c.ImageURL,
				}
			} else if c.ImageBase64 != "" {
				mediaType := c.ImageType
				if mediaType == "" {
					mediaType = "jpeg"
				}
				imgContent["image_url"] = map[string]string{
					"url": "data:image/" + mediaType + ";base64," + c.ImageBase64,
				}
			}
			content = append(content, imgContent)
		}
	}

	if len(content) > 0 {
		result["content"] = content
	}

	// Add tool calls
	if len(msg.ToolCalls) > 0 {
		result["tool_calls"] = msg.ToolCalls
	}

	// Add tool call ID for tool responses
	if msg.ToolCallID != "" {
		result["tool_call_id"] = msg.ToolCallID
	}

	return result
}

func (a *Adapter) convertChatResponse(resp *chatCompletionResponse) *ai.ChatResponse {
	if len(resp.Choices) == 0 {
		return &ai.ChatResponse{
			ID:       resp.ID,
			Model:    resp.Model,
			Provider: ai.ProviderOpenAI,
			Created:  time.Unix(resp.Created, 0),
		}
	}

	choice := resp.Choices[0]
	message := ai.Message{
		Role: ai.Role(choice.Message.Role),
		Content: []ai.Content{
			{Type: ai.ContentTypeText, Text: choice.Message.Content},
		},
	}

	// Convert tool calls
	if len(choice.Message.ToolCalls) > 0 {
		message.ToolCalls = make([]ai.ToolCall, len(choice.Message.ToolCalls))
		for i, tc := range choice.Message.ToolCalls {
			message.ToolCalls[i] = ai.ToolCall{
				ID:   tc.ID,
				Type: tc.Type,
			}
			message.ToolCalls[i].Function.Name = tc.Function.Name
			message.ToolCalls[i].Function.Arguments = tc.Function.Arguments
		}
	}

	usage := ai.Usage{
		InputTokens:  resp.Usage.PromptTokens,
		OutputTokens: resp.Usage.CompletionTokens,
		TotalTokens:  resp.Usage.TotalTokens,
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
		Created:      time.Unix(resp.Created, 0),
		Provider:     ai.ProviderOpenAI,
		FinishReason: choice.FinishReason,
		CostUSD:      costUSD,
	}
}

func (a *Adapter) parseError(statusCode int, body []byte) error {
	var errResp errorResponse
	if err := json.Unmarshal(body, &errResp); err != nil {
		return ai.NewProviderError(ai.ProviderOpenAI, statusCode, "unknown", string(body), false)
	}

	retryable := statusCode == 429 || statusCode >= 500
	return ai.NewProviderError(ai.ProviderOpenAI, statusCode, errResp.Error.Code, errResp.Error.Message, retryable)
}

// Helper to encode image to base64
func EncodeImageBase64(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}
