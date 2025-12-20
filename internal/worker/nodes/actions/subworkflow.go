package actions

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/domain/models"
	"github.com/linkflow-ai/linkflow/internal/pkg/queue"
	"github.com/linkflow-ai/linkflow/internal/worker/nodes"
	"github.com/redis/go-redis/v9"
)

// SubWorkflowNode executes another workflow as part of the current workflow
type SubWorkflowNode struct {
	queueClient *queue.Client
	redisClient *redis.Client
}

func NewSubWorkflowNode(queueClient *queue.Client, redisClient *redis.Client) *SubWorkflowNode {
	return &SubWorkflowNode{
		queueClient: queueClient,
		redisClient: redisClient,
	}
}

func (n *SubWorkflowNode) Type() string {
	return "action.sub_workflow"
}

func (n *SubWorkflowNode) Execute(ctx context.Context, execCtx *nodes.ExecutionContext) (map[string]interface{}, error) {
	config := execCtx.Config
	input := execCtx.Input

	workflowIDStr := getString(config, "workflowId", "")
	if workflowIDStr == "" {
		return nil, fmt.Errorf("workflowId is required")
	}

	workflowID, err := uuid.Parse(workflowIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid workflowId: %w", err)
	}

	mode := getString(config, "mode", "wait") // wait, fire_and_forget
	timeout := getInt(config, "timeout", 300) // seconds
	passInput := getBool(config, "passInput", true)

	// Prepare input data for sub-workflow
	var inputData models.JSON
	if passInput {
		inputData = input
	}
	if customInput, ok := config["inputData"].(map[string]interface{}); ok {
		if inputData == nil {
			inputData = make(models.JSON)
		}
		for k, v := range customInput {
			inputData[k] = v
		}
	}

	// Generate correlation ID for tracking
	correlationID := uuid.New().String()

	// Queue the sub-workflow execution
	payload := queue.WorkflowExecutionPayload{
		WorkflowID:  workflowID,
		WorkspaceID: execCtx.WorkspaceID,
		TriggerType: models.TriggerSubWorkflow,
		InputData:   inputData,
		TriggerData: models.JSON{
			"parentExecutionId": execCtx.ExecutionID.String(),
			"correlationId":     correlationID,
		},
	}

	taskInfo, err := n.queueClient.EnqueueWorkflowExecution(ctx, payload)
	if err != nil {
		return nil, fmt.Errorf("failed to queue sub-workflow: %w", err)
	}

	if mode == "fire_and_forget" {
		return map[string]interface{}{
			"queued":        true,
			"workflowId":    workflowID.String(),
			"taskId":        taskInfo.ID,
			"correlationId": correlationID,
		}, nil
	}

	// Wait for result
	result, err := n.waitForResult(ctx, correlationID, time.Duration(timeout)*time.Second)
	if err != nil {
		return nil, fmt.Errorf("sub-workflow execution failed: %w", err)
	}

	return result, nil
}

func (n *SubWorkflowNode) waitForResult(ctx context.Context, correlationID string, timeout time.Duration) (map[string]interface{}, error) {
	resultKey := fmt.Sprintf("subworkflow:result:%s", correlationID)

	deadline := time.Now().Add(timeout)
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
			if time.Now().After(deadline) {
				return nil, fmt.Errorf("timeout waiting for sub-workflow result")
			}

			data, err := n.redisClient.Get(ctx, resultKey).Bytes()
			if err == redis.Nil {
				continue
			}
			if err != nil {
				return nil, err
			}

			var result SubWorkflowResult
			if err := json.Unmarshal(data, &result); err != nil {
				return nil, err
			}

			// Clean up
			n.redisClient.Del(ctx, resultKey)

			if result.Status == "failed" {
				return nil, fmt.Errorf("sub-workflow failed: %s", result.Error)
			}

			return result.Output, nil
		}
	}
}

type SubWorkflowResult struct {
	Status  string                 `json:"status"`
	Output  map[string]interface{} `json:"output"`
	Error   string                 `json:"error,omitempty"`
}

// PublishSubWorkflowResult publishes the result of a sub-workflow execution
// This should be called by the executor when a sub-workflow completes
func PublishSubWorkflowResult(ctx context.Context, redisClient *redis.Client, correlationID string, status string, output map[string]interface{}, errMsg string) error {
	resultKey := fmt.Sprintf("subworkflow:result:%s", correlationID)

	result := SubWorkflowResult{
		Status: status,
		Output: output,
		Error:  errMsg,
	}

	data, err := json.Marshal(result)
	if err != nil {
		return err
	}

	return redisClient.Set(ctx, resultKey, data, 5*time.Minute).Err()
}

// ExecuteWorkflowNode is similar but allows dynamic workflow selection
type ExecuteWorkflowNode struct {
	queueClient *queue.Client
	redisClient *redis.Client
}

func NewExecuteWorkflowNode(queueClient *queue.Client, redisClient *redis.Client) *ExecuteWorkflowNode {
	return &ExecuteWorkflowNode{
		queueClient: queueClient,
		redisClient: redisClient,
	}
}

func (n *ExecuteWorkflowNode) Type() string {
	return "action.execute_workflow"
}

func (n *ExecuteWorkflowNode) Execute(ctx context.Context, execCtx *nodes.ExecutionContext) (map[string]interface{}, error) {
	config := execCtx.Config
	input := execCtx.Input

	// Workflow ID can come from config or input
	workflowIDStr := getString(config, "workflowId", "")
	if workflowIDStr == "" {
		if id, ok := input["workflowId"].(string); ok {
			workflowIDStr = id
		}
	}

	if workflowIDStr == "" {
		return nil, fmt.Errorf("workflowId is required")
	}

	workflowID, err := uuid.Parse(workflowIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid workflowId: %w", err)
	}

	mode := getString(config, "mode", "wait")
	timeout := getInt(config, "timeout", 300)

	// Get input data
	var inputData models.JSON
	if data, ok := input["inputData"].(map[string]interface{}); ok {
		inputData = data
	} else if data, ok := config["inputData"].(map[string]interface{}); ok {
		inputData = data
	}

	correlationID := uuid.New().String()

	payload := queue.WorkflowExecutionPayload{
		WorkflowID:  workflowID,
		WorkspaceID: execCtx.WorkspaceID,
		TriggerType: models.TriggerSubWorkflow,
		InputData:   inputData,
		TriggerData: models.JSON{
			"parentExecutionId": execCtx.ExecutionID.String(),
			"correlationId":     correlationID,
		},
	}

	taskInfo, err := n.queueClient.EnqueueWorkflowExecution(ctx, payload)
	if err != nil {
		return nil, fmt.Errorf("failed to queue workflow: %w", err)
	}

	if mode == "fire_and_forget" {
		return map[string]interface{}{
			"queued":        true,
			"workflowId":    workflowID.String(),
			"taskId":        taskInfo.ID,
			"correlationId": correlationID,
		}, nil
	}

	// Wait for result using same mechanism as SubWorkflowNode
	resultKey := fmt.Sprintf("subworkflow:result:%s", correlationID)
	deadline := time.Now().Add(time.Duration(timeout) * time.Second)
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
			if time.Now().After(deadline) {
				return nil, fmt.Errorf("timeout waiting for workflow result")
			}

			data, err := n.redisClient.Get(ctx, resultKey).Bytes()
			if err == redis.Nil {
				continue
			}
			if err != nil {
				return nil, err
			}

			var result SubWorkflowResult
			if err := json.Unmarshal(data, &result); err != nil {
				return nil, err
			}

			n.redisClient.Del(ctx, resultKey)

			if result.Status == "failed" {
				return nil, fmt.Errorf("workflow failed: %s", result.Error)
			}

			return result.Output, nil
		}
	}
}

// Helper functions - use the ones from code.go (same package)
