package ai

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/linkflow-ai/linkflow/internal/adapters/worker/executor"
	wtypes "github.com/linkflow-ai/linkflow/internal/adapters/worker/types"
	"github.com/linkflow-ai/linkflow/internal/core/domain/ai"
	"github.com/linkflow-ai/linkflow/internal/infrastructure/ai/providers"
	"github.com/linkflow-ai/linkflow/internal/shared/types"
)

// VisionNode analyzes images using AI vision models
type VisionNode struct {
	factory    *providers.Factory
	httpClient *http.Client
}

// NewVisionNode creates a new vision node
func NewVisionNode() *VisionNode {
	return &VisionNode{
		factory:    providers.NewFactory(),
		httpClient: &http.Client{Timeout: 60 * time.Second},
	}
}

func (n *VisionNode) Execute(ctx context.Context, runtime *executor.Runtime, node map[string]interface{}) (types.JSON, error) {
	params, _ := node["parameters"].(map[string]interface{})

	// Get credentials
	apiKey, _ := params["api_key"].(string)
	if apiKey == "" {
		if credRef, ok := params["credential"].(map[string]interface{}); ok {
			if key, ok := credRef["api_key"].(string); ok {
				apiKey = key
			}
		}
	}

	if apiKey == "" {
		return nil, fmt.Errorf("API key is required")
	}

	// Get provider
	providerStr, _ := params["provider"].(string)
	provider := ai.ProviderOpenAI
	if providerStr != "" {
		if p, ok := ai.ParseProvider(providerStr); ok {
			provider = p
		}
	}

	// Get model (must support vision)
	model, _ := params["model"].(string)
	if model == "" {
		switch provider {
		case ai.ProviderOpenAI:
			model = "gpt-4o"
		case ai.ProviderAnthropic:
			model = "claude-3-5-sonnet-20241022"
		case ai.ProviderGoogle:
			model = "gemini-1.5-flash"
		default:
			model = "gpt-4o"
		}
	}

	// Get image input
	imageURL, _ := params["image_url"].(string)
	imageBase64, _ := params["image_base64"].(string)

	// If URL provided, optionally fetch and convert to base64
	if imageURL != "" && imageBase64 == "" {
		fetchImage, _ := params["fetch_image"].(bool)
		if fetchImage {
			data, mediaType, err := n.fetchImage(ctx, imageURL)
			if err != nil {
				return nil, fmt.Errorf("failed to fetch image: %w", err)
			}
			imageBase64 = base64.StdEncoding.EncodeToString(data)
			params["image_type"] = strings.TrimPrefix(mediaType, "image/")
		}
	}

	if imageURL == "" && imageBase64 == "" {
		return nil, fmt.Errorf("either image_url or image_base64 is required")
	}

	// Get prompt
	prompt, _ := params["prompt"].(string)
	if prompt == "" {
		prompt = "Describe this image in detail."
	}

	// Build message content
	var content []ai.Content
	content = append(content, ai.Content{Type: ai.ContentTypeText, Text: prompt})

	if imageBase64 != "" {
		imageType, _ := params["image_type"].(string)
		if imageType == "" {
			imageType = "jpeg"
		}
		content = append(content, ai.Content{
			Type:        ai.ContentTypeImage,
			ImageBase64: imageBase64,
			ImageType:   imageType,
		})
	} else if imageURL != "" {
		content = append(content, ai.Content{
			Type:     ai.ContentTypeImage,
			ImageURL: imageURL,
		})
	}

	// Build messages
	var messages []ai.Message

	// Add system message if provided
	if systemMsg, ok := params["system_message"].(string); ok && systemMsg != "" {
		messages = append(messages, ai.NewSystemMessage(systemMsg))
	}

	messages = append(messages, ai.Message{
		Role:    ai.RoleUser,
		Content: content,
	})

	// Build request
	req := &ai.VisionRequest{
		Messages: messages,
		Model:    model,
	}

	if maxTokens, ok := params["max_tokens"].(float64); ok {
		req.MaxTokens = int(maxTokens)
	}
	if temp, ok := params["temperature"].(float64); ok {
		req.Temperature = &temp
	}

	// Create provider adapter
	config := &ai.ProviderConfig{
		Provider: provider,
		APIKey:   apiKey,
	}

	if orgID, ok := params["org_id"].(string); ok {
		config.OrgID = orgID
	}

	adapter, err := n.factory.CreateAdapter(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create provider adapter: %w", err)
	}

	// Execute request
	resp, err := adapter.AnalyzeImage(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("vision request failed: %w", err)
	}

	return types.JSON{
		"analysis": resp.GetText(),
		"model":    resp.Model,
		"provider": resp.Provider.String(),
		"usage": map[string]interface{}{
			"input_tokens":  resp.Usage.InputTokens,
			"output_tokens": resp.Usage.OutputTokens,
			"total_tokens":  resp.Usage.TotalTokens,
		},
		"cost_usd":   resp.CostUSD,
		"latency_ms": resp.LatencyMS,
	}, nil
}

func (n *VisionNode) fetchImage(ctx context.Context, url string) ([]byte, string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, "", err
	}

	resp, err := n.httpClient.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("failed to fetch image: status %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", err
	}

	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "image/jpeg"
	}

	return data, contentType, nil
}

func (n *VisionNode) Metadata() wtypes.NodeMetadata {
	return wtypes.NodeMetadata{
		Type:        "ai.vision",
		Name:        "AI Vision",
		Description: "Analyze images using AI vision models",
		Category:    "ai",
		Version:     "1.0.0",
		Icon:        "Eye",
		Color:       "#EC4899",
		Inputs:      []wtypes.NodePort{{Name: "main", Type: "any"}},
		Outputs:     []wtypes.NodePort{{Name: "main", Type: "object"}},
		Parameters: []wtypes.NodeParameter{
			{
				Name:        "provider",
				DisplayName: "Provider",
				Type:        "options",
				Required:    true,
				Default:     "openai",
				Options: []wtypes.ParamOption{
					{Name: "OpenAI", Value: "openai"},
					{Name: "Anthropic", Value: "anthropic"},
					{Name: "Google", Value: "google"},
				},
			},
			{
				Name:        "api_key",
				DisplayName: "API Key",
				Type:        "credential",
				Required:    true,
			},
			{
				Name:        "model",
				DisplayName: "Model",
				Type:        "options",
				Required:    false,
				Options: []wtypes.ParamOption{
					{Name: "GPT-4o (OpenAI)", Value: "gpt-4o"},
					{Name: "GPT-4o Mini (OpenAI)", Value: "gpt-4o-mini"},
					{Name: "Claude 3.5 Sonnet (Anthropic)", Value: "claude-3-5-sonnet-20241022"},
					{Name: "Claude 3.5 Haiku (Anthropic)", Value: "claude-3-5-haiku-20241022"},
					{Name: "Gemini 1.5 Pro (Google)", Value: "gemini-1.5-pro"},
					{Name: "Gemini 1.5 Flash (Google)", Value: "gemini-1.5-flash"},
				},
			},
			{
				Name:        "image_url",
				DisplayName: "Image URL",
				Type:        "string",
				Required:    false,
				Description: "URL of the image to analyze",
			},
			{
				Name:        "image_base64",
				DisplayName: "Image Base64",
				Type:        "string",
				Required:    false,
				Description: "Base64-encoded image data",
			},
			{
				Name:        "image_type",
				DisplayName: "Image Type",
				Type:        "options",
				Required:    false,
				Default:     "jpeg",
				Options: []wtypes.ParamOption{
					{Name: "JPEG", Value: "jpeg"},
					{Name: "PNG", Value: "png"},
					{Name: "GIF", Value: "gif"},
					{Name: "WebP", Value: "webp"},
				},
			},
			{
				Name:        "fetch_image",
				DisplayName: "Fetch Image",
				Type:        "boolean",
				Required:    false,
				Default:     false,
				Description: "Fetch image from URL and convert to base64",
			},
			{
				Name:        "prompt",
				DisplayName: "Prompt",
				Type:        "string",
				Required:    false,
				Default:     "Describe this image in detail.",
				Description: "What to analyze about the image",
			},
			{
				Name:        "system_message",
				DisplayName: "System Message",
				Type:        "string",
				Required:    false,
			},
			{
				Name:        "max_tokens",
				DisplayName: "Max Tokens",
				Type:        "number",
				Required:    false,
				Default:     1024,
			},
		},
		Credentials: []string{"openai", "anthropic", "google_ai"},
	}
}
