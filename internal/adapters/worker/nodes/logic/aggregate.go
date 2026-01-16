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
		Description: "Aggregate values (sum, avg, min, max, count)",
		Category:    "logic",
		Version:     "1.0.0",
		Inputs:      []wtypes.NodePort{{Name: "main", Type: "array"}},
		Outputs:     []wtypes.NodePort{{Name: "main", Type: "object"}},
		Parameters: []wtypes.NodeParameter{
			{Name: "field", DisplayName: "Field", Type: "string", Required: true},
			{Name: "operation", DisplayName: "Operation", Type: "options", Required: true},
		},
	}
}
