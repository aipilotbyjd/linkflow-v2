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
		Description: "Route data to multiple branches based on matching values. Like a multi-way if statement with multiple possible outputs.",
		Category:    "logic",
		Version:     "1.0.0",
		Icon:        "shuffle",
		Color:       "#EC4899",
		Inputs: []wtypes.NodePort{
			{Name: "main", Type: "any", Description: "Input data to evaluate"},
		},
		Outputs: []wtypes.NodePort{
			{Name: "output_0", Type: "any", Description: "First case output"},
			{Name: "output_1", Type: "any", Description: "Second case output"},
			{Name: "output_2", Type: "any", Description: "Third case output"},
			{Name: "output_3", Type: "any", Description: "Fourth case output"},
			{Name: "default", Type: "any", Description: "Default output when no case matches"},
		},
		Parameters: []wtypes.NodeParameter{
			{
				Name:        "mode",
				DisplayName: "Switch Mode",
				Type:        "options",
				Required:    true,
				Default:     "value",
				Description: "How to determine which branch to take",
				Options: []wtypes.ParamOption{
					{Name: "Value Match", Value: "value", Description: "Match against specific values"},
					{Name: "Expression", Value: "expression", Description: "Evaluate expressions for each case"},
					{Name: "Routing Rules", Value: "rules", Description: "Use complex routing rules"},
				},
			},
			{
				Name:        "value",
				DisplayName: "Value to Match",
				Type:        "string",
				Required:    true,
				Description: "Value to compare against cases (supports expressions)",
				Placeholder: "{{$input.status}}",
				ShowIf:      "mode === 'value'",
			},
			{
				Name:        "data_type",
				DisplayName: "Data Type",
				Type:        "options",
				Required:    false,
				Default:     "string",
				Description: "Type of data being matched",
				ShowIf:      "mode === 'value'",
				Options: []wtypes.ParamOption{
					{Name: "String", Value: "string"},
					{Name: "Number", Value: "number"},
					{Name: "Boolean", Value: "boolean"},
				},
			},
			{
				Name:        "cases",
				DisplayName: "Cases",
				Type:        "json",
				Required:    true,
				Description: "Array of cases with output mapping",
				Placeholder: `[{"value": "active", "output": 0}, {"value": "pending", "output": 1}]`,
			},
			{
				Name:        "fallthrough",
				DisplayName: "Fall Through",
				Type:        "boolean",
				Required:    false,
				Default:     false,
				Description: "Continue to next case after match (like switch without break)",
			},
			{
				Name:        "case_sensitive",
				DisplayName: "Case Sensitive",
				Type:        "boolean",
				Required:    false,
				Default:     true,
				Description: "Whether string matching is case sensitive",
				ShowIf:      "data_type === 'string'",
			},
			{
				Name:        "multiple_match",
				DisplayName: "Multiple Match Mode",
				Type:        "options",
				Required:    false,
				Default:     "first",
				Description: "What to do when multiple cases match",
				Options: []wtypes.ParamOption{
					{Name: "First Match", Value: "first", Description: "Use first matching case"},
					{Name: "All Matches", Value: "all", Description: "Output to all matching cases"},
					{Name: "Last Match", Value: "last", Description: "Use last matching case"},
				},
			},
		},
	}
}
