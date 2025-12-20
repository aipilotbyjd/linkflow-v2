package actions

import (
	"context"

	"github.com/linkflow-ai/linkflow/internal/worker/core"
)

func init() {
	// Register all action nodes
	core.Register(&HTTPRequestNode{}, core.NodeMeta{
		Name:        "HTTP Request",
		Description: "Make HTTP requests to external APIs",
		Category:    "actions",
		Icon:        "globe",
		Version:     "1.0.0",
	})

	core.Register(&CodeNode{}, core.NodeMeta{
		Name:        "Code",
		Description: "Execute JavaScript code",
		Category:    "actions",
		Icon:        "code",
		Version:     "1.0.0",
	})

	core.Register(&SetVariableNode{}, core.NodeMeta{
		Name:        "Set Variable",
		Description: "Set a workflow variable",
		Category:    "actions",
		Icon:        "variable",
		Version:     "1.0.0",
	})

	core.Register(&RespondNode{}, core.NodeMeta{
		Name:        "Respond to Webhook",
		Description: "Send response back to webhook caller",
		Category:    "actions",
		Icon:        "send",
		Version:     "1.0.0",
	})

	// Register sub-workflow node (needs dependencies)
	core.Register(&SubWorkflowNode{}, core.NodeMeta{
		Name:        "Sub-Workflow",
		Description: "Execute another workflow",
		Category:    "actions",
		Icon:        "git-branch",
		Version:     "1.0.0",
	})
}

// SetVariableNode sets a variable in the execution context
type SetVariableNode struct{}

func (n *SetVariableNode) Type() string { return "action.set" }

func (n *SetVariableNode) Execute(ctx context.Context, execCtx *core.ExecutionContext) (map[string]interface{}, error) {
	name := core.GetString(execCtx.Config, "name", "")
	value := execCtx.Config["value"]
	execCtx.Variables[name] = value
	return map[string]interface{}{"name": name, "value": value}, nil
}

// RespondNode sends response back for webhook triggers
type RespondNode struct{}

func (n *RespondNode) Type() string { return "action.respond" }

func (n *RespondNode) Execute(ctx context.Context, execCtx *core.ExecutionContext) (map[string]interface{}, error) {
	statusCode := core.GetInt(execCtx.Config, "statusCode", 200)
	body := execCtx.Config["body"]
	headers := core.GetMap(execCtx.Config, "headers")

	return map[string]interface{}{
		"responded":  true,
		"statusCode": statusCode,
		"body":       body,
		"headers":    headers,
	}, nil
}
