package logic

import (
	"context"

	"github.com/linkflow-ai/linkflow/internal/adapters/worker/executor"
	wtypes "github.com/linkflow-ai/linkflow/internal/adapters/worker/types"
	"github.com/linkflow-ai/linkflow/internal/shared/types"
)

type LoopNode struct{}

func NewLoopNode() *LoopNode {
	return &LoopNode{}
}

func (n *LoopNode) Execute(ctx context.Context, runtime *executor.Runtime, node map[string]interface{}) (types.JSON, error) {
	inputData := runtime.GetInputData()
	items, _ := inputData["items"].([]interface{})

	// Store loop context
	return types.JSON{
		"items":         items,
		"current_index": 0,
		"total":         len(items),
		"is_loop":       true,
	}, nil
}

func (n *LoopNode) Metadata() wtypes.NodeMetadata {
	return wtypes.NodeMetadata{
		Type:        "logic.loop",
		Name:        "Loop",
		Description: "Iterate over arrays or execute a fixed number of times. Process each item individually with access to index and item data.",
		Category:    "logic",
		Version:     "1.0.0",
		Icon:        "repeat",
		Color:       "#8B5CF6",
		Inputs: []wtypes.NodePort{
			{Name: "main", Type: "array", Description: "Array of items to iterate over"},
		},
		Outputs: []wtypes.NodePort{
			{Name: "loop", Type: "any", Description: "Current item in iteration (connect nodes to process each item)"},
			{Name: "done", Type: "any", Description: "Output after all iterations complete"},
		},
		Parameters: []wtypes.NodeParameter{
			{
				Name:        "mode",
				DisplayName: "Loop Mode",
				Type:        "options",
				Required:    true,
				Default:     "each",
				Description: "How to iterate",
				Options: []wtypes.ParamOption{
					{Name: "For Each", Value: "each", Description: "Iterate over each item in array"},
					{Name: "Fixed Count", Value: "count", Description: "Execute a fixed number of times"},
					{Name: "While", Value: "while", Description: "Loop while condition is true"},
				},
			},
			{
				Name:        "items",
				DisplayName: "Items",
				Type:        "string",
				Required:    false,
				Description: "Expression for array to iterate (e.g., {{$input.data}})",
				Placeholder: "{{$input.items}}",
				ShowIf:      "mode === 'each'",
			},
			{
				Name:        "count",
				DisplayName: "Iterations",
				Type:        "number",
				Required:    true,
				Default:     10,
				Description: "Number of times to execute the loop",
				ShowIf:      "mode === 'count'",
			},
			{
				Name:        "condition",
				DisplayName: "Continue While",
				Type:        "string",
				Required:    true,
				Description: "Expression that returns true to continue (e.g., $index < 10)",
				Placeholder: "$index < 100 && $item.hasMore",
				ShowIf:      "mode === 'while'",
			},
			{
				Name:        "max_iterations",
				DisplayName: "Max Iterations",
				Type:        "number",
				Required:    false,
				Default:     1000,
				Description: "Maximum iterations to prevent infinite loops",
			},
			{
				Name:        "batch_size",
				DisplayName: "Batch Size",
				Type:        "number",
				Required:    false,
				Default:     1,
				Description: "Process items in batches of this size (1 = one at a time)",
			},
			{
				Name:        "parallel",
				DisplayName: "Parallel Execution",
				Type:        "boolean",
				Required:    false,
				Default:     false,
				Description: "Execute iterations in parallel (order not guaranteed)",
			},
			{
				Name:        "concurrency",
				DisplayName: "Max Concurrency",
				Type:        "number",
				Required:    false,
				Default:     5,
				Description: "Maximum parallel executions",
				ShowIf:      "parallel === true",
			},
			{
				Name:        "continue_on_error",
				DisplayName: "Continue on Error",
				Type:        "boolean",
				Required:    false,
				Default:     false,
				Description: "Continue loop even if an iteration fails",
			},
			{
				Name:        "collect_results",
				DisplayName: "Collect Results",
				Type:        "boolean",
				Required:    false,
				Default:     true,
				Description: "Collect all iteration outputs into an array",
			},
		},
	}
}
