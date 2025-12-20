package logic

import (
	"context"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/linkflow-ai/linkflow/internal/worker/core"
)

type ConditionNode struct{}

func NewConditionNode() *ConditionNode {
	return &ConditionNode{}
}

func (n *ConditionNode) Type() string {
	return "logic.condition"
}

func (n *ConditionNode) Execute(ctx context.Context, execCtx *core.ExecutionContext) (map[string]interface{}, error) {
	conditions, _ := execCtx.Config["conditions"].([]interface{})
	combineWith, _ := execCtx.Config["combineWith"].(string)
	if combineWith == "" {
		combineWith = "and"
	}

	if len(conditions) == 0 {
		condition, _ := execCtx.Config["condition"].(map[string]interface{})
		if condition != nil {
			conditions = []interface{}{condition}
		}
	}

	results := make([]bool, 0, len(conditions))
	for _, cond := range conditions {
		condMap, ok := cond.(map[string]interface{})
		if !ok {
			continue
		}
		result := n.evaluateCondition(condMap, execCtx.Input)
		results = append(results, result)
	}

	var finalResult bool
	if combineWith == "or" {
		for _, r := range results {
			if r {
				finalResult = true
				break
			}
		}
	} else {
		finalResult = true
		for _, r := range results {
			if !r {
				finalResult = false
				break
			}
		}
	}

	branch := "false"
	if finalResult {
		branch = "true"
	}

	return map[string]interface{}{
		"result": finalResult,
		"branch": branch,
		"data":   execCtx.Input["$json"],
	}, nil
}

func (n *ConditionNode) evaluateCondition(cond map[string]interface{}, input map[string]interface{}) bool {
	leftValue := n.resolveValue(cond["leftValue"], input)
	rightValue := n.resolveValue(cond["rightValue"], input)
	operator, _ := cond["operator"].(string)

	switch operator {
	case "equal", "equals", "==", "eq":
		return fmt.Sprintf("%v", leftValue) == fmt.Sprintf("%v", rightValue)
	case "notEqual", "!=", "ne":
		return fmt.Sprintf("%v", leftValue) != fmt.Sprintf("%v", rightValue)
	case "greater", ">", "gt":
		return toFloat(leftValue) > toFloat(rightValue)
	case "greaterEqual", ">=", "gte":
		return toFloat(leftValue) >= toFloat(rightValue)
	case "less", "<", "lt":
		return toFloat(leftValue) < toFloat(rightValue)
	case "lessEqual", "<=", "lte":
		return toFloat(leftValue) <= toFloat(rightValue)
	case "contains":
		return strings.Contains(fmt.Sprintf("%v", leftValue), fmt.Sprintf("%v", rightValue))
	case "notContains":
		return !strings.Contains(fmt.Sprintf("%v", leftValue), fmt.Sprintf("%v", rightValue))
	case "startsWith":
		return strings.HasPrefix(fmt.Sprintf("%v", leftValue), fmt.Sprintf("%v", rightValue))
	case "endsWith":
		return strings.HasSuffix(fmt.Sprintf("%v", leftValue), fmt.Sprintf("%v", rightValue))
	case "regex", "matches":
		re, err := regexp.Compile(fmt.Sprintf("%v", rightValue))
		if err != nil {
			return false
		}
		return re.MatchString(fmt.Sprintf("%v", leftValue))
	case "isEmpty":
		return isEmpty(leftValue)
	case "isNotEmpty":
		return !isEmpty(leftValue)
	case "isTrue":
		return toBool(leftValue)
	case "isFalse":
		return !toBool(leftValue)
	case "isNull":
		return leftValue == nil
	case "isNotNull":
		return leftValue != nil
	case "in":
		return isIn(leftValue, rightValue)
	case "notIn":
		return !isIn(leftValue, rightValue)
	case "between":
		min, _ := cond["min"]
		max, _ := cond["max"]
		v := toFloat(leftValue)
		return v >= toFloat(min) && v <= toFloat(max)
	default:
		return false
	}
}

func (n *ConditionNode) resolveValue(value interface{}, input map[string]interface{}) interface{} {
	if value == nil {
		return nil
	}
	str, ok := value.(string)
	if !ok {
		return value
	}
	if strings.HasPrefix(str, "{{") && strings.HasSuffix(str, "}}") {
		path := strings.TrimSpace(str[2 : len(str)-2])
		return core.GetNestedValue(input, path)
	}
	return value
}



func toFloat(v interface{}) float64 {
	switch val := v.(type) {
	case int:
		return float64(val)
	case int64:
		return float64(val)
	case float64:
		return val
	case string:
		f, _ := strconv.ParseFloat(val, 64)
		return f
	}
	return 0
}

func toBool(v interface{}) bool {
	switch val := v.(type) {
	case bool:
		return val
	case string:
		return val != "" && val != "false" && val != "0"
	case int, int64, float64:
		return toFloat(v) != 0
	}
	return v != nil
}

func isEmpty(v interface{}) bool {
	if v == nil {
		return true
	}
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.String, reflect.Array, reflect.Slice, reflect.Map:
		return rv.Len() == 0
	}
	return false
}

func isIn(value interface{}, list interface{}) bool {
	arr, ok := list.([]interface{})
	if !ok {
		return false
	}
	strVal := fmt.Sprintf("%v", value)
	for _, item := range arr {
		if fmt.Sprintf("%v", item) == strVal {
			return true
		}
	}
	return false
}
