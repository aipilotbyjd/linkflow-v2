package executor

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/core/domain/execution"
	"github.com/linkflow-ai/linkflow/internal/core/domain/workflow"
	"github.com/linkflow-ai/linkflow/internal/infrastructure/observability/logger"
)

type Executor struct {
	workflowRepo  workflow.Repository
	executionRepo execution.Repository
	nodeExecRepo  execution.NodeExecutionRepository
	processor     *Processor
	logger        logger.Logger
}

func NewExecutor(
	workflowRepo workflow.Repository,
	executionRepo execution.Repository,
	nodeExecRepo execution.NodeExecutionRepository,
	processor *Processor,
	log logger.Logger,
) *Executor {
	return &Executor{
		workflowRepo:  workflowRepo,
		executionRepo: executionRepo,
		nodeExecRepo:  nodeExecRepo,
		processor:     processor,
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

	exec.Start()
	if err := e.executionRepo.Update(ctx, exec); err != nil {
		return fmt.Errorf("failed to update execution: %w", err)
	}

	e.logger.Info().
		Str("execution_id", executionID.String()).
		Str("workflow_id", wf.ID.String()).
		Msg("Starting workflow execution")

	runtime := NewRuntime(exec, wf, e.logger)

	err = e.executeWorkflow(ctx, runtime)

	if err != nil {
		exec.Fail(err.Error(), nil)
		e.logger.Error().Err(err).
			Str("execution_id", executionID.String()).
			Msg("Workflow execution failed")
	} else {
		exec.Complete(runtime.GetOutputData())
		e.logger.Info().
			Str("execution_id", executionID.String()).
			Dur("duration", time.Since(*exec.StartedAt)).
			Msg("Workflow execution completed")
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

	nodeExec := execution.NewNodeExecution(runtime.Execution.ID, nodeID, nodeType, nil)
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

	if err != nil {
		return err
	}

	nextNodes := runtime.GetNextNodes(nodeID)
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
