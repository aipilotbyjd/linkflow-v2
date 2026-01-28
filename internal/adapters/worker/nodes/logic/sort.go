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
		Description: "Sort array items by one or more fields in ascending or descending order",
		Category:    "logic",
		Version:     "1.0.0",
		Icon:        "SortingAZ01",
		Color:       "#14B8A6",
		Inputs: []wtypes.NodePort{
			{Name: "main", Type: "array", Description: "Array of items to sort"},
		},
		Outputs: []wtypes.NodePort{
			{Name: "main", Type: "array", Description: "Sorted array"},
		},
		Parameters: []wtypes.NodeParameter{
			{
				Name:        "mode",
				DisplayName: "Sort Mode",
				Type:        "options",
				Required:    true,
				Default:     "simple",
				Description: "How to define sort criteria",
				Options: []wtypes.ParamOption{
					{Name: "Simple", Value: "simple", Description: "Sort by single field"},
					{Name: "Multiple Fields", Value: "multiple", Description: "Sort by multiple fields"},
					{Name: "Custom", Value: "custom", Description: "Custom sort expression"},
				},
			},
			{
				Name:        "field",
				DisplayName: "Sort Field",
				Type:        "string",
				Required:    true,
				Description: "Field to sort by (supports dot notation: user.createdAt)",
				Placeholder: "createdAt",
				ShowIf:      "mode === 'simple'",
			},
			{
				Name:        "order",
				DisplayName: "Order",
				Type:        "options",
				Required:    false,
				Default:     "asc",
				Description: "Sort direction",
				ShowIf:      "mode === 'simple'",
				Options: []wtypes.ParamOption{
					{Name: "Ascending (A-Z, 0-9)", Value: "asc"},
					{Name: "Descending (Z-A, 9-0)", Value: "desc"},
				},
			},
			{
				Name:        "sort_fields",
				DisplayName: "Sort Fields",
				Type:        "json",
				Required:    true,
				Description: "Array of fields to sort by with order",
				Placeholder: `[{"field": "lastName", "order": "asc"}, {"field": "firstName", "order": "asc"}]`,
				ShowIf:      "mode === 'multiple'",
			},
			{
				Name:        "expression",
				DisplayName: "Sort Expression",
				Type:        "code",
				Required:    true,
				Description: "JavaScript comparator function (use $a and $b for items, return -1, 0, or 1)",
				Placeholder: "$a.priority - $b.priority || $a.name.localeCompare($b.name)",
				ShowIf:      "mode === 'custom'",
			},
			{
				Name:        "data_type",
				DisplayName: "Data Type",
				Type:        "options",
				Required:    false,
				Default:     "auto",
				Description: "How to interpret values for sorting",
				ShowIf:      "mode === 'simple'",
				Options: []wtypes.ParamOption{
					{Name: "Auto Detect", Value: "auto"},
					{Name: "String", Value: "string"},
					{Name: "Number", Value: "number"},
					{Name: "Date", Value: "date"},
					{Name: "Boolean", Value: "boolean"},
				},
			},
			{
				Name:        "case_sensitive",
				DisplayName: "Case Sensitive",
				Type:        "boolean",
				Required:    false,
				Default:     false,
				Description: "Whether string sorting is case sensitive",
			},
			{
				Name:        "nulls_position",
				DisplayName: "Null Values",
				Type:        "options",
				Required:    false,
				Default:     "last",
				Description: "Where to place null/undefined values",
				Options: []wtypes.ParamOption{
					{Name: "At End", Value: "last"},
					{Name: "At Start", Value: "first"},
				},
			},
		},
	}
}
