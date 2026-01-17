package communication

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/linkflow-ai/linkflow/internal/adapters/worker/executor"
	wtypes "github.com/linkflow-ai/linkflow/internal/adapters/worker/types"
	"github.com/linkflow-ai/linkflow/internal/shared/types"
)

type DiscordNode struct {
	httpClient *http.Client
}

func NewDiscordNode() *DiscordNode {
	return &DiscordNode{httpClient: &http.Client{}}
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
	webhookURL, _ := params["webhook_url"].(string)
	content, _ := params["content"].(string)
	username, _ := params["username"].(string)
	avatarURL, _ := params["avatar_url"].(string)

	if webhookURL == "" {
		return nil, fmt.Errorf("webhook_url is required")
	}
	if content == "" {
		return nil, fmt.Errorf("content is required")
	}

	payload := map[string]interface{}{
		"content": content,
	}
	if username != "" {
		payload["username"] = username
	}
	if avatarURL != "" {
		payload["avatar_url"] = avatarURL
	}

	if err := n.sendWebhook(ctx, webhookURL, payload); err != nil {
		return nil, err
	}

	return types.JSON{
		"success":  true,
		"content":  content,
		"username": username,
	}, nil
}

func (n *DiscordNode) sendEmbed(ctx context.Context, params map[string]interface{}) (types.JSON, error) {
	webhookURL, _ := params["webhook_url"].(string)
	title, _ := params["title"].(string)
	description, _ := params["description"].(string)
	color, _ := params["color"].(float64)
	url, _ := params["url"].(string)
	username, _ := params["username"].(string)

	if webhookURL == "" {
		return nil, fmt.Errorf("webhook_url is required")
	}

	embed := map[string]interface{}{
		"title":       title,
		"description": description,
	}
	if color > 0 {
		embed["color"] = int(color)
	}
	if url != "" {
		embed["url"] = url
	}

	payload := map[string]interface{}{
		"embeds": []map[string]interface{}{embed},
	}
	if username != "" {
		payload["username"] = username
	}

	if err := n.sendWebhook(ctx, webhookURL, payload); err != nil {
		return nil, err
	}

	return types.JSON{
		"success":     true,
		"title":       title,
		"description": description,
	}, nil
}

func (n *DiscordNode) sendWebhook(ctx context.Context, webhookURL string, payload map[string]interface{}) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, webhookURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := n.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send webhook: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("discord webhook failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	return nil
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
