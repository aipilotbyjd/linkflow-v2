package logic

import (
	"context"

	"github.com/linkflow-ai/linkflow/internal/adapters/worker/executor"
	wtypes "github.com/linkflow-ai/linkflow/internal/adapters/worker/types"
	"github.com/linkflow-ai/linkflow/internal/shared/types"
)

type IfNode struct{}

func NewIfNode() *IfNode {
	return &IfNode{}
}

func (n *IfNode) Execute(ctx context.Context, runtime *executor.Runtime, node map[string]interface{}) (types.JSON, error) {
	params, _ := node["parameters"].(map[string]interface{})
	condition, _ := params["condition"].(bool)

	result := types.JSON{
		"condition": condition,
		"branch":    "false",
	}

	if condition {
		result["branch"] = "true"
	}

	return result, nil
}

func (n *IfNode) Metadata() wtypes.NodeMetadata {
	return wtypes.NodeMetadata{
		Type:        "logic.if",
		Name:        "If",
		Description: "Conditional branching based on a condition",
		Category:    "logic",
		Version:     "1.0.0",
		Inputs:      []wtypes.NodePort{{Name: "main", Type: "any"}},
		Outputs:     []wtypes.NodePort{{Name: "true", Type: "any"}, {Name: "false", Type: "any"}},
		Parameters: []wtypes.NodeParameter{
			{Name: "condition", DisplayName: "Condition", Type: "boolean", Required: true},
		},
	}
}
