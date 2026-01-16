package logic

import (
	"context"

	"github.com/linkflow-ai/linkflow/internal/adapters/worker/executor"
	wtypes "github.com/linkflow-ai/linkflow/internal/adapters/worker/types"
	"github.com/linkflow-ai/linkflow/internal/shared/types"
)

type MergeNode struct{}

func NewMergeNode() *MergeNode {
	return &MergeNode{}
}

func (n *MergeNode) Execute(ctx context.Context, runtime *executor.Runtime, node map[string]interface{}) (types.JSON, error) {
	result := types.JSON{
		"merged": true,
	}
	return result, nil
}

func (n *MergeNode) Metadata() wtypes.NodeMetadata {
	return wtypes.NodeMetadata{
		Type:        "logic.merge",
		Name:        "Merge",
		Description: "Merge multiple inputs into one",
		Category:    "logic",
		Version:     "1.0.0",
		Inputs:      []wtypes.NodePort{{Name: "input1", Type: "any"}, {Name: "input2", Type: "any"}},
		Outputs:     []wtypes.NodePort{{Name: "main", Type: "any"}},
		Parameters:  []wtypes.NodeParameter{},
	}
}
