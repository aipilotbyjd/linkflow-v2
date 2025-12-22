package executor

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/domain/models"
	"github.com/linkflow-ai/linkflow/internal/domain/services"
	"github.com/linkflow-ai/linkflow/internal/pkg/queue"
	"github.com/linkflow-ai/linkflow/internal/worker/cache"
	"github.com/linkflow-ai/linkflow/internal/worker/events"
	"github.com/linkflow-ai/linkflow/internal/worker/processor"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

// Executor handles workflow execution jobs from the queue
type Executor struct {
	processor     *processor.Processor
	executionSvc  *services.ExecutionService
	credentialSvc *services.CredentialService
	workflowSvc   *services.WorkflowService
	publisher     *events.Publisher
	cancellation  *processor.CancellationManager
	credCache     *cache.CredentialCache
	redis         *redis.Client
}

// ExecutorConfig configures the executor
type ExecutorConfig struct {
	MaxParallelNodes   int
	DefaultNodeTimeout time.Duration
	WorkflowTimeout    time.Duration
	EnableCaching      bool
}

// DefaultExecutorConfig returns default configuration
func DefaultExecutorConfig() ExecutorConfig {
	return ExecutorConfig{
		MaxParallelNodes:   10,
		DefaultNodeTimeout: 5 * time.Minute,
		WorkflowTimeout:    30 * time.Minute,
		EnableCaching:      true,
	}
}

// New creates a new executor
func New(
	proc *processor.Processor,
	executionSvc *services.ExecutionService,
	credentialSvc *services.CredentialService,
	workflowSvc *services.WorkflowService,
	publisher *events.Publisher,
	cancellation *processor.CancellationManager,
	credCache *cache.CredentialCache,
	redisClient *redis.Client,
) *Executor {
	return &Executor{
		processor:     proc,
		executionSvc:  executionSvc,
		credentialSvc: credentialSvc,
		workflowSvc:   workflowSvc,
		publisher:     publisher,
		cancellation:  cancellation,
		credCache:     credCache,
		redis:         redisClient,
	}
}

// Execute handles a workflow execution job
func (e *Executor) Execute(ctx context.Context, payload queue.WorkflowExecutionPayload) error {
	return e.ExecuteWithOptions(ctx, payload, DefaultExecutorConfig())
}

// ExecuteWithOptions handles a workflow execution job with custom options
func (e *Executor) ExecuteWithOptions(ctx context.Context, payload queue.WorkflowExecutionPayload, cfg ExecutorConfig) error {
	// Create execution record
	execution, err := e.executionSvc.Create(ctx, services.CreateExecutionInput{
		WorkflowID:  payload.WorkflowID,
		WorkspaceID: payload.WorkspaceID,
		TriggeredBy: payload.TriggeredBy,
		TriggerType: payload.TriggerType,
		TriggerData: payload.TriggerData,
		InputData:   payload.InputData,
	})
	if err != nil {
		return fmt.Errorf("failed to create execution: %w", err)
	}

	log.Info().
		Str("execution_id", execution.ID.String()).
		Str("workflow_id", payload.WorkflowID.String()).
		Str("workspace_id", payload.WorkspaceID.String()).
		Msg("Starting workflow execution")

	// Register for cancellation
	execCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	if e.cancellation != nil {
		e.cancellation.Register(execution.ID, cancel)
		defer e.cancellation.Unregister(execution.ID)
	}

	// Get workflow
	workflow, err := e.workflowSvc.GetByID(ctx, payload.WorkflowID)
	if err != nil {
		e.handleExecutionError(ctx, execution, payload, "Workflow not found", nil)
		return fmt.Errorf("workflow not found: %w", err)
	}

	// Parse workflow definition
	workflowDef, err := processor.ParseWorkflow(workflow)
	if err != nil {
		e.handleExecutionError(ctx, execution, payload, "Invalid workflow definition", nil)
		return fmt.Errorf("invalid workflow definition: %w", err)
	}

	// Start execution
	if err := e.executionSvc.Start(ctx, execution.ID); err != nil {
		return err
	}

	// Publish execution started event
	e.publishExecutionStarted(ctx, payload.WorkspaceID, payload.WorkflowID, execution.ID, payload.TriggerType)

	// Prepare input
	input := processor.Input(payload.InputData)
	if input == nil {
		input = make(processor.Input)
	}

	// Add trigger data
	if payload.TriggerData != nil {
		input["$trigger"] = payload.TriggerData
	}

	// Build execution options
	opts := processor.ExecutionOptions{
		MaxParallelNodes:   cfg.MaxParallelNodes,
		DefaultNodeTimeout: cfg.DefaultNodeTimeout,
		WorkflowTimeout:    cfg.WorkflowTimeout,
		EnableCaching:      cfg.EnableCaching,
	}

	// Create credential resolver
	getCredential := e.createCredentialResolver(ctx)

	// Execute workflow via processor
	result, err := e.processor.Execute(
		execCtx,
		workflowDef,
		input,
		opts,
		execution.ID,
		getCredential,
		e.publisher,
	)

	// Handle result
	if err != nil {
		e.handleExecutionError(ctx, execution, payload, err.Error(), result)
		return err
	}

	// Check for processor-level failure
	if result.Status == processor.StatusFailed {
		nodeID := &result.ErrorNodeID
		if result.ErrorNodeID == "" {
			nodeID = nil
		}
		_ = e.executionSvc.Fail(ctx, execution.ID, result.Error, nodeID)
		e.publishExecutionFailed(ctx, payload.WorkspaceID, payload.WorkflowID, execution.ID, result.Error, nodeID)
		return fmt.Errorf("workflow failed: %s", result.Error)
	}

	if result.Status == processor.StatusCancelled {
		_ = e.executionSvc.Fail(ctx, execution.ID, "Execution cancelled", nil)
		e.publishExecutionFailed(ctx, payload.WorkspaceID, payload.WorkflowID, execution.ID, "Execution cancelled", nil)
		return nil
	}

	// Complete execution
	outputJSON := models.JSON(result.Output)
	if err := e.executionSvc.Complete(ctx, execution.ID, outputJSON); err != nil {
		return err
	}

	e.publishExecutionCompleted(ctx, payload.WorkspaceID, payload.WorkflowID, execution.ID, result.Duration.Milliseconds(), result.NodesExecuted)

	// Handle sub-workflow result publishing
	e.publishSubWorkflowResult(ctx, payload, result)

	log.Info().
		Str("execution_id", execution.ID.String()).
		Int64("duration_ms", result.Duration.Milliseconds()).
		Int("nodes_executed", result.NodesExecuted).
		Msg("Workflow execution completed")

	return nil
}

// Preview performs a dry-run of the workflow
func (e *Executor) Preview(ctx context.Context, workflowID uuid.UUID) (*processor.PreviewResult, error) {
	workflow, err := e.workflowSvc.GetByID(ctx, workflowID)
	if err != nil {
		return nil, fmt.Errorf("workflow not found: %w", err)
	}

	workflowDef, err := processor.ParseWorkflow(workflow)
	if err != nil {
		return nil, fmt.Errorf("invalid workflow definition: %w", err)
	}

	return e.processor.Preview(ctx, workflowDef, nil)
}

// Cancel cancels a running execution
func (e *Executor) Cancel(ctx context.Context, executionID uuid.UUID, reason, requestedBy string) error {
	if e.cancellation == nil {
		return fmt.Errorf("cancellation not supported")
	}
	return e.cancellation.Cancel(ctx, executionID, reason, requestedBy)
}

// GetProgress returns execution progress
func (e *Executor) GetProgress(ctx context.Context, executionID uuid.UUID) (map[string]interface{}, error) {
	return processor.GetProgressByID(ctx, e.redis, executionID)
}

func (e *Executor) createCredentialResolver(ctx context.Context) func(uuid.UUID) (*models.CredentialData, error) {
	return func(credID uuid.UUID) (*models.CredentialData, error) {
		// Try cache first
		if e.credCache != nil {
			if data, ok := e.credCache.Get(ctx, credID); ok {
				return data, nil
			}
		}

		// Fetch from service
		_, data, err := e.credentialSvc.GetDecrypted(ctx, credID)
		if err != nil {
			return nil, err
		}

		// Cache the result
		if e.credCache != nil {
			_ = e.credCache.Set(ctx, credID, data)
		}

		return data, nil
	}
}

func (e *Executor) handleExecutionError(ctx context.Context, execution *models.Execution, payload queue.WorkflowExecutionPayload, errMsg string, result *processor.Result) {
	var nodeID *string
	if result != nil && result.ErrorNodeID != "" {
		nodeID = &result.ErrorNodeID
	}

	_ = e.executionSvc.Fail(ctx, execution.ID, errMsg, nodeID)
	e.publishExecutionFailed(ctx, payload.WorkspaceID, payload.WorkflowID, execution.ID, errMsg, nodeID)
}

func (e *Executor) publishSubWorkflowResult(ctx context.Context, payload queue.WorkflowExecutionPayload, result *processor.Result) {
	if payload.TriggerData == nil {
		return
	}

	correlationID, ok := payload.TriggerData["correlationId"].(string)
	if !ok || correlationID == "" {
		return
	}

	// Publish result for sub-workflow
	status := "completed"
	errMsg := ""
	if result.Status != processor.StatusCompleted {
		status = "failed"
		errMsg = result.Error
	}

	resultData := map[string]interface{}{
		"status": status,
		"output": result.Output,
		"error":  errMsg,
	}

	data, _ := json.Marshal(resultData)
	key := fmt.Sprintf("subworkflow:result:%s", correlationID)
	e.redis.Set(ctx, key, data, 5*time.Minute)
}

func (e *Executor) publishExecutionStarted(ctx context.Context, workspaceID, workflowID, executionID uuid.UUID, triggerType string) {
	if e.publisher != nil {
		_ = e.publisher.ExecutionStarted(ctx, workspaceID, workflowID, executionID, triggerType)
	}
}

func (e *Executor) publishExecutionCompleted(ctx context.Context, workspaceID, workflowID, executionID uuid.UUID, durationMs int64, nodesCompleted int) {
	if e.publisher != nil {
		_ = e.publisher.ExecutionCompleted(ctx, workspaceID, workflowID, executionID, durationMs, nodesCompleted)
	}
}

func (e *Executor) publishExecutionFailed(ctx context.Context, workspaceID, workflowID, executionID uuid.UUID, errorMsg string, errorNodeID *string) {
	if e.publisher != nil {
		_ = e.publisher.ExecutionFailed(ctx, workspaceID, workflowID, executionID, errorMsg, errorNodeID)
	}
}

// NodeExecutionError wraps a node-specific error
type NodeExecutionError struct {
	NodeID string
	Err    error
}

func (e *NodeExecutionError) Error() string {
	return fmt.Sprintf("node %s: %v", e.NodeID, e.Err)
}

// ExtractNodeID extracts node ID from an error if available
func ExtractNodeID(err error) *string {
	if nodeErr, ok := err.(*NodeExecutionError); ok {
		return &nodeErr.NodeID
	}
	return nil
}
