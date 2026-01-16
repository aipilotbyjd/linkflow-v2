package communication

import (
	"context"
	"fmt"

	"github.com/linkflow-ai/linkflow/internal/adapters/worker/executor"
	wtypes "github.com/linkflow-ai/linkflow/internal/adapters/worker/types"
	"github.com/linkflow-ai/linkflow/internal/shared/types"
)

type DiscordNode struct{}

func NewDiscordNode() *DiscordNode {
	return &DiscordNode{}
}

func (n *DiscordNode) Execute(ctx context.Context, runtime *executor.Runtime, node map[string]interface{}) (types.JSON, error) {
	params, _ := node["parameters"].(map[string]interface{})
	operation, _ := params["operation"].(string)

	switch operation {
	case "send_message":
		return n.sendMessage(ctx, params)
	case "send_embed":
		return n.sendEmbed(ctx, params)
	default:
		return nil, fmt.Errorf("unsupported Discord operation: %s", operation)
	}
}

func (n *DiscordNode) sendMessage(ctx context.Context, params map[string]interface{}) (types.JSON, error) {
	webhookUrl, _ := params["webhook_url"].(string)
	content, _ := params["content"].(string)
	username, _ := params["username"].(string)

	// TODO: Implement Discord webhook message sending
	_ = webhookUrl
	return types.JSON{
		"success":  true,
		"content":  content,
		"username": username,
	}, nil
}

func (n *DiscordNode) sendEmbed(ctx context.Context, params map[string]interface{}) (types.JSON, error) {
	webhookUrl, _ := params["webhook_url"].(string)
	title, _ := params["title"].(string)
	description, _ := params["description"].(string)

	// TODO: Implement Discord webhook embed sending
	_ = webhookUrl
	return types.JSON{
		"success":     true,
		"title":       title,
		"description": description,
	}, nil
}

func (n *DiscordNode) Metadata() wtypes.NodeMetadata {
	return wtypes.NodeMetadata{
		Type:        "integration.discord",
		Name:        "Discord",
		Description: "Send messages to Discord",
		Category:    "integration",
		Version:     "1.0.0",
		Inputs:      []wtypes.NodePort{{Name: "main", Type: "any"}},
		Outputs:     []wtypes.NodePort{{Name: "main", Type: "object"}},
		Parameters:  []wtypes.NodeParameter{},
	}
}
