package executor

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/adapters/websocket"
	"github.com/linkflow-ai/linkflow/internal/core/domain/execution"
	"github.com/linkflow-ai/linkflow/internal/core/domain/workflow"
	"github.com/linkflow-ai/linkflow/internal/infrastructure/observability/logger"
)

const (
	MaxNodesPerExecution = 500
)

type Executor struct {
	workflowRepo  workflow.Repository
	executionRepo execution.Repository
	nodeExecRepo  execution.NodeExecutionRepository
	processor     *Processor
	usageTracker  *UsageTracker
	streamService *websocket.ExecutionStreamService
	logger        logger.Logger
}

func NewExecutor(
	workflowRepo workflow.Repository,
	executionRepo execution.Repository,
	nodeExecRepo execution.NodeExecutionRepository,
	processor *Processor,
	usageTracker *UsageTracker,
	streamService *websocket.ExecutionStreamService,
	log logger.Logger,
) *Executor {
	return &Executor{
		workflowRepo:  workflowRepo,
		executionRepo: executionRepo,
		nodeExecRepo:  nodeExecRepo,
		processor:     processor,
		usageTracker:  usageTracker,
		streamService: streamService,
		logger:        log,
	}
}

func (e *Executor) Execute(ctx context.Context, executionID uuid.UUID) error {
	exec, err := e.executionRepo.FindByID(ctx, executionID)
	if err != nil {
		return fmt.Errorf("execution not found: %w", err)
	}

	wf, err := e.workflowRepo.FindByID(ctx, exec.WorkflowID)
	if err != nil {
		return fmt.Errorf("workflow not found: %w", err)
	}

	// Pre-execution billing check
	if e.usageTracker != nil {
		estimatedOps := int64(len(wf.Nodes))
		if estimatedOps < 1 {
			estimatedOps = 1
		}
		if err := e.usageTracker.PreExecutionCheck(ctx, exec.WorkspaceID, estimatedOps); err != nil {
			if failErr := exec.Fail(err.Error(), nil); failErr != nil {
				e.logger.Error().Err(failErr).Msg("Failed to mark execution as failed during billing check")
			}
			if updateErr := e.executionRepo.Update(ctx, exec); updateErr != nil {
				e.logger.Error().Err(updateErr).Msg("Failed to update execution after billing check failure")
			}
			return fmt.Errorf("billing check failed: %w", err)
		}
	}

	if err := exec.Start(); err != nil {
		return fmt.Errorf("failed to start execution: %w", err)
	}
	if err := e.executionRepo.Update(ctx, exec); err != nil {
		return fmt.Errorf("failed to update execution: %w", err)
	}

	e.logger.Info().
		Str("execution_id", executionID.String()).
		Str("workflow_id", wf.ID.String()).
		Msg("Starting workflow execution")

	// Emit initial progress
	if e.streamService != nil {
		e.streamService.SendProgress(ctx, exec.WorkspaceID, websocket.ExecutionProgressData{
			ExecutionID: exec.ID,
			WorkflowID:  wf.ID,
			Status:      "running",
			Progress:    0,
			NodesTotal:  len(wf.Nodes),
			StartedAt:   *exec.StartedAt,
		})
	}

	runtime := NewRuntime(exec, wf, e.logger)

	err = e.executeWorkflow(ctx, runtime)

	if err != nil {
		if failErr := exec.Fail(err.Error(), nil); failErr != nil {
			e.logger.Error().Err(failErr).Str("execution_id", executionID.String()).Msg("Failed to mark execution as failed")
		}
		e.logger.Error().Err(err).
			Str("execution_id", executionID.String()).
			Msg("Workflow execution failed")
	} else {
		if completeErr := exec.Complete(runtime.GetOutputData()); completeErr != nil {
			e.logger.Error().Err(completeErr).Str("execution_id", executionID.String()).Msg("Failed to mark execution as completed")
			return fmt.Errorf("failed to complete execution: %w", completeErr)
		}
		e.logger.Info().
			Str("execution_id", executionID.String()).
			Dur("duration", time.Since(*exec.StartedAt)).
			Msg("Workflow execution completed")

		// Emit final progress
		if e.streamService != nil {
			e.streamService.SendProgress(ctx, exec.WorkspaceID, websocket.ExecutionProgressData{
				ExecutionID:    exec.ID,
				WorkflowID:     wf.ID,
				Status:         "completed",
				Progress:       100,
				NodesCompleted: int(runtime.GetNodeCount()),
				NodesTotal:     len(wf.Nodes),
				StartedAt:      *exec.StartedAt,
				ElapsedMs:      time.Since(*exec.StartedAt).Milliseconds(),
			})
		}
	}

	if err := e.executionRepo.Update(ctx, exec); err != nil {
		e.logger.Error().Err(err).Msg("Failed to update execution status")
	}

	return err
}

func (e *Executor) executeWorkflow(ctx context.Context, runtime *Runtime) error {
	nodes := runtime.GetNodes()

	triggerNode := findTriggerNode(nodes)
	if triggerNode == nil {
		return fmt.Errorf("no trigger node found")
	}

	return e.executeNode(ctx, runtime, triggerNode)
}

func (e *Executor) executeNode(ctx context.Context, runtime *Runtime, node map[string]interface{}) error {
	nodeID, _ := node["id"].(string)
	nodeType, _ := node["type"].(string)

	// Check execution limit
	if runtime.IncrementNodeCount() > MaxNodesPerExecution {
		return fmt.Errorf("execution limit exceeded: maximum of %d nodes allowed", MaxNodesPerExecution)
	}

	// Track node execution for billing
	if e.usageTracker != nil {
		if err := e.usageTracker.TrackNodeExecution(ctx, runtime.Execution.WorkspaceID, runtime.Execution.ID, nodeType); err != nil {
			e.logger.Warn().Err(err).Str("node_type", nodeType).Msg("Usage limit exceeded during node execution")
			return fmt.Errorf("usage limit exceeded: %w", err)
		}
	}

	nodeExec := execution.NewNodeExecution(runtime.Execution.ID, nodeID, nodeType, nil)

	// Resolve parameters and set as input data
	if params, ok := node["parameters"].(map[string]interface{}); ok {
		resolvedParams := runtime.ResolveParameters(params)
		nodeExec.InputData = resolvedParams
	}

	nodeExec.Start()

	if err := e.nodeExecRepo.Create(ctx, nodeExec); err != nil {
		e.logger.Warn().Err(err).Msg("Failed to create node execution")
	}

	result, err := e.processor.Process(ctx, runtime, node)

	if err != nil {
		nodeExec.Fail(err.Error())
	} else {
		nodeExec.Complete(result)
		runtime.SetNodeOutput(nodeID, result)
	}

	if err := e.nodeExecRepo.Update(ctx, nodeExec); err != nil {
		e.logger.Warn().Err(err).Msg("Failed to update node execution")
	}

	// Emit node output event
	if e.streamService != nil {
		var startedAt time.Time
		if nodeExec.StartedAt != nil {
			startedAt = *nodeExec.StartedAt
		}

		var durationMs int64
		if nodeExec.DurationMs != nil {
			durationMs = int64(*nodeExec.DurationMs)
		}

		e.streamService.SendNodeOutput(ctx, runtime.Execution.WorkspaceID, websocket.NodeOutputData{
			ExecutionID: runtime.Execution.ID,
			NodeID:      nodeID,
			NodeName:    fmt.Sprintf("%v", node["name"]),
			NodeType:    nodeType,
			Input:       nodeExec.InputData,
			Output:      result,
			StartedAt:   startedAt,
			CompletedAt: nodeExec.CompletedAt,
			DurationMs:  durationMs,
		})
	}

	if err != nil {
		return err
	}

	// Determine output port (branch)
	sourcePort := "main"
	if branch, ok := result["branch"].(string); ok {
		sourcePort = branch
	}

	nextNodes := runtime.GetNextNodes(nodeID, sourcePort)
	for _, next := range nextNodes {
		if err := e.executeNode(ctx, runtime, next); err != nil {
			return err
		}
	}

	return nil
}

func findTriggerNode(nodes []map[string]interface{}) map[string]interface{} {
	for _, node := range nodes {
		if nodeType, ok := node["type"].(string); ok {
			if len(nodeType) > 8 && nodeType[:8] == "trigger." {
				return node
			}
		}
	}
	return nil
}
