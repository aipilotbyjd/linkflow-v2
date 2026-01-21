package logic

import (
	"context"

	"github.com/linkflow-ai/linkflow/internal/adapters/worker/executor"
	wtypes "github.com/linkflow-ai/linkflow/internal/adapters/worker/types"
	"github.com/linkflow-ai/linkflow/internal/shared/types"
)

type AggregateNode struct{}

func NewAggregateNode() *AggregateNode {
	return &AggregateNode{}
}

func (n *AggregateNode) Execute(ctx context.Context, runtime *executor.Runtime, node map[string]interface{}) (types.JSON, error) {
	params, _ := node["parameters"].(map[string]interface{})
	field, _ := params["field"].(string)
	operation, _ := params["operation"].(string)

	inputData := runtime.GetInputData()
	items, _ := inputData["items"].([]interface{})

	var result float64
	var values []float64

	for _, item := range items {
		if itemMap, ok := item.(map[string]interface{}); ok {
			if val, ok := itemMap[field].(float64); ok {
				values = append(values, val)
			}
		}
	}

	switch operation {
	case "sum":
		for _, v := range values {
			result += v
		}
	case "avg", "average":
		if len(values) > 0 {
			for _, v := range values {
				result += v
			}
			result /= float64(len(values))
		}
	case "min":
		if len(values) > 0 {
			result = values[0]
			for _, v := range values[1:] {
				if v < result {
					result = v
				}
			}
		}
	case "max":
		if len(values) > 0 {
			result = values[0]
			for _, v := range values[1:] {
				if v > result {
					result = v
				}
			}
		}
	case "count":
		result = float64(len(values))
	}

	return types.JSON{
		"result":    result,
		"operation": operation,
		"field":     field,
		"count":     len(values),
	}, nil
}

func (n *AggregateNode) Metadata() wtypes.NodeMetadata {
	return wtypes.NodeMetadata{
		Type:        "logic.aggregate",
		Name:        "Aggregate",
		Description: "Perform aggregation operations on arrays: sum, average, min, max, count, and group by",
		Category:    "logic",
		Version:     "1.0.0",
		Icon:        "calculator",
		Color:       "#0EA5E9",
		Inputs: []wtypes.NodePort{
			{Name: "main", Type: "array", Description: "Array of items to aggregate"},
		},
		Outputs: []wtypes.NodePort{
			{Name: "main", Type: "object", Description: "Aggregation result with computed value"},
		},
		Parameters: []wtypes.NodeParameter{
			{
				Name:        "operation",
				DisplayName: "Operation",
				Type:        "options",
				Required:    true,
				Default:     "sum",
				Description: "Aggregation operation to perform",
				Options: []wtypes.ParamOption{
					{Name: "Sum", Value: "sum", Description: "Add all values together"},
					{Name: "Average", Value: "avg", Description: "Calculate mean value"},
					{Name: "Minimum", Value: "min", Description: "Find smallest value"},
					{Name: "Maximum", Value: "max", Description: "Find largest value"},
					{Name: "Count", Value: "count", Description: "Count number of items"},
					{Name: "Count Distinct", Value: "count_distinct", Description: "Count unique values"},
					{Name: "First", Value: "first", Description: "Get first item"},
					{Name: "Last", Value: "last", Description: "Get last item"},
					{Name: "Concatenate", Value: "concat", Description: "Join string values"},
					{Name: "Group By", Value: "group", Description: "Group items by a field"},
				},
			},
			{
				Name:        "field",
				DisplayName: "Field",
				Type:        "string",
				Required:    true,
				Description: "Field to aggregate (supports dot notation: item.price)",
				Placeholder: "amount",
			},
			{
				Name:        "group_by",
				DisplayName: "Group By Field",
				Type:        "string",
				Required:    false,
				Description: "Field to group by before aggregating",
				Placeholder: "category",
				ShowIf:      "operation === 'group'",
			},
			{
				Name:        "separator",
				DisplayName: "Separator",
				Type:        "string",
				Required:    false,
				Default:     ", ",
				Description: "Separator for concatenation",
				ShowIf:      "operation === 'concat'",
			},
			{
				Name:        "include_nulls",
				DisplayName: "Include Null Values",
				Type:        "boolean",
				Required:    false,
				Default:     false,
				Description: "Include null/undefined values in calculation",
			},
			{
				Name:        "precision",
				DisplayName: "Decimal Precision",
				Type:        "number",
				Required:    false,
				Default:     2,
				Description: "Number of decimal places for numeric results",
			},
		},
	}
}
