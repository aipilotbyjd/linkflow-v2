package logic

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/linkflow-ai/linkflow/internal/worker/core"
)

// JSONTransformNode performs JSON transformations
type JSONTransformNode struct{}

func (n *JSONTransformNode) Type() string {
	return "logic.json_transform"
}

func (n *JSONTransformNode) Execute(ctx context.Context, execCtx *core.ExecutionContext) (map[string]interface{}, error) {
	config := execCtx.Config
	input := execCtx.Input

	operation := core.GetString(config, "operation", "get")

	switch operation {
	case "get":
		return n.getValue(config, input)
	case "set":
		return n.setValue(config, input)
	case "delete":
		return n.deleteKey(config, input)
	case "merge":
		return n.mergeObjects(config, input)
	case "flatten":
		return n.flattenObject(config, input)
	case "unflatten":
		return n.unflattenObject(config, input)
	case "pick":
		return n.pickKeys(config, input)
	case "omit":
		return n.omitKeys(config, input)
	case "rename":
		return n.renameKeys(config, input)
	case "map":
		return n.mapValues(config, input)
	case "stringify":
		return n.stringify(config, input)
	case "parse":
		return n.parse(config, input)
	default:
		return n.getValue(config, input)
	}
}

func (n *JSONTransformNode) getValue(config map[string]interface{}, input map[string]interface{}) (map[string]interface{}, error) {
	path := core.GetString(config, "path", "")
	if path == "" {
		return nil, fmt.Errorf("path is required")
	}

	data := getJSONInput(config, input)
	value := getNestedValue(data, path)

	return map[string]interface{}{
		"value": value,
		"found": value != nil,
		"path":  path,
	}, nil
}

func (n *JSONTransformNode) setValue(config map[string]interface{}, input map[string]interface{}) (map[string]interface{}, error) {
	path := core.GetString(config, "path", "")
	if path == "" {
		return nil, fmt.Errorf("path is required")
	}

	value := config["value"]
	if value == nil {
		value = input["value"]
	}

	data := getJSONInput(config, input)
	result := deepCopy(data)

	setNestedValue(result, path, value)

	return map[string]interface{}{
		"data": result,
	}, nil
}

func (n *JSONTransformNode) deleteKey(config map[string]interface{}, input map[string]interface{}) (map[string]interface{}, error) {
	path := core.GetString(config, "path", "")
	if path == "" {
		return nil, fmt.Errorf("path is required")
	}

	data := getJSONInput(config, input)
	result := deepCopy(data)

	deleteNestedKey(result, path)

	return map[string]interface{}{
		"data": result,
	}, nil
}

func (n *JSONTransformNode) mergeObjects(config map[string]interface{}, input map[string]interface{}) (map[string]interface{}, error) {
	data := getJSONInput(config, input)

	mergeWith := config["mergeWith"]
	if mergeWith == nil {
		mergeWith = input["mergeWith"]
	}

	if mergeWith == nil {
		return nil, fmt.Errorf("mergeWith object is required")
	}

	result := deepCopy(data)
	deep := core.GetBool(config, "deep", true)

	if mergeMap, ok := mergeWith.(map[string]interface{}); ok {
		if deep {
			deepMerge(result, mergeMap)
		} else {
			for k, v := range mergeMap {
				result[k] = v
			}
		}
	}

	return map[string]interface{}{
		"data": result,
	}, nil
}

func (n *JSONTransformNode) flattenObject(config map[string]interface{}, input map[string]interface{}) (map[string]interface{}, error) {
	data := getJSONInput(config, input)
	delimiter := core.GetString(config, "delimiter", ".")

	result := make(map[string]interface{})
	flattenMap("", data, result, delimiter)

	return map[string]interface{}{
		"data": result,
	}, nil
}

func (n *JSONTransformNode) unflattenObject(config map[string]interface{}, input map[string]interface{}) (map[string]interface{}, error) {
	data := getJSONInput(config, input)
	delimiter := core.GetString(config, "delimiter", ".")

	result := make(map[string]interface{})
	for key, value := range data {
		setNestedValueDelim(result, key, value, delimiter)
	}

	return map[string]interface{}{
		"data": result,
	}, nil
}

func (n *JSONTransformNode) pickKeys(config map[string]interface{}, input map[string]interface{}) (map[string]interface{}, error) {
	data := getJSONInput(config, input)

	keys := getStringArray(config, "keys")
	if len(keys) == 0 {
		return nil, fmt.Errorf("keys array is required")
	}

	result := make(map[string]interface{})
	for _, key := range keys {
		if val, ok := data[key]; ok {
			result[key] = val
		}
	}

	return map[string]interface{}{
		"data": result,
	}, nil
}

func (n *JSONTransformNode) omitKeys(config map[string]interface{}, input map[string]interface{}) (map[string]interface{}, error) {
	data := getJSONInput(config, input)

	keys := getStringArray(config, "keys")
	if len(keys) == 0 {
		return nil, fmt.Errorf("keys array is required")
	}

	keysMap := make(map[string]bool)
	for _, key := range keys {
		keysMap[key] = true
	}

	result := make(map[string]interface{})
	for key, val := range data {
		if !keysMap[key] {
			result[key] = val
		}
	}

	return map[string]interface{}{
		"data": result,
	}, nil
}

func (n *JSONTransformNode) renameKeys(config map[string]interface{}, input map[string]interface{}) (map[string]interface{}, error) {
	data := getJSONInput(config, input)

	mapping, ok := config["mapping"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("mapping object is required")
	}

	result := make(map[string]interface{})
	for key, val := range data {
		if newKey, ok := mapping[key].(string); ok {
			result[newKey] = val
		} else {
			result[key] = val
		}
	}

	return map[string]interface{}{
		"data": result,
	}, nil
}

func (n *JSONTransformNode) mapValues(config map[string]interface{}, input map[string]interface{}) (map[string]interface{}, error) {
	data := getJSONInput(config, input)

	// Simple value transformation
	transform := core.GetString(config, "transform", "")

	result := make(map[string]interface{})
	for key, val := range data {
		switch transform {
		case "string":
			result[key] = fmt.Sprintf("%v", val)
		case "uppercase":
			result[key] = strings.ToUpper(fmt.Sprintf("%v", val))
		case "lowercase":
			result[key] = strings.ToLower(fmt.Sprintf("%v", val))
		case "trim":
			result[key] = strings.TrimSpace(fmt.Sprintf("%v", val))
		default:
			result[key] = val
		}
	}

	return map[string]interface{}{
		"data": result,
	}, nil
}

func (n *JSONTransformNode) stringify(config map[string]interface{}, input map[string]interface{}) (map[string]interface{}, error) {
	data := getJSONInput(config, input)
	pretty := core.GetBool(config, "pretty", false)

	var jsonBytes []byte
	var err error

	if pretty {
		jsonBytes, err = json.MarshalIndent(data, "", "  ")
	} else {
		jsonBytes, err = json.Marshal(data)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to stringify: %w", err)
	}

	return map[string]interface{}{
		"json": string(jsonBytes),
	}, nil
}

func (n *JSONTransformNode) parse(config map[string]interface{}, input map[string]interface{}) (map[string]interface{}, error) {
	jsonStr := core.GetString(config, "json", "")
	if jsonStr == "" {
		if j, ok := input["json"].(string); ok {
			jsonStr = j
		} else if j, ok := input["data"].(string); ok {
			jsonStr = j
		}
	}

	if jsonStr == "" {
		return nil, fmt.Errorf("json string is required")
	}

	var result interface{}
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	return map[string]interface{}{
		"data": result,
	}, nil
}

// Helper functions

func getJSONInput(config map[string]interface{}, input map[string]interface{}) map[string]interface{} {
	if data, ok := config["data"].(map[string]interface{}); ok {
		return data
	}
	return input
}

func getNestedValue(data map[string]interface{}, path string) interface{} {
	parts := strings.Split(path, ".")
	current := interface{}(data)

	for _, part := range parts {
		// Handle array index
		if idx := extractArrayIndex(part); idx >= 0 {
			key := part[:strings.Index(part, "[")]
			if m, ok := current.(map[string]interface{}); ok {
				if arr, ok := m[key].([]interface{}); ok && idx < len(arr) {
					current = arr[idx]
					continue
				}
			}
			return nil
		}

		if m, ok := current.(map[string]interface{}); ok {
			if val, exists := m[part]; exists {
				current = val
			} else {
				return nil
			}
		} else {
			return nil
		}
	}

	return current
}

func setNestedValue(data map[string]interface{}, path string, value interface{}) {
	parts := strings.Split(path, ".")
	current := data

	for i, part := range parts[:len(parts)-1] {
		if _, exists := current[part]; !exists {
			// Check if next part is array index
			if i+1 < len(parts) && extractArrayIndex(parts[i+1]) >= 0 {
				current[part] = []interface{}{}
			} else {
				current[part] = make(map[string]interface{})
			}
		}

		if m, ok := current[part].(map[string]interface{}); ok {
			current = m
		} else {
			return
		}
	}

	current[parts[len(parts)-1]] = value
}

func setNestedValueDelim(data map[string]interface{}, path string, value interface{}, delimiter string) {
	parts := strings.Split(path, delimiter)
	current := data

	for _, part := range parts[:len(parts)-1] {
		if _, exists := current[part]; !exists {
			current[part] = make(map[string]interface{})
		}

		if m, ok := current[part].(map[string]interface{}); ok {
			current = m
		} else {
			return
		}
	}

	current[parts[len(parts)-1]] = value
}

func deleteNestedKey(data map[string]interface{}, path string) {
	parts := strings.Split(path, ".")
	if len(parts) == 1 {
		delete(data, parts[0])
		return
	}

	current := data
	for _, part := range parts[:len(parts)-1] {
		if m, ok := current[part].(map[string]interface{}); ok {
			current = m
		} else {
			return
		}
	}

	delete(current, parts[len(parts)-1])
}

func extractArrayIndex(part string) int {
	re := regexp.MustCompile(`\[(\d+)\]`)
	if matches := re.FindStringSubmatch(part); len(matches) > 1 {
		if idx, err := strconv.Atoi(matches[1]); err == nil {
			return idx
		}
	}
	return -1
}

func deepCopy(data map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range data {
		switch val := v.(type) {
		case map[string]interface{}:
			result[k] = deepCopy(val)
		case []interface{}:
			arr := make([]interface{}, len(val))
			copy(arr, val)
			result[k] = arr
		default:
			result[k] = v
		}
	}
	return result
}

func deepMerge(target, source map[string]interface{}) {
	for key, srcVal := range source {
		if targetVal, exists := target[key]; exists {
			targetMap, targetOk := targetVal.(map[string]interface{})
			srcMap, srcOk := srcVal.(map[string]interface{})
			if targetOk && srcOk {
				deepMerge(targetMap, srcMap)
				continue
			}
		}
		target[key] = srcVal
	}
}

func flattenMap(prefix string, data map[string]interface{}, result map[string]interface{}, delimiter string) {
	for key, val := range data {
		newKey := key
		if prefix != "" {
			newKey = prefix + delimiter + key
		}

		if nestedMap, ok := val.(map[string]interface{}); ok {
			flattenMap(newKey, nestedMap, result, delimiter)
		} else {
			result[newKey] = val
		}
	}
}

func getStringArray(config map[string]interface{}, key string) []string {
	if arr, ok := config[key].([]interface{}); ok {
		result := make([]string, len(arr))
		for i, v := range arr {
			result[i] = fmt.Sprintf("%v", v)
		}
		return result
	}
	if arr, ok := config[key].([]string); ok {
		return arr
	}
	if str, ok := config[key].(string); ok {
		return strings.Split(str, ",")
	}
	return nil
}

// SplitDataNode splits data into batches
type SplitDataNode struct{}

func (n *SplitDataNode) Type() string {
	return "logic.splitData"
}

func (n *SplitDataNode) Execute(ctx context.Context, execCtx *core.ExecutionContext) (map[string]interface{}, error) {
	config := execCtx.Config
	input := execCtx.Input

	// Get array to split
	var items []interface{}
	if arr, ok := input["items"].([]interface{}); ok {
		items = arr
	} else if arr, ok := config["items"].([]interface{}); ok {
		items = arr
	} else {
		// Try to convert input to array
		v := reflect.ValueOf(input)
		if v.Kind() == reflect.Map {
			items = []interface{}{input}
		}
	}

	batchSize := core.GetInt(config, "batchSize", 1)
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
	}, nil
}

// MergeDataNode merges multiple inputs (JSON version)
type MergeDataNode struct{}

func (n *MergeDataNode) Type() string {
	return "logic.mergeData"
}

func (n *MergeDataNode) Execute(ctx context.Context, execCtx *core.ExecutionContext) (map[string]interface{}, error) {
	config := execCtx.Config
	input := execCtx.Input

	mode := core.GetString(config, "mode", "append")

	// Get inputs to merge
	inputs := []interface{}{}
	if arr, ok := config["inputs"].([]interface{}); ok {
		inputs = arr
	} else if arr, ok := input["inputs"].([]interface{}); ok {
		inputs = arr
	}

	switch mode {
	case "append":
		// Append arrays
		var result []interface{}
		for _, item := range inputs {
			if arr, ok := item.([]interface{}); ok {
				result = append(result, arr...)
			} else {
				result = append(result, item)
			}
		}
		return map[string]interface{}{
			"data":  result,
			"count": len(result),
		}, nil

	case "combine":
		// Combine objects
		result := make(map[string]interface{})
		for _, item := range inputs {
			if obj, ok := item.(map[string]interface{}); ok {
				for k, v := range obj {
					result[k] = v
				}
			}
		}
		return map[string]interface{}{
			"data": result,
		}, nil

	case "zip":
		// Zip arrays together
		var result []map[string]interface{}
		maxLen := 0
		arrays := [][]interface{}{}
		for _, item := range inputs {
			if arr, ok := item.([]interface{}); ok {
				arrays = append(arrays, arr)
				if len(arr) > maxLen {
					maxLen = len(arr)
				}
			}
		}

		for i := 0; i < maxLen; i++ {
			row := make(map[string]interface{})
			for j, arr := range arrays {
				if i < len(arr) {
					row[fmt.Sprintf("input%d", j)] = arr[i]
				}
			}
			result = append(result, row)
		}

		return map[string]interface{}{
			"data":  result,
			"count": len(result),
		}, nil

	default:
		return map[string]interface{}{
			"data": inputs,
		}, nil
	}
}

// Note: Nodes are registered in logic/init.go
