package ai

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

type OpenAINode struct {
	client *http.Client
}

func NewOpenAINode() *OpenAINode {
	return &OpenAINode{
		client: &http.Client{},
	}
}

func (n *OpenAINode) Execute(ctx context.Context, runtime *executor.Runtime, node map[string]interface{}) (types.JSON, error) {
	params, _ := node["parameters"].(map[string]interface{})

	apiKey, _ := params["api_key"].(string)
	model, _ := params["model"].(string)
	prompt, _ := params["prompt"].(string)

	if apiKey == "" {
		return nil, fmt.Errorf("OpenAI API key is required")
	}
	if prompt == "" {
		return nil, fmt.Errorf("prompt is required")
	}
	if model == "" {
		model = "gpt-3.5-turbo"
	}

	payload := map[string]interface{}{
		"model": model,
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
	}

	body, _ := json.Marshal(payload)

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.openai.com/v1/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := n.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call OpenAI: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	var result map[string]interface{}
	json.Unmarshal(respBody, &result)

	// Extract response text
	var responseText string
	if choices, ok := result["choices"].([]interface{}); ok && len(choices) > 0 {
		if choice, ok := choices[0].(map[string]interface{}); ok {
			if message, ok := choice["message"].(map[string]interface{}); ok {
				responseText, _ = message["content"].(string)
			}
		}
	}

	return types.JSON{
		"model":        model,
		"response":     responseText,
		"status_code":  resp.StatusCode,
		"raw_response": result,
	}, nil
}

func (n *OpenAINode) Metadata() wtypes.NodeMetadata {
	return wtypes.NodeMetadata{
		Type:        "integration.openai",
		Name:        "OpenAI",
		Description: "Integrate with OpenAI API",
		Category:    "integration",
		Version:     "1.0.0",
		Inputs:      []wtypes.NodePort{{Name: "main", Type: "any"}},
		Outputs:     []wtypes.NodePort{{Name: "main", Type: "object"}},
		Parameters: []wtypes.NodeParameter{
			{Name: "api_key", DisplayName: "API Key", Type: "credential", Required: true},
			{Name: "model", DisplayName: "Model", Type: "string", Required: false, Default: "gpt-3.5-turbo"},
			{Name: "prompt", DisplayName: "Prompt", Type: "string", Required: true},
		},
		Credentials: []string{"openai"},
	}
}
