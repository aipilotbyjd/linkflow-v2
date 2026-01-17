package code

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/dop251/goja"
	"github.com/linkflow-ai/linkflow/internal/adapters/worker/executor"
	wtypes "github.com/linkflow-ai/linkflow/internal/adapters/worker/types"
	"github.com/linkflow-ai/linkflow/internal/shared/types"
)

type JavaScriptNode struct {
	timeout time.Duration
}

func NewJavaScriptNode() *JavaScriptNode {
	return &JavaScriptNode{timeout: 30 * time.Second}
}

func (n *JavaScriptNode) Execute(ctx context.Context, runtime *executor.Runtime, node map[string]interface{}) (types.JSON, error) {
	params, _ := node["parameters"].(map[string]interface{})

	code, _ := params["code"].(string)
	if code == "" {
		return nil, fmt.Errorf("JavaScript code is required")
	}

	// Get input data from previous nodes
	inputData := runtime.GetInputData()

	// Create goja runtime
	vm := goja.New()

	// Set up context cancellation
	done := make(chan struct{})
	go func() {
		select {
		case <-ctx.Done():
			vm.Interrupt("execution cancelled")
		case <-time.After(n.timeout):
			vm.Interrupt("execution timeout")
		case <-done:
		}
	}()
	defer close(done)

	// Inject input data as $input
	if err := vm.Set("$input", inputData); err != nil {
		return nil, fmt.Errorf("failed to set input data: %w", err)
	}

	// Inject console.log
	console := vm.NewObject()
	logs := make([]interface{}, 0)
	if err := console.Set("log", func(call goja.FunctionCall) goja.Value {
		args := make([]interface{}, len(call.Arguments))
		for i, arg := range call.Arguments {
			args[i] = arg.Export()
		}
		logs = append(logs, args)
		return goja.Undefined()
	}); err != nil {
		return nil, fmt.Errorf("failed to set console.log: %w", err)
	}
	if err := vm.Set("console", console); err != nil {
		return nil, fmt.Errorf("failed to set console: %w", err)
	}

	// Inject JSON helpers
	if err := vm.Set("JSON", map[string]interface{}{
		"parse": func(s string) (interface{}, error) {
			var result interface{}
			if err := json.Unmarshal([]byte(s), &result); err != nil {
				return nil, err
			}
			return result, nil
		},
		"stringify": func(v interface{}) (string, error) {
			b, err := json.Marshal(v)
			if err != nil {
				return "", err
			}
			return string(b), nil
		},
	}); err != nil {
		return nil, fmt.Errorf("failed to set JSON: %w", err)
	}

	// Execute the code
	result, err := vm.RunString(code)
	if err != nil {
		return nil, fmt.Errorf("JavaScript execution error: %w", err)
	}

	// Export result
	var output interface{}
	if result != nil && !goja.IsUndefined(result) && !goja.IsNull(result) {
		output = result.Export()
	}

	return types.JSON{
		"result":   output,
		"logs":     logs,
		"executed": true,
		"language": "javascript",
	}, nil
}

func (n *JavaScriptNode) Metadata() wtypes.NodeMetadata {
	return wtypes.NodeMetadata{
		Type:        "action.code",
		Name:        "JavaScript",
		Description: "Execute JavaScript code",
		Category:    "action",
		Version:     "1.0.0",
		Inputs:      []wtypes.NodePort{{Name: "main", Type: "any"}},
		Outputs:     []wtypes.NodePort{{Name: "main", Type: "any"}},
		Parameters: []wtypes.NodeParameter{
			{Name: "code", DisplayName: "Code", Type: "code", Required: true},
		},
	}
}
