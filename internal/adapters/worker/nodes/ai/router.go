package ai

import (
	"context"
	"fmt"

	"github.com/linkflow-ai/linkflow/internal/adapters/worker/executor"
	wtypes "github.com/linkflow-ai/linkflow/internal/adapters/worker/types"
	"github.com/linkflow-ai/linkflow/internal/core/domain/ai"
	"github.com/linkflow-ai/linkflow/internal/infrastructure/ai/gateway"
	"github.com/linkflow-ai/linkflow/internal/shared/types"
)

// RouterNode routes requests to the optimal model based on criteria
type RouterNode struct{}

// NewRouterNode creates a new AI router node
func NewRouterNode() *RouterNode {
	return &RouterNode{}
}

func (n *RouterNode) Execute(ctx context.Context, runtime *executor.Runtime, node map[string]interface{}) (types.JSON, error) {
	params, _ := node["parameters"].(map[string]interface{})

	// Get the message/prompt to analyze
	message, _ := params["message"].(string)
	if message == "" {
		message, _ = params["prompt"].(string)
	}

	// Build messages for routing analysis
	var messages []ai.Message
	if message != "" {
		messages = append(messages, ai.NewUserMessage(message))
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

	if len(messages) == 0 {
		return nil, fmt.Errorf("message or messages is required")
	}

	// Build router request
	req := &ai.RouterRequest{
		Messages: messages,
	}

	// Parse preferences
	if maxCost, ok := params["max_cost_per_1m"].(float64); ok {
		req.MaxCostPer1M = maxCost
	}

	if preferSpeed, ok := params["prefer_speed"].(bool); ok {
		req.PreferSpeed = preferSpeed
	}

	if preferQuality, ok := params["prefer_quality"].(bool); ok {
		req.PreferQuality = preferQuality
	}

	if requireVision, ok := params["require_vision"].(bool); ok {
		req.RequireVision = requireVision
	}

	if requireTools, ok := params["require_tools"].(bool); ok {
		req.RequireTools = requireTools
	}

	// Parse preferred providers
	if providers, ok := params["preferred_providers"].([]interface{}); ok {
		for _, p := range providers {
			if pStr, ok := p.(string); ok {
				if provider, valid := ai.ParseProvider(pStr); valid {
					req.PreferredProviders = append(req.PreferredProviders, provider)
				}
			}
		}
	}

	// Use the gateway router
	router := gateway.NewRouter()
	resp, err := router.Route(ctx, req, nil)
	if err != nil {
		return nil, fmt.Errorf("routing failed: %w", err)
	}

	// Build alternatives list
	alternatives := make([]map[string]interface{}, len(resp.Alternatives))
	for i, alt := range resp.Alternatives {
		alternatives[i] = map[string]interface{}{
			"id":       alt.ID,
			"name":     alt.Name,
			"provider": alt.Provider.String(),
		}
	}

	return types.JSON{
		"selected_model":    resp.SelectedModel.ID,
		"selected_provider": resp.SelectedProvider.String(),
		"model_name":        resp.SelectedModel.Name,
		"reason":            resp.Reason,
		"alternatives":      alternatives,
		"model_info": map[string]interface{}{
			"context_window":      resp.SelectedModel.ContextWindow,
			"max_output_tokens":   resp.SelectedModel.MaxOutputTokens,
			"input_price_per_1m":  resp.SelectedModel.InputPricePer1M,
			"output_price_per_1m": resp.SelectedModel.OutputPricePer1M,
			"supports_vision":     resp.SelectedModel.SupportsVision,
			"supports_tools":      resp.SelectedModel.SupportsTools,
			"supports_json":       resp.SelectedModel.SupportsJSON,
		},
	}, nil
}

func (n *RouterNode) Metadata() wtypes.NodeMetadata {
	return wtypes.NodeMetadata{
		Type:        "ai.router",
		Name:        "AI Model Router",
		Description: "Intelligently route to the optimal AI model based on cost, speed, and capability requirements",
		Category:    "ai",
		Version:     "1.0.0",
		Icon:        "Route01",
		Color:       "#A855F7",
		Inputs:      []wtypes.NodePort{{Name: "main", Type: "any"}},
		Outputs:     []wtypes.NodePort{{Name: "main", Type: "object"}},
		Parameters: []wtypes.NodeParameter{
			{
				Name:        "message",
				DisplayName: "Message",
				Type:        "string",
				Required:    false,
				Description: "Message to analyze for routing",
			},
			{
				Name:        "messages",
				DisplayName: "Conversation",
				Type:        "json",
				Required:    false,
				Description: "Full conversation history",
			},
			{
				Name:        "preferred_providers",
				DisplayName: "Preferred Providers",
				Type:        "json",
				Required:    false,
				Description: "List of preferred providers (e.g., [\"openai\", \"anthropic\"])",
			},
			{
				Name:        "max_cost_per_1m",
				DisplayName: "Max Cost per 1M tokens",
				Type:        "number",
				Required:    false,
				Description: "Maximum acceptable cost per million tokens",
			},
			{
				Name:        "prefer_speed",
				DisplayName: "Prefer Speed",
				Type:        "boolean",
				Required:    false,
				Default:     false,
				Description: "Prioritize faster models",
			},
			{
				Name:        "prefer_quality",
				DisplayName: "Prefer Quality",
				Type:        "boolean",
				Required:    false,
				Default:     false,
				Description: "Prioritize higher quality models",
			},
			{
				Name:        "require_vision",
				DisplayName: "Require Vision",
				Type:        "boolean",
				Required:    false,
				Default:     false,
				Description: "Only select models with vision capability",
			},
			{
				Name:        "require_tools",
				DisplayName: "Require Tools",
				Type:        "boolean",
				Required:    false,
				Default:     false,
				Description: "Only select models with tool/function calling",
			},
		},
	}
}
