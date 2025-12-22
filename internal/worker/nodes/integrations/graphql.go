package integrations

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/linkflow-ai/linkflow/internal/worker/core"
)

// GraphQLNode executes GraphQL queries and mutations
type GraphQLNode struct{}

func (n *GraphQLNode) Type() string {
	return "integrations.graphql"
}

func (n *GraphQLNode) Execute(ctx context.Context, execCtx *core.ExecutionContext) (map[string]interface{}, error) {
	config := execCtx.Config
	input := execCtx.Input

	endpoint := core.GetString(config, "endpoint", "")
	if endpoint == "" {
		return nil, fmt.Errorf("GraphQL endpoint is required")
	}

	query := core.GetString(config, "query", "")
	if query == "" {
		if q, ok := input["query"].(string); ok {
			query = q
		}
	}
	if query == "" {
		return nil, fmt.Errorf("GraphQL query is required")
	}

	variables := config["variables"]
	if variables == nil {
		variables = input["variables"]
	}

	operationName := core.GetString(config, "operationName", "")
	timeout := core.GetInt(config, "timeout", 30)

	// Build request body
	requestBody := map[string]interface{}{
		"query": query,
	}
	if variables != nil {
		requestBody["variables"] = variables
	}
	if operationName != "" {
		requestBody["operationName"] = operationName
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Add authorization header
	if auth := core.GetString(config, "authorization", ""); auth != "" {
		req.Header.Set("Authorization", auth)
	}
	if bearer := core.GetString(config, "bearerToken", ""); bearer != "" {
		req.Header.Set("Authorization", "Bearer "+bearer)
	}

	// Add custom headers
	if headers, ok := config["headers"].(map[string]interface{}); ok {
		for k, v := range headers {
			req.Header.Set(k, fmt.Sprintf("%v", v))
		}
	}

	// Execute request
	client := &http.Client{Timeout: time.Duration(timeout) * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Check for GraphQL errors
	var errors []interface{}
	if errs, ok := result["errors"].([]interface{}); ok {
		errors = errs
	}

	return map[string]interface{}{
		"data":       result["data"],
		"errors":     errors,
		"hasErrors":  len(errors) > 0,
		"statusCode": resp.StatusCode,
		"extensions": result["extensions"],
	}, nil
}

// Note: GraphQLNode is registered in integrations/init.go
