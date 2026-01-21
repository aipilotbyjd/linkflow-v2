package logic

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/adapters/worker/executor"
	wtypes "github.com/linkflow-ai/linkflow/internal/adapters/worker/types"
	"github.com/linkflow-ai/linkflow/internal/shared/types"
)

// ExecuteWorkflowNode executes another workflow as a sub-workflow
type ExecuteWorkflowNode struct{}

// NewExecuteWorkflowNode creates a new execute workflow node
func NewExecuteWorkflowNode() *ExecuteWorkflowNode {
	return &ExecuteWorkflowNode{}
}

// Execute runs the sub-workflow and returns its output
func (n *ExecuteWorkflowNode) Execute(ctx context.Context, runtime *executor.Runtime, node map[string]interface{}) (types.JSON, error) {
	params, _ := node["parameters"].(map[string]interface{})
	
	workflowIDStr, _ := params["workflow_id"].(string)
	if workflowIDStr == "" {
		return nil, fmt.Errorf("workflow_id is required")
	}

	workflowID, err := uuid.Parse(workflowIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid workflow_id: %w", err)
	}

	// Get input data to pass to sub-workflow
	inputData := runtime.GetInputData()
	passInput, _ := params["pass_input"].(bool)
	customInput, _ := params["input"].(map[string]interface{})

	var subWorkflowInput map[string]interface{}
	if passInput {
		subWorkflowInput = inputData
	} else if customInput != nil {
		subWorkflowInput = customInput
	} else {
		subWorkflowInput = make(map[string]interface{})
	}

	// Execute mode: sync or async
	mode, _ := params["mode"].(string)
	if mode == "" {
		mode = "sync"
	}

	// Get wait for completion setting
	waitForCompletion := true
	if wait, ok := params["wait_for_completion"].(bool); ok {
		waitForCompletion = wait
	}

	// Execute sub-workflow through runtime
	result, err := runtime.ExecuteSubWorkflow(ctx, workflowID, subWorkflowInput, waitForCompletion)
	if err != nil {
		return nil, fmt.Errorf("sub-workflow execution failed: %w", err)
	}

	return types.JSON{
		"execution_id": result.ExecutionID,
		"status":       result.Status,
		"output":       result.Output,
		"started_at":   result.StartedAt,
		"completed_at": result.CompletedAt,
	}, nil
}

// Metadata returns node metadata
func (n *ExecuteWorkflowNode) Metadata() wtypes.NodeMetadata {
	return wtypes.NodeMetadata{
		Type:        "logic.execute_workflow",
		Name:        "Execute Workflow",
		Description: "Execute another workflow as a sub-workflow",
		Category:    "logic",
		Version:     "1.0.0",
		Inputs:      []wtypes.NodePort{{Name: "main", Type: "any"}},
		Outputs:     []wtypes.NodePort{{Name: "main", Type: "any"}, {Name: "error", Type: "any"}},
		Parameters: []wtypes.NodeParameter{
			{Name: "workflow_id", Type: "string", Description: "ID of the workflow to execute", Required: true},
			{Name: "mode", Type: "select", Description: "Execution mode", Default: "sync", Options: []wtypes.ParamOption{{Value: "sync", Name: "Synchronous"}, {Value: "async", Name: "Asynchronous"}}},
			{Name: "pass_input", Type: "boolean", Description: "Pass current input to sub-workflow", Default: true},
			{Name: "input", Type: "json", Description: "Custom input data for sub-workflow"},
			{Name: "wait_for_completion", Type: "boolean", Description: "Wait for sub-workflow to complete", Default: true},
		},
	}
}
