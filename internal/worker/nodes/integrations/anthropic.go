package integrations

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/linkflow-ai/linkflow/internal/worker/nodes"
)

type AnthropicNode struct{}

func NewAnthropicNode() *AnthropicNode {
	return &AnthropicNode{}
}

func (n *AnthropicNode) Type() string {
	return "integration.anthropic"
}

func (n *AnthropicNode) Execute(ctx context.Context, execCtx *nodes.ExecutionContext) (map[string]interface{}, error) {
	credID := getString(execCtx.Config, "credentialId", "")
	if credID == "" {
		return nil, fmt.Errorf("credential ID is required")
	}

	cred, err := execCtx.GetCredential(parseUUID(credID))
	if err != nil {
		return nil, fmt.Errorf("failed to get credential: %w", err)
	}

	apiKey := cred.APIKey
	operation := getString(execCtx.Config, "operation", "chat")

	switch operation {
	case "chat":
		return n.chat(ctx, apiKey, execCtx.Config)
	default:
		return nil, fmt.Errorf("unknown operation: %s", operation)
	}
}

func (n *AnthropicNode) chat(ctx context.Context, apiKey string, config map[string]interface{}) (map[string]interface{}, error) {
	model := getString(config, "model", "claude-3-sonnet-20240229")
	maxTokens := getInt(config, "maxTokens", 1024)
	temperature := getFloat(config, "temperature", 1.0)
	systemPrompt := getString(config, "system", "")

	var messages []map[string]interface{}
	if msgConfig, ok := config["messages"].([]interface{}); ok {
		for _, m := range msgConfig {
			if msg, ok := m.(map[string]interface{}); ok {
				messages = append(messages, msg)
			}
		}
	} else {
		userMessage := getString(config, "message", "")
		if userMessage == "" {
			userMessage = getString(config, "prompt", "")
		}
		messages = []map[string]interface{}{
			{"role": "user", "content": userMessage},
		}
	}

	payload := map[string]interface{}{
		"model":      model,
		"max_tokens": maxTokens,
		"messages":   messages,
	}

	if temperature != 1.0 {
		payload["temperature"] = temperature
	}
	if systemPrompt != "" {
		payload["system"] = systemPrompt
	}

	if topP := getFloat(config, "topP", 0); topP > 0 {
		payload["top_p"] = topP
	}
	if topK := getInt(config, "topK", 0); topK > 0 {
		payload["top_k"] = topK
	}
	if stopSequences, ok := config["stopSequences"].([]interface{}); ok && len(stopSequences) > 0 {
		payload["stop_sequences"] = stopSequences
	}

	result, err := n.makeRequest(ctx, apiKey, "https://api.anthropic.com/v1/messages", payload)
	if err != nil {
		return nil, err
	}

	content := ""
	if contentArr, ok := result["content"].([]interface{}); ok && len(contentArr) > 0 {
		if textBlock, ok := contentArr[0].(map[string]interface{}); ok {
			if text, ok := textBlock["text"].(string); ok {
				content = text
			}
		}
	}

	return map[string]interface{}{
		"id":           result["id"],
		"type":         result["type"],
		"role":         result["role"],
		"content":      content,
		"model":        result["model"],
		"stopReason":   result["stop_reason"],
		"usage":        result["usage"],
		"fullResponse": result,
	}, nil
}

func (n *AnthropicNode) makeRequest(ctx context.Context, apiKey, url string, payload map[string]interface{}) (map[string]interface{}, error) {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode >= 400 {
		var errResp map[string]interface{}
		json.Unmarshal(body, &errResp)
		if errMsg, ok := errResp["error"].(map[string]interface{}); ok {
			return nil, fmt.Errorf("Anthropic API error: %v", errMsg["message"])
		}
		return nil, fmt.Errorf("Anthropic API error: %s", string(body))
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	return result, nil
}
