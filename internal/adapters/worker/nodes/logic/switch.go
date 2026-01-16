package logic

import (
	"context"

	"github.com/linkflow-ai/linkflow/internal/adapters/worker/executor"
	wtypes "github.com/linkflow-ai/linkflow/internal/adapters/worker/types"
	"github.com/linkflow-ai/linkflow/internal/shared/types"
)

type SwitchNode struct{}

func NewSwitchNode() *SwitchNode {
	return &SwitchNode{}
}

func (n *SwitchNode) Execute(ctx context.Context, runtime *executor.Runtime, node map[string]interface{}) (types.JSON, error) {
	params, _ := node["parameters"].(map[string]interface{})
	value, _ := params["value"].(string)
	cases, _ := params["cases"].([]interface{})

	matchedCase := "default"
	for _, c := range cases {
		if caseMap, ok := c.(map[string]interface{}); ok {
			if caseValue, ok := caseMap["value"].(string); ok && caseValue == value {
				matchedCase = caseValue
				break
			}
		}
	}

	result := types.JSON{
		"value":       value,
		"matchedCase": matchedCase,
	}

	return result, nil
}

func (n *SwitchNode) Metadata() wtypes.NodeMetadata {
	return wtypes.NodeMetadata{
		Type:        "logic.switch",
		Name:        "Switch",
		Description: "Route execution based on value matching",
		Category:    "logic",
		Version:     "1.0.0",
		Inputs:      []wtypes.NodePort{{Name: "main", Type: "any"}},
		Outputs:     []wtypes.NodePort{{Name: "default", Type: "any"}},
		Parameters: []wtypes.NodeParameter{
			{Name: "value", DisplayName: "Value", Type: "string", Required: true},
			{Name: "cases", DisplayName: "Cases", Type: "json", Required: true},
		},
	}
}
