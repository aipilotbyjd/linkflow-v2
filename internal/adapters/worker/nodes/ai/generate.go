package ai

import (
	"context"
	"fmt"

	"github.com/linkflow-ai/linkflow/internal/adapters/worker/executor"
	wtypes "github.com/linkflow-ai/linkflow/internal/adapters/worker/types"
	"github.com/linkflow-ai/linkflow/internal/core/domain/ai"
	"github.com/linkflow-ai/linkflow/internal/infrastructure/ai/providers"
	"github.com/linkflow-ai/linkflow/internal/shared/types"
)

// GenerateNode handles single-prompt text generation
type GenerateNode struct {
	factory *providers.Factory
}

// NewGenerateNode creates a new AI generate node
func NewGenerateNode() *GenerateNode {
	return &GenerateNode{
		factory: providers.NewFactory(),
	}
}

func (n *GenerateNode) Execute(ctx context.Context, runtime *executor.Runtime, node map[string]interface{}) (types.JSON, error) {
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

	// Get model
	model, _ := params["model"].(string)
	if model == "" {
		model = n.factory.GetDefaultModel(provider)
	}

	// Get prompt
	prompt, _ := params["prompt"].(string)
	if prompt == "" {
		return nil, fmt.Errorf("prompt is required")
	}

	// Optional system message
	systemMessage, _ := params["system_message"].(string)

	// Build messages
	var messages []ai.Message
	if systemMessage != "" {
		messages = append(messages, ai.NewSystemMessage(systemMessage))
	}
	messages = append(messages, ai.NewUserMessage(prompt))

	// Build request
	req := &ai.ChatRequest{
		Messages: messages,
		Model:    model,
	}

	// Optional parameters
	if maxTokens, ok := params["max_tokens"].(float64); ok {
		req.MaxTokens = int(maxTokens)
	}
	if temp, ok := params["temperature"].(float64); ok {
		req.Temperature = &temp
	}
	if topP, ok := params["top_p"].(float64); ok {
		req.TopP = &topP
	}

	// Stop sequences
	if stop, ok := params["stop"].([]interface{}); ok {
		stopSeqs := make([]string, len(stop))
		for i, s := range stop {
			stopSeqs[i], _ = s.(string)
		}
		req.Stop = stopSeqs
	}

	// JSON mode
	if jsonMode, ok := params["json_mode"].(bool); ok && jsonMode {
		req.ResponseFormat = &ai.ResponseFormat{Type: "json_object"}
	}

	// Create provider adapter
	config := &ai.ProviderConfig{
		Provider: provider,
		APIKey:   apiKey,
	}

	if orgID, ok := params["org_id"].(string); ok {
		config.OrgID = orgID
	}
	if baseURL, ok := params["base_url"].(string); ok {
		config.BaseURL = baseURL
	}

	adapter, err := n.factory.CreateAdapter(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create provider adapter: %w", err)
	}

	// Execute request
	resp, err := adapter.Chat(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("generation request failed: %w", err)
	}

	return types.JSON{
		"text":          resp.GetText(),
		"model":         resp.Model,
		"provider":      resp.Provider.String(),
		"finish_reason": resp.FinishReason,
		"usage": map[string]interface{}{
			"input_tokens":  resp.Usage.InputTokens,
			"output_tokens": resp.Usage.OutputTokens,
			"total_tokens":  resp.Usage.TotalTokens,
		},
		"cost_usd":   resp.CostUSD,
		"latency_ms": resp.LatencyMS,
	}, nil
}

func (n *GenerateNode) Metadata() wtypes.NodeMetadata {
	return wtypes.NodeMetadata{
		Type:        "ai.generate",
		Name:        "AI Generate",
		Description: "Generate text from a single prompt",
		Category:    "ai",
		Version:     "1.0.0",
		Icon:        "Sparkles",
		Color:       "#8B5CF6",
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
					{Name: "Ollama (Local)", Value: "ollama"},
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
				Type:        "string",
				Required:    false,
				Description: "Model to use (defaults to provider's default)",
			},
			{
				Name:        "prompt",
				DisplayName: "Prompt",
				Type:        "string",
				Required:    true,
				Description: "The prompt for text generation",
			},
			{
				Name:        "system_message",
				DisplayName: "System Message",
				Type:        "string",
				Required:    false,
				Description: "System instructions for the AI",
			},
			{
				Name:        "max_tokens",
				DisplayName: "Max Tokens",
				Type:        "number",
				Required:    false,
				Default:     1024,
			},
			{
				Name:        "temperature",
				DisplayName: "Temperature",
				Type:        "number",
				Required:    false,
				Default:     0.7,
			},
			{
				Name:        "stop",
				DisplayName: "Stop Sequences",
				Type:        "json",
				Required:    false,
				Description: "Sequences that stop generation",
			},
			{
				Name:        "json_mode",
				DisplayName: "JSON Mode",
				Type:        "boolean",
				Required:    false,
				Default:     false,
			},
		},
		Credentials: []string{"openai", "anthropic", "google_ai"},
	}
}
