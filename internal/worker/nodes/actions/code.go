package actions

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/dop251/goja"
	"github.com/linkflow-ai/linkflow/internal/worker/nodes"
)

type CodeNode struct{}

func NewCodeNode() *CodeNode {
	return &CodeNode{}
}

func (n *CodeNode) Type() string {
	return "action.code"
}

func (n *CodeNode) Execute(ctx context.Context, execCtx *nodes.ExecutionContext) (map[string]interface{}, error) {
	code := getString(execCtx.Config, "code", "")
	if code == "" {
		return nil, fmt.Errorf("code is required")
	}

	language := getString(execCtx.Config, "language", "javascript")
	timeout := getInt(execCtx.Config, "timeout", 10)

	switch language {
	case "javascript", "js":
		return n.executeJavaScript(ctx, code, execCtx.Input, timeout)
	default:
		return nil, fmt.Errorf("unsupported language: %s", language)
	}
}

func (n *CodeNode) executeJavaScript(ctx context.Context, code string, input map[string]interface{}, timeoutSec int) (map[string]interface{}, error) {
	vm := goja.New()

	// Set up timeout
	time.AfterFunc(time.Duration(timeoutSec)*time.Second, func() {
		vm.Interrupt("execution timeout")
	})

	// Inject input data
	if err := vm.Set("$input", input); err != nil {
		return nil, fmt.Errorf("failed to set input: %w", err)
	}

	if jsonData, ok := input["$json"]; ok {
		if err := vm.Set("$json", jsonData); err != nil {
			return nil, fmt.Errorf("failed to set $json: %w", err)
		}
	}

	if vars, ok := input["$vars"]; ok {
		if err := vm.Set("$vars", vars); err != nil {
			return nil, fmt.Errorf("failed to set $vars: %w", err)
		}
	}

	// Add console.log
	console := map[string]interface{}{
		"log": func(args ...interface{}) {
			// Log to execution output
		},
	}
	if err := vm.Set("console", console); err != nil {
		return nil, fmt.Errorf("failed to set console: %w", err)
	}

	// Add helper functions
	helpers := `
		function items(data) {
			if (Array.isArray(data)) return data;
			if (data && typeof data === 'object') return [data];
			return [];
		}
		
		function item(index) {
			var data = $json;
			if (Array.isArray(data)) return data[index || 0];
			return data;
		}
	`

	// Wrap code in async-compatible function
	wrappedCode := fmt.Sprintf(`
		%s
		(function() {
			%s
		})()
	`, helpers, code)

	result, err := vm.RunString(wrappedCode)
	if err != nil {
		if interrupted, ok := err.(*goja.InterruptedError); ok {
			return nil, fmt.Errorf("code execution timeout: %v", interrupted.Value())
		}
		return nil, fmt.Errorf("code execution error: %w", err)
	}

	// Convert result to Go types
	output := n.convertValue(result.Export())

	if outputMap, ok := output.(map[string]interface{}); ok {
		return outputMap, nil
	}

	return map[string]interface{}{
		"result": output,
	}, nil
}

func (n *CodeNode) convertValue(val interface{}) interface{} {
	if val == nil {
		return nil
	}

	switch v := val.(type) {
	case map[string]interface{}:
		result := make(map[string]interface{})
		for key, value := range v {
			result[key] = n.convertValue(value)
		}
		return result
	case []interface{}:
		result := make([]interface{}, len(v))
		for i, value := range v {
			result[i] = n.convertValue(value)
		}
		return result
	case goja.Value:
		return n.convertValue(v.Export())
	default:
		return v
	}
}

type FunctionNode struct{}

func NewFunctionNode() *FunctionNode {
	return &FunctionNode{}
}

func (n *FunctionNode) Type() string {
	return "action.function"
}

func (n *FunctionNode) Execute(ctx context.Context, execCtx *nodes.ExecutionContext) (map[string]interface{}, error) {
	code := getString(execCtx.Config, "code", "")
	if code == "" {
		code = getString(execCtx.Config, "function", "")
	}
	if code == "" {
		return nil, fmt.Errorf("function code is required")
	}

	codeNode := &CodeNode{}
	wrappedCode := fmt.Sprintf(`
		var items = $json;
		if (!Array.isArray(items)) items = [items];
		
		var results = [];
		for (var i = 0; i < items.length; i++) {
			var item = items[i];
			var result = (function(item, index) {
				%s
			})(item, i);
			if (result !== undefined) {
				results.push(result);
			}
		}
		
		({ items: results, count: results.length })
	`, code)

	return codeNode.executeJavaScript(ctx, wrappedCode, execCtx.Input, 30)
}

type TransformNode struct{}

func NewTransformNode() *TransformNode {
	return &TransformNode{}
}

func (n *TransformNode) Type() string {
	return "action.transform"
}

func (n *TransformNode) Execute(ctx context.Context, execCtx *nodes.ExecutionContext) (map[string]interface{}, error) {
	mode := getString(execCtx.Config, "mode", "map")

	var items []interface{}
	if jsonData, ok := execCtx.Input["$json"].([]interface{}); ok {
		items = jsonData
	} else if jsonData, ok := execCtx.Input["$json"].(map[string]interface{}); ok {
		items = []interface{}{jsonData}
	}

	fields, _ := execCtx.Config["fields"].(map[string]interface{})

	switch mode {
	case "map":
		return n.mapFields(items, fields)
	case "rename":
		return n.renameFields(items, fields)
	case "remove":
		return n.removeFields(items, fields)
	case "keep":
		return n.keepFields(items, fields)
	default:
		return map[string]interface{}{"items": items}, nil
	}
}

func (n *TransformNode) mapFields(items []interface{}, fields map[string]interface{}) (map[string]interface{}, error) {
	results := make([]interface{}, 0, len(items))

	for _, item := range items {
		itemMap, ok := item.(map[string]interface{})
		if !ok {
			results = append(results, item)
			continue
		}

		newItem := make(map[string]interface{})
		for k, v := range itemMap {
			newItem[k] = v
		}
		for k, v := range fields {
			newItem[k] = v
		}
		results = append(results, newItem)
	}

	return map[string]interface{}{
		"items": results,
		"count": len(results),
	}, nil
}

func (n *TransformNode) renameFields(items []interface{}, fields map[string]interface{}) (map[string]interface{}, error) {
	results := make([]interface{}, 0, len(items))

	for _, item := range items {
		itemMap, ok := item.(map[string]interface{})
		if !ok {
			results = append(results, item)
			continue
		}

		newItem := make(map[string]interface{})
		for k, v := range itemMap {
			if newKey, exists := fields[k]; exists {
				newItem[newKey.(string)] = v
			} else {
				newItem[k] = v
			}
		}
		results = append(results, newItem)
	}

	return map[string]interface{}{
		"items": results,
		"count": len(results),
	}, nil
}

func (n *TransformNode) removeFields(items []interface{}, fields map[string]interface{}) (map[string]interface{}, error) {
	results := make([]interface{}, 0, len(items))

	for _, item := range items {
		itemMap, ok := item.(map[string]interface{})
		if !ok {
			results = append(results, item)
			continue
		}

		newItem := make(map[string]interface{})
		for k, v := range itemMap {
			if _, remove := fields[k]; !remove {
				newItem[k] = v
			}
		}
		results = append(results, newItem)
	}

	return map[string]interface{}{
		"items": results,
		"count": len(results),
	}, nil
}

func (n *TransformNode) keepFields(items []interface{}, fields map[string]interface{}) (map[string]interface{}, error) {
	results := make([]interface{}, 0, len(items))

	for _, item := range items {
		itemMap, ok := item.(map[string]interface{})
		if !ok {
			results = append(results, item)
			continue
		}

		newItem := make(map[string]interface{})
		for k := range fields {
			if v, exists := itemMap[k]; exists {
				newItem[k] = v
			}
		}
		results = append(results, newItem)
	}

	return map[string]interface{}{
		"items": results,
		"count": len(results),
	}, nil
}

func getString(config map[string]interface{}, key, defaultVal string) string {
	if v, ok := config[key].(string); ok {
		return v
	}
	return defaultVal
}

func getInt(config map[string]interface{}, key string, defaultVal int) int {
	if v, ok := config[key]; ok {
		switch val := v.(type) {
		case int:
			return val
		case float64:
			return int(val)
		case json.Number:
			if i, err := val.Int64(); err == nil {
				return int(i)
			}
		}
	}
	return defaultVal
}

func getBool(config map[string]interface{}, key string, defaultVal bool) bool {
	if v, ok := config[key].(bool); ok {
		return v
	}
	return defaultVal
}

var _ nodes.Node = (*CodeNode)(nil)
var _ nodes.Node = (*FunctionNode)(nil)
var _ nodes.Node = (*TransformNode)(nil)
