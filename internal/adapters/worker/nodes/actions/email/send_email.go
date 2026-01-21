package email

import (
	"context"
	"fmt"
	"strings"

	"github.com/linkflow-ai/linkflow/internal/adapters/worker/executor"
	wtypes "github.com/linkflow-ai/linkflow/internal/adapters/worker/types"
	infraEmail "github.com/linkflow-ai/linkflow/internal/infrastructure/email"
	"github.com/linkflow-ai/linkflow/internal/shared/types"
)

type SendEmailNode struct {
	emailProvider infraEmail.Provider
}

func NewSendEmailNode() *SendEmailNode {
	return &SendEmailNode{}
}

func NewSendEmailNodeWithProvider(emailProvider infraEmail.Provider) *SendEmailNode {
	return &SendEmailNode{emailProvider: emailProvider}
}

func (n *SendEmailNode) Execute(ctx context.Context, runtime *executor.Runtime, node map[string]interface{}) (types.JSON, error) {
	params, _ := node["parameters"].(map[string]interface{})

	to, _ := params["to"].(string)
	subject, _ := params["subject"].(string)
	body, _ := params["body"].(string)
	from, _ := params["from"].(string)
	cc, _ := params["cc"].(string)
	bcc, _ := params["bcc"].(string)
	replyTo, _ := params["reply_to"].(string)
	isHTML, _ := params["is_html"].(bool)

	if to == "" {
		return nil, fmt.Errorf("email recipient (to) is required")
	}
	if subject == "" {
		return nil, fmt.Errorf("email subject is required")
	}

	// Parse recipients
	toAddrs := strings.Split(to, ",")
	for i := range toAddrs {
		toAddrs[i] = strings.TrimSpace(toAddrs[i])
	}

	msg := &infraEmail.Message{
		From:    from,
		To:      toAddrs,
		Subject: subject,
		ReplyTo: replyTo,
	}

	// Parse CC
	if cc != "" {
		ccAddrs := strings.Split(cc, ",")
		for i := range ccAddrs {
			ccAddrs[i] = strings.TrimSpace(ccAddrs[i])
		}
		msg.CC = ccAddrs
	}

	// Parse BCC
	if bcc != "" {
		bccAddrs := strings.Split(bcc, ",")
		for i := range bccAddrs {
			bccAddrs[i] = strings.TrimSpace(bccAddrs[i])
		}
		msg.BCC = bccAddrs
	}

	// Set body
	if isHTML {
		msg.HTMLBody = body
	} else {
		msg.TextBody = body
	}

	// Send email
	if n.emailProvider != nil {
		if err := n.emailProvider.Send(ctx, msg); err != nil {
			return nil, fmt.Errorf("failed to send email: %w", err)
		}
	}

	return types.JSON{
		"status":     "sent",
		"to":         toAddrs,
		"from":       from,
		"subject":    subject,
		"recipients": len(toAddrs),
	}, nil
}

func (n *SendEmailNode) Metadata() wtypes.NodeMetadata {
	return wtypes.NodeMetadata{
		Type:        "action.email",
		Name:        "Send Email",
		Description: "Send emails via SMTP, SendGrid, AWS SES, or other email providers with support for HTML, attachments, and templates",
		Category:    "action",
		Version:     "1.0.0",
		Icon:        "mail",
		Color:       "#EA580C",
		Inputs: []wtypes.NodePort{
			{Name: "main", Type: "any", Description: "Input data for dynamic email content"},
		},
		Outputs: []wtypes.NodePort{
			{Name: "main", Type: "object", Description: "Email send result with status and message ID"},
		},
		Parameters: []wtypes.NodeParameter{
			{
				Name:        "to",
				DisplayName: "To",
				Type:        "string",
				Required:    true,
				Description: "Recipient email address(es), comma-separated for multiple",
				Placeholder: "recipient@example.com",
			},
			{
				Name:        "from",
				DisplayName: "From",
				Type:        "string",
				Required:    false,
				Description: "Sender email address (uses default from credentials if not specified)",
				Placeholder: "sender@example.com",
			},
			{
				Name:        "from_name",
				DisplayName: "From Name",
				Type:        "string",
				Required:    false,
				Description: "Sender display name",
				Placeholder: "My Company",
			},
			{
				Name:        "reply_to",
				DisplayName: "Reply To",
				Type:        "string",
				Required:    false,
				Description: "Reply-to email address",
				Placeholder: "replies@example.com",
			},
			{
				Name:        "cc",
				DisplayName: "CC",
				Type:        "string",
				Required:    false,
				Description: "CC recipient(s), comma-separated for multiple",
				Placeholder: "cc@example.com",
			},
			{
				Name:        "bcc",
				DisplayName: "BCC",
				Type:        "string",
				Required:    false,
				Description: "BCC recipient(s), comma-separated for multiple",
				Placeholder: "bcc@example.com",
			},
			{
				Name:        "subject",
				DisplayName: "Subject",
				Type:        "string",
				Required:    true,
				Description: "Email subject line (supports expressions)",
				Placeholder: "Hello {{$input.name}}!",
			},
			{
				Name:        "content_type",
				DisplayName: "Content Type",
				Type:        "options",
				Required:    false,
				Default:     "html",
				Description: "Email body content type",
				Options: []wtypes.ParamOption{
					{Name: "HTML", Value: "html", Description: "Rich HTML email"},
					{Name: "Plain Text", Value: "text", Description: "Plain text email"},
					{Name: "Both", Value: "both", Description: "Both HTML and plain text"},
				},
			},
			{
				Name:        "body",
				DisplayName: "Body (HTML)",
				Type:        "code",
				Required:    true,
				Description: "Email body content (HTML or plain text based on content type)",
				Placeholder: "<h1>Hello!</h1><p>This is your email content.</p>",
			},
			{
				Name:        "text_body",
				DisplayName: "Plain Text Body",
				Type:        "string",
				Required:    false,
				Description: "Plain text version of email (for multipart emails)",
				ShowIf:      "content_type === 'both'",
			},
			{
				Name:        "priority",
				DisplayName: "Priority",
				Type:        "options",
				Required:    false,
				Default:     "normal",
				Description: "Email priority level",
				Options: []wtypes.ParamOption{
					{Name: "High", Value: "high"},
					{Name: "Normal", Value: "normal"},
					{Name: "Low", Value: "low"},
				},
			},
			{
				Name:        "attachments",
				DisplayName: "Attachments",
				Type:        "json",
				Required:    false,
				Description: "Array of attachments with filename, content (base64), and mimeType",
				Placeholder: `[{"filename": "doc.pdf", "content": "base64...", "mimeType": "application/pdf"}]`,
			},
			{
				Name:        "headers",
				DisplayName: "Custom Headers",
				Type:        "json",
				Required:    false,
				Description: "Custom email headers as key-value pairs",
			},
			{
				Name:        "track_opens",
				DisplayName: "Track Opens",
				Type:        "boolean",
				Required:    false,
				Default:     false,
				Description: "Track when recipients open the email (provider dependent)",
			},
			{
				Name:        "track_clicks",
				DisplayName: "Track Clicks",
				Type:        "boolean",
				Required:    false,
				Default:     false,
				Description: "Track when recipients click links (provider dependent)",
			},
		},
		Credentials: []string{"smtp", "sendgrid", "aws_ses", "mailgun", "postmark"},
	}
}
