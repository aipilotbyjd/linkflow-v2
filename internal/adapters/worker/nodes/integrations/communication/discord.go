package communication

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/linkflow-ai/linkflow/internal/adapters/worker/executor"
	wtypes "github.com/linkflow-ai/linkflow/internal/adapters/worker/types"
	"github.com/linkflow-ai/linkflow/internal/shared/types"
)

type DiscordNode struct {
	httpClient *http.Client
}

func NewDiscordNode() *DiscordNode {
	return &DiscordNode{httpClient: &http.Client{Timeout: 30 * time.Second}}
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
		Description: "Send messages, embeds, and interact with Discord servers via webhooks or bot API",
		Category:    "integration",
		Version:     "1.0.0",
		Icon:        "message-circle",
		Color:       "#5865F2",
		Inputs: []wtypes.NodePort{
			{Name: "main", Type: "any", Description: "Input data for message content"},
		},
		Outputs: []wtypes.NodePort{
			{Name: "main", Type: "object", Description: "Discord API response"},
		},
		Parameters: []wtypes.NodeParameter{
			{
				Name:        "operation",
				DisplayName: "Operation",
				Type:        "options",
				Required:    true,
				Default:     "send_message",
				Description: "Discord operation to perform",
				Options: []wtypes.ParamOption{
					{Name: "Send Message", Value: "send_message", Description: "Send a text message"},
					{Name: "Send Embed", Value: "send_embed", Description: "Send a rich embed message"},
					{Name: "Send File", Value: "send_file", Description: "Upload and send a file"},
					{Name: "Edit Message", Value: "edit_message", Description: "Edit an existing message"},
					{Name: "Delete Message", Value: "delete_message", Description: "Delete a message"},
				},
			},
			{
				Name:        "webhook_url",
				DisplayName: "Webhook URL",
				Type:        "string",
				Required:    true,
				Description: "Discord webhook URL",
				Placeholder: "https://discord.com/api/webhooks/...",
			},
			{
				Name:        "content",
				DisplayName: "Message Content",
				Type:        "string",
				Required:    false,
				Description: "Text content of the message (max 2000 characters)",
				Placeholder: "Hello from LinkFlow!",
				ShowIf:      "operation === 'send_message'",
			},
			{
				Name:        "username",
				DisplayName: "Username Override",
				Type:        "string",
				Required:    false,
				Description: "Override the webhook's default username",
				Placeholder: "Workflow Bot",
			},
			{
				Name:        "avatar_url",
				DisplayName: "Avatar URL",
				Type:        "string",
				Required:    false,
				Description: "Override the webhook's default avatar",
				Placeholder: "https://example.com/avatar.png",
			},
			{
				Name:        "title",
				DisplayName: "Embed Title",
				Type:        "string",
				Required:    false,
				Description: "Title of the embed",
				ShowIf:      "operation === 'send_embed'",
			},
			{
				Name:        "description",
				DisplayName: "Embed Description",
				Type:        "string",
				Required:    false,
				Description: "Description text of the embed",
				ShowIf:      "operation === 'send_embed'",
			},
			{
				Name:        "color",
				DisplayName: "Embed Color",
				Type:        "number",
				Required:    false,
				Description: "Embed color as decimal (e.g., 5814783 for blue)",
				ShowIf:      "operation === 'send_embed'",
			},
			{
				Name:        "url",
				DisplayName: "Embed URL",
				Type:        "string",
				Required:    false,
				Description: "URL for the embed title",
				ShowIf:      "operation === 'send_embed'",
			},
			{
				Name:        "thumbnail_url",
				DisplayName: "Thumbnail URL",
				Type:        "string",
				Required:    false,
				Description: "URL for embed thumbnail image",
				ShowIf:      "operation === 'send_embed'",
			},
			{
				Name:        "image_url",
				DisplayName: "Image URL",
				Type:        "string",
				Required:    false,
				Description: "URL for large embed image",
				ShowIf:      "operation === 'send_embed'",
			},
			{
				Name:        "footer_text",
				DisplayName: "Footer Text",
				Type:        "string",
				Required:    false,
				Description: "Text for embed footer",
				ShowIf:      "operation === 'send_embed'",
			},
			{
				Name:        "fields",
				DisplayName: "Embed Fields",
				Type:        "json",
				Required:    false,
				Description: "Array of embed fields",
				Placeholder: `[{"name": "Field 1", "value": "Value 1", "inline": true}]`,
				ShowIf:      "operation === 'send_embed'",
			},
			{
				Name:        "tts",
				DisplayName: "Text-to-Speech",
				Type:        "boolean",
				Required:    false,
				Default:     false,
				Description: "Send as text-to-speech message",
			},
		},
		Credentials: []string{"discord_webhook", "discord_bot"},
	}
}
