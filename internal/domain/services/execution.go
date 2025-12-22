package services

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/domain/models"
	"github.com/linkflow-ai/linkflow/internal/domain/repositories"
)

var (
	ErrExecutionNotFound = errors.New("execution not found")
	ErrExecutionNotRunning = errors.New("execution is not running")
)

type ExecutionService struct {
	executionRepo     *repositories.ExecutionRepository
	nodeExecutionRepo *repositories.NodeExecutionRepository
	workflowRepo      *repositories.WorkflowRepository
}

func NewExecutionService(
	executionRepo *repositories.ExecutionRepository,
	nodeExecutionRepo *repositories.NodeExecutionRepository,
	workflowRepo *repositories.WorkflowRepository,
) *ExecutionService {
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

func (s *ExecutionService) Create(ctx context.Context, input CreateExecutionInput) (*models.Execution, error) {
	workflow, err := s.workflowRepo.FindByID(ctx, input.WorkflowID)
	if err != nil {
		return nil, err
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
		return nil, err
	}

	_ = s.workflowRepo.IncrementExecutionCount(ctx, input.WorkflowID)

	return execution, nil
}

func (s *ExecutionService) GetByID(ctx context.Context, id uuid.UUID) (*models.Execution, error) {
	return s.executionRepo.FindByID(ctx, id)
}

func (s *ExecutionService) GetByWorkflow(ctx context.Context, workflowID uuid.UUID, opts *repositories.ListOptions) ([]models.Execution, int64, error) {
	return s.executionRepo.FindByWorkflowID(ctx, workflowID, opts)
}

func (s *ExecutionService) GetByWorkspace(ctx context.Context, workspaceID uuid.UUID, opts *repositories.ListOptions) ([]models.Execution, int64, error) {
	return s.executionRepo.FindByWorkspaceID(ctx, workspaceID, opts)
}

func (s *ExecutionService) GetNodeExecutions(ctx context.Context, executionID uuid.UUID) ([]models.NodeExecution, error) {
	return s.nodeExecutionRepo.FindByExecutionID(ctx, executionID)
}

func (s *ExecutionService) Start(ctx context.Context, executionID uuid.UUID) error {
	return s.executionRepo.UpdateStatus(ctx, executionID, models.ExecutionStatusRunning)
}

func (s *ExecutionService) Complete(ctx context.Context, executionID uuid.UUID, output models.JSON) error {
	if err := s.executionRepo.SetOutput(ctx, executionID, output); err != nil {
		return err
	}
	return s.executionRepo.UpdateStatus(ctx, executionID, models.ExecutionStatusCompleted)
}

func (s *ExecutionService) Fail(ctx context.Context, executionID uuid.UUID, errorMessage string, errorNodeID *string) error {
	return s.executionRepo.SetError(ctx, executionID, errorMessage, errorNodeID)
}

func (s *ExecutionService) Cancel(ctx context.Context, executionID uuid.UUID) error {
	execution, err := s.executionRepo.FindByID(ctx, executionID)
	if err != nil {
		return err
	}

	if execution.Status != models.ExecutionStatusQueued && execution.Status != models.ExecutionStatusRunning {
		return ErrExecutionNotRunning
	}

	return s.executionRepo.UpdateStatus(ctx, executionID, models.ExecutionStatusCancelled)
}

func (s *ExecutionService) UpdateProgress(ctx context.Context, executionID uuid.UUID, nodesCompleted int) error {
	return s.executionRepo.UpdateProgress(ctx, executionID, nodesCompleted)
}

func (s *ExecutionService) CreateNodeExecution(ctx context.Context, executionID uuid.UUID, nodeID, nodeType, nodeName string) (*models.NodeExecution, error) {
	nodeExec := &models.NodeExecution{
		ExecutionID: executionID,
		NodeID:      nodeID,
		NodeType:    nodeType,
		NodeName:    &nodeName,
		Status:      models.NodeStatusPending,
	}

	if err := s.nodeExecutionRepo.Create(ctx, nodeExec); err != nil {
		return nil, err
	}

	return nodeExec, nil
}

func (s *ExecutionService) StartNodeExecution(ctx context.Context, nodeExecutionID uuid.UUID, input models.JSON) error {
	if err := s.nodeExecutionRepo.DB().WithContext(ctx).Model(&models.NodeExecution{}).
		Where("id = ?", nodeExecutionID).
		Update("input_data", input).Error; err != nil {
		return err
	}
	return s.nodeExecutionRepo.UpdateStatus(ctx, nodeExecutionID, models.NodeStatusRunning)
}

func (s *ExecutionService) CompleteNodeExecution(ctx context.Context, nodeExecutionID uuid.UUID, output models.JSON, durationMs int) error {
	return s.nodeExecutionRepo.SetResult(ctx, nodeExecutionID, models.NodeStatusCompleted, output, durationMs)
}

func (s *ExecutionService) FailNodeExecution(ctx context.Context, nodeExecutionID uuid.UUID, errorMessage string) error {
	return s.nodeExecutionRepo.SetError(ctx, nodeExecutionID, errorMessage)
}

func (s *ExecutionService) SkipNodeExecution(ctx context.Context, nodeExecutionID uuid.UUID) error {
	return s.nodeExecutionRepo.UpdateStatus(ctx, nodeExecutionID, models.NodeStatusSkipped)
}

func (s *ExecutionService) Retry(ctx context.Context, executionID uuid.UUID, triggeredBy *uuid.UUID) (*models.Execution, error) {
	original, err := s.executionRepo.FindByID(ctx, executionID)
	if err != nil {
		return nil, err
	}

	return s.Create(ctx, CreateExecutionInput{
		WorkflowID:  original.WorkflowID,
		WorkspaceID: original.WorkspaceID,
		TriggeredBy: triggeredBy,
		TriggerType: original.TriggerType,
		TriggerData: original.TriggerData,
		InputData:   original.InputData,
	})
}

func (s *ExecutionService) GetStaleExecutions(ctx context.Context, staleAfter time.Duration) ([]models.Execution, error) {
	return s.executionRepo.FindStale(ctx, staleAfter)
}

func (s *ExecutionService) DeleteOlderThan(ctx context.Context, cutoff time.Time) (int64, error) {
	return s.executionRepo.DeleteOlderThan(ctx, cutoff)
}

func (s *ExecutionService) GetHourlyStats(ctx context.Context, start, end time.Time) (map[uuid.UUID]int64, error) {
	return s.executionRepo.GetHourlyStatsByWorkspace(ctx, start, end)
}

// Search searches executions with filters
func (s *ExecutionService) Search(ctx context.Context, filter repositories.ExecutionFilter, opts *repositories.ListOptions) ([]models.Execution, int64, error) {
	return s.executionRepo.Search(ctx, filter, opts)
}

// DeleteByIDs deletes executions by their IDs
func (s *ExecutionService) DeleteByIDs(ctx context.Context, workspaceID uuid.UUID, ids []uuid.UUID) (int64, error) {
	return s.executionRepo.DeleteByIDs(ctx, workspaceID, ids)
}

// GetStats returns execution statistics
func (s *ExecutionService) GetStats(ctx context.Context, workspaceID uuid.UUID, start, end time.Time) (map[string]interface{}, error) {
	return s.executionRepo.GetStats(ctx, workspaceID, start, end)
}
