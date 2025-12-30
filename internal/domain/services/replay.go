package services

import (
	"context"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/domain/models"
	"github.com/linkflow-ai/linkflow/internal/domain/repositories"
)

type ExecutionReplayService struct {
	execService *ExecutionService
	execRepo    *repositories.ExecutionRepository
}

// NewExecutionReplayService creates a new ExecutionReplayService for replaying workflow executions.
func NewExecutionReplayService(execService *ExecutionService, execRepo *repositories.ExecutionRepository) *ExecutionReplayService {
	return &ExecutionReplayService{execService: execService, execRepo: execRepo}
}

func (s *ExecutionReplayService) Replay(ctx context.Context, executionID uuid.UUID, triggeredBy *uuid.UUID) (*models.Execution, error) {
	original, err := s.execRepo.FindByID(ctx, executionID)
	if err != nil {
		return nil, err
	}

	return s.execService.Create(ctx, CreateExecutionInput{
		WorkflowID:  original.WorkflowID,
		WorkspaceID: original.WorkspaceID,
		TriggeredBy: triggeredBy,
		TriggerType: "replay",
		TriggerData: original.TriggerData,
		InputData:   original.InputData,
	})
}

func (s *ExecutionReplayService) ReplayFromNode(ctx context.Context, executionID uuid.UUID, startNodeID string, triggeredBy *uuid.UUID) (*models.Execution, error) {
	original, err := s.execRepo.FindByID(ctx, executionID)
	if err != nil {
		return nil, err
	}

	// Create execution with partial flag
	triggerData := models.JSON{
		"replay_from":        startNodeID,
		"original_execution": executionID.String(),
	}

	return s.execService.Create(ctx, CreateExecutionInput{
		WorkflowID:  original.WorkflowID,
		WorkspaceID: original.WorkspaceID,
		TriggeredBy: triggeredBy,
		TriggerType: "partial_replay",
		TriggerData: triggerData,
		InputData:   original.InputData,
	})
}
