package http

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

type HTTPRequestNode struct {
	client *http.Client
}

func NewHTTPRequestNode() *HTTPRequestNode {
	return &HTTPRequestNode{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (n *HTTPRequestNode) Execute(ctx context.Context, runtime *executor.Runtime, node map[string]interface{}) (types.JSON, error) {
	params, _ := node["parameters"].(map[string]interface{})

	url, _ := params["url"].(string)
	method, _ := params["method"].(string)
	if method == "" {
		method = "GET"
	}

	var body io.Reader
	if bodyData, ok := params["body"]; ok {
		bodyBytes, _ := json.Marshal(bodyData)
		body = bytes.NewReader(bodyBytes)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if headers, ok := params["headers"].(map[string]interface{}); ok {
		for k, v := range headers {
			if vs, ok := v.(string); ok {
				req.Header.Set(k, vs)
			}
		}
	}

	if req.Header.Get("Content-Type") == "" && body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := n.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	result := types.JSON{
		"statusCode": resp.StatusCode,
		"headers":    resp.Header,
		"body":       string(respBody),
	}

	var jsonBody interface{}
	if json.Unmarshal(respBody, &jsonBody) == nil {
		result["json"] = jsonBody
	}

	return result, nil
}

func (n *HTTPRequestNode) Metadata() wtypes.NodeMetadata {
	return wtypes.NodeMetadata{
		Type:        "action.http",
		Name:        "HTTP Request",
		Description: "Make HTTP requests",
		Category:    "action",
		Version:     "1.0.0",
		Inputs:      []wtypes.NodePort{{Name: "main", Type: "any"}},
		Outputs:     []wtypes.NodePort{{Name: "main", Type: "object"}},
		Parameters: []wtypes.NodeParameter{
			{Name: "url", DisplayName: "URL", Type: "string", Required: true},
			{Name: "method", DisplayName: "Method", Type: "options", Required: false, Default: "GET"},
			{Name: "headers", DisplayName: "Headers", Type: "json", Required: false},
			{Name: "body", DisplayName: "Body", Type: "json", Required: false},
		},
	}
}
