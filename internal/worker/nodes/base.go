package nodes

import (
	"context"
	"time"

	"github.com/linkflow-ai/linkflow/internal/pkg/queue"
	"github.com/redis/go-redis/v9"
)

// Trigger nodes

type ManualTrigger struct{}

func (n *ManualTrigger) Type() string { return "trigger.manual" }

func (n *ManualTrigger) Execute(ctx context.Context, execCtx *ExecutionContext) (map[string]interface{}, error) {
	return map[string]interface{}{
		"triggered": true,
		"input":     execCtx.Input["$input"],
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}, nil
}

type WebhookTrigger struct{}

func (n *WebhookTrigger) Type() string { return "trigger.webhook" }

func (n *WebhookTrigger) Execute(ctx context.Context, execCtx *ExecutionContext) (map[string]interface{}, error) {
	return map[string]interface{}{
		"triggered": true,
		"data":      execCtx.Input["$input"],
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}, nil
}

type ScheduleTrigger struct{}

func (n *ScheduleTrigger) Type() string { return "trigger.schedule" }

func (n *ScheduleTrigger) Execute(ctx context.Context, execCtx *ExecutionContext) (map[string]interface{}, error) {
	return map[string]interface{}{
		"triggered": true,
		"input":     execCtx.Input["$input"],
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}, nil
}

// SetVariableNode sets a variable in the execution context
type SetVariableNode struct{}

func (n *SetVariableNode) Type() string { return "action.set" }

func (n *SetVariableNode) Execute(ctx context.Context, execCtx *ExecutionContext) (map[string]interface{}, error) {
	name, _ := execCtx.Config["name"].(string)
	value := execCtx.Config["value"]
	execCtx.Variables[name] = value
	return map[string]interface{}{"name": name, "value": value}, nil
}

// RespondNode sends response back for webhook triggers
type RespondNode struct{}

func (n *RespondNode) Type() string { return "action.respond" }

func (n *RespondNode) Execute(ctx context.Context, execCtx *ExecutionContext) (map[string]interface{}, error) {
	statusCode := getIntBase(execCtx.Config, "statusCode", 200)
	body := execCtx.Config["body"]
	headers := getMapBase(execCtx.Config, "headers")

	return map[string]interface{}{
		"responded":  true,
		"statusCode": statusCode,
		"body":       body,
		"headers":    headers,
	}, nil
}

// SubWorkflowNode executes another workflow
type SubWorkflowNode struct {
	queueClient *queue.Client
	redisClient *redis.Client
}

func (n *SubWorkflowNode) Type() string { return "action.sub_workflow" }

func (n *SubWorkflowNode) Execute(ctx context.Context, execCtx *ExecutionContext) (map[string]interface{}, error) {
	// Implementation delegated to actions package
	return map[string]interface{}{"error": "sub_workflow requires dependencies"}, nil
}

// ExecuteWorkflowNode executes a workflow dynamically
type ExecuteWorkflowNode struct {
	queueClient *queue.Client
	redisClient *redis.Client
}

func (n *ExecuteWorkflowNode) Type() string { return "action.execute_workflow" }

func (n *ExecuteWorkflowNode) Execute(ctx context.Context, execCtx *ExecutionContext) (map[string]interface{}, error) {
	return map[string]interface{}{"error": "execute_workflow requires dependencies"}, nil
}

// FunctionNode executes a function on items
type FunctionNode struct{}

func (n *FunctionNode) Type() string { return "action.function" }

func (n *FunctionNode) Execute(ctx context.Context, execCtx *ExecutionContext) (map[string]interface{}, error) {
	return map[string]interface{}{"result": execCtx.Input}, nil
}

// TransformNode transforms data
type TransformNode struct{}

func (n *TransformNode) Type() string { return "action.transform" }

func (n *TransformNode) Execute(ctx context.Context, execCtx *ExecutionContext) (map[string]interface{}, error) {
	return map[string]interface{}{"items": execCtx.Input["$json"]}, nil
}

// Helper functions for base nodes
func getStringBase(m map[string]interface{}, key, defaultVal string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return defaultVal
}

func getIntBase(m map[string]interface{}, key string, defaultVal int) int {
	if v, ok := m[key].(float64); ok {
		return int(v)
	}
	if v, ok := m[key].(int); ok {
		return v
	}
	return defaultVal
}

func getBoolBase(m map[string]interface{}, key string, defaultVal bool) bool {
	if v, ok := m[key].(bool); ok {
		return v
	}
	return defaultVal
}

func getMapBase(m map[string]interface{}, key string) map[string]interface{} {
	if v, ok := m[key].(map[string]interface{}); ok {
		return v
	}
	return make(map[string]interface{})
}
