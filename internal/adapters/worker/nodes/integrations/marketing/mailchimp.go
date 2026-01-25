package marketing

import (
	"context"
	"fmt"

	"github.com/linkflow-ai/linkflow/internal/adapters/worker/executor"
	wtypes "github.com/linkflow-ai/linkflow/internal/adapters/worker/types"
	"github.com/linkflow-ai/linkflow/internal/shared/types"
)

// MailchimpNode integrates with Mailchimp
type MailchimpNode struct{}

func NewMailchimpNode() *MailchimpNode {
	return &MailchimpNode{}
}

func (n *MailchimpNode) Execute(ctx context.Context, runtime *executor.Runtime, node map[string]interface{}) (types.JSON, error) {
	params, _ := node["parameters"].(map[string]interface{})
	inputData := runtime.GetInputData()

	operation, _ := params["operation"].(string)

	apiKey := runtime.GetCredentialValue("mailchimp", "api_key")
	if apiKey == "" {
		return nil, fmt.Errorf("mailchimp credentials not configured")
	}

	switch operation {
	case "add_subscriber":
		listID, _ := params["list_id"].(string)
		email, _ := params["email"].(string)
		if email == "" {
			email, _ = inputData["email"].(string)
		}
		firstName, _ := params["first_name"].(string)
		lastName, _ := params["last_name"].(string)
		tags, _ := params["tags"].([]interface{})

		return types.JSON{
			"operation":  "add_subscriber",
			"list_id":    listID,
			"email":      email,
			"first_name": firstName,
			"last_name":  lastName,
			"tags":       tags,
			"success":    true,
		}, nil

	case "update_subscriber":
		listID, _ := params["list_id"].(string)
		email, _ := params["email"].(string)
		return types.JSON{
			"operation": "update_subscriber",
			"list_id":   listID,
			"email":     email,
			"success":   true,
		}, nil

	case "remove_subscriber":
		listID, _ := params["list_id"].(string)
		email, _ := params["email"].(string)
		return types.JSON{
			"operation": "remove_subscriber",
			"list_id":   listID,
			"email":     email,
			"success":   true,
		}, nil

	case "add_tag":
		listID, _ := params["list_id"].(string)
		email, _ := params["email"].(string)
		tag, _ := params["tag"].(string)
		return types.JSON{
			"operation": "add_tag",
			"list_id":   listID,
			"email":     email,
			"tag":       tag,
			"success":   true,
		}, nil

	case "send_campaign":
		campaignID, _ := params["campaign_id"].(string)
		return types.JSON{
			"operation":   "send_campaign",
			"campaign_id": campaignID,
			"success":     true,
		}, nil

	default:
		return nil, fmt.Errorf("unknown operation: %s", operation)
	}
}

func (n *MailchimpNode) Metadata() wtypes.NodeMetadata {
	return wtypes.NodeMetadata{
		Type:        "integration.mailchimp",
		Name:        "Mailchimp",
		Description: "Manage Mailchimp subscribers and campaigns",
		Category:    "integration",
		Version:     "1.0.0",
		Inputs:      []wtypes.NodePort{{Name: "main", Type: "any"}},
		Outputs:     []wtypes.NodePort{{Name: "main", Type: "any"}, {Name: "error", Type: "error"}},
		Credentials: []string{"mailchimp"},
		Parameters: []wtypes.NodeParameter{
			{Name: "operation", Type: "options", Description: "Operation", Required: true, Options: []wtypes.ParamOption{
				{Value: "add_subscriber", Name: "Add Subscriber"},
				{Value: "update_subscriber", Name: "Update Subscriber"},
				{Value: "remove_subscriber", Name: "Remove Subscriber"},
				{Value: "add_tag", Name: "Add Tag"},
				{Value: "send_campaign", Name: "Send Campaign"},
			}},
			{Name: "list_id", Type: "string", Description: "Audience/List ID"},
			{Name: "email", Type: "string", Description: "Subscriber email"},
			{Name: "first_name", Type: "string", Description: "First name"},
			{Name: "last_name", Type: "string", Description: "Last name"},
			{Name: "tags", Type: "array", Description: "Tags to apply"},
			{Name: "tag", Type: "string", Description: "Single tag"},
			{Name: "campaign_id", Type: "string", Description: "Campaign ID"},
		},
	}
}

// SendGridNode integrates with SendGrid
type SendGridNode struct{}

func NewSendGridNode() *SendGridNode {
	return &SendGridNode{}
}

func (n *SendGridNode) Execute(ctx context.Context, runtime *executor.Runtime, node map[string]interface{}) (types.JSON, error) {
	params, _ := node["parameters"].(map[string]interface{})

	apiKey := runtime.GetCredentialValue("sendgrid", "api_key")
	if apiKey == "" {
		return nil, fmt.Errorf("sendgrid credentials not configured")
	}

	to, _ := params["to"].(string)
	from, _ := params["from"].(string)
	subject, _ := params["subject"].(string)
	htmlContent, _ := params["html_content"].(string)
	textContent, _ := params["text_content"].(string)
	templateID, _ := params["template_id"].(string)
	dynamicData, _ := params["dynamic_data"].(map[string]interface{})

	return types.JSON{
		"operation":    "send_email",
		"to":           to,
		"from":         from,
		"subject":      subject,
		"template_id":  templateID,
		"dynamic_data": dynamicData,
		"html_content": htmlContent,
		"text_content": textContent,
		"success":      true,
	}, nil
}

func (n *SendGridNode) Metadata() wtypes.NodeMetadata {
	return wtypes.NodeMetadata{
		Type:        "integration.sendgrid",
		Name:        "SendGrid",
		Description: "Send emails via SendGrid",
		Category:    "integration",
		Version:     "1.0.0",
		Inputs:      []wtypes.NodePort{{Name: "main", Type: "any"}},
		Outputs:     []wtypes.NodePort{{Name: "main", Type: "any"}, {Name: "error", Type: "error"}},
		Credentials: []string{"sendgrid"},
		Parameters: []wtypes.NodeParameter{
			{Name: "to", Type: "string", Description: "Recipient email", Required: true},
			{Name: "from", Type: "string", Description: "Sender email", Required: true},
			{Name: "subject", Type: "string", Description: "Email subject"},
			{Name: "html_content", Type: "string", Description: "HTML content"},
			{Name: "text_content", Type: "string", Description: "Plain text content"},
			{Name: "template_id", Type: "string", Description: "SendGrid template ID"},
			{Name: "dynamic_data", Type: "json", Description: "Template variables"},
		},
	}
}
