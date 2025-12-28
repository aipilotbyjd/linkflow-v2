package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/domain/models"
	"github.com/linkflow-ai/linkflow/internal/domain/repositories"
	"github.com/robfig/cron/v3"
	"github.com/rs/zerolog/log"
)

// Schedule errors
var (
	ErrScheduleNotFound  = errors.New("schedule not found")
	ErrInvalidCron       = errors.New("invalid cron expression")
	ErrScheduleNameRequired = errors.New("schedule name is required")
)

// ScheduleService handles workflow scheduling operations.
type ScheduleService struct {
	scheduleRepo *repositories.ScheduleRepository
	cronParser   cron.Parser
}

// NewScheduleService creates a new ScheduleService with required dependencies.
func NewScheduleService(scheduleRepo *repositories.ScheduleRepository) *ScheduleService {
	if scheduleRepo == nil {
		panic("schedule service: scheduleRepo is required")
	}
	return &ScheduleService{
		scheduleRepo: scheduleRepo,
		cronParser:   cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow),
	}
}

// CreateScheduleInput holds the input for creating a schedule.
type CreateScheduleInput struct {
	WorkflowID     uuid.UUID
	WorkspaceID    uuid.UUID
	CreatedBy      uuid.UUID
	Name           string
	Description    *string
	CronExpression string
	Timezone       string
	InputData      models.JSON
}

// Create creates a new workflow schedule.
func (s *ScheduleService) Create(ctx context.Context, input CreateScheduleInput) (*models.Schedule, error) {
	if input.Name == "" {
		return nil, ErrScheduleNameRequired
	}

	nextRun, err := s.calculateNextRun(input.CronExpression, input.Timezone)
	if err != nil {
		return nil, ErrInvalidCron
	}

	schedule := &models.Schedule{
		WorkflowID:     input.WorkflowID,
		WorkspaceID:    input.WorkspaceID,
		CreatedBy:      input.CreatedBy,
		Name:           input.Name,
		Description:    input.Description,
		CronExpression: input.CronExpression,
		Timezone:       input.Timezone,
		IsActive:       true,
		InputData:      input.InputData,
		NextRunAt:      &nextRun,
	}

	if err := s.scheduleRepo.Create(ctx, schedule); err != nil {
		return nil, fmt.Errorf("failed to create schedule: %w", err)
	}

	log.Info().
		Str("schedule_id", schedule.ID.String()).
		Str("workflow_id", input.WorkflowID.String()).
		Str("cron", input.CronExpression).
		Msg("Schedule created")

	return schedule, nil
}

// GetByID returns a schedule by its ID.
func (s *ScheduleService) GetByID(ctx context.Context, id uuid.UUID) (*models.Schedule, error) {
	schedule, err := s.scheduleRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrScheduleNotFound, id)
	}
	return schedule, nil
}

// GetByWorkflow returns all schedules for a workflow.
func (s *ScheduleService) GetByWorkflow(ctx context.Context, workflowID uuid.UUID) ([]models.Schedule, error) {
	schedules, err := s.scheduleRepo.FindByWorkflowID(ctx, workflowID)
	if err != nil {
		return nil, fmt.Errorf("failed to get schedules: %w", err)
	}
	return schedules, nil
}

// GetByWorkspace returns paginated schedules for a workspace.
func (s *ScheduleService) GetByWorkspace(ctx context.Context, workspaceID uuid.UUID, opts *repositories.ListOptions) ([]models.Schedule, int64, error) {
	schedules, total, err := s.scheduleRepo.FindByWorkspaceID(ctx, workspaceID, opts)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get schedules: %w", err)
	}
	return schedules, total, nil
}

// GetDue returns all schedules that are due to run.
func (s *ScheduleService) GetDue(ctx context.Context) ([]models.Schedule, error) {
	schedules, err := s.scheduleRepo.FindDue(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get due schedules: %w", err)
	}
	return schedules, nil
}

// UpdateScheduleInput holds the input for updating a schedule.
type UpdateScheduleInput struct {
	Name           *string
	Description    *string
	CronExpression *string
	Timezone       *string
	InputData      models.JSON
}

// Update updates a schedule's configuration.
func (s *ScheduleService) Update(ctx context.Context, scheduleID uuid.UUID, input UpdateScheduleInput) (*models.Schedule, error) {
	schedule, err := s.scheduleRepo.FindByID(ctx, scheduleID)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrScheduleNotFound, scheduleID)
	}

	if input.Name != nil {
		schedule.Name = *input.Name
	}
	if input.Description != nil {
		schedule.Description = input.Description
	}
	if input.CronExpression != nil {
		schedule.CronExpression = *input.CronExpression
	}
	if input.Timezone != nil {
		schedule.Timezone = *input.Timezone
	}
	if input.InputData != nil {
		schedule.InputData = input.InputData
	}

	if input.CronExpression != nil || input.Timezone != nil {
		nextRun, err := s.calculateNextRun(schedule.CronExpression, schedule.Timezone)
		if err != nil {
			return nil, ErrInvalidCron
		}
		schedule.NextRunAt = &nextRun
	}

	if err := s.scheduleRepo.Update(ctx, schedule); err != nil {
		return nil, fmt.Errorf("failed to update schedule: %w", err)
	}

	log.Info().Str("schedule_id", scheduleID.String()).Msg("Schedule updated")

	return schedule, nil
}

// Delete deletes a schedule.
func (s *ScheduleService) Delete(ctx context.Context, scheduleID uuid.UUID) error {
	if err := s.scheduleRepo.Delete(ctx, scheduleID); err != nil {
		return fmt.Errorf("failed to delete schedule: %w", err)
	}
	log.Info().Str("schedule_id", scheduleID.String()).Msg("Schedule deleted")
	return nil
}

// Pause pauses a schedule.
func (s *ScheduleService) Pause(ctx context.Context, scheduleID uuid.UUID) error {
	if err := s.scheduleRepo.SetActive(ctx, scheduleID, false); err != nil {
		return fmt.Errorf("failed to pause schedule: %w", err)
	}
	log.Info().Str("schedule_id", scheduleID.String()).Msg("Schedule paused")
	return nil
}

// Resume resumes a paused schedule.
func (s *ScheduleService) Resume(ctx context.Context, scheduleID uuid.UUID) error {
	schedule, err := s.scheduleRepo.FindByID(ctx, scheduleID)
	if err != nil {
		return fmt.Errorf("%w: %s", ErrScheduleNotFound, scheduleID)
	}

	nextRun, err := s.calculateNextRun(schedule.CronExpression, schedule.Timezone)
	if err != nil {
		return fmt.Errorf("failed to calculate next run: %w", err)
	}

	if err := s.scheduleRepo.UpdateNextRun(ctx, scheduleID, nextRun); err != nil {
		return fmt.Errorf("failed to update next run: %w", err)
	}

	if err := s.scheduleRepo.SetActive(ctx, scheduleID, true); err != nil {
		return fmt.Errorf("failed to resume schedule: %w", err)
	}

	log.Info().Str("schedule_id", scheduleID.String()).Time("next_run", nextRun).Msg("Schedule resumed")

	return nil
}

// RecordRun records a schedule run and calculates the next run time.
func (s *ScheduleService) RecordRun(ctx context.Context, scheduleID, executionID uuid.UUID) error {
	schedule, err := s.scheduleRepo.FindByID(ctx, scheduleID)
	if err != nil {
		return fmt.Errorf("%w: %s", ErrScheduleNotFound, scheduleID)
	}

	nextRun, err := s.calculateNextRun(schedule.CronExpression, schedule.Timezone)
	if err != nil {
		return fmt.Errorf("failed to calculate next run: %w", err)
	}

	if err := s.scheduleRepo.RecordRun(ctx, scheduleID, executionID, nextRun); err != nil {
		return fmt.Errorf("failed to record run: %w", err)
	}

	log.Debug().
		Str("schedule_id", scheduleID.String()).
		Str("execution_id", executionID.String()).
		Msg("Schedule run recorded")

	return nil
}

// calculateNextRun calculates the next run time for a cron expression.
func (s *ScheduleService) calculateNextRun(cronExpr, timezone string) (time.Time, error) {
	schedule, err := s.cronParser.Parse(cronExpr)
	if err != nil {
		return time.Time{}, err
	}

	loc, err := time.LoadLocation(timezone)
	if err != nil {
		loc = time.UTC
	}

	now := time.Now().In(loc)
	return schedule.Next(now), nil
}

// ValidateCron validates a cron expression.
func (s *ScheduleService) ValidateCron(cronExpr string) error {
	_, err := s.cronParser.Parse(cronExpr)
	if err != nil {
		return fmt.Errorf("%w: %s", ErrInvalidCron, err)
	}
	return nil
}

// GetDueBatch returns due schedules in batches for processing.
func (s *ScheduleService) GetDueBatch(ctx context.Context, limit, offset int) ([]models.Schedule, error) {
	schedules, err := s.scheduleRepo.FindDueBatch(ctx, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get due batch: %w", err)
	}
	return schedules, nil
}

// GetDueByPriority returns due schedules filtered by priority.
func (s *ScheduleService) GetDueByPriority(ctx context.Context, priority string) ([]models.Schedule, error) {
	schedules, err := s.scheduleRepo.FindDueByPriority(ctx, priority)
	if err != nil {
		return nil, fmt.Errorf("failed to get due by priority: %w", err)
	}
	return schedules, nil
}
