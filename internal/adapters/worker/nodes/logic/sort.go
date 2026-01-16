package logic

import (
	"context"
	"sort"

	"github.com/linkflow-ai/linkflow/internal/adapters/worker/executor"
	wtypes "github.com/linkflow-ai/linkflow/internal/adapters/worker/types"
	"github.com/linkflow-ai/linkflow/internal/shared/types"
)

type SortNode struct{}

func NewSortNode() *SortNode {
	return &SortNode{}
}

func (n *SortNode) Execute(ctx context.Context, runtime *executor.Runtime, node map[string]interface{}) (types.JSON, error) {
	params, _ := node["parameters"].(map[string]interface{})
	field, _ := params["field"].(string)
	order, _ := params["order"].(string)
	if order == "" {
		order = "asc"
	}

	inputData := runtime.GetInputData()
	items, _ := inputData["items"].([]interface{})

	sorted := make([]interface{}, len(items))
	copy(sorted, items)

	sort.Slice(sorted, func(i, j int) bool {
		iMap, _ := sorted[i].(map[string]interface{})
		jMap, _ := sorted[j].(map[string]interface{})

		iVal := iMap[field]
		jVal := jMap[field]

		less := compareValues(iVal, jVal)
		if order == "desc" {
			return !less
		}
		return less
	})

	return types.JSON{
		"items": sorted,
		"count": len(sorted),
	}, nil
}

func compareValues(a, b interface{}) bool {
	switch av := a.(type) {
	case string:
		if bv, ok := b.(string); ok {
			return av < bv
		}
	case float64:
		if bv, ok := b.(float64); ok {
			return av < bv
		}
	case int:
		if bv, ok := b.(int); ok {
			return av < bv
		}
	}
	return false
}

func (n *SortNode) Metadata() wtypes.NodeMetadata {
	return wtypes.NodeMetadata{
		Type:        "logic.sort",
		Name:        "Sort",
		Description: "Sort items by a field",
		Category:    "logic",
		Version:     "1.0.0",
		Inputs:      []wtypes.NodePort{{Name: "main", Type: "array"}},
		Outputs:     []wtypes.NodePort{{Name: "main", Type: "array"}},
		Parameters: []wtypes.NodeParameter{
			{Name: "field", DisplayName: "Field", Type: "string", Required: true},
			{Name: "order", DisplayName: "Order", Type: "options", Required: false, Default: "asc"},
		},
	}
}
