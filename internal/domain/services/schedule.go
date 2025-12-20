package services

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/domain/models"
	"github.com/linkflow-ai/linkflow/internal/domain/repositories"
	"github.com/robfig/cron/v3"
)

var (
	ErrScheduleNotFound = errors.New("schedule not found")
	ErrInvalidCron      = errors.New("invalid cron expression")
)

type ScheduleService struct {
	scheduleRepo *repositories.ScheduleRepository
	cronParser   cron.Parser
}

func NewScheduleService(scheduleRepo *repositories.ScheduleRepository) *ScheduleService {
	return &ScheduleService{
		scheduleRepo: scheduleRepo,
		cronParser:   cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow),
	}
}

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

func (s *ScheduleService) Create(ctx context.Context, input CreateScheduleInput) (*models.Schedule, error) {
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
		return nil, err
	}

	return schedule, nil
}

func (s *ScheduleService) GetByID(ctx context.Context, id uuid.UUID) (*models.Schedule, error) {
	return s.scheduleRepo.FindByID(ctx, id)
}

func (s *ScheduleService) GetByWorkflow(ctx context.Context, workflowID uuid.UUID) ([]models.Schedule, error) {
	return s.scheduleRepo.FindByWorkflowID(ctx, workflowID)
}

func (s *ScheduleService) GetByWorkspace(ctx context.Context, workspaceID uuid.UUID, opts *repositories.ListOptions) ([]models.Schedule, int64, error) {
	return s.scheduleRepo.FindByWorkspaceID(ctx, workspaceID, opts)
}

func (s *ScheduleService) GetDue(ctx context.Context) ([]models.Schedule, error) {
	return s.scheduleRepo.FindDue(ctx)
}

type UpdateScheduleInput struct {
	Name           *string
	Description    *string
	CronExpression *string
	Timezone       *string
	InputData      models.JSON
}

func (s *ScheduleService) Update(ctx context.Context, scheduleID uuid.UUID, input UpdateScheduleInput) (*models.Schedule, error) {
	schedule, err := s.scheduleRepo.FindByID(ctx, scheduleID)
	if err != nil {
		return nil, err
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
		return nil, err
	}

	return schedule, nil
}

func (s *ScheduleService) Delete(ctx context.Context, scheduleID uuid.UUID) error {
	return s.scheduleRepo.Delete(ctx, scheduleID)
}

func (s *ScheduleService) Pause(ctx context.Context, scheduleID uuid.UUID) error {
	return s.scheduleRepo.SetActive(ctx, scheduleID, false)
}

func (s *ScheduleService) Resume(ctx context.Context, scheduleID uuid.UUID) error {
	schedule, err := s.scheduleRepo.FindByID(ctx, scheduleID)
	if err != nil {
		return err
	}

	nextRun, err := s.calculateNextRun(schedule.CronExpression, schedule.Timezone)
	if err != nil {
		return err
	}

	if err := s.scheduleRepo.UpdateNextRun(ctx, scheduleID, nextRun); err != nil {
		return err
	}

	return s.scheduleRepo.SetActive(ctx, scheduleID, true)
}

func (s *ScheduleService) RecordRun(ctx context.Context, scheduleID, executionID uuid.UUID) error {
	schedule, err := s.scheduleRepo.FindByID(ctx, scheduleID)
	if err != nil {
		return err
	}

	nextRun, err := s.calculateNextRun(schedule.CronExpression, schedule.Timezone)
	if err != nil {
		return err
	}

	return s.scheduleRepo.RecordRun(ctx, scheduleID, executionID, nextRun)
}

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

func (s *ScheduleService) ValidateCron(cronExpr string) error {
	_, err := s.cronParser.Parse(cronExpr)
	return err
}

func (s *ScheduleService) GetDueBatch(ctx context.Context, limit, offset int) ([]models.Schedule, error) {
	return s.scheduleRepo.FindDueBatch(ctx, limit, offset)
}

func (s *ScheduleService) GetDueByPriority(ctx context.Context, priority string) ([]models.Schedule, error) {
	return s.scheduleRepo.FindDueByPriority(ctx, priority)
}
