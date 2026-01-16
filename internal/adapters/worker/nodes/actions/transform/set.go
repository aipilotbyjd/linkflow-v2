package transform

import (
	"context"

	"github.com/linkflow-ai/linkflow/internal/adapters/worker/executor"
	wtypes "github.com/linkflow-ai/linkflow/internal/adapters/worker/types"
	"github.com/linkflow-ai/linkflow/internal/shared/types"
)

type SetNode struct{}

func NewSetNode() *SetNode {
	return &SetNode{}
}

func (n *SetNode) Execute(ctx context.Context, runtime *executor.Runtime, node map[string]interface{}) (types.JSON, error) {
	params, _ := node["parameters"].(map[string]interface{})

	values, _ := params["values"].(map[string]interface{})

	result := make(types.JSON)
	for k, v := range values {
		result[k] = v
	}

	return result, nil
}

func (n *SetNode) Metadata() wtypes.NodeMetadata {
	return wtypes.NodeMetadata{
		Type:        "action.set",
		Name:        "Set",
		Description: "Set variable values",
		Category:    "action",
		Version:     "1.0.0",
		Inputs:      []wtypes.NodePort{{Name: "main", Type: "any"}},
		Outputs:     []wtypes.NodePort{{Name: "main", Type: "object"}},
		Parameters: []wtypes.NodeParameter{
			{Name: "values", DisplayName: "Values", Type: "json", Required: true},
		},
	}
}
