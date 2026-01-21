package transform

import (
	"context"
	"fmt"
	"math"
	"sort"

	"github.com/linkflow-ai/linkflow/internal/adapters/worker/executor"
	wtypes "github.com/linkflow-ai/linkflow/internal/adapters/worker/types"
	"github.com/linkflow-ai/linkflow/internal/shared/types"
)

// MathOperationNode performs math operations
type MathOperationNode struct{}

func NewMathOperationNode() *MathOperationNode {
	return &MathOperationNode{}
}

func (n *MathOperationNode) Execute(ctx context.Context, runtime *executor.Runtime, node map[string]interface{}) (types.JSON, error) {
	params, _ := node["parameters"].(map[string]interface{})

	operation, _ := params["operation"].(string)
	a := toFloat(params["a"])
	b := toFloat(params["b"])

	var result float64
	switch operation {
	case "add":
		result = a + b
	case "subtract":
		result = a - b
	case "multiply":
		result = a * b
	case "divide":
		if b == 0 {
			return types.JSON{"error": "division by zero", "success": false}, nil
		}
		result = a / b
	case "modulo":
		result = math.Mod(a, b)
	case "power":
		result = math.Pow(a, b)
	case "sqrt":
		result = math.Sqrt(a)
	case "abs":
		result = math.Abs(a)
	case "ceil":
		result = math.Ceil(a)
	case "floor":
		result = math.Floor(a)
	case "round":
		result = math.Round(a)
	case "min":
		result = math.Min(a, b)
	case "max":
		result = math.Max(a, b)
	case "log":
		result = math.Log(a)
	case "log10":
		result = math.Log10(a)
	case "sin":
		result = math.Sin(a)
	case "cos":
		result = math.Cos(a)
	case "tan":
		result = math.Tan(a)
	case "random":
		// Note: In production, use crypto/rand for better randomness
		result = math.Floor(a + (b-a+1)*0.5) // Placeholder
	default:
		result = a
	}

	return types.JSON{"result": result, "success": true}, nil
}

func (n *MathOperationNode) Metadata() wtypes.NodeMetadata {
	return wtypes.NodeMetadata{
		Type:        "transform.math",
		Name:        "Math Operations",
		Description: "Perform mathematical operations",
		Category:    "transform",
		Version:     "1.0.0",
		Inputs:      []wtypes.NodePort{{Name: "main", Type: "any"}},
		Outputs:     []wtypes.NodePort{{Name: "main", Type: "number"}},
		Parameters: []wtypes.NodeParameter{
			{Name: "operation", Type: "select", Description: "Math operation", Required: true, Options: []wtypes.ParamOption{
				{Value: "add", Name: "Add (+)"},
				{Value: "subtract", Name: "Subtract (-)"},
				{Value: "multiply", Name: "Multiply (*)"},
				{Value: "divide", Name: "Divide (/)"},
				{Value: "modulo", Name: "Modulo (%)"},
				{Value: "power", Name: "Power (^)"},
				{Value: "sqrt", Name: "Square Root"},
				{Value: "abs", Name: "Absolute Value"},
				{Value: "ceil", Name: "Ceiling"},
				{Value: "floor", Name: "Floor"},
				{Value: "round", Name: "Round"},
				{Value: "min", Name: "Minimum"},
				{Value: "max", Name: "Maximum"},
			}},
			{Name: "a", Type: "number", Description: "First operand", Required: true},
			{Name: "b", Type: "number", Description: "Second operand"},
		},
	}
}

// AggregateNumbersNode performs aggregate operations on arrays
type AggregateNumbersNode struct{}

func NewAggregateNumbersNode() *AggregateNumbersNode {
	return &AggregateNumbersNode{}
}

func (n *AggregateNumbersNode) Execute(ctx context.Context, runtime *executor.Runtime, node map[string]interface{}) (types.JSON, error) {
	params, _ := node["parameters"].(map[string]interface{})
	inputData := runtime.GetInputData()

	items, _ := inputData["items"].([]interface{})
	if items == nil {
		if arr, ok := inputData["data"].([]interface{}); ok {
			items = arr
		}
	}

	field, _ := params["field"].(string)

	var numbers []float64
	for _, item := range items {
		var val float64
		if field != "" {
			if m, ok := item.(map[string]interface{}); ok {
				val = toFloat(m[field])
			}
		} else {
			val = toFloat(item)
		}
		numbers = append(numbers, val)
	}

	if len(numbers) == 0 {
		return types.JSON{"error": "no numbers to aggregate", "success": false}, nil
	}

	operation, _ := params["operation"].(string)

	var result float64
	switch operation {
	case "sum":
		for _, n := range numbers {
			result += n
		}
	case "average":
		for _, n := range numbers {
			result += n
		}
		result /= float64(len(numbers))
	case "min":
		result = numbers[0]
		for _, n := range numbers[1:] {
			if n < result {
				result = n
			}
		}
	case "max":
		result = numbers[0]
		for _, n := range numbers[1:] {
			if n > result {
				result = n
			}
		}
	case "count":
		result = float64(len(numbers))
	case "median":
		sorted := make([]float64, len(numbers))
		copy(sorted, numbers)
		sort.Float64s(sorted)
		mid := len(sorted) / 2
		if len(sorted)%2 == 0 {
			result = (sorted[mid-1] + sorted[mid]) / 2
		} else {
			result = sorted[mid]
		}
	default:
		result = 0
	}

	return types.JSON{"result": result, "count": len(numbers), "success": true}, nil
}

func (n *AggregateNumbersNode) Metadata() wtypes.NodeMetadata {
	return wtypes.NodeMetadata{
		Type:        "transform.aggregate_numbers",
		Name:        "Aggregate Numbers",
		Description: "Perform aggregate operations on number arrays",
		Category:    "transform",
		Version:     "1.0.0",
		Inputs:      []wtypes.NodePort{{Name: "main", Type: "array"}},
		Outputs:     []wtypes.NodePort{{Name: "main", Type: "number"}},
		Parameters: []wtypes.NodeParameter{
			{Name: "operation", Type: "select", Description: "Aggregation operation", Required: true, Options: []wtypes.ParamOption{
				{Value: "sum", Name: "Sum"},
				{Value: "average", Name: "Average"},
				{Value: "min", Name: "Minimum"},
				{Value: "max", Name: "Maximum"},
				{Value: "count", Name: "Count"},
				{Value: "median", Name: "Median"},
			}},
			{Name: "field", Type: "string", Description: "Field to aggregate (for objects)"},
		},
	}
}

func toFloat(v interface{}) float64 {
	switch n := v.(type) {
	case float64:
		return n
	case float32:
		return float64(n)
	case int:
		return float64(n)
	case int64:
		return float64(n)
	case string:
		var f float64
		fmt.Sscanf(n, "%f", &f)
		return f
	default:
		return 0
	}
}
