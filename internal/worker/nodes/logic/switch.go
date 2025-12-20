package logic

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/linkflow-ai/linkflow/internal/worker/core"
)

type SwitchNode struct{}

func NewSwitchNode() *SwitchNode {
	return &SwitchNode{}
}

func (n *SwitchNode) Type() string {
	return "logic.switch"
}

func (n *SwitchNode) Execute(ctx context.Context, execCtx *core.ExecutionContext) (map[string]interface{}, error) {
	mode, _ := execCtx.Config["mode"].(string)
	if mode == "" {
		mode = "rules"
	}

	var matchedCase string
	var matchedIndex int = -1

	switch mode {
	case "expression":
		matchedCase, matchedIndex = n.evaluateExpressionMode(execCtx)
	default:
		matchedCase, matchedIndex = n.evaluateRulesMode(execCtx)
	}

	if matchedCase == "" {
		matchedCase = "default"
	}

	return map[string]interface{}{
		"case":       matchedCase,
		"caseIndex":  matchedIndex,
		"data":       execCtx.Input["$json"],
		"matched":    matchedIndex >= 0,
		"outputIndex": matchedIndex,
	}, nil
}

func (n *SwitchNode) evaluateExpressionMode(execCtx *core.ExecutionContext) (string, int) {
	value := n.resolveValue(execCtx.Config["value"], execCtx.Input)
	cases, _ := execCtx.Config["cases"].([]interface{})

	for i, c := range cases {
		caseConfig, ok := c.(map[string]interface{})
		if !ok {
			continue
		}
		caseValue := n.resolveValue(caseConfig["value"], execCtx.Input)
		caseName, _ := caseConfig["name"].(string)
		if caseName == "" {
			caseName = fmt.Sprintf("case_%d", i)
		}

		if fmt.Sprintf("%v", value) == fmt.Sprintf("%v", caseValue) {
			return caseName, i
		}
	}
	return "default", -1
}

func (n *SwitchNode) evaluateRulesMode(execCtx *core.ExecutionContext) (string, int) {
	rules, _ := execCtx.Config["rules"].([]interface{})

	for i, r := range rules {
		rule, ok := r.(map[string]interface{})
		if !ok {
			continue
		}

		conditions, _ := rule["conditions"].([]interface{})
		combineWith, _ := rule["combineWith"].(string)
		if combineWith == "" {
			combineWith = "and"
		}
		outputName, _ := rule["output"].(string)
		if outputName == "" {
			outputName = fmt.Sprintf("output_%d", i)
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
			return outputName, i
		}
	}

	return "default", -1
}

func (n *SwitchNode) evaluateCondition(cond map[string]interface{}, input map[string]interface{}) bool {
	leftValue := n.resolveValue(cond["leftValue"], input)
	rightValue := n.resolveValue(cond["rightValue"], input)
	operator, _ := cond["operator"].(string)

	switch operator {
	case "equal", "equals", "==":
		return fmt.Sprintf("%v", leftValue) == fmt.Sprintf("%v", rightValue)
	case "notEqual", "!=":
		return fmt.Sprintf("%v", leftValue) != fmt.Sprintf("%v", rightValue)
	case "greater", ">":
		return toFloat(leftValue) > toFloat(rightValue)
	case "greaterEqual", ">=":
		return toFloat(leftValue) >= toFloat(rightValue)
	case "less", "<":
		return toFloat(leftValue) < toFloat(rightValue)
	case "lessEqual", "<=":
		return toFloat(leftValue) <= toFloat(rightValue)
	case "contains":
		return strings.Contains(fmt.Sprintf("%v", leftValue), fmt.Sprintf("%v", rightValue))
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
	default:
		return false
	}
}

func (n *SwitchNode) resolveValue(value interface{}, input map[string]interface{}) interface{} {
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
