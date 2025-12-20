package integrations

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"

	"github.com/linkflow-ai/linkflow/internal/worker/nodes"
)

type TelegramNode struct{}

func NewTelegramNode() *TelegramNode {
	return &TelegramNode{}
}

func (n *TelegramNode) Type() string {
	return "integration.telegram"
}

func (n *TelegramNode) Execute(ctx context.Context, execCtx *nodes.ExecutionContext) (map[string]interface{}, error) {
	credID := getString(execCtx.Config, "credentialId", "")
	if credID == "" {
		return nil, fmt.Errorf("credential ID is required")
	}

	cred, err := execCtx.GetCredential(parseUUID(credID))
	if err != nil {
		return nil, fmt.Errorf("failed to get credential: %w", err)
	}

	botToken := cred.Token
	if botToken == "" {
		botToken = cred.APIKey
	}
	operation := getString(execCtx.Config, "operation", "sendMessage")

	switch operation {
	case "sendMessage":
		return n.sendMessage(ctx, botToken, execCtx.Config)
	case "editMessage":
		return n.editMessage(ctx, botToken, execCtx.Config)
	case "deleteMessage":
		return n.deleteMessage(ctx, botToken, execCtx.Config)
	case "sendPhoto":
		return n.sendPhoto(ctx, botToken, execCtx.Config)
	case "sendDocument":
		return n.sendDocument(ctx, botToken, execCtx.Config)
	case "sendLocation":
		return n.sendLocation(ctx, botToken, execCtx.Config)
	case "getChat":
		return n.getChat(ctx, botToken, execCtx.Config)
	case "getChatMember":
		return n.getChatMember(ctx, botToken, execCtx.Config)
	case "getChatMemberCount":
		return n.getChatMemberCount(ctx, botToken, execCtx.Config)
	case "sendPoll":
		return n.sendPoll(ctx, botToken, execCtx.Config)
	case "pinMessage":
		return n.pinMessage(ctx, botToken, execCtx.Config)
	case "getMe":
		return n.getMe(ctx, botToken)
	default:
		return nil, fmt.Errorf("unknown operation: %s", operation)
	}
}

func (n *TelegramNode) sendMessage(ctx context.Context, token string, config map[string]interface{}) (map[string]interface{}, error) {
	payload := map[string]interface{}{
		"chat_id": getString(config, "chatId", ""),
		"text":    getString(config, "text", ""),
	}

	if parseMode := getString(config, "parseMode", ""); parseMode != "" {
		payload["parse_mode"] = parseMode
	}
	if disablePreview := getBool(config, "disableLinkPreview", false); disablePreview {
		payload["disable_web_page_preview"] = true
	}
	if disableNotification := getBool(config, "disableNotification", false); disableNotification {
		payload["disable_notification"] = true
	}
	if replyTo := getInt(config, "replyToMessageId", 0); replyTo > 0 {
		payload["reply_to_message_id"] = replyTo
	}

	if buttons, ok := config["inlineKeyboard"].([]interface{}); ok && len(buttons) > 0 {
		payload["reply_markup"] = map[string]interface{}{
			"inline_keyboard": buttons,
		}
	}

	return n.makeRequest(ctx, token, "sendMessage", payload)
}

func (n *TelegramNode) editMessage(ctx context.Context, token string, config map[string]interface{}) (map[string]interface{}, error) {
	payload := map[string]interface{}{
		"chat_id":    getString(config, "chatId", ""),
		"message_id": getInt(config, "messageId", 0),
		"text":       getString(config, "text", ""),
	}

	if parseMode := getString(config, "parseMode", ""); parseMode != "" {
		payload["parse_mode"] = parseMode
	}

	return n.makeRequest(ctx, token, "editMessageText", payload)
}

func (n *TelegramNode) deleteMessage(ctx context.Context, token string, config map[string]interface{}) (map[string]interface{}, error) {
	payload := map[string]interface{}{
		"chat_id":    getString(config, "chatId", ""),
		"message_id": getInt(config, "messageId", 0),
	}

	return n.makeRequest(ctx, token, "deleteMessage", payload)
}

func (n *TelegramNode) sendPhoto(ctx context.Context, token string, config map[string]interface{}) (map[string]interface{}, error) {
	photoURL := getString(config, "photo", "")

	if photoURL != "" {
		payload := map[string]interface{}{
			"chat_id": getString(config, "chatId", ""),
			"photo":   photoURL,
		}
		if caption := getString(config, "caption", ""); caption != "" {
			payload["caption"] = caption
		}
		return n.makeRequest(ctx, token, "sendPhoto", payload)
	}

	return nil, fmt.Errorf("photo URL is required")
}

func (n *TelegramNode) sendDocument(ctx context.Context, token string, config map[string]interface{}) (map[string]interface{}, error) {
	docURL := getString(config, "document", "")

	if docURL != "" {
		payload := map[string]interface{}{
			"chat_id":  getString(config, "chatId", ""),
			"document": docURL,
		}
		if caption := getString(config, "caption", ""); caption != "" {
			payload["caption"] = caption
		}
		return n.makeRequest(ctx, token, "sendDocument", payload)
	}

	return nil, fmt.Errorf("document URL is required")
}

func (n *TelegramNode) sendLocation(ctx context.Context, token string, config map[string]interface{}) (map[string]interface{}, error) {
	payload := map[string]interface{}{
		"chat_id":   getString(config, "chatId", ""),
		"latitude":  getFloat(config, "latitude", 0),
		"longitude": getFloat(config, "longitude", 0),
	}

	return n.makeRequest(ctx, token, "sendLocation", payload)
}

func (n *TelegramNode) getChat(ctx context.Context, token string, config map[string]interface{}) (map[string]interface{}, error) {
	payload := map[string]interface{}{
		"chat_id": getString(config, "chatId", ""),
	}

	return n.makeRequest(ctx, token, "getChat", payload)
}

func (n *TelegramNode) getChatMember(ctx context.Context, token string, config map[string]interface{}) (map[string]interface{}, error) {
	payload := map[string]interface{}{
		"chat_id": getString(config, "chatId", ""),
		"user_id": getInt(config, "userId", 0),
	}

	return n.makeRequest(ctx, token, "getChatMember", payload)
}

func (n *TelegramNode) getChatMemberCount(ctx context.Context, token string, config map[string]interface{}) (map[string]interface{}, error) {
	payload := map[string]interface{}{
		"chat_id": getString(config, "chatId", ""),
	}

	return n.makeRequest(ctx, token, "getChatMemberCount", payload)
}

func (n *TelegramNode) sendPoll(ctx context.Context, token string, config map[string]interface{}) (map[string]interface{}, error) {
	payload := map[string]interface{}{
		"chat_id":  getString(config, "chatId", ""),
		"question": getString(config, "question", ""),
		"options":  config["options"],
	}

	if isAnonymous := getBool(config, "isAnonymous", true); !isAnonymous {
		payload["is_anonymous"] = false
	}
	if pollType := getString(config, "type", ""); pollType != "" {
		payload["type"] = pollType
	}
	if multipleAnswers := getBool(config, "allowsMultipleAnswers", false); multipleAnswers {
		payload["allows_multiple_answers"] = true
	}

	return n.makeRequest(ctx, token, "sendPoll", payload)
}

func (n *TelegramNode) pinMessage(ctx context.Context, token string, config map[string]interface{}) (map[string]interface{}, error) {
	payload := map[string]interface{}{
		"chat_id":    getString(config, "chatId", ""),
		"message_id": getInt(config, "messageId", 0),
	}

	if disableNotification := getBool(config, "disableNotification", false); disableNotification {
		payload["disable_notification"] = true
	}

	return n.makeRequest(ctx, token, "pinChatMessage", payload)
}

func (n *TelegramNode) getMe(ctx context.Context, token string) (map[string]interface{}, error) {
	return n.makeRequest(ctx, token, "getMe", nil)
}

func (n *TelegramNode) makeRequest(ctx context.Context, token, method string, payload map[string]interface{}) (map[string]interface{}, error) {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/%s", token, method)

	var body io.Reader
	if payload != nil {
		jsonData, err := json.Marshal(payload)
		if err != nil {
			return nil, err
		}
		body = bytes.NewReader(jsonData)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if ok, exists := result["ok"].(bool); exists && !ok {
		description := result["description"]
		return nil, fmt.Errorf("telegram API error: %v", description)
	}

	return map[string]interface{}{
		"ok":     result["ok"],
		"result": result["result"],
	}, nil
}

func (n *TelegramNode) sendFileWithMultipart(ctx context.Context, token, method string, config map[string]interface{}, fileField string, fileData []byte, filename string) (map[string]interface{}, error) {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/%s", token, method)

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	if err := writer.WriteField("chat_id", getString(config, "chatId", "")); err != nil {
		return nil, err
	}

	if caption := getString(config, "caption", ""); caption != "" {
		if err := writer.WriteField("caption", caption); err != nil {
			return nil, err
		}
	}

	part, err := writer.CreateFormFile(fileField, filename)
	if err != nil {
		return nil, err
	}
	if _, err := part.Write(fileData); err != nil {
		return nil, err
	}

	if err := writer.Close(); err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, &buf)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, err
	}

	return result, nil
}
