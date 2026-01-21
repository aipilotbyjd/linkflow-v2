package triggers

import (
	"context"

	"github.com/linkflow-ai/linkflow/internal/adapters/worker/executor"
	wtypes "github.com/linkflow-ai/linkflow/internal/adapters/worker/types"
	"github.com/linkflow-ai/linkflow/internal/shared/types"
)

type ManualTrigger struct{}

func NewManualTrigger() *ManualTrigger {
	return &ManualTrigger{}
}

func (t *ManualTrigger) Execute(ctx context.Context, runtime *executor.Runtime, node map[string]interface{}) (types.JSON, error) {
	return runtime.GetInputData(), nil
}

func (t *ManualTrigger) Metadata() wtypes.NodeMetadata {
	return wtypes.NodeMetadata{
		Type:        "trigger.manual",
		Name:        "Manual Trigger",
		Description: "Start workflow execution manually via UI or API with optional input data",
		Category:    "trigger",
		Version:     "1.0.0",
		Icon:        "play-circle",
		Color:       "#22C55E",
		Inputs:      []wtypes.NodePort{},
		Outputs: []wtypes.NodePort{
			{Name: "main", Type: "object", Description: "Input data provided when triggering the workflow"},
		},
		Parameters: []wtypes.NodeParameter{
			{
				Name:        "input_schema",
				DisplayName: "Input Schema",
				Type:        "json",
				Required:    false,
				Description: "JSON Schema defining expected input data structure (for validation and UI generation)",
				Placeholder: `{"type": "object", "properties": {"name": {"type": "string"}}}`,
			},
			{
				Name:        "default_data",
				DisplayName: "Default Data",
				Type:        "json",
				Required:    false,
				Description: "Default input data when no data is provided",
				Default:     map[string]interface{}{},
			},
			{
				Name:        "require_confirmation",
				DisplayName: "Require Confirmation",
				Type:        "boolean",
				Required:    false,
				Default:     false,
				Description: "Show confirmation dialog before execution",
			},
		},
	}
}
