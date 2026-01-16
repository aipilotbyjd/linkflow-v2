package triggers

import (
	"context"

	"github.com/linkflow-ai/linkflow/internal/adapters/worker/executor"
	wtypes "github.com/linkflow-ai/linkflow/internal/adapters/worker/types"
	"github.com/linkflow-ai/linkflow/internal/shared/types"
)

type ScheduleTrigger struct{}

func NewScheduleTrigger() *ScheduleTrigger {
	return &ScheduleTrigger{}
}

func (t *ScheduleTrigger) Execute(ctx context.Context, runtime *executor.Runtime, node map[string]interface{}) (types.JSON, error) {
	return runtime.GetInputData(), nil
}

func (t *ScheduleTrigger) Metadata() wtypes.NodeMetadata {
	return wtypes.NodeMetadata{
		Type:        "trigger.schedule",
		Name:        "Schedule Trigger",
		Description: "Trigger workflow on a schedule (cron)",
		Category:    "trigger",
		Version:     "1.0.0",
		Inputs:      []wtypes.NodePort{},
		Outputs:     []wtypes.NodePort{{Name: "main", Type: "any"}},
		Parameters:  []wtypes.NodeParameter{},
	}
}
