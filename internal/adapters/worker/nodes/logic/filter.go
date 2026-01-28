package logic

import (
	"context"

	"github.com/linkflow-ai/linkflow/internal/adapters/worker/executor"
	wtypes "github.com/linkflow-ai/linkflow/internal/adapters/worker/types"
	"github.com/linkflow-ai/linkflow/internal/shared/types"
)

type FilterNode struct{}

func NewFilterNode() *FilterNode {
	return &FilterNode{}
}

func (n *FilterNode) Execute(ctx context.Context, runtime *executor.Runtime, node map[string]interface{}) (types.JSON, error) {
	params, _ := node["parameters"].(map[string]interface{})
	field, _ := params["field"].(string)
	op, _ := params["operator"].(string)
	value := params["value"]

	inputData := runtime.GetInputData()
	items, _ := inputData["items"].([]interface{})

	var filtered []interface{}
	for _, item := range items {
		if itemMap, ok := item.(map[string]interface{}); ok {
			if matchesFilter(itemMap, field, op, value) {
				filtered = append(filtered, item)
			}
		}
	}

	return types.JSON{
		"items": filtered,
		"count": len(filtered),
	}, nil
}

func matchesFilter(item map[string]interface{}, field, op string, value interface{}) bool {
	fieldVal := item[field]
	switch op {
	case "eq", "equals":
		return fieldVal == value
	case "neq", "not_equals":
		return fieldVal != value
	case "contains":
		if s, ok := fieldVal.(string); ok {
			if v, ok := value.(string); ok {
				return len(s) > 0 && len(v) > 0 && contains(s, v)
			}
		}
	case "exists":
		return fieldVal != nil
	}
	return true
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func (n *FilterNode) Metadata() wtypes.NodeMetadata {
	return wtypes.NodeMetadata{
		Type:        "logic.filter",
		Name:        "Filter",
		Description: "Filter array items based on conditions. Keep only items that match specified criteria.",
		Category:    "logic",
		Version:     "1.0.0",
		Icon:        "FilterHorizontal",
		Color:       "#06B6D4",
		Inputs: []wtypes.NodePort{
			{Name: "main", Type: "array", Description: "Array of items to filter"},
		},
		Outputs: []wtypes.NodePort{
			{Name: "kept", Type: "array", Description: "Items that match the filter"},
			{Name: "removed", Type: "array", Description: "Items that don't match the filter"},
		},
		Parameters: []wtypes.NodeParameter{
			{
				Name:        "mode",
				DisplayName: "Filter Mode",
				Type:        "options",
				Required:    true,
				Default:     "simple",
				Description: "How to define filter conditions",
				Options: []wtypes.ParamOption{
					{Name: "Simple", Value: "simple", Description: "Single field comparison"},
					{Name: "Expression", Value: "expression", Description: "JavaScript expression"},
					{Name: "Multiple Conditions", Value: "multiple", Description: "Combine multiple conditions"},
				},
			},
			{
				Name:        "field",
				DisplayName: "Field",
				Type:        "string",
				Required:    true,
				Description: "Field name to filter by (supports dot notation: user.name)",
				Placeholder: "status",
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
					{Name: "Is Null", Value: "is_null"},
					{Name: "Is Not Null", Value: "is_not_null"},
					{Name: "In List", Value: "in"},
					{Name: "Not In List", Value: "not_in"},
				},
			},
			{
				Name:        "value",
				DisplayName: "Value",
				Type:        "string",
				Required:    false,
				Description: "Value to compare against",
				Placeholder: "active",
				ShowIf:      "mode === 'simple' && operator !== 'is_empty' && operator !== 'is_not_empty' && operator !== 'is_null' && operator !== 'is_not_null'",
			},
			{
				Name:        "expression",
				DisplayName: "Expression",
				Type:        "code",
				Required:    true,
				Description: "JavaScript expression returning true to keep item (use $item to access current item)",
				Placeholder: "$item.price > 100 && $item.inStock === true",
				ShowIf:      "mode === 'expression'",
			},
			{
				Name:        "conditions",
				DisplayName: "Conditions",
				Type:        "json",
				Required:    true,
				Description: "Array of filter conditions",
				Placeholder: `[{"field": "status", "operator": "equals", "value": "active"}]`,
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
				Name:        "output_removed",
				DisplayName: "Output Removed Items",
				Type:        "boolean",
				Required:    false,
				Default:     false,
				Description: "Also output items that didn't match the filter",
			},
		},
	}
}
