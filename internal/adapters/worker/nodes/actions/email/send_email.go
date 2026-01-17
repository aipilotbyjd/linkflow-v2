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
		Description: "Send an email",
		Category:    "action",
		Version:     "1.0.0",
		Inputs:      []wtypes.NodePort{{Name: "main", Type: "any"}},
		Outputs:     []wtypes.NodePort{{Name: "main", Type: "object"}},
		Parameters: []wtypes.NodeParameter{
			{Name: "to", DisplayName: "To", Type: "string", Required: true},
			{Name: "subject", DisplayName: "Subject", Type: "string", Required: true},
			{Name: "body", DisplayName: "Body", Type: "string", Required: true},
			{Name: "from", DisplayName: "From", Type: "string", Required: false},
		},
	}
}
