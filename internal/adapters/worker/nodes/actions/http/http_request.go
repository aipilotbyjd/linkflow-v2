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

	// Handle Authentication
	authType, _ := params["authentication"].(string)
	switch authType {
	case "basic":
		username, _ := params["username"].(string)
		password, _ := params["password"].(string)
		if username != "" || password != "" {
			req.SetBasicAuth(username, password)
		}
	case "bearer":
		token, _ := params["bearer_token"].(string)
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
	case "api_key":
		keyName, _ := params["api_key_name"].(string)
		keyValue, _ := params["api_key_value"].(string)
		location, _ := params["api_key_location"].(string)
		if keyName != "" && keyValue != "" {
			if location == "query" {
				q := req.URL.Query()
				q.Add(keyName, keyValue)
				req.URL.RawQuery = q.Encode()
			} else {
				req.Header.Set(keyName, keyValue)
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
		Description: "Make HTTP requests to external APIs and web services with full control over method, headers, body, and authentication",
		Category:    "action",
		Version:     "1.0.0",
		Icon:        "globe",
		Color:       "#3B82F6",
		Inputs: []wtypes.NodePort{
			{Name: "main", Type: "any", Description: "Input data available for use in request parameters"},
		},
		Outputs: []wtypes.NodePort{
			{Name: "main", Type: "object", Description: "Response with statusCode, headers, body, and json (if parseable)"},
		},
		Parameters: []wtypes.NodeParameter{
			{
				Name:        "method",
				DisplayName: "Method",
				Type:        "options",
				Required:    true,
				Default:     "GET",
				Description: "HTTP method for the request",
				Options: []wtypes.ParamOption{
					{Name: "GET", Value: "GET", Description: "Retrieve data"},
					{Name: "POST", Value: "POST", Description: "Create/submit data"},
					{Name: "PUT", Value: "PUT", Description: "Update/replace data"},
					{Name: "PATCH", Value: "PATCH", Description: "Partial update"},
					{Name: "DELETE", Value: "DELETE", Description: "Delete data"},
					{Name: "HEAD", Value: "HEAD", Description: "Get headers only"},
					{Name: "OPTIONS", Value: "OPTIONS", Description: "Get supported methods"},
				},
			},
			{
				Name:        "url",
				DisplayName: "URL",
				Type:        "string",
				Required:    true,
				Description: "The URL to send the request to (supports expressions like {{$input.url}})",
				Placeholder: "https://api.example.com/endpoint",
			},
			{
				Name:        "authentication",
				DisplayName: "Authentication",
				Type:        "options",
				Required:    false,
				Default:     "none",
				Description: "Authentication method for the request",
				Options: []wtypes.ParamOption{
					{Name: "None", Value: "none"},
					{Name: "Basic Auth", Value: "basic"},
					{Name: "Bearer Token", Value: "bearer"},
					{Name: "API Key", Value: "api_key"},
					{Name: "OAuth2", Value: "oauth2"},
					{Name: "Digest Auth", Value: "digest"},
				},
			},
			{
				Name:        "username",
				DisplayName: "Username",
				Type:        "string",
				Required:    false,
				Description: "Username for Basic/Digest authentication",
				ShowIf:      "authentication === 'basic' || authentication === 'digest'",
			},
			{
				Name:        "password",
				DisplayName: "Password",
				Type:        "string",
				Required:    false,
				Description: "Password for Basic/Digest authentication",
				ShowIf:      "authentication === 'basic' || authentication === 'digest'",
			},
			{
				Name:        "bearer_token",
				DisplayName: "Bearer Token",
				Type:        "string",
				Required:    false,
				Description: "Bearer token for Authorization header",
				ShowIf:      "authentication === 'bearer'",
			},
			{
				Name:        "api_key_name",
				DisplayName: "API Key Name",
				Type:        "string",
				Required:    false,
				Default:     "X-API-Key",
				Description: "Header or query parameter name for API key",
				ShowIf:      "authentication === 'api_key'",
			},
			{
				Name:        "api_key_value",
				DisplayName: "API Key Value",
				Type:        "string",
				Required:    false,
				Description: "API key value",
				ShowIf:      "authentication === 'api_key'",
			},
			{
				Name:        "api_key_location",
				DisplayName: "API Key Location",
				Type:        "options",
				Required:    false,
				Default:     "header",
				Description: "Where to send the API key",
				ShowIf:      "authentication === 'api_key'",
				Options: []wtypes.ParamOption{
					{Name: "Header", Value: "header"},
					{Name: "Query Parameter", Value: "query"},
				},
			},
			{
				Name:        "headers",
				DisplayName: "Headers",
				Type:        "json",
				Required:    false,
				Description: "Custom HTTP headers as key-value pairs",
				Placeholder: `{"Content-Type": "application/json", "Accept": "application/json"}`,
			},
			{
				Name:        "query_params",
				DisplayName: "Query Parameters",
				Type:        "json",
				Required:    false,
				Description: "URL query parameters as key-value pairs",
				Placeholder: `{"page": 1, "limit": 10}`,
			},
			{
				Name:        "body_type",
				DisplayName: "Body Type",
				Type:        "options",
				Required:    false,
				Default:     "json",
				Description: "Content type of the request body",
				ShowIf:      "method !== 'GET' && method !== 'HEAD'",
				Options: []wtypes.ParamOption{
					{Name: "JSON", Value: "json"},
					{Name: "Form Data", Value: "form"},
					{Name: "Form URL Encoded", Value: "urlencoded"},
					{Name: "Raw Text", Value: "raw"},
					{Name: "Binary", Value: "binary"},
					{Name: "None", Value: "none"},
				},
			},
			{
				Name:        "body",
				DisplayName: "Body",
				Type:        "json",
				Required:    false,
				Description: "Request body (JSON object or array)",
				ShowIf:      "body_type === 'json' && method !== 'GET' && method !== 'HEAD'",
			},
			{
				Name:        "body_raw",
				DisplayName: "Raw Body",
				Type:        "string",
				Required:    false,
				Description: "Raw request body content",
				ShowIf:      "body_type === 'raw' && method !== 'GET' && method !== 'HEAD'",
			},
			{
				Name:        "timeout",
				DisplayName: "Timeout (seconds)",
				Type:        "number",
				Required:    false,
				Default:     30,
				Description: "Request timeout in seconds (0 for no timeout)",
			},
			{
				Name:        "follow_redirects",
				DisplayName: "Follow Redirects",
				Type:        "boolean",
				Required:    false,
				Default:     true,
				Description: "Automatically follow HTTP redirects",
			},
			{
				Name:        "ignore_ssl_errors",
				DisplayName: "Ignore SSL Errors",
				Type:        "boolean",
				Required:    false,
				Default:     false,
				Description: "Ignore SSL certificate validation errors (not recommended for production)",
			},
			{
				Name:        "response_format",
				DisplayName: "Response Format",
				Type:        "options",
				Required:    false,
				Default:     "auto",
				Description: "How to parse the response",
				Options: []wtypes.ParamOption{
					{Name: "Auto Detect", Value: "auto"},
					{Name: "JSON", Value: "json"},
					{Name: "Text", Value: "text"},
					{Name: "Binary", Value: "binary"},
				},
			},
		},
		Credentials: []string{"http_basic", "http_bearer", "http_api_key", "oauth2"},
	}
}
