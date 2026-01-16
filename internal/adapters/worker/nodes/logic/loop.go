package logic

import (
	"context"

	"github.com/linkflow-ai/linkflow/internal/adapters/worker/executor"
	wtypes "github.com/linkflow-ai/linkflow/internal/adapters/worker/types"
	"github.com/linkflow-ai/linkflow/internal/shared/types"
)

type LoopNode struct{}

func NewLoopNode() *LoopNode {
	return &LoopNode{}
}

func (n *LoopNode) Execute(ctx context.Context, runtime *executor.Runtime, node map[string]interface{}) (types.JSON, error) {
	inputData := runtime.GetInputData()
	items, _ := inputData["items"].([]interface{})

	// Store loop context
	return types.JSON{
		"items":         items,
		"current_index": 0,
		"total":         len(items),
		"is_loop":       true,
	}, nil
}

func (n *LoopNode) Metadata() wtypes.NodeMetadata {
	return wtypes.NodeMetadata{
		Type:        "logic.loop",
		Name:        "Loop",
		Description: "Iterate over items",
		Category:    "logic",
		Version:     "1.0.0",
		Inputs:      []wtypes.NodePort{{Name: "main", Type: "array"}},
		Outputs:     []wtypes.NodePort{{Name: "loop", Type: "any"}, {Name: "done", Type: "any"}},
		Parameters:  []wtypes.NodeParameter{},
	}
}
