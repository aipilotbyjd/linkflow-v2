package actions

import (
	"context"

	"github.com/linkflow-ai/linkflow/internal/adapters/worker/executor"
	wtypes "github.com/linkflow-ai/linkflow/internal/adapters/worker/types"
	"github.com/linkflow-ai/linkflow/internal/shared/types"
)

// WebhookResponseNode customizes the webhook response
type WebhookResponseNode struct{}

func NewWebhookResponseNode() *WebhookResponseNode {
	return &WebhookResponseNode{}
}

func (n *WebhookResponseNode) Execute(ctx context.Context, runtime *executor.Runtime, node map[string]interface{}) (types.JSON, error) {
	params, _ := node["parameters"].(map[string]interface{})
	inputData := runtime.GetInputData()

	// Get status code
	statusCode := 200
	if code, ok := params["status_code"].(float64); ok {
		statusCode = int(code)
	}

	// Get headers
	headers := make(map[string]string)
	if h, ok := params["headers"].(map[string]interface{}); ok {
		for k, v := range h {
			if str, ok := v.(string); ok {
				headers[k] = str
			}
		}
	}

	// Get content type
	contentType, _ := params["content_type"].(string)
	if contentType == "" {
		contentType = "application/json"
	}
	headers["Content-Type"] = contentType

	// Get response body
	var body interface{}
	if b, ok := params["body"]; ok && b != nil {
		body = b
	} else if template, ok := params["body_template"].(string); ok && template != "" {
		// Use template - in production would use template engine
		body = map[string]interface{}{
			"template": template,
			"data":     inputData,
		}
	} else {
		// Use input data as body
		body = inputData
	}

	// Return webhook response configuration
	return types.JSON{
		"_webhook_response": true,
		"status_code":       statusCode,
		"headers":           headers,
		"content_type":      contentType,
		"body":              body,
	}, nil
}

func (n *WebhookResponseNode) Metadata() wtypes.NodeMetadata {
	return wtypes.NodeMetadata{
		Type:        "action.webhook_response",
		Name:        "Webhook Response",
		Description: "Customize the HTTP response for webhook triggers",
		Category:    "action",
		Version:     "1.0.0",
		Inputs:      []wtypes.NodePort{{Name: "main", Type: "any"}},
		Outputs:     []wtypes.NodePort{{Name: "main", Type: "any"}},
		Parameters: []wtypes.NodeParameter{
			{Name: "status_code", Type: "number", Description: "HTTP status code", Default: 200},
			{Name: "content_type", Type: "select", Description: "Response content type", Default: "application/json", Options: []wtypes.ParamOption{
				{Value: "application/json", Name: "JSON"},
				{Value: "text/plain", Name: "Plain Text"},
				{Value: "text/html", Name: "HTML"},
				{Value: "application/xml", Name: "XML"},
			}},
			{Name: "headers", Type: "json", Description: "Custom response headers"},
			{Name: "body", Type: "json", Description: "Response body (overrides input)"},
			{Name: "body_template", Type: "string", Description: "Response body template"},
		},
	}
}

// RespondToWebhookNode is an alias for immediate webhook response
type RespondToWebhookNode struct{}

func NewRespondToWebhookNode() *RespondToWebhookNode {
	return &RespondToWebhookNode{}
}

func (n *RespondToWebhookNode) Execute(ctx context.Context, runtime *executor.Runtime, node map[string]interface{}) (types.JSON, error) {
	params, _ := node["parameters"].(map[string]interface{})
	inputData := runtime.GetInputData()

	respondWith, _ := params["respond_with"].(string)

	var body interface{}
	switch respondWith {
	case "first_input":
		body = inputData
	case "last_node":
		body = runtime.GetOutputData()
	case "custom":
		body = params["body"]
	case "empty":
		body = nil
	default:
		body = map[string]string{"status": "ok"}
	}

	statusCode := 200
	if code, ok := params["status_code"].(float64); ok {
		statusCode = int(code)
	}

	return types.JSON{
		"_webhook_response":    true,
		"_respond_immediately": true,
		"status_code":          statusCode,
		"body":                 body,
		"headers":              params["headers"],
	}, nil
}

func (n *RespondToWebhookNode) Metadata() wtypes.NodeMetadata {
	return wtypes.NodeMetadata{
		Type:        "action.respond_to_webhook",
		Name:        "Respond to Webhook",
		Description: "Send immediate response to webhook caller",
		Category:    "action",
		Version:     "1.0.0",
		Inputs:      []wtypes.NodePort{{Name: "main", Type: "any"}},
		Outputs:     []wtypes.NodePort{{Name: "main", Type: "any"}},
		Parameters: []wtypes.NodeParameter{
			{Name: "respond_with", Type: "select", Description: "What to respond with", Default: "first_input", Options: []wtypes.ParamOption{
				{Value: "first_input", Name: "First Input Data"},
				{Value: "last_node", Name: "Last Node Output"},
				{Value: "custom", Name: "Custom Body"},
				{Value: "empty", Name: "Empty Response"},
			}},
			{Name: "status_code", Type: "number", Description: "HTTP status code", Default: 200},
			{Name: "body", Type: "json", Description: "Custom response body"},
			{Name: "headers", Type: "json", Description: "Custom headers"},
		},
	}
}
