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
		Beta:        false,
		Tags:        []string{"http", "api", "integration", "trigger"},
		Examples: []wtypes.NodeExample{
			{
				Name:        "Basic Webhook",
				Description: "Simple webhook that accepts POST requests",
				Parameters: map[string]interface{}{
					"http_method": "POST",
					"path":        "/my-webhook",
				},
			},
		},
		Links: []wtypes.NodeLink{
			{Name: "Webhook Documentation", URL: "https://docs.linkflow.ai/nodes/webhook-trigger", Type: "docs"},
		},
		Inputs: []wtypes.NodePort{},
		Outputs: []wtypes.NodePort{
			{
				Name:        "main", 
				Type:        "object", 
				Description: "Webhook request data including headers, body, and query parameters",
				Example: map[string]interface{}{
					"headers": map[string]string{"content-type": "application/json"},
					"body":    map[string]interface{}{"message": "Hello World"},
					"query":   map[string]string{"param": "value"},
				},
				Schema: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"headers": map[string]interface{}{"type": "object"},
						"body":    map[string]interface{}{"type": "object"},
						"query":   map[string]interface{}{"type": "object"},
					},
				},
			},
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
				LongDescription: "Specify a custom endpoint path for your webhook. If left empty, a unique path will be auto-generated. The full webhook URL will be: https://your-domain.com/webhooks/[path]",
				Placeholder: "/my-webhook",
				Validation: &wtypes.Validation{
					Pattern: `^\/[a-zA-Z0-9\-_\/]*$`,
					MinLength: wtypes.IntPtr(1),
					MaxLength: wtypes.IntPtr(100),
				},
				Tooltip: "Enter a custom path or leave empty for auto-generation",
				Width:   "full",
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
					{
						Name:        "None", 
						Value:       "none", 
						Description: "No authentication required",
						Icon:        "unlock",
						Color:       "#6B7280",
					},
					{
						Name:        "Basic Auth", 
						Value:       "basic", 
						Description: "HTTP Basic Authentication",
						Icon:        "lock",
						Color:       "#F59E0B",
						Badge:       "Legacy",
						BadgeColor:  "#FBBF24",
					},
					{
						Name:        "Header Auth", 
						Value:       "header", 
						Description: "Custom header authentication",
						Icon:        "key",
						Color:       "#10B981",
						Badge:       "Recommended",
						BadgeColor:  "#34D399",
					},
					{
						Name:        "JWT", 
						Value:       "jwt", 
						Description: "JSON Web Token authentication",
						Icon:        "shield",
						Color:       "#3B82F6",
						SearchTerms: []string{"bearer", "token", "oauth"},
					},
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
				LongDescription: "The secret token or API key that must be provided in the specified header for authentication. Store sensitive values in credentials for better security.",
				ShowIf:      "authentication === 'header'",
				Sensitive:   true,
				Warning:     "Consider using Credentials instead of storing secrets directly in parameters",
				Group:       "Authentication",
				Order:       1,
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
				LongDescription:    "The HTTP status code to return when responding immediately to the webhook request. Common codes: 200 (OK), 201 (Created), 202 (Accepted), 400 (Bad Request), 401 (Unauthorized), 500 (Internal Server Error)",
				ShowIf:             "response_mode === 'immediately'",
				ExpressionDisabled: true,
				Validation: &wtypes.Validation{
					Min:          wtypes.Float64Ptr(100),
					Max:          wtypes.Float64Ptr(599),
					ExclusiveMin: true,
					ExclusiveMax: true,
					Enum:         []interface{}{float64(200), float64(201), float64(202), float64(400), float64(401), float64(500)},
				},
				Tooltip: "Select HTTP status code for immediate response",
				Width:   "half",
				Step:    1,
			},
		},
	}
}
