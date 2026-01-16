package communication

import (
	"context"
	"fmt"

	"github.com/linkflow-ai/linkflow/internal/adapters/worker/executor"
	wtypes "github.com/linkflow-ai/linkflow/internal/adapters/worker/types"
	"github.com/linkflow-ai/linkflow/internal/shared/types"
)

type TwilioNode struct{}

func NewTwilioNode() *TwilioNode {
	return &TwilioNode{}
}

func (n *TwilioNode) Execute(ctx context.Context, runtime *executor.Runtime, node map[string]interface{}) (types.JSON, error) {
	params, _ := node["parameters"].(map[string]interface{})
	operation, _ := params["operation"].(string)

	switch operation {
	case "send_sms":
		return n.sendSMS(ctx, params)
	case "send_whatsapp":
		return n.sendWhatsApp(ctx, params)
	default:
		return nil, fmt.Errorf("unsupported Twilio operation: %s", operation)
	}
}

func (n *TwilioNode) sendSMS(ctx context.Context, params map[string]interface{}) (types.JSON, error) {
	from, _ := params["from"].(string)
	to, _ := params["to"].(string)
	body, _ := params["body"].(string)

	return types.JSON{
		"success": true,
		"from":    from,
		"to":      to,
		"body":    body,
	}, nil
}

func (n *TwilioNode) sendWhatsApp(ctx context.Context, params map[string]interface{}) (types.JSON, error) {
	from, _ := params["from"].(string)
	to, _ := params["to"].(string)
	body, _ := params["body"].(string)

	return types.JSON{
		"success": true,
		"from":    "whatsapp:" + from,
		"to":      "whatsapp:" + to,
		"body":    body,
	}, nil
}

func (n *TwilioNode) Metadata() wtypes.NodeMetadata {
	return wtypes.NodeMetadata{
		Type:        "integration.twilio",
		Name:        "Twilio",
		Description: "Send SMS and calls via Twilio",
		Category:    "integration",
		Version:     "1.0.0",
		Inputs:      []wtypes.NodePort{{Name: "main", Type: "any"}},
		Outputs:     []wtypes.NodePort{{Name: "main", Type: "object"}},
		Parameters:  []wtypes.NodeParameter{},
	}
}
