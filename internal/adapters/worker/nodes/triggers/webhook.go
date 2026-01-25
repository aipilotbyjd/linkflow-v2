package triggers

import (
	"context"

	"github.com/linkflow-ai/linkflow/internal/adapters/worker/executor"
	wtypes "github.com/linkflow-ai/linkflow/internal/adapters/worker/types"
	"github.com/linkflow-ai/linkflow/internal/shared/types"
)

type WebhookTrigger struct{}

func NewWebhookTrigger() *WebhookTrigger {
	return &WebhookTrigger{}
}

func (t *WebhookTrigger) Execute(ctx context.Context, runtime *executor.Runtime, node map[string]interface{}) (types.JSON, error) {
	return runtime.GetInputData(), nil
}

func (t *WebhookTrigger) Metadata() wtypes.NodeMetadata {
	return wtypes.NodeMetadata{
		Type:        "trigger.webhook",
		Name:        "Webhook Trigger",
		Description: "Trigger workflow when an HTTP request is received at the webhook URL",
		Category:    "trigger",
		Version:     "1.0.0",
		Icon:        "webhook",
		Color:       "#6366F1",
		Inputs:      []wtypes.NodePort{},
		Outputs: []wtypes.NodePort{
			{Name: "main", Type: "object", Description: "Webhook request data including headers, body, and query parameters"},
		},
		Parameters: []wtypes.NodeParameter{
			{
				Name:               "http_method",
				DisplayName:        "HTTP Method",
				Type:               "options",
				Required:           false,
				Default:            "POST",
				Description:        "Allowed HTTP method for the webhook",
				ExpressionDisabled: true,
				Options: []wtypes.ParamOption{
					{Name: "GET", Value: "GET", Description: "Accept GET requests"},
					{Name: "POST", Value: "POST", Description: "Accept POST requests"},
					{Name: "PUT", Value: "PUT", Description: "Accept PUT requests"},
					{Name: "DELETE", Value: "DELETE", Description: "Accept DELETE requests"},
					{Name: "PATCH", Value: "PATCH", Description: "Accept PATCH requests"},
					{Name: "All Methods", Value: "ALL", Description: "Accept any HTTP method"},
				},
			},
			{
				Name:        "path",
				DisplayName: "Webhook Path",
				Type:        "string",
				Required:    false,
				Description: "Custom path for the webhook URL (auto-generated if empty)",
				Placeholder: "/my-webhook",
			},
			{
				Name:               "authentication",
				DisplayName:        "Authentication",
				Type:               "options",
				Required:           false,
				Default:            "none",
				Description:        "Authentication method for incoming requests",
				ExpressionDisabled: true,
				Options: []wtypes.ParamOption{
					{Name: "None", Value: "none", Description: "No authentication required"},
					{Name: "Basic Auth", Value: "basic", Description: "HTTP Basic Authentication"},
					{Name: "Header Auth", Value: "header", Description: "Custom header authentication"},
					{Name: "JWT", Value: "jwt", Description: "JSON Web Token authentication"},
				},
			},
			{
				Name:        "auth_header_name",
				DisplayName: "Auth Header Name",
				Type:        "string",
				Required:    false,
				Default:     "X-API-Key",
				Description: "Header name for authentication",
				ShowIf:      "authentication === 'header'",
			},
			{
				Name:        "auth_header_value",
				DisplayName: "Auth Header Value",
				Type:        "string",
				Required:    false,
				Description: "Expected value for the auth header",
				ShowIf:      "authentication === 'header'",
			},
			{
				Name:               "response_mode",
				DisplayName:        "Response Mode",
				Type:               "options",
				Required:           false,
				Default:            "last_node",
				Description:        "How to respond to the webhook request",
				ExpressionDisabled: true,
				Options: []wtypes.ParamOption{
					{Name: "Respond Immediately", Value: "immediately", Description: "Return 200 OK immediately"},
					{Name: "Wait for Workflow", Value: "last_node", Description: "Wait and return last node output"},
					{Name: "Custom Response Node", Value: "response_node", Description: "Use Webhook Response node"},
				},
			},
			{
				Name:               "response_code",
				DisplayName:        "Response Code",
				Type:               "number",
				Required:           false,
				Default:            200,
				Description:        "HTTP status code for immediate response",
				ShowIf:             "response_mode === 'immediately'",
				ExpressionDisabled: true,
				Validation: &wtypes.Validation{
					Min: wtypes.Float64Ptr(100),
					Max: wtypes.Float64Ptr(599),
				},
			},
		},
	}
}
