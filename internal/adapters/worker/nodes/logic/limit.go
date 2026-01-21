package logic

import (
	"context"

	"github.com/linkflow-ai/linkflow/internal/adapters/worker/executor"
	wtypes "github.com/linkflow-ai/linkflow/internal/adapters/worker/types"
	"github.com/linkflow-ai/linkflow/internal/shared/types"
)

type LimitNode struct{}

func NewLimitNode() *LimitNode {
	return &LimitNode{}
}

func (n *LimitNode) Execute(ctx context.Context, runtime *executor.Runtime, node map[string]interface{}) (types.JSON, error) {
	params, _ := node["parameters"].(map[string]interface{})
	limit := 10
	offset := 0

	if l, ok := params["limit"].(float64); ok {
		limit = int(l)
	}
	if o, ok := params["offset"].(float64); ok {
		offset = int(o)
	}

	inputData := runtime.GetInputData()
	items, _ := inputData["items"].([]interface{})

	start := offset
	if start > len(items) {
		start = len(items)
	}

	end := start + limit
	if end > len(items) {
		end = len(items)
	}

	limited := items[start:end]

	return types.JSON{
		"items":       limited,
		"count":       len(limited),
		"total":       len(items),
		"has_more":    end < len(items),
		"next_offset": end,
	}, nil
}

func (n *LimitNode) Metadata() wtypes.NodeMetadata {
	return wtypes.NodeMetadata{
		Type:        "logic.limit",
		Name:        "Limit",
		Description: "Limit the number of items and implement pagination with offset support",
		Category:    "logic",
		Version:     "1.0.0",
		Icon:        "list",
		Color:       "#84CC16",
		Inputs: []wtypes.NodePort{
			{Name: "main", Type: "array", Description: "Array of items to limit"},
		},
		Outputs: []wtypes.NodePort{
			{Name: "main", Type: "array", Description: "Limited array with pagination info"},
		},
		Parameters: []wtypes.NodeParameter{
			{
				Name:        "limit",
				DisplayName: "Limit",
				Type:        "number",
				Required:    true,
				Default:     10,
				Description: "Maximum number of items to return",
			},
			{
				Name:        "offset",
				DisplayName: "Offset",
				Type:        "number",
				Required:    false,
				Default:     0,
				Description: "Number of items to skip from the beginning",
			},
			{
				Name:        "mode",
				DisplayName: "Mode",
				Type:        "options",
				Required:    false,
				Default:     "head",
				Description: "Which items to select",
				Options: []wtypes.ParamOption{
					{Name: "First N Items", Value: "head", Description: "Take from the beginning"},
					{Name: "Last N Items", Value: "tail", Description: "Take from the end"},
					{Name: "Random Sample", Value: "random", Description: "Random selection"},
				},
			},
			{
				Name:        "include_pagination",
				DisplayName: "Include Pagination Info",
				Type:        "boolean",
				Required:    false,
				Default:     true,
				Description: "Include total count, hasMore, and nextOffset in output",
			},
		},
	}
}
