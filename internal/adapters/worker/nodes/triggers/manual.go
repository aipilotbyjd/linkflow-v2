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
		Description: "Manually trigger workflow execution",
		Category:    "trigger",
		Version:     "1.0.0",
		Inputs:      []wtypes.NodePort{},
		Outputs:     []wtypes.NodePort{{Name: "main", Type: "any"}},
		Parameters:  []wtypes.NodeParameter{},
	}
}
