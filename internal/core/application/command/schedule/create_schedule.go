package schedule

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	billingapp "github.com/linkflow-ai/linkflow/internal/core/application/billing"
	"github.com/linkflow-ai/linkflow/internal/core/domain/schedule"
	"github.com/linkflow-ai/linkflow/internal/shared/events"
	"github.com/linkflow-ai/linkflow/internal/shared/types"
)

type CreateScheduleCommand struct {
	WorkflowID     uuid.UUID
	WorkspaceID    uuid.UUID
	CreatedBy      uuid.UUID
	Name           string
	Description    *string
	CronExpression string
	Timezone       string
	InputData      types.JSON
}

type CreateScheduleHandler struct {
	scheduleRepo schedule.Repository
	usageService *billingapp.UsageService
	eventBus     events.Bus
}

func NewCreateScheduleHandler(scheduleRepo schedule.Repository, usageService *billingapp.UsageService, eventBus events.Bus) *CreateScheduleHandler {
	return &CreateScheduleHandler{scheduleRepo: scheduleRepo, usageService: usageService, eventBus: eventBus}
}

func (h *CreateScheduleHandler) Handle(ctx context.Context, cmd CreateScheduleCommand) (*schedule.Schedule, error) {
	// Check minimum interval from billing plan
	if h.usageService != nil {
		minInterval, err := h.usageService.GetMinInterval(ctx, cmd.WorkspaceID)
		if err != nil {
			return nil, fmt.Errorf("failed to get plan limits: %w", err)
		}

		// Validate cron expression against minimum interval
		if err := schedule.ValidateCronMinInterval(cmd.CronExpression, minInterval); err != nil {
			return nil, fmt.Errorf("schedule interval too frequent: minimum allowed is %d minutes on your plan", minInterval)
		}
	}

	// Create schedule (includes validation)
	sched, err := schedule.NewSchedule(cmd.WorkflowID, cmd.WorkspaceID, cmd.CreatedBy, cmd.Name, cmd.CronExpression)
	if err != nil {
		return nil, err
	}

	if cmd.Timezone != "" {
		sched.WithTimezone(cmd.Timezone)
	}
	if cmd.Description != nil {
		sched.WithDescription(*cmd.Description)
	}
	if cmd.InputData != nil {
		sched.WithInputData(cmd.InputData)
	}

	nextRun, err := sched.CalculateNextRun()
	if err != nil {
		return nil, fmt.Errorf("invalid cron expression: %w", err)
	}
	sched.SetNextRunAt(nextRun)

	if err := h.scheduleRepo.Create(ctx, sched); err != nil {
		return nil, fmt.Errorf("failed to create schedule: %w", err)
	}

	if h.eventBus != nil {
		_ = h.eventBus.Publish(ctx, events.ScheduleCreated{
			BaseEvent:   events.NewBaseEvent("schedule.created", sched.ID, "schedule"),
			ScheduleID:  sched.ID,
			WorkflowID:  sched.WorkflowID,
			WorkspaceID: sched.WorkspaceID,
		})
	}

	return sched, nil
}
