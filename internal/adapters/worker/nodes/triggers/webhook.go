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
		Description: "Trigger workflow via HTTP webhook",
		Category:    "trigger",
		Version:     "1.0.0",
		Inputs:      []wtypes.NodePort{},
		Outputs:     []wtypes.NodePort{{Name: "main", Type: "any"}},
		Parameters:  []wtypes.NodeParameter{},
	}
}
