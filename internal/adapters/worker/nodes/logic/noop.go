package logic

import (
	"context"

	"github.com/linkflow-ai/linkflow/internal/adapters/worker/executor"
	wtypes "github.com/linkflow-ai/linkflow/internal/adapters/worker/types"
	"github.com/linkflow-ai/linkflow/internal/shared/types"
)

type NoopNode struct{}

func NewNoopNode() *NoopNode {
	return &NoopNode{}
}

func (n *NoopNode) Execute(ctx context.Context, runtime *executor.Runtime, node map[string]interface{}) (types.JSON, error) {
	// Pass through - return input as output
	return runtime.GetInputData(), nil
}

func (n *NoopNode) Metadata() wtypes.NodeMetadata {
	return wtypes.NodeMetadata{
		Type:        "logic.noop",
		Name:        "No Operation",
		Description: "Pass-through node that does nothing",
		Category:    "logic",
		Version:     "1.0.0",
		Inputs:      []wtypes.NodePort{{Name: "main", Type: "any"}},
		Outputs:     []wtypes.NodePort{{Name: "main", Type: "any"}},
		Parameters:  []wtypes.NodeParameter{},
	}
}
