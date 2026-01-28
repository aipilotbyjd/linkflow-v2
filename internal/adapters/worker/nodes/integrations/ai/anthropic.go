package ai

import (
	"context"
	"fmt"

	"github.com/linkflow-ai/linkflow/internal/adapters/worker/executor"
	wtypes "github.com/linkflow-ai/linkflow/internal/adapters/worker/types"
	"github.com/linkflow-ai/linkflow/internal/shared/types"
)

type AnthropicNode struct{}

func NewAnthropicNode() *AnthropicNode {
	return &AnthropicNode{}
}

func (n *AnthropicNode) Execute(ctx context.Context, runtime *executor.Runtime, node map[string]interface{}) (types.JSON, error) {
	params, _ := node["parameters"].(map[string]interface{})
	operation, _ := params["operation"].(string)

	switch operation {
	case "message", "":
		return n.createMessage(ctx, params)
	default:
		return nil, fmt.Errorf("unsupported Anthropic operation: %s", operation)
	}
}

func (n *AnthropicNode) createMessage(ctx context.Context, params map[string]interface{}) (types.JSON, error) {
	model, _ := params["model"].(string)
	messages, _ := params["messages"].([]interface{})
	maxTokens, _ := params["max_tokens"].(float64)
	systemPrompt, _ := params["system"].(string)

	if model == "" {
		model = "claude-3-sonnet-20240229"
	}
	if maxTokens == 0 {
		maxTokens = 1024
	}

	// Anthropic API integration requires API key from credentials
	// This returns a placeholder - full implementation requires HTTP client and API key
	return types.JSON{
		"model":      model,
		"messages":   messages,
		"system":     systemPrompt,
		"max_tokens": int(maxTokens),
		"content":    "",
		"usage": map[string]interface{}{
			"input_tokens":  0,
			"output_tokens": 0,
		},
		"message": "Anthropic API requires API key credential configuration",
	}, nil
}

func (n *AnthropicNode) Metadata() wtypes.NodeMetadata {
	return wtypes.NodeMetadata{
		Type:        "integration.anthropic",
		Name:        "Anthropic",
		Description: "Integrate with Anthropic Claude API",
		Category:    "integration",
		Version:     "1.0.0",
		Icon:        "AiBrain01",
		Inputs:      []wtypes.NodePort{{Name: "main", Type: "any"}},
		Outputs:     []wtypes.NodePort{{Name: "main", Type: "object"}},
		Parameters:  []wtypes.NodeParameter{},
	}
}
