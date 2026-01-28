package communication

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/linkflow-ai/linkflow/internal/adapters/worker/executor"
	wtypes "github.com/linkflow-ai/linkflow/internal/adapters/worker/types"
	"github.com/linkflow-ai/linkflow/internal/shared/types"
)

type SlackNode struct {
	client *http.Client
}

func NewSlackNode() *SlackNode {
	return &SlackNode{
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

func (n *SlackNode) Execute(ctx context.Context, runtime *executor.Runtime, node map[string]interface{}) (types.JSON, error) {
	params, _ := node["parameters"].(map[string]interface{})

	webhookURL, _ := params["webhook_url"].(string)
	channel, _ := params["channel"].(string)
	message, _ := params["message"].(string)

	if webhookURL == "" {
		return nil, fmt.Errorf("Slack webhook URL is required")
	}
	if message == "" {
		return nil, fmt.Errorf("message is required")
	}

	payload := map[string]interface{}{
		"text": message,
	}
	if channel != "" {
		payload["channel"] = channel
	}

	body, _ := json.Marshal(payload)

	req, err := http.NewRequestWithContext(ctx, "POST", webhookURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := n.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send Slack message: %w", err)
	}
	defer resp.Body.Close()

	return types.JSON{
		"status":      "sent",
		"status_code": resp.StatusCode,
		"channel":     channel,
	}, nil
}

func (n *SlackNode) Metadata() wtypes.NodeMetadata {
	return wtypes.NodeMetadata{
		Type:        "integration.slack",
		Name:        "Slack",
		Description: "Send messages, upload files, and interact with Slack channels, users, and workflows",
		Category:    "integration",
		Version:     "1.0.0",
		Icon:        "Slack",
		Color:       "#4A154B",
		Inputs: []wtypes.NodePort{
			{Name: "main", Type: "any", Description: "Input data for message content"},
		},
		Outputs: []wtypes.NodePort{
			{Name: "main", Type: "object", Description: "Slack API response with message details"},
		},
		Parameters: []wtypes.NodeParameter{
			{
				Name:        "operation",
				DisplayName: "Operation",
				Type:        "options",
				Required:    true,
				Default:     "send_message",
				Description: "Slack operation to perform",
				Options: []wtypes.ParamOption{
					{Name: "Send Message", Value: "send_message", Description: "Send a message to a channel or user"},
					{Name: "Update Message", Value: "update_message", Description: "Update an existing message"},
					{Name: "Delete Message", Value: "delete_message", Description: "Delete a message"},
					{Name: "React to Message", Value: "add_reaction", Description: "Add emoji reaction to a message"},
					{Name: "Upload File", Value: "upload_file", Description: "Upload a file to a channel"},
					{Name: "Get User Info", Value: "get_user", Description: "Get information about a user"},
					{Name: "List Channels", Value: "list_channels", Description: "List available channels"},
					{Name: "Create Channel", Value: "create_channel", Description: "Create a new channel"},
				},
			},
			{
				Name:        "auth_type",
				DisplayName: "Authentication Type",
				Type:        "options",
				Required:    true,
				Default:     "webhook",
				Description: "How to authenticate with Slack",
				Options: []wtypes.ParamOption{
					{Name: "Webhook URL", Value: "webhook", Description: "Use incoming webhook URL"},
					{Name: "Bot Token", Value: "bot_token", Description: "Use Slack Bot OAuth token"},
					{Name: "User Token", Value: "user_token", Description: "Use Slack User OAuth token"},
				},
			},
			{
				Name:        "webhook_url",
				DisplayName: "Webhook URL",
				Type:        "string",
				Required:    true,
				Description: "Slack incoming webhook URL",
				Placeholder: "https://hooks.slack.com/services/T.../B.../...",
				ShowIf:      "auth_type === 'webhook'",
			},
			{
				Name:        "channel",
				DisplayName: "Channel",
				Type:        "string",
				Required:    false,
				Description: "Channel ID or name (e.g., #general or C1234567890)",
				Placeholder: "#general",
				ShowIf:      "auth_type !== 'webhook' && (operation === 'send_message' || operation === 'upload_file' || operation === 'create_channel')",
			},
			{
				Name:        "user",
				DisplayName: "User",
				Type:        "string",
				Required:    false,
				Description: "User ID for direct messages (e.g., U1234567890)",
				Placeholder: "U1234567890",
				ShowIf:      "operation === 'send_message' || operation === 'get_user'",
			},
			{
				Name:        "message",
				DisplayName: "Message",
				Type:        "string",
				Required:    true,
				Description: "Message text (supports Slack markdown)",
				Placeholder: "Hello *world*! :wave:",
				ShowIf:      "operation === 'send_message' || operation === 'update_message'",
			},
			{
				Name:        "blocks",
				DisplayName: "Blocks",
				Type:        "json",
				Required:    false,
				Description: "Slack Block Kit blocks for rich message formatting",
				Placeholder: `[{"type": "section", "text": {"type": "mrkdwn", "text": "Hello!"}}]`,
				ShowIf:      "operation === 'send_message' || operation === 'update_message'",
			},
			{
				Name:        "attachments",
				DisplayName: "Attachments",
				Type:        "json",
				Required:    false,
				Description: "Legacy message attachments (prefer blocks for new messages)",
				ShowIf:      "operation === 'send_message'",
			},
			{
				Name:        "thread_ts",
				DisplayName: "Thread Timestamp",
				Type:        "string",
				Required:    false,
				Description: "Reply in thread (timestamp of parent message)",
				Placeholder: "1234567890.123456",
				ShowIf:      "operation === 'send_message' || operation === 'update_message'",
			},
			{
				Name:        "message_ts",
				DisplayName: "Message Timestamp",
				Type:        "string",
				Required:    true,
				Description: "Timestamp of message to update/delete/react to",
				Placeholder: "1234567890.123456",
				ShowIf:      "operation === 'update_message' || operation === 'delete_message' || operation === 'add_reaction'",
			},
			{
				Name:        "reaction",
				DisplayName: "Emoji Reaction",
				Type:        "string",
				Required:    true,
				Description: "Emoji name without colons (e.g., thumbsup)",
				Placeholder: "thumbsup",
				ShowIf:      "operation === 'add_reaction'",
			},
			{
				Name:        "file_content",
				DisplayName: "File Content",
				Type:        "string",
				Required:    false,
				Description: "File content (text or base64 for binary)",
				ShowIf:      "operation === 'upload_file'",
			},
			{
				Name:        "file_name",
				DisplayName: "File Name",
				Type:        "string",
				Required:    false,
				Description: "Name of the file to upload",
				Placeholder: "report.txt",
				ShowIf:      "operation === 'upload_file'",
			},
			{
				Name:        "file_title",
				DisplayName: "File Title",
				Type:        "string",
				Required:    false,
				Description: "Title of the file",
				ShowIf:      "operation === 'upload_file'",
			},
			{
				Name:        "username",
				DisplayName: "Bot Username",
				Type:        "string",
				Required:    false,
				Description: "Override bot username for this message",
				Placeholder: "My Bot",
			},
			{
				Name:        "icon_emoji",
				DisplayName: "Icon Emoji",
				Type:        "string",
				Required:    false,
				Description: "Override bot icon with emoji",
				Placeholder: ":robot_face:",
			},
			{
				Name:        "icon_url",
				DisplayName: "Icon URL",
				Type:        "string",
				Required:    false,
				Description: "Override bot icon with image URL",
				Placeholder: "https://example.com/icon.png",
			},
			{
				Name:        "unfurl_links",
				DisplayName: "Unfurl Links",
				Type:        "boolean",
				Required:    false,
				Default:     true,
				Description: "Enable link previews",
			},
			{
				Name:        "unfurl_media",
				DisplayName: "Unfurl Media",
				Type:        "boolean",
				Required:    false,
				Default:     true,
				Description: "Enable media previews",
			},
		},
		Credentials: []string{"slack_webhook", "slack_oauth"},
	}
}
