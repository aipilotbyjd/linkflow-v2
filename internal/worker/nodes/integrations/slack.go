package integrations

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/worker/core"
)

type SlackNode struct{}

func (n *SlackNode) Type() string {
	return "integration.slack"
}

func (n *SlackNode) Execute(ctx context.Context, execCtx *core.ExecutionContext) (map[string]interface{}, error) {
	config := execCtx.Config
	operation := getString(config, "operation", "sendMessage")

	// Get credential
	credIDStr := getString(config, "credentialId", "")
	if credIDStr == "" {
		return nil, fmt.Errorf("credential is required")
	}

	credID, err := uuid.Parse(credIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid credential ID")
	}

	cred, err := execCtx.GetCredential(credID)
	if err != nil {
		return nil, fmt.Errorf("failed to get credential: %w", err)
	}

	token := cred.Token
	if token == "" {
		token = cred.AccessToken
	}

	switch operation {
	case "sendMessage":
		return n.sendMessage(ctx, token, config)
	case "updateMessage":
		return n.updateMessage(ctx, token, config)
	case "deleteMessage":
		return n.deleteMessage(ctx, token, config)
	case "uploadFile":
		return n.uploadFile(ctx, token, config)
	case "getChannel":
		return n.getChannel(ctx, token, config)
	case "listChannels":
		return n.listChannels(ctx, token, config)
	case "getUser":
		return n.getUser(ctx, token, config)
	case "listUsers":
		return n.listUsers(ctx, token, config)
	case "addReaction":
		return n.addReaction(ctx, token, config)
	default:
		return nil, fmt.Errorf("unknown operation: %s", operation)
	}
}

func (n *SlackNode) sendMessage(ctx context.Context, token string, config map[string]interface{}) (map[string]interface{}, error) {
	channel := getString(config, "channel", "")
	text := getString(config, "text", "")
	blocks := config["blocks"]
	attachments := config["attachments"]
	threadTs := getString(config, "threadTs", "")

	payload := map[string]interface{}{
		"channel": channel,
		"text":    text,
	}

	if blocks != nil {
		payload["blocks"] = blocks
	}
	if attachments != nil {
		payload["attachments"] = attachments
	}
	if threadTs != "" {
		payload["thread_ts"] = threadTs
	}

	return n.makeRequest(ctx, token, "POST", "https://slack.com/api/chat.postMessage", payload)
}

func (n *SlackNode) updateMessage(ctx context.Context, token string, config map[string]interface{}) (map[string]interface{}, error) {
	channel := getString(config, "channel", "")
	ts := getString(config, "ts", "")
	text := getString(config, "text", "")
	blocks := config["blocks"]

	payload := map[string]interface{}{
		"channel": channel,
		"ts":      ts,
		"text":    text,
	}

	if blocks != nil {
		payload["blocks"] = blocks
	}

	return n.makeRequest(ctx, token, "POST", "https://slack.com/api/chat.update", payload)
}

func (n *SlackNode) deleteMessage(ctx context.Context, token string, config map[string]interface{}) (map[string]interface{}, error) {
	channel := getString(config, "channel", "")
	ts := getString(config, "ts", "")

	payload := map[string]interface{}{
		"channel": channel,
		"ts":      ts,
	}

	return n.makeRequest(ctx, token, "POST", "https://slack.com/api/chat.delete", payload)
}

func (n *SlackNode) uploadFile(ctx context.Context, token string, config map[string]interface{}) (map[string]interface{}, error) {
	channels := getString(config, "channels", "")
	content := getString(config, "content", "")
	filename := getString(config, "filename", "file.txt")
	title := getString(config, "title", "")

	payload := map[string]interface{}{
		"channels": channels,
		"content":  content,
		"filename": filename,
	}

	if title != "" {
		payload["title"] = title
	}

	return n.makeRequest(ctx, token, "POST", "https://slack.com/api/files.upload", payload)
}

func (n *SlackNode) getChannel(ctx context.Context, token string, config map[string]interface{}) (map[string]interface{}, error) {
	channel := getString(config, "channel", "")

	return n.makeRequest(ctx, token, "GET", fmt.Sprintf("https://slack.com/api/conversations.info?channel=%s", channel), nil)
}

func (n *SlackNode) listChannels(ctx context.Context, token string, config map[string]interface{}) (map[string]interface{}, error) {
	limit := getInt(config, "limit", 100)
	types := getString(config, "types", "public_channel,private_channel")

	return n.makeRequest(ctx, token, "GET", fmt.Sprintf("https://slack.com/api/conversations.list?limit=%d&types=%s", limit, types), nil)
}

func (n *SlackNode) getUser(ctx context.Context, token string, config map[string]interface{}) (map[string]interface{}, error) {
	user := getString(config, "user", "")

	return n.makeRequest(ctx, token, "GET", fmt.Sprintf("https://slack.com/api/users.info?user=%s", user), nil)
}

func (n *SlackNode) listUsers(ctx context.Context, token string, config map[string]interface{}) (map[string]interface{}, error) {
	limit := getInt(config, "limit", 100)

	return n.makeRequest(ctx, token, "GET", fmt.Sprintf("https://slack.com/api/users.list?limit=%d", limit), nil)
}

func (n *SlackNode) addReaction(ctx context.Context, token string, config map[string]interface{}) (map[string]interface{}, error) {
	channel := getString(config, "channel", "")
	ts := getString(config, "timestamp", "")
	name := getString(config, "name", "")

	payload := map[string]interface{}{
		"channel":   channel,
		"timestamp": ts,
		"name":      name,
	}

	return n.makeRequest(ctx, token, "POST", "https://slack.com/api/reactions.add", payload)
}

func (n *SlackNode) makeRequest(ctx context.Context, token, method, url string, payload map[string]interface{}) (map[string]interface{}, error) {
	var body io.Reader
	if payload != nil {
		jsonBody, _ := json.Marshal(payload)
		body = bytes.NewReader(jsonBody)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	var result map[string]interface{}
	json.Unmarshal(respBody, &result)

	if ok, _ := result["ok"].(bool); !ok {
		errMsg, _ := result["error"].(string)
		return result, fmt.Errorf("Slack API error: %s", errMsg)
	}

	return result, nil
}

var _ core.Node = (*SlackNode)(nil)
