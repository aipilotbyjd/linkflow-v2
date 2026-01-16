package email

import (
	"context"
	"fmt"

	"github.com/linkflow-ai/linkflow/internal/adapters/worker/executor"
	wtypes "github.com/linkflow-ai/linkflow/internal/adapters/worker/types"
	"github.com/linkflow-ai/linkflow/internal/shared/types"
)

type SendEmailNode struct{}

func NewSendEmailNode() *SendEmailNode {
	return &SendEmailNode{}
}

func (n *SendEmailNode) Execute(ctx context.Context, runtime *executor.Runtime, node map[string]interface{}) (types.JSON, error) {
	params, _ := node["parameters"].(map[string]interface{})

	to, _ := params["to"].(string)
	subject, _ := params["subject"].(string)
	_, _ = params["body"].(string)
	from, _ := params["from"].(string)

	if to == "" {
		return nil, fmt.Errorf("email recipient (to) is required")
	}
	if subject == "" {
		return nil, fmt.Errorf("email subject is required")
	}

	// TODO: Integrate with actual email service
	return types.JSON{
		"status":  "sent",
		"to":      to,
		"from":    from,
		"subject": subject,
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
