package logic

import (
	"context"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"

	"github.com/linkflow-ai/linkflow/internal/worker/core"
)

// ExpressionNode evaluates expressions and performs calculations
type ExpressionNode struct{}

func (n *ExpressionNode) Type() string {
	return "logic.expression"
}

func (n *ExpressionNode) Execute(ctx context.Context, execCtx *core.ExecutionContext) (map[string]interface{}, error) {
	config := execCtx.Config
	input := execCtx.Input

	expression := core.GetString(config, "expression", "")
	if expression == "" {
		return nil, fmt.Errorf("expression is required")
	}

	// Replace variables in expression
	expression = replaceVariables(expression, config, input)

	// Evaluate the expression
	result, err := evaluateExpression(expression)
	if err != nil {
		return nil, fmt.Errorf("expression evaluation failed: %w", err)
	}

	return map[string]interface{}{
		"result":     result,
		"expression": expression,
	}, nil
}

func replaceVariables(expr string, config map[string]interface{}, input map[string]interface{}) string {
	// Replace {{ variable }} patterns
	re := regexp.MustCompile(`\{\{\s*(\w+(?:\.\w+)*)\s*\}\}`)
	return re.ReplaceAllStringFunc(expr, func(match string) string {
		// Extract variable name
		varName := strings.TrimSpace(match[2 : len(match)-2])
		
		// Look up in input first, then config
		if val := lookupVariable(varName, input); val != nil {
			return fmt.Sprintf("%v", val)
		}
		if val := lookupVariable(varName, config); val != nil {
			return fmt.Sprintf("%v", val)
		}
		return match
	})
}

func lookupVariable(path string, data map[string]interface{}) interface{} {
	parts := strings.Split(path, ".")
	current := interface{}(data)

	for _, part := range parts {
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

func evaluateExpression(expr string) (interface{}, error) {
	expr = strings.TrimSpace(expr)

	// Boolean expressions
	if expr == "true" {
		return true, nil
	}
	if expr == "false" {
		return false, nil
	}

	// Try to evaluate as math expression
	result, err := evaluateMath(expr)
	if err == nil {
		return result, nil
	}

	// Try to evaluate as comparison
	result2, err := evaluateComparison(expr)
	if err == nil {
		return result2, nil
	}

	// Return as string
	return expr, nil
}

func evaluateMath(expr string) (float64, error) {
	// Remove spaces
	expr = strings.ReplaceAll(expr, " ", "")

	// Handle parentheses recursively
	for strings.Contains(expr, "(") {
		start := strings.LastIndex(expr, "(")
		end := strings.Index(expr[start:], ")") + start
		if end < start {
			return 0, fmt.Errorf("mismatched parentheses")
		}

		innerExpr := expr[start+1 : end]
		innerResult, err := evaluateMath(innerExpr)
		if err != nil {
			return 0, err
		}
		expr = expr[:start] + strconv.FormatFloat(innerResult, 'f', -1, 64) + expr[end+1:]
	}

	// Handle operators in order of precedence
	// Addition and subtraction (left to right)
	result, err := evaluateAddSub(expr)
	if err != nil {
		return 0, err
	}

	return result, nil
}

func evaluateAddSub(expr string) (float64, error) {
	// Find the rightmost + or - (that's not part of a number)
	lastOp := -1
	lastOpChar := byte(0)
	parenDepth := 0

	for i := len(expr) - 1; i >= 0; i-- {
		c := expr[i]
		if c == ')' {
			parenDepth++
		} else if c == '(' {
			parenDepth--
		} else if parenDepth == 0 && (c == '+' || c == '-') {
			// Make sure it's an operator, not a sign
			if i > 0 {
				prev := expr[i-1]
				if prev != '*' && prev != '/' && prev != '+' && prev != '-' && prev != '^' {
					lastOp = i
					lastOpChar = c
					break
				}
			}
		}
	}

	if lastOp > 0 {
		left, err := evaluateAddSub(expr[:lastOp])
		if err != nil {
			return 0, err
		}
		right, err := evaluateMulDiv(expr[lastOp+1:])
		if err != nil {
			return 0, err
		}

		if lastOpChar == '+' {
			return left + right, nil
		}
		return left - right, nil
	}

	return evaluateMulDiv(expr)
}

func evaluateMulDiv(expr string) (float64, error) {
	// Find the rightmost * or /
	lastOp := -1
	lastOpChar := byte(0)
	parenDepth := 0

	for i := len(expr) - 1; i >= 0; i-- {
		c := expr[i]
		if c == ')' {
			parenDepth++
		} else if c == '(' {
			parenDepth--
		} else if parenDepth == 0 && (c == '*' || c == '/') {
			lastOp = i
			lastOpChar = c
			break
		}
	}

	if lastOp > 0 {
		left, err := evaluateMulDiv(expr[:lastOp])
		if err != nil {
			return 0, err
		}
		right, err := evaluatePow(expr[lastOp+1:])
		if err != nil {
			return 0, err
		}

		if lastOpChar == '*' {
			return left * right, nil
		}
		if right == 0 {
			return 0, fmt.Errorf("division by zero")
		}
		return left / right, nil
	}

	return evaluatePow(expr)
}

func evaluatePow(expr string) (float64, error) {
	// Find ^
	idx := strings.Index(expr, "^")
	if idx > 0 {
		base, err := evaluatePow(expr[:idx])
		if err != nil {
			return 0, err
		}
		exp, err := evaluatePow(expr[idx+1:])
		if err != nil {
			return 0, err
		}
		return math.Pow(base, exp), nil
	}

	return evaluateNumber(expr)
}

func evaluateNumber(expr string) (float64, error) {
	expr = strings.TrimSpace(expr)

	// Handle math functions
	if strings.HasPrefix(expr, "sqrt(") && strings.HasSuffix(expr, ")") {
		inner := expr[5 : len(expr)-1]
		val, err := evaluateMath(inner)
		if err != nil {
			return 0, err
		}
		return math.Sqrt(val), nil
	}
	if strings.HasPrefix(expr, "abs(") && strings.HasSuffix(expr, ")") {
		inner := expr[4 : len(expr)-1]
		val, err := evaluateMath(inner)
		if err != nil {
			return 0, err
		}
		return math.Abs(val), nil
	}
	if strings.HasPrefix(expr, "floor(") && strings.HasSuffix(expr, ")") {
		inner := expr[6 : len(expr)-1]
		val, err := evaluateMath(inner)
		if err != nil {
			return 0, err
		}
		return math.Floor(val), nil
	}
	if strings.HasPrefix(expr, "ceil(") && strings.HasSuffix(expr, ")") {
		inner := expr[5 : len(expr)-1]
		val, err := evaluateMath(inner)
		if err != nil {
			return 0, err
		}
		return math.Ceil(val), nil
	}
	if strings.HasPrefix(expr, "round(") && strings.HasSuffix(expr, ")") {
		inner := expr[6 : len(expr)-1]
		val, err := evaluateMath(inner)
		if err != nil {
			return 0, err
		}
		return math.Round(val), nil
	}
	if strings.HasPrefix(expr, "sin(") && strings.HasSuffix(expr, ")") {
		inner := expr[4 : len(expr)-1]
		val, err := evaluateMath(inner)
		if err != nil {
			return 0, err
		}
		return math.Sin(val), nil
	}
	if strings.HasPrefix(expr, "cos(") && strings.HasSuffix(expr, ")") {
		inner := expr[4 : len(expr)-1]
		val, err := evaluateMath(inner)
		if err != nil {
			return 0, err
		}
		return math.Cos(val), nil
	}
	if strings.HasPrefix(expr, "log(") && strings.HasSuffix(expr, ")") {
		inner := expr[4 : len(expr)-1]
		val, err := evaluateMath(inner)
		if err != nil {
			return 0, err
		}
		return math.Log(val), nil
	}
	if strings.HasPrefix(expr, "log10(") && strings.HasSuffix(expr, ")") {
		inner := expr[6 : len(expr)-1]
		val, err := evaluateMath(inner)
		if err != nil {
			return 0, err
		}
		return math.Log10(val), nil
	}

	// Handle constants
	if expr == "pi" || expr == "PI" {
		return math.Pi, nil
	}
	if expr == "e" || expr == "E" {
		return math.E, nil
	}

	// Parse number
	return strconv.ParseFloat(expr, 64)
}

func evaluateComparison(expr string) (bool, error) {
	// Handle comparison operators
	operators := []string{">=", "<=", "!=", "==", ">", "<"}
	for _, op := range operators {
		if idx := strings.Index(expr, op); idx >= 0 {
			left := strings.TrimSpace(expr[:idx])
			right := strings.TrimSpace(expr[idx+len(op):])

			// Try numeric comparison
			leftNum, leftErr := strconv.ParseFloat(left, 64)
			rightNum, rightErr := strconv.ParseFloat(right, 64)

			if leftErr == nil && rightErr == nil {
				switch op {
				case "==":
					return leftNum == rightNum, nil
				case "!=":
					return leftNum != rightNum, nil
				case ">":
					return leftNum > rightNum, nil
				case "<":
					return leftNum < rightNum, nil
				case ">=":
					return leftNum >= rightNum, nil
				case "<=":
					return leftNum <= rightNum, nil
				}
			}

			// String comparison
			switch op {
			case "==":
				return left == right, nil
			case "!=":
				return left != right, nil
			}
		}
	}

	// Handle logical operators
	if idx := strings.Index(expr, "&&"); idx >= 0 {
		left, err := evaluateComparison(expr[:idx])
		if err != nil {
			return false, err
		}
		right, err := evaluateComparison(expr[idx+2:])
		if err != nil {
			return false, err
		}
		return left && right, nil
	}

	if idx := strings.Index(expr, "||"); idx >= 0 {
		left, err := evaluateComparison(expr[:idx])
		if err != nil {
			return false, err
		}
		right, err := evaluateComparison(expr[idx+2:])
		if err != nil {
			return false, err
		}
		return left || right, nil
	}

	return false, fmt.Errorf("invalid comparison expression")
}

// MathNode performs math operations
type MathNode struct{}

func (n *MathNode) Type() string {
	return "logic.math"
}

func (n *MathNode) Execute(ctx context.Context, execCtx *core.ExecutionContext) (map[string]interface{}, error) {
	config := execCtx.Config
	input := execCtx.Input

	operation := core.GetString(config, "operation", "add")

	a := getNumber(config, input, "a")
	b := getNumber(config, input, "b")

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
			return nil, fmt.Errorf("division by zero")
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
	case "floor":
		result = math.Floor(a)
	case "ceil":
		result = math.Ceil(a)
	case "round":
		result = math.Round(a)
	case "min":
		result = math.Min(a, b)
	case "max":
		result = math.Max(a, b)
	case "random":
		// Note: This is deterministic for reproducibility in workflows
		// Use crypto/rand for true randomness
		result = math.Mod(a*1103515245+12345, math.Pow(2, 31)) / math.Pow(2, 31)
	default:
		result = a + b
	}

	return map[string]interface{}{
		"result":    result,
		"operation": operation,
		"a":         a,
		"b":         b,
	}, nil
}

func getNumber(config map[string]interface{}, input map[string]interface{}, key string) float64 {
	if val, ok := config[key]; ok {
		return toFloat(val)
	}
	if val, ok := input[key]; ok {
		return toFloat(val)
	}
	return 0
}

// Note: ExpressionNode and MathNode are registered in logic/init.go
