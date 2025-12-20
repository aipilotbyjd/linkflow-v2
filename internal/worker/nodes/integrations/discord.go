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

const discordAPIBase = "https://discord.com/api/v10"

type DiscordNode struct{}

func (n *DiscordNode) Type() string {
	return "integration.discord"
}

func (n *DiscordNode) Execute(ctx context.Context, execCtx *core.ExecutionContext) (map[string]interface{}, error) {
	config := execCtx.Config
	operation := getString(config, "operation", "sendMessage")

	// Get credential
	credIDStr := getString(config, "credentialId", "")
	if credIDStr == "" {
		// Check for webhook URL (no auth needed)
		if webhookURL := getString(config, "webhookUrl", ""); webhookURL != "" {
			return n.sendWebhook(ctx, webhookURL, config)
		}
		return nil, fmt.Errorf("credential or webhook URL is required")
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
	case "editMessage":
		return n.editMessage(ctx, token, config)
	case "deleteMessage":
		return n.deleteMessage(ctx, token, config)
	case "sendWebhook":
		return n.sendWebhook(ctx, getString(config, "webhookUrl", ""), config)
	case "getChannel":
		return n.getChannel(ctx, token, config)
	case "listChannels":
		return n.listChannels(ctx, token, config)
	case "getUser":
		return n.getUser(ctx, token, config)
	case "getGuild":
		return n.getGuild(ctx, token, config)
	case "addReaction":
		return n.addReaction(ctx, token, config)
	case "createThread":
		return n.createThread(ctx, token, config)
	default:
		return nil, fmt.Errorf("unknown operation: %s", operation)
	}
}

func (n *DiscordNode) sendMessage(ctx context.Context, token string, config map[string]interface{}) (map[string]interface{}, error) {
	channelID := getString(config, "channelId", "")
	content := getString(config, "content", "")
	embeds := config["embeds"]
	tts := getBool(config, "tts", false)

	payload := map[string]interface{}{
		"content": content,
		"tts":     tts,
	}

	if embeds != nil {
		payload["embeds"] = embeds
	}

	return n.makeRequest(ctx, token, "POST", fmt.Sprintf("%s/channels/%s/messages", discordAPIBase, channelID), payload)
}

func (n *DiscordNode) editMessage(ctx context.Context, token string, config map[string]interface{}) (map[string]interface{}, error) {
	channelID := getString(config, "channelId", "")
	messageID := getString(config, "messageId", "")
	content := getString(config, "content", "")
	embeds := config["embeds"]

	payload := map[string]interface{}{
		"content": content,
	}

	if embeds != nil {
		payload["embeds"] = embeds
	}

	return n.makeRequest(ctx, token, "PATCH", fmt.Sprintf("%s/channels/%s/messages/%s", discordAPIBase, channelID, messageID), payload)
}

func (n *DiscordNode) deleteMessage(ctx context.Context, token string, config map[string]interface{}) (map[string]interface{}, error) {
	channelID := getString(config, "channelId", "")
	messageID := getString(config, "messageId", "")

	return n.makeRequest(ctx, token, "DELETE", fmt.Sprintf("%s/channels/%s/messages/%s", discordAPIBase, channelID, messageID), nil)
}

func (n *DiscordNode) sendWebhook(ctx context.Context, webhookURL string, config map[string]interface{}) (map[string]interface{}, error) {
	if webhookURL == "" {
		return nil, fmt.Errorf("webhook URL is required")
	}

	content := getString(config, "content", "")
	username := getString(config, "username", "")
	avatarURL := getString(config, "avatarUrl", "")
	embeds := config["embeds"]

	payload := map[string]interface{}{
		"content": content,
	}

	if username != "" {
		payload["username"] = username
	}
	if avatarURL != "" {
		payload["avatar_url"] = avatarURL
	}
	if embeds != nil {
		payload["embeds"] = embeds
	}

	jsonBody, _ := json.Marshal(payload)

	req, err := http.NewRequestWithContext(ctx, "POST", webhookURL, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("webhook error %d: %s", resp.StatusCode, string(body))
	}

	return map[string]interface{}{
		"sent":   true,
		"status": resp.StatusCode,
	}, nil
}

func (n *DiscordNode) getChannel(ctx context.Context, token string, config map[string]interface{}) (map[string]interface{}, error) {
	channelID := getString(config, "channelId", "")
	return n.makeRequest(ctx, token, "GET", fmt.Sprintf("%s/channels/%s", discordAPIBase, channelID), nil)
}

func (n *DiscordNode) listChannels(ctx context.Context, token string, config map[string]interface{}) (map[string]interface{}, error) {
	guildID := getString(config, "guildId", "")
	return n.makeRequest(ctx, token, "GET", fmt.Sprintf("%s/guilds/%s/channels", discordAPIBase, guildID), nil)
}

func (n *DiscordNode) getUser(ctx context.Context, token string, config map[string]interface{}) (map[string]interface{}, error) {
	userID := getString(config, "userId", "@me")
	return n.makeRequest(ctx, token, "GET", fmt.Sprintf("%s/users/%s", discordAPIBase, userID), nil)
}

func (n *DiscordNode) getGuild(ctx context.Context, token string, config map[string]interface{}) (map[string]interface{}, error) {
	guildID := getString(config, "guildId", "")
	return n.makeRequest(ctx, token, "GET", fmt.Sprintf("%s/guilds/%s", discordAPIBase, guildID), nil)
}

func (n *DiscordNode) addReaction(ctx context.Context, token string, config map[string]interface{}) (map[string]interface{}, error) {
	channelID := getString(config, "channelId", "")
	messageID := getString(config, "messageId", "")
	emoji := getString(config, "emoji", "👍")

	_, err := n.makeRequest(ctx, token, "PUT", fmt.Sprintf("%s/channels/%s/messages/%s/reactions/%s/@me", discordAPIBase, channelID, messageID, emoji), nil)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"added": true,
		"emoji": emoji,
	}, nil
}

func (n *DiscordNode) createThread(ctx context.Context, token string, config map[string]interface{}) (map[string]interface{}, error) {
	channelID := getString(config, "channelId", "")
	messageID := getString(config, "messageId", "")
	name := getString(config, "name", "")
	autoArchiveDuration := getInt(config, "autoArchiveDuration", 1440) // 24 hours

	payload := map[string]interface{}{
		"name":                  name,
		"auto_archive_duration": autoArchiveDuration,
	}

	var url string
	if messageID != "" {
		url = fmt.Sprintf("%s/channels/%s/messages/%s/threads", discordAPIBase, channelID, messageID)
	} else {
		url = fmt.Sprintf("%s/channels/%s/threads", discordAPIBase, channelID)
	}

	return n.makeRequest(ctx, token, "POST", url, payload)
}

func (n *DiscordNode) makeRequest(ctx context.Context, token, method, url string, payload map[string]interface{}) (map[string]interface{}, error) {
	var body io.Reader
	if payload != nil {
		jsonBody, _ := json.Marshal(payload)
		body = bytes.NewReader(jsonBody)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bot "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode >= 400 {
		var errResp map[string]interface{}
		json.Unmarshal(respBody, &errResp)
		return errResp, fmt.Errorf("Discord API error %d: %s", resp.StatusCode, string(respBody))
	}

	var result map[string]interface{}
	json.Unmarshal(respBody, &result)

	return result, nil
}

var _ core.Node = (*DiscordNode)(nil)
