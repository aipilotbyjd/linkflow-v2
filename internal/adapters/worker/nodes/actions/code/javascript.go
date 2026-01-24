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
			vm.Interrupt("execution canceled")
		case <-time.After(n.timeout):
			vm.Interrupt("execution timeout")
		case <-done:
		}
	}()
	defer close(done)

	// Inject input data as $input
	m := make(map[string]interface{})
	for k, v := range inputData {
		m[k] = v
	}
	if err := vm.Set("$input", m); err != nil {
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
	// Wrap in an IIFE to support top-level return
	wrappedCode := fmt.Sprintf("(function() {\n%s\n})()", code)
	result, err := vm.RunString(wrappedCode)
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
		Description: "Execute custom JavaScript code to transform data, perform calculations, or implement custom logic",
		Category:    "action",
		Version:     "1.0.0",
		Icon:        "code",
		Color:       "#F7DF1E",
		Inputs: []wtypes.NodePort{
			{Name: "main", Type: "any", Description: "Input data available as $input variable"},
		},
		Outputs: []wtypes.NodePort{
			{Name: "main", Type: "any", Description: "The value returned by your code"},
		},
		Parameters: []wtypes.NodeParameter{
			{
				Name:        "code",
				DisplayName: "JavaScript Code",
				Type:        "code",
				Required:    true,
				Description: "JavaScript code to execute. Access input data via $input. Return value becomes node output.",
				Placeholder: `// Access input data with $input
const items = $input.items || [];

// Transform data
const result = items.map(item => ({
  ...item,
  processed: true,
  timestamp: new Date().toISOString()
}));

// Return the result
return result;`,
			},
			{
				Name:        "mode",
				DisplayName: "Execution Mode",
				Type:        "options",
				Required:    false,
				Default:     "expression",
				Description: "How to execute the code",
				Options: []wtypes.ParamOption{
					{Name: "Expression", Value: "expression", Description: "Last expression value is returned"},
					{Name: "Function", Value: "function", Description: "Must use return statement"},
					{Name: "Each Item", Value: "each_item", Description: "Run code for each item in array"},
				},
			},
			{
				Name:        "timeout",
				DisplayName: "Timeout (seconds)",
				Type:        "number",
				Required:    false,
				Default:     30,
				Description: "Maximum execution time in seconds",
			},
			{
				Name:        "continue_on_error",
				DisplayName: "Continue on Error",
				Type:        "boolean",
				Required:    false,
				Default:     false,
				Description: "Continue workflow even if code throws an error",
			},
		},
	}
}
