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
		Description: "Filter items based on conditions",
		Category:    "logic",
		Version:     "1.0.0",
		Inputs:      []wtypes.NodePort{{Name: "main", Type: "array"}},
		Outputs:     []wtypes.NodePort{{Name: "main", Type: "array"}},
		Parameters: []wtypes.NodeParameter{
			{Name: "field", DisplayName: "Field", Type: "string", Required: true},
			{Name: "operator", DisplayName: "Operator", Type: "options", Required: true},
			{Name: "value", DisplayName: "Value", Type: "string", Required: true},
		},
	}
}
