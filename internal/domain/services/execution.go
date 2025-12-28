package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/domain/models"
	"github.com/linkflow-ai/linkflow/internal/domain/repositories"
	"github.com/rs/zerolog/log"
)

// Execution errors
var (
	ErrExecutionNotFound   = errors.New("execution not found")
	ErrExecutionNotRunning = errors.New("execution is not running")
)

// ExecutionService handles workflow execution management.
type ExecutionService struct {
	executionRepo     *repositories.ExecutionRepository
	nodeExecutionRepo *repositories.NodeExecutionRepository
	workflowRepo      *repositories.WorkflowRepository
}

// NewExecutionService creates a new ExecutionService with required dependencies.
func NewExecutionService(
	executionRepo *repositories.ExecutionRepository,
	nodeExecutionRepo *repositories.NodeExecutionRepository,
	workflowRepo *repositories.WorkflowRepository,
) *ExecutionService {
	if executionRepo == nil || nodeExecutionRepo == nil || workflowRepo == nil {
		panic("execution service: all repositories are required")
	}
	return &ExecutionService{
		executionRepo:     executionRepo,
		nodeExecutionRepo: nodeExecutionRepo,
		workflowRepo:      workflowRepo,
	}
}

type CreateExecutionInput struct {
	WorkflowID  uuid.UUID
	WorkspaceID uuid.UUID
	TriggeredBy *uuid.UUID
	TriggerType string
	TriggerData models.JSON
	InputData   models.JSON
}

// Create creates a new workflow execution.
func (s *ExecutionService) Create(ctx context.Context, input CreateExecutionInput) (*models.Execution, error) {
	workflow, err := s.workflowRepo.FindByID(ctx, input.WorkflowID)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrWorkflowNotFound, input.WorkflowID)
	}

	execution := &models.Execution{
		WorkflowID:      input.WorkflowID,
		WorkspaceID:     input.WorkspaceID,
		TriggeredBy:     input.TriggeredBy,
		WorkflowVersion: workflow.Version,
		Status:          models.ExecutionStatusQueued,
		TriggerType:     input.TriggerType,
		TriggerData:     input.TriggerData,
		InputData:       input.InputData,
	}

	if err := s.executionRepo.Create(ctx, execution); err != nil {
		return nil, fmt.Errorf("failed to create execution: %w", err)
	}

	if err := s.workflowRepo.IncrementExecutionCount(ctx, input.WorkflowID); err != nil {
		log.Warn().Err(err).Str("workflow_id", input.WorkflowID.String()).Msg("Failed to increment execution count")
	}

	log.Info().
		Str("execution_id", execution.ID.String()).
		Str("workflow_id", input.WorkflowID.String()).
		Str("trigger_type", input.TriggerType).
		Msg("Execution created")

	return execution, nil
}

// GetByID returns an execution by its ID.
func (s *ExecutionService) GetByID(ctx context.Context, id uuid.UUID) (*models.Execution, error) {
	execution, err := s.executionRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrExecutionNotFound, id)
	}
	return execution, nil
}

// GetByWorkflow returns paginated executions for a workflow.
func (s *ExecutionService) GetByWorkflow(ctx context.Context, workflowID uuid.UUID, opts *repositories.ListOptions) ([]models.Execution, int64, error) {
	executions, total, err := s.executionRepo.FindByWorkflowID(ctx, workflowID, opts)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get executions: %w", err)
	}
	return executions, total, nil
}

// GetByWorkspace returns paginated executions for a workspace.
func (s *ExecutionService) GetByWorkspace(ctx context.Context, workspaceID uuid.UUID, opts *repositories.ListOptions) ([]models.Execution, int64, error) {
	executions, total, err := s.executionRepo.FindByWorkspaceID(ctx, workspaceID, opts)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get executions: %w", err)
	}
	return executions, total, nil
}

// GetNodeExecutions returns all node executions for an execution.
func (s *ExecutionService) GetNodeExecutions(ctx context.Context, executionID uuid.UUID) ([]models.NodeExecution, error) {
	nodeExecs, err := s.nodeExecutionRepo.FindByExecutionID(ctx, executionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get node executions: %w", err)
	}
	return nodeExecs, nil
}

// Start marks an execution as running.
func (s *ExecutionService) Start(ctx context.Context, executionID uuid.UUID) error {
	if err := s.executionRepo.UpdateStatus(ctx, executionID, models.ExecutionStatusRunning); err != nil {
		return fmt.Errorf("failed to start execution: %w", err)
	}
	log.Debug().Str("execution_id", executionID.String()).Msg("Execution started")
	return nil
}

// Complete marks an execution as completed with output.
func (s *ExecutionService) Complete(ctx context.Context, executionID uuid.UUID, output models.JSON) error {
	if err := s.executionRepo.SetOutput(ctx, executionID, output); err != nil {
		return fmt.Errorf("failed to set execution output: %w", err)
	}
	if err := s.executionRepo.UpdateStatus(ctx, executionID, models.ExecutionStatusCompleted); err != nil {
		return fmt.Errorf("failed to complete execution: %w", err)
	}
	log.Info().Str("execution_id", executionID.String()).Msg("Execution completed")
	return nil
}

// Fail marks an execution as failed with an error message.
func (s *ExecutionService) Fail(ctx context.Context, executionID uuid.UUID, errorMessage string, errorNodeID *string) error {
	if err := s.executionRepo.SetError(ctx, executionID, errorMessage, errorNodeID); err != nil {
		return fmt.Errorf("failed to set execution error: %w", err)
	}
	log.Warn().Str("execution_id", executionID.String()).Str("error", errorMessage).Msg("Execution failed")
	return nil
}

// Cancel cancels a running or queued execution.
func (s *ExecutionService) Cancel(ctx context.Context, executionID uuid.UUID) error {
	execution, err := s.executionRepo.FindByID(ctx, executionID)
	if err != nil {
		return fmt.Errorf("%w: %s", ErrExecutionNotFound, executionID)
	}

	if execution.Status != models.ExecutionStatusQueued && execution.Status != models.ExecutionStatusRunning {
		return ErrExecutionNotRunning
	}

	if err := s.executionRepo.UpdateStatus(ctx, executionID, models.ExecutionStatusCancelled); err != nil {
		return fmt.Errorf("failed to cancel execution: %w", err)
	}
	log.Info().Str("execution_id", executionID.String()).Msg("Execution cancelled")
	return nil
}

// UpdateProgress updates the progress of an execution.
func (s *ExecutionService) UpdateProgress(ctx context.Context, executionID uuid.UUID, nodesCompleted int) error {
	if err := s.executionRepo.UpdateProgress(ctx, executionID, nodesCompleted); err != nil {
		return fmt.Errorf("failed to update progress: %w", err)
	}
	return nil
}

// CreateNodeExecution creates a new node execution record.
func (s *ExecutionService) CreateNodeExecution(ctx context.Context, executionID uuid.UUID, nodeID, nodeType, nodeName string) (*models.NodeExecution, error) {
	nodeExec := &models.NodeExecution{
		ExecutionID: executionID,
		NodeID:      nodeID,
		NodeType:    nodeType,
		NodeName:    &nodeName,
		Status:      models.NodeStatusPending,
	}

	if err := s.nodeExecutionRepo.Create(ctx, nodeExec); err != nil {
		return nil, fmt.Errorf("failed to create node execution: %w", err)
	}

	return nodeExec, nil
}

// StartNodeExecution marks a node execution as running with input data.
func (s *ExecutionService) StartNodeExecution(ctx context.Context, nodeExecutionID uuid.UUID, input models.JSON) error {
	// Update input data
	if err := s.nodeExecutionRepo.DB().WithContext(ctx).
		Model(&models.NodeExecution{}).
		Where("id = ?", nodeExecutionID).
		Update("input_data", input).Error; err != nil {
		return fmt.Errorf("failed to set node input: %w", err)
	}
	if err := s.nodeExecutionRepo.UpdateStatus(ctx, nodeExecutionID, models.NodeStatusRunning); err != nil {
		return fmt.Errorf("failed to start node execution: %w", err)
	}
	return nil
}

// CompleteNodeExecution marks a node execution as completed with output.
func (s *ExecutionService) CompleteNodeExecution(ctx context.Context, nodeExecutionID uuid.UUID, output models.JSON, durationMs int) error {
	if err := s.nodeExecutionRepo.SetResult(ctx, nodeExecutionID, models.NodeStatusCompleted, output, durationMs); err != nil {
		return fmt.Errorf("failed to complete node execution: %w", err)
	}
	return nil
}

// FailNodeExecution marks a node execution as failed with an error message.
func (s *ExecutionService) FailNodeExecution(ctx context.Context, nodeExecutionID uuid.UUID, errorMessage string) error {
	if err := s.nodeExecutionRepo.SetError(ctx, nodeExecutionID, errorMessage); err != nil {
		return fmt.Errorf("failed to fail node execution: %w", err)
	}
	return nil
}

// SkipNodeExecution marks a node execution as skipped.
func (s *ExecutionService) SkipNodeExecution(ctx context.Context, nodeExecutionID uuid.UUID) error {
	if err := s.nodeExecutionRepo.UpdateStatus(ctx, nodeExecutionID, models.NodeStatusSkipped); err != nil {
		return fmt.Errorf("failed to skip node execution: %w", err)
	}
	return nil
}

// Retry creates a new execution based on a previous one.
func (s *ExecutionService) Retry(ctx context.Context, executionID uuid.UUID, triggeredBy *uuid.UUID) (*models.Execution, error) {
	original, err := s.executionRepo.FindByID(ctx, executionID)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrExecutionNotFound, executionID)
	}

	newExec, err := s.Create(ctx, CreateExecutionInput{
		WorkflowID:  original.WorkflowID,
		WorkspaceID: original.WorkspaceID,
		TriggeredBy: triggeredBy,
		TriggerType: original.TriggerType,
		TriggerData: original.TriggerData,
		InputData:   original.InputData,
	})
	if err != nil {
		return nil, err
	}

	log.Info().
		Str("original_id", executionID.String()).
		Str("new_id", newExec.ID.String()).
		Msg("Execution retried")

	return newExec, nil
}

// GetStaleExecutions returns executions that have been running longer than expected.
func (s *ExecutionService) GetStaleExecutions(ctx context.Context, staleAfter time.Duration) ([]models.Execution, error) {
	executions, err := s.executionRepo.FindStale(ctx, staleAfter)
	if err != nil {
		return nil, fmt.Errorf("failed to get stale executions: %w", err)
	}
	return executions, nil
}

// DeleteOlderThan deletes executions older than the given cutoff time.
func (s *ExecutionService) DeleteOlderThan(ctx context.Context, cutoff time.Time) (int64, error) {
	count, err := s.executionRepo.DeleteOlderThan(ctx, cutoff)
	if err != nil {
		return 0, fmt.Errorf("failed to delete old executions: %w", err)
	}
	if count > 0 {
		log.Info().Int64("count", count).Time("cutoff", cutoff).Msg("Deleted old executions")
	}
	return count, nil
}

// GetHourlyStats returns hourly execution statistics by workspace.
func (s *ExecutionService) GetHourlyStats(ctx context.Context, start, end time.Time) (map[uuid.UUID]int64, error) {
	stats, err := s.executionRepo.GetHourlyStatsByWorkspace(ctx, start, end)
	if err != nil {
		return nil, fmt.Errorf("failed to get hourly stats: %w", err)
	}
	return stats, nil
}

// Search searches executions with filters.
func (s *ExecutionService) Search(ctx context.Context, filter repositories.ExecutionFilter, opts *repositories.ListOptions) ([]models.Execution, int64, error) {
	executions, total, err := s.executionRepo.Search(ctx, filter, opts)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to search executions: %w", err)
	}
	return executions, total, nil
}

// DeleteByIDs deletes executions by their IDs.
func (s *ExecutionService) DeleteByIDs(ctx context.Context, workspaceID uuid.UUID, ids []uuid.UUID) (int64, error) {
	count, err := s.executionRepo.DeleteByIDs(ctx, workspaceID, ids)
	if err != nil {
		return 0, fmt.Errorf("failed to delete executions: %w", err)
	}
	return count, nil
}

// GetStats returns execution statistics for a workspace.
func (s *ExecutionService) GetStats(ctx context.Context, workspaceID uuid.UUID, start, end time.Time) (map[string]interface{}, error) {
	stats, err := s.executionRepo.GetStats(ctx, workspaceID, start, end)
	if err != nil {
		return nil, fmt.Errorf("failed to get execution stats: %w", err)
	}
	return stats, nil
}
