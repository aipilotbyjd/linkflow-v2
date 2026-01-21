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
		Description: "Route data to different branches based on conditions. Supports multiple condition types including comparisons, regex, and expressions.",
		Category:    "logic",
		Version:     "1.0.0",
		Icon:        "git-branch",
		Color:       "#F59E0B",
		Inputs: []wtypes.NodePort{
			{Name: "main", Type: "any", Description: "Input data to evaluate"},
		},
		Outputs: []wtypes.NodePort{
			{Name: "true", Type: "any", Description: "Output when condition is true"},
			{Name: "false", Type: "any", Description: "Output when condition is false"},
		},
		Parameters: []wtypes.NodeParameter{
			{
				Name:        "mode",
				DisplayName: "Condition Mode",
				Type:        "options",
				Required:    true,
				Default:     "simple",
				Description: "How to define the condition",
				Options: []wtypes.ParamOption{
					{Name: "Simple", Value: "simple", Description: "Compare two values"},
					{Name: "Expression", Value: "expression", Description: "JavaScript expression"},
					{Name: "Multiple Conditions", Value: "multiple", Description: "AND/OR multiple conditions"},
				},
			},
			{
				Name:        "value1",
				DisplayName: "Value 1",
				Type:        "string",
				Required:    true,
				Description: "First value to compare (supports expressions like {{$input.field}})",
				Placeholder: "{{$input.status}}",
				ShowIf:      "mode === 'simple'",
			},
			{
				Name:        "operator",
				DisplayName: "Operator",
				Type:        "options",
				Required:    true,
				Default:     "equals",
				Description: "Comparison operator",
				ShowIf:      "mode === 'simple'",
				Options: []wtypes.ParamOption{
					{Name: "Equals", Value: "equals"},
					{Name: "Not Equals", Value: "not_equals"},
					{Name: "Greater Than", Value: "greater_than"},
					{Name: "Greater Than or Equal", Value: "greater_equal"},
					{Name: "Less Than", Value: "less_than"},
					{Name: "Less Than or Equal", Value: "less_equal"},
					{Name: "Contains", Value: "contains"},
					{Name: "Not Contains", Value: "not_contains"},
					{Name: "Starts With", Value: "starts_with"},
					{Name: "Ends With", Value: "ends_with"},
					{Name: "Matches Regex", Value: "regex"},
					{Name: "Is Empty", Value: "is_empty"},
					{Name: "Is Not Empty", Value: "is_not_empty"},
					{Name: "Is True", Value: "is_true"},
					{Name: "Is False", Value: "is_false"},
					{Name: "Is Null", Value: "is_null"},
					{Name: "Is Not Null", Value: "is_not_null"},
					{Name: "Type Is", Value: "type_is"},
				},
			},
			{
				Name:        "value2",
				DisplayName: "Value 2",
				Type:        "string",
				Required:    false,
				Description: "Second value to compare against",
				Placeholder: "active",
				ShowIf:      "mode === 'simple' && operator !== 'is_empty' && operator !== 'is_not_empty' && operator !== 'is_true' && operator !== 'is_false' && operator !== 'is_null' && operator !== 'is_not_null'",
			},
			{
				Name:        "expression",
				DisplayName: "Expression",
				Type:        "code",
				Required:    true,
				Description: "JavaScript expression that returns true or false",
				Placeholder: "$input.amount > 100 && $input.status === 'active'",
				ShowIf:      "mode === 'expression'",
			},
			{
				Name:        "conditions",
				DisplayName: "Conditions",
				Type:        "json",
				Required:    true,
				Description: "Array of conditions with logic operator",
				Placeholder: `[{"value1": "{{$input.x}}", "operator": "equals", "value2": "y"}]`,
				ShowIf:      "mode === 'multiple'",
			},
			{
				Name:        "combine_operator",
				DisplayName: "Combine With",
				Type:        "options",
				Required:    false,
				Default:     "and",
				Description: "How to combine multiple conditions",
				ShowIf:      "mode === 'multiple'",
				Options: []wtypes.ParamOption{
					{Name: "AND (all must match)", Value: "and"},
					{Name: "OR (any must match)", Value: "or"},
				},
			},
			{
				Name:        "case_sensitive",
				DisplayName: "Case Sensitive",
				Type:        "boolean",
				Required:    false,
				Default:     true,
				Description: "Whether string comparisons are case sensitive",
			},
			{
				Name:        "condition",
				DisplayName: "Condition (Legacy)",
				Type:        "boolean",
				Required:    false,
				Description: "Direct boolean condition value (legacy support)",
			},
		},
	}
}
