package ai

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/linkflow-ai/linkflow/internal/adapters/worker/executor"
	wtypes "github.com/linkflow-ai/linkflow/internal/adapters/worker/types"
	"github.com/linkflow-ai/linkflow/internal/core/domain/ai"
	"github.com/linkflow-ai/linkflow/internal/infrastructure/ai/gateway"
	"github.com/linkflow-ai/linkflow/internal/infrastructure/ai/providers"
	"github.com/linkflow-ai/linkflow/internal/shared/types"
)

// ChatNode handles multi-turn AI chat conversations
type ChatNode struct {
	gateway *gateway.Gateway
	factory *providers.Factory
}

// NewChatNode creates a new AI chat node
func NewChatNode() *ChatNode {
	return &ChatNode{
		factory: providers.NewFactory(),
	}
}

// NewChatNodeWithGateway creates a new AI chat node with a gateway
func NewChatNodeWithGateway(gw *gateway.Gateway) *ChatNode {
	return &ChatNode{
		gateway: gw,
		factory: providers.NewFactory(),
	}
}

func (n *ChatNode) Execute(ctx context.Context, runtime *executor.Runtime, node map[string]interface{}) (types.JSON, error) {
	params, _ := node["parameters"].(map[string]interface{})

	// Get credentials
	apiKey, _ := params["api_key"].(string)
	if apiKey == "" {
		// Try to get from credential reference
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

	// Build messages
	var messages []ai.Message

	// Add system message if provided
	if systemMsg, ok := params["system_message"].(string); ok && systemMsg != "" {
		messages = append(messages, ai.NewSystemMessage(systemMsg))
	}

	// Add conversation history if provided
	if history, ok := params["messages"].([]interface{}); ok {
		for _, m := range history {
			if msgMap, ok := m.(map[string]interface{}); ok {
				role, _ := msgMap["role"].(string)
				content, _ := msgMap["content"].(string)
				messages = append(messages, ai.NewTextMessage(ai.Role(role), content))
			}
		}
	}

	// Add user message
	userMessage, _ := params["message"].(string)
	if userMessage == "" {
		userMessage, _ = params["prompt"].(string)
	}
	if userMessage != "" {
		messages = append(messages, ai.NewUserMessage(userMessage))
	}

	if len(messages) == 0 {
		return nil, fmt.Errorf("at least one message is required")
	}

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

	// Add tools if provided
	if toolsData, ok := params["tools"].([]interface{}); ok && len(toolsData) > 0 {
		tools := make([]ai.Tool, len(toolsData))
		for i, t := range toolsData {
			if toolMap, ok := t.(map[string]interface{}); ok {
				name, _ := toolMap["name"].(string)
				desc, _ := toolMap["description"].(string)
				parameters, _ := toolMap["parameters"].(map[string]interface{})
				tools[i] = ai.NewTool(name, desc, parameters)
			}
		}
		req.Tools = tools
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
		return nil, fmt.Errorf("chat request failed: %w", err)
	}

	// Build response
	result := types.JSON{
		"response":      resp.GetText(),
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
	}

	// Include tool calls if present
	if resp.HasToolCalls() {
		toolCalls := make([]map[string]interface{}, len(resp.Message.ToolCalls))
		for i, tc := range resp.Message.ToolCalls {
			var args map[string]interface{}
			json.Unmarshal([]byte(tc.Function.Arguments), &args)
			toolCalls[i] = map[string]interface{}{
				"id":        tc.ID,
				"name":      tc.Function.Name,
				"arguments": args,
			}
		}
		result["tool_calls"] = toolCalls
	}

	return result, nil
}

func (n *ChatNode) Metadata() wtypes.NodeMetadata {
	return wtypes.NodeMetadata{
		Type:        "ai.chat",
		Name:        "AI Chat",
		Description: "Multi-turn AI chat conversation with support for tools and vision",
		Category:    "ai",
		Version:     "1.0.0",
		Icon:        "MessageChatCircle",
		Color:       "#10B981",
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
					{Name: "Azure OpenAI", Value: "azure_openai"},
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
				Type:        "options",
				Required:    false,
				Description: "Model to use (defaults to provider's default)",
				Options: []wtypes.ParamOption{
					// OpenAI
					{Name: "GPT-4o", Value: "gpt-4o"},
					{Name: "GPT-4o Mini", Value: "gpt-4o-mini"},
					{Name: "O1", Value: "o1"},
					{Name: "O1 Mini", Value: "o1-mini"},
					// Anthropic
					{Name: "Claude 3.5 Sonnet", Value: "claude-3-5-sonnet-20241022"},
					{Name: "Claude 3.5 Haiku", Value: "claude-3-5-haiku-20241022"},
					{Name: "Claude 3 Opus", Value: "claude-3-opus-20240229"},
					// Google
					{Name: "Gemini 1.5 Pro", Value: "gemini-1.5-pro"},
					{Name: "Gemini 1.5 Flash", Value: "gemini-1.5-flash"},
					{Name: "Gemini 2.0 Flash", Value: "gemini-2.0-flash-exp"},
				},
			},
			{
				Name:        "system_message",
				DisplayName: "System Message",
				Type:        "string",
				Required:    false,
				Description: "System instructions for the AI",
			},
			{
				Name:        "message",
				DisplayName: "User Message",
				Type:        "string",
				Required:    true,
				Description: "The message to send to the AI",
			},
			{
				Name:        "messages",
				DisplayName: "Conversation History",
				Type:        "json",
				Required:    false,
				Description: "Previous messages for multi-turn conversation",
			},
			{
				Name:        "max_tokens",
				DisplayName: "Max Tokens",
				Type:        "number",
				Required:    false,
				Default:     4096,
				Description: "Maximum tokens in response",
			},
			{
				Name:        "temperature",
				DisplayName: "Temperature",
				Type:        "number",
				Required:    false,
				Default:     0.7,
				Description: "Randomness (0-2, lower = more focused)",
			},
			{
				Name:        "tools",
				DisplayName: "Tools",
				Type:        "json",
				Required:    false,
				Description: "Function tools the AI can call",
			},
			{
				Name:        "json_mode",
				DisplayName: "JSON Mode",
				Type:        "boolean",
				Required:    false,
				Default:     false,
				Description: "Force JSON output",
			},
		},
		Credentials: []string{"openai", "anthropic", "google_ai"},
	}
}
