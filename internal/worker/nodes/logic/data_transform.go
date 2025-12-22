package logic

import (
	"context"
	"fmt"
	"reflect"
	"sort"
	"strings"

	"github.com/linkflow-ai/linkflow/internal/worker/core"
)

// DataFilterNode filters array items based on conditions (extended version)
type DataFilterNode struct{}

func (n *DataFilterNode) Type() string {
	return "logic.dataFilter"
}

func (n *DataFilterNode) Execute(ctx context.Context, execCtx *core.ExecutionContext) (map[string]interface{}, error) {
	config := execCtx.Config
	input := execCtx.Input

	// Get items to filter
	items := getArrayFromInput(input, config)
	if items == nil {
		return map[string]interface{}{"items": []interface{}{}, "count": 0}, nil
	}

	// Get filter conditions
	field := core.GetString(config, "field", "")
	operator := core.GetString(config, "operator", "equals")
	value := config["value"]
	keepMatching := core.GetBool(config, "keepMatching", true)

	var result []interface{}
	for _, item := range items {
		itemMap, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		fieldValue := getNestedField(itemMap, field)
		matches := compareValues(fieldValue, operator, value)

		if (keepMatching && matches) || (!keepMatching && !matches) {
			result = append(result, item)
		}
	}

	return map[string]interface{}{
		"items": result,
		"count": len(result),
	}, nil
}

// DataSortNode sorts array items (extended version)
type DataSortNode struct{}

func (n *DataSortNode) Type() string {
	return "logic.dataSort"
}

func (n *DataSortNode) Execute(ctx context.Context, execCtx *core.ExecutionContext) (map[string]interface{}, error) {
	config := execCtx.Config
	input := execCtx.Input

	items := getArrayFromInput(input, config)
	if len(items) == 0 {
		return map[string]interface{}{"items": []interface{}{}, "count": 0}, nil
	}

	field := core.GetString(config, "field", "")
	order := core.GetString(config, "order", "asc") // asc or desc

	// Create a copy to sort
	result := make([]interface{}, len(items))
	copy(result, items)

	sort.Slice(result, func(i, j int) bool {
		iMap, iOk := result[i].(map[string]interface{})
		jMap, jOk := result[j].(map[string]interface{})
		if !iOk || !jOk {
			return false
		}

		iVal := getNestedField(iMap, field)
		jVal := getNestedField(jMap, field)

		less := compareLess(iVal, jVal)
		if order == "desc" {
			return !less
		}
		return less
	})

	return map[string]interface{}{
		"items": result,
		"count": len(result),
	}, nil
}

// DataLimitNode limits the number of items (extended version)
type DataLimitNode struct{}

func (n *DataLimitNode) Type() string {
	return "logic.dataLimit"
}

func (n *DataLimitNode) Execute(ctx context.Context, execCtx *core.ExecutionContext) (map[string]interface{}, error) {
	config := execCtx.Config
	input := execCtx.Input

	items := getArrayFromInput(input, config)
	if items == nil {
		return map[string]interface{}{"items": []interface{}{}, "count": 0}, nil
	}

	limit := core.GetInt(config, "limit", 10)
	offset := core.GetInt(config, "offset", 0)

	// Apply offset
	if offset >= len(items) {
		return map[string]interface{}{"items": []interface{}{}, "count": 0}, nil
	}
	items = items[offset:]

	// Apply limit
	if limit > 0 && limit < len(items) {
		items = items[:limit]
	}

	return map[string]interface{}{
		"items": items,
		"count": len(items),
	}, nil
}

// RemoveDuplicatesNode removes duplicate items
type RemoveDuplicatesNode struct{}

func (n *RemoveDuplicatesNode) Type() string {
	return "logic.remove_duplicates"
}

func (n *RemoveDuplicatesNode) Execute(ctx context.Context, execCtx *core.ExecutionContext) (map[string]interface{}, error) {
	config := execCtx.Config
	input := execCtx.Input

	items := getArrayFromInput(input, config)
	if items == nil {
		return map[string]interface{}{"items": []interface{}{}, "count": 0, "removed": 0}, nil
	}

	field := core.GetString(config, "field", "") // Empty means compare whole object
	seen := make(map[string]bool)
	var result []interface{}

	for _, item := range items {
		var key string
		if field == "" {
			key = fmt.Sprintf("%v", item)
		} else if itemMap, ok := item.(map[string]interface{}); ok {
			key = fmt.Sprintf("%v", getNestedField(itemMap, field))
		} else {
			key = fmt.Sprintf("%v", item)
		}

		if !seen[key] {
			seen[key] = true
			result = append(result, item)
		}
	}

	return map[string]interface{}{
		"items":   result,
		"count":   len(result),
		"removed": len(items) - len(result),
	}, nil
}

// AggregateNode performs aggregation operations
type AggregateNode struct{}

func (n *AggregateNode) Type() string {
	return "logic.aggregate"
}

func (n *AggregateNode) Execute(ctx context.Context, execCtx *core.ExecutionContext) (map[string]interface{}, error) {
	config := execCtx.Config
	input := execCtx.Input

	items := getArrayFromInput(input, config)
	if len(items) == 0 {
		return map[string]interface{}{"result": nil, "count": 0}, nil
	}

	operation := core.GetString(config, "operation", "count")
	field := core.GetString(config, "field", "")
	groupBy := core.GetString(config, "groupBy", "")

	// If groupBy is specified, group items first
	if groupBy != "" {
		return n.aggregateWithGroupBy(items, operation, field, groupBy)
	}

	// Simple aggregation
	result := n.performAggregation(items, operation, field)

	return map[string]interface{}{
		"result": result,
		"count":  len(items),
	}, nil
}

func (n *AggregateNode) aggregateWithGroupBy(items []interface{}, operation, field, groupBy string) (map[string]interface{}, error) {
	groups := make(map[string][]interface{})

	for _, item := range items {
		itemMap, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		key := fmt.Sprintf("%v", getNestedField(itemMap, groupBy))
		groups[key] = append(groups[key], item)
	}

	results := make([]map[string]interface{}, 0, len(groups))
	for groupKey, groupItems := range groups {
		result := n.performAggregation(groupItems, operation, field)
		results = append(results, map[string]interface{}{
			"group": groupKey,
			"value": result,
			"count": len(groupItems),
		})
	}

	return map[string]interface{}{
		"groups": results,
		"count":  len(results),
	}, nil
}

func (n *AggregateNode) performAggregation(items []interface{}, operation, field string) interface{} {
	switch operation {
	case "count":
		return len(items)

	case "sum":
		var sum float64
		for _, item := range items {
			sum += getFloatValue(item, field)
		}
		return sum

	case "avg":
		if len(items) == 0 {
			return 0.0
		}
		var sum float64
		for _, item := range items {
			sum += getFloatValue(item, field)
		}
		return sum / float64(len(items))

	case "min":
		var min float64
		for i, item := range items {
			val := getFloatValue(item, field)
			if i == 0 || val < min {
				min = val
			}
		}
		return min

	case "max":
		var max float64
		for i, item := range items {
			val := getFloatValue(item, field)
			if i == 0 || val > max {
				max = val
			}
		}
		return max

	case "first":
		if len(items) > 0 {
			if field != "" {
				if m, ok := items[0].(map[string]interface{}); ok {
					return getNestedField(m, field)
				}
			}
			return items[0]
		}
		return nil

	case "last":
		if len(items) > 0 {
			if field != "" {
				if m, ok := items[len(items)-1].(map[string]interface{}); ok {
					return getNestedField(m, field)
				}
			}
			return items[len(items)-1]
		}
		return nil

	case "concat":
		var values []string
		for _, item := range items {
			if field != "" {
				if m, ok := item.(map[string]interface{}); ok {
					values = append(values, fmt.Sprintf("%v", getNestedField(m, field)))
				}
			} else {
				values = append(values, fmt.Sprintf("%v", item))
			}
		}
		return strings.Join(values, ", ")

	default:
		return len(items)
	}
}

// Helper functions

func getArrayFromInput(input map[string]interface{}, config map[string]interface{}) []interface{} {
	// Check config for items path
	if itemsPath, ok := config["items"].(string); ok && itemsPath != "" {
		if resolved := core.GetNestedValue(input, itemsPath); resolved != nil {
			if arr, ok := resolved.([]interface{}); ok {
				return arr
			}
		}
	}

	// Check for direct items in config
	if items, ok := config["items"].([]interface{}); ok {
		return items
	}

	// Check input for common array locations
	if items, ok := input["items"].([]interface{}); ok {
		return items
	}
	if items, ok := input["data"].([]interface{}); ok {
		return items
	}
	if items, ok := input["$json"].([]interface{}); ok {
		return items
	}

	return nil
}

func getNestedField(obj map[string]interface{}, path string) interface{} {
	if path == "" {
		return obj
	}

	parts := strings.Split(path, ".")
	var current interface{} = obj

	for _, part := range parts {
		if m, ok := current.(map[string]interface{}); ok {
			current = m[part]
		} else {
			return nil
		}
	}

	return current
}

func compareValues(fieldValue interface{}, operator string, compareValue interface{}) bool {
	switch operator {
	case "equals", "eq", "==":
		return fmt.Sprintf("%v", fieldValue) == fmt.Sprintf("%v", compareValue)
	case "notEquals", "ne", "!=":
		return fmt.Sprintf("%v", fieldValue) != fmt.Sprintf("%v", compareValue)
	case "contains":
		return strings.Contains(fmt.Sprintf("%v", fieldValue), fmt.Sprintf("%v", compareValue))
	case "notContains":
		return !strings.Contains(fmt.Sprintf("%v", fieldValue), fmt.Sprintf("%v", compareValue))
	case "startsWith":
		return strings.HasPrefix(fmt.Sprintf("%v", fieldValue), fmt.Sprintf("%v", compareValue))
	case "endsWith":
		return strings.HasSuffix(fmt.Sprintf("%v", fieldValue), fmt.Sprintf("%v", compareValue))
	case "greaterThan", "gt", ">":
		return toFloat(fieldValue) > toFloat(compareValue)
	case "lessThan", "lt", "<":
		return toFloat(fieldValue) < toFloat(compareValue)
	case "greaterOrEqual", "gte", ">=":
		return toFloat(fieldValue) >= toFloat(compareValue)
	case "lessOrEqual", "lte", "<=":
		return toFloat(fieldValue) <= toFloat(compareValue)
	case "isEmpty":
		return fieldValue == nil || fmt.Sprintf("%v", fieldValue) == ""
	case "isNotEmpty":
		return fieldValue != nil && fmt.Sprintf("%v", fieldValue) != ""
	case "regex":
		// TODO: implement regex matching
		return false
	default:
		return false
	}
}

func compareLess(a, b interface{}) bool {
	// Try numeric comparison first
	aFloat, aOk := toFloatSafeTransform(a)
	bFloat, bOk := toFloatSafeTransform(b)
	if aOk && bOk {
		return aFloat < bFloat
	}

	// Fall back to string comparison
	return fmt.Sprintf("%v", a) < fmt.Sprintf("%v", b)
}

func toFloatTransform(v interface{}) float64 {
	f, _ := toFloatSafeTransform(v)
	return f
}

func toFloatSafeTransform(v interface{}) (float64, bool) {
	switch val := v.(type) {
	case float64:
		return val, true
	case float32:
		return float64(val), true
	case int:
		return float64(val), true
	case int64:
		return float64(val), true
	case int32:
		return float64(val), true
	case string:
		var f float64
		_, err := fmt.Sscanf(val, "%f", &f)
		return f, err == nil
	default:
		return 0, false
	}
}

func getFloatValue(item interface{}, field string) float64 {
	if field == "" {
		return toFloatTransform(item)
	}
	if m, ok := item.(map[string]interface{}); ok {
		return toFloatTransform(getNestedField(m, field))
	}
	return 0
}

// Suppress unused import warning
var _ = reflect.TypeOf(nil)
