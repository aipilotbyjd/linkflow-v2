package logic

import (
	"context"
	"fmt"
	"sort"

	"github.com/linkflow-ai/linkflow/internal/worker/nodes"
)

type MergeNode struct{}

func NewMergeNode() *MergeNode {
	return &MergeNode{}
}

func (n *MergeNode) Type() string {
	return "logic.merge"
}

func (n *MergeNode) Execute(ctx context.Context, execCtx *nodes.ExecutionContext) (map[string]interface{}, error) {
	mode, _ := execCtx.Config["mode"].(string)
	if mode == "" {
		mode = "append"
	}

	inputs := n.collectInputs(execCtx)

	switch mode {
	case "append":
		return n.appendMode(inputs)
	case "combine":
		return n.combineMode(inputs)
	case "multiplex":
		return n.multiplexMode(inputs)
	case "chooseBranch":
		return n.chooseBranchMode(inputs, execCtx.Config)
	case "wait":
		return n.waitMode(inputs)
	default:
		return n.appendMode(inputs)
	}
}

func (n *MergeNode) collectInputs(execCtx *nodes.ExecutionContext) []interface{} {
	var inputs []interface{}

	for key, value := range execCtx.Input {
		if key == "$vars" || key == "$env" {
			continue
		}
		inputs = append(inputs, value)
	}

	if input1, ok := execCtx.Config["input1"]; ok {
		inputs = append(inputs, input1)
	}
	if input2, ok := execCtx.Config["input2"]; ok {
		inputs = append(inputs, input2)
	}

	return inputs
}

func (n *MergeNode) appendMode(inputs []interface{}) (map[string]interface{}, error) {
	var result []interface{}
	for _, input := range inputs {
		if arr, ok := input.([]interface{}); ok {
			result = append(result, arr...)
		} else if input != nil {
			result = append(result, input)
		}
	}
	return map[string]interface{}{
		"data":  result,
		"count": len(result),
	}, nil
}

func (n *MergeNode) combineMode(inputs []interface{}) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	for i, input := range inputs {
		if obj, ok := input.(map[string]interface{}); ok {
			for k, v := range obj {
				result[k] = v
			}
		} else {
			result[fmt.Sprintf("input%d", i)] = input
		}
	}

	return map[string]interface{}{
		"data":  result,
		"count": len(result),
	}, nil
}

func (n *MergeNode) multiplexMode(inputs []interface{}) (map[string]interface{}, error) {
	if len(inputs) < 2 {
		return map[string]interface{}{
			"data":  inputs,
			"count": len(inputs),
		}, nil
	}

	arr1, ok1 := inputs[0].([]interface{})
	arr2, ok2 := inputs[1].([]interface{})

	if !ok1 || !ok2 {
		return n.combineMode(inputs)
	}

	var result []interface{}
	maxLen := len(arr1)
	if len(arr2) > maxLen {
		maxLen = len(arr2)
	}

	for i := 0; i < maxLen; i++ {
		item := make(map[string]interface{})
		if i < len(arr1) {
			item["input1"] = arr1[i]
		}
		if i < len(arr2) {
			item["input2"] = arr2[i]
		}
		result = append(result, item)
	}

	return map[string]interface{}{
		"data":  result,
		"count": len(result),
	}, nil
}

func (n *MergeNode) chooseBranchMode(inputs []interface{}, config map[string]interface{}) (map[string]interface{}, error) {
	outputType, _ := config["outputType"].(string)

	switch outputType {
	case "first":
		if len(inputs) > 0 {
			return map[string]interface{}{
				"data":        inputs[0],
				"branchIndex": 0,
			}, nil
		}
	case "last":
		if len(inputs) > 0 {
			return map[string]interface{}{
				"data":        inputs[len(inputs)-1],
				"branchIndex": len(inputs) - 1,
			}, nil
		}
	case "specified":
		idx := getInt(config, "branchIndex", 0)
		if idx >= 0 && idx < len(inputs) {
			return map[string]interface{}{
				"data":        inputs[idx],
				"branchIndex": idx,
			}, nil
		}
	}

	return map[string]interface{}{
		"data":        nil,
		"branchIndex": -1,
	}, nil
}

func (n *MergeNode) waitMode(inputs []interface{}) (map[string]interface{}, error) {
	return map[string]interface{}{
		"data":        inputs,
		"count":       len(inputs),
		"allReceived": true,
	}, nil
}

type FilterNode struct{}

func NewFilterNode() *FilterNode {
	return &FilterNode{}
}

func (n *FilterNode) Type() string {
	return "logic.filter"
}

func (n *FilterNode) Execute(ctx context.Context, execCtx *nodes.ExecutionContext) (map[string]interface{}, error) {
	var items []interface{}

	if arr, ok := execCtx.Input["$json"].([]interface{}); ok {
		items = arr
	} else if arr, ok := execCtx.Config["items"].([]interface{}); ok {
		items = arr
	}

	conditions, _ := execCtx.Config["conditions"].([]interface{})
	combineWith, _ := execCtx.Config["combineWith"].(string)
	if combineWith == "" {
		combineWith = "and"
	}

	var kept []interface{}
	var removed []interface{}

	condNode := &ConditionNode{}
	for _, item := range items {
		itemMap := map[string]interface{}{"$json": item}

		results := make([]bool, 0, len(conditions))
		for _, cond := range conditions {
			condMap, ok := cond.(map[string]interface{})
			if !ok {
				continue
			}
			result := condNode.evaluateCondition(condMap, itemMap)
			results = append(results, result)
		}

		var matched bool
		if combineWith == "or" {
			for _, r := range results {
				if r {
					matched = true
					break
				}
			}
		} else {
			matched = len(results) > 0
			for _, r := range results {
				if !r {
					matched = false
					break
				}
			}
		}

		if matched {
			kept = append(kept, item)
		} else {
			removed = append(removed, item)
		}
	}

	return map[string]interface{}{
		"kept":         kept,
		"removed":      removed,
		"keptCount":    len(kept),
		"removedCount": len(removed),
	}, nil
}

type SortNode struct{}

func NewSortNode() *SortNode {
	return &SortNode{}
}

func (n *SortNode) Type() string {
	return "logic.sort"
}

func (n *SortNode) Execute(ctx context.Context, execCtx *nodes.ExecutionContext) (map[string]interface{}, error) {
	var items []interface{}

	if arr, ok := execCtx.Input["$json"].([]interface{}); ok {
		items = make([]interface{}, len(arr))
		copy(items, arr)
	} else if arr, ok := execCtx.Config["items"].([]interface{}); ok {
		items = make([]interface{}, len(arr))
		copy(items, arr)
	}

	sortKey, _ := execCtx.Config["sortKey"].(string)
	order, _ := execCtx.Config["order"].(string)
	if order == "" {
		order = "asc"
	}

	sort.SliceStable(items, func(i, j int) bool {
		var valI, valJ interface{}

		if sortKey != "" {
			if objI, ok := items[i].(map[string]interface{}); ok {
				valI = objI[sortKey]
			}
			if objJ, ok := items[j].(map[string]interface{}); ok {
				valJ = objJ[sortKey]
			}
		} else {
			valI = items[i]
			valJ = items[j]
		}

		strI := fmt.Sprintf("%v", valI)
		strJ := fmt.Sprintf("%v", valJ)

		if order == "desc" {
			return strI > strJ
		}
		return strI < strJ
	})

	return map[string]interface{}{
		"data":  items,
		"count": len(items),
	}, nil
}

type LimitNode struct{}

func NewLimitNode() *LimitNode {
	return &LimitNode{}
}

func (n *LimitNode) Type() string {
	return "logic.limit"
}

func (n *LimitNode) Execute(ctx context.Context, execCtx *nodes.ExecutionContext) (map[string]interface{}, error) {
	var items []interface{}

	if arr, ok := execCtx.Input["$json"].([]interface{}); ok {
		items = arr
	} else if arr, ok := execCtx.Config["items"].([]interface{}); ok {
		items = arr
	}

	limit := getInt(execCtx.Config, "limit", 10)
	offset := getInt(execCtx.Config, "offset", 0)

	if offset > len(items) {
		offset = len(items)
	}
	items = items[offset:]

	if limit > len(items) {
		limit = len(items)
	}
	items = items[:limit]

	return map[string]interface{}{
		"data":   items,
		"count":  len(items),
		"offset": offset,
		"limit":  limit,
	}, nil
}

type UniqueNode struct{}

func NewUniqueNode() *UniqueNode {
	return &UniqueNode{}
}

func (n *UniqueNode) Type() string {
	return "logic.unique"
}

func (n *UniqueNode) Execute(ctx context.Context, execCtx *nodes.ExecutionContext) (map[string]interface{}, error) {
	var items []interface{}

	if arr, ok := execCtx.Input["$json"].([]interface{}); ok {
		items = arr
	} else if arr, ok := execCtx.Config["items"].([]interface{}); ok {
		items = arr
	}

	uniqueKey, _ := execCtx.Config["key"].(string)

	seen := make(map[string]bool)
	var unique []interface{}
	var duplicates []interface{}

	for _, item := range items {
		var key string
		if uniqueKey != "" {
			if obj, ok := item.(map[string]interface{}); ok {
				key = fmt.Sprintf("%v", obj[uniqueKey])
			} else {
				key = fmt.Sprintf("%v", item)
			}
		} else {
			key = fmt.Sprintf("%v", item)
		}

		if !seen[key] {
			seen[key] = true
			unique = append(unique, item)
		} else {
			duplicates = append(duplicates, item)
		}
	}

	return map[string]interface{}{
		"unique":         unique,
		"duplicates":     duplicates,
		"uniqueCount":    len(unique),
		"duplicateCount": len(duplicates),
	}, nil
}

type SplitBatchesNode struct{}

func NewSplitBatchesNode() *SplitBatchesNode {
	return &SplitBatchesNode{}
}

func (n *SplitBatchesNode) Type() string {
	return "logic.splitBatches"
}

func (n *SplitBatchesNode) Execute(ctx context.Context, execCtx *nodes.ExecutionContext) (map[string]interface{}, error) {
	var items []interface{}

	if arr, ok := execCtx.Input["$json"].([]interface{}); ok {
		items = arr
	} else if arr, ok := execCtx.Config["items"].([]interface{}); ok {
		items = arr
	}

	batchSize := getInt(execCtx.Config, "batchSize", 10)
	if batchSize < 1 {
		batchSize = 1
	}

	var batches [][]interface{}
	for i := 0; i < len(items); i += batchSize {
		end := i + batchSize
		if end > len(items) {
			end = len(items)
		}
		batches = append(batches, items[i:end])
	}

	return map[string]interface{}{
		"batches":    batches,
		"batchCount": len(batches),
		"totalItems": len(items),
		"batchSize":  batchSize,
	}, nil
}
