package logic

import (
	"context"

	"github.com/linkflow-ai/linkflow/internal/worker/core"
)

type LoopNode struct{}

func NewLoopNode() *LoopNode {
	return &LoopNode{}
}

func (n *LoopNode) Type() string {
	return "logic.loop"
}

func (n *LoopNode) Execute(ctx context.Context, execCtx *core.ExecutionContext) (map[string]interface{}, error) {
	mode, _ := execCtx.Config["mode"].(string)
	if mode == "" {
		mode = "forEach"
	}

	switch mode {
	case "forEach":
		return n.executeForEach(ctx, execCtx)
	case "times":
		return n.executeTimes(ctx, execCtx)
	case "while":
		return n.executeWhile(ctx, execCtx)
	default:
		return n.executeForEach(ctx, execCtx)
	}
}

func (n *LoopNode) executeForEach(ctx context.Context, execCtx *core.ExecutionContext) (map[string]interface{}, error) {
	var items []interface{}

	if itemsConfig, ok := execCtx.Config["items"].([]interface{}); ok {
		items = itemsConfig
	} else if itemsPath, ok := execCtx.Config["items"].(string); ok {
		resolved := core.GetNestedValue(execCtx.Input, itemsPath)
		if arr, ok := resolved.([]interface{}); ok {
			items = arr
		}
	} else if jsonData, ok := execCtx.Input["$json"].([]interface{}); ok {
		items = jsonData
	}

	limit := core.GetInt(execCtx.Config, "limit", 1000)
	if len(items) > limit {
		items = items[:limit]
	}

	results := make([]map[string]interface{}, 0, len(items))
	for i, item := range items {
		select {
		case <-ctx.Done():
			return map[string]interface{}{
				"items":       results,
				"count":       len(results),
				"interrupted": true,
			}, ctx.Err()
		default:
		}

		results = append(results, map[string]interface{}{
			"item":  item,
			"index": i,
			"first": i == 0,
			"last":  i == len(items)-1,
		})
	}

	return map[string]interface{}{
		"items":       results,
		"count":       len(results),
		"interrupted": false,
	}, nil
}

func (n *LoopNode) executeTimes(ctx context.Context, execCtx *core.ExecutionContext) (map[string]interface{}, error) {
	times := core.GetInt(execCtx.Config, "times", 1)
	limit := core.GetInt(execCtx.Config, "limit", 1000)
	if times > limit {
		times = limit
	}

	results := make([]map[string]interface{}, 0, times)
	for i := 0; i < times; i++ {
		select {
		case <-ctx.Done():
			return map[string]interface{}{
				"items":       results,
				"count":       len(results),
				"interrupted": true,
			}, ctx.Err()
		default:
		}

		results = append(results, map[string]interface{}{
			"index": i,
			"first": i == 0,
			"last":  i == times-1,
		})
	}

	return map[string]interface{}{
		"items":       results,
		"count":       len(results),
		"interrupted": false,
	}, nil
}

func (n *LoopNode) executeWhile(ctx context.Context, execCtx *core.ExecutionContext) (map[string]interface{}, error) {
	limit := core.GetInt(execCtx.Config, "limit", 1000)
	results := make([]map[string]interface{}, 0)
	iteration := 0

	for iteration < limit {
		select {
		case <-ctx.Done():
			return map[string]interface{}{
				"items":       results,
				"count":       len(results),
				"interrupted": true,
			}, ctx.Err()
		default:
		}

		shouldContinue, _ := execCtx.Config["continue"].(bool)
		if !shouldContinue && iteration > 0 {
			break
		}

		results = append(results, map[string]interface{}{
			"index": iteration,
		})
		iteration++
	}

	return map[string]interface{}{
		"items":       results,
		"count":       len(results),
		"interrupted": iteration >= limit,
	}, nil
}

// core.GetInt is defined in error_handling.go
