package schedule

import (
	"context"
	"fmt"

	"github.com/google/uuid"
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
	eventBus     events.Bus
}

func NewCreateScheduleHandler(scheduleRepo schedule.Repository, eventBus events.Bus) *CreateScheduleHandler {
	return &CreateScheduleHandler{scheduleRepo: scheduleRepo, eventBus: eventBus}
}

func (h *CreateScheduleHandler) Handle(ctx context.Context, cmd CreateScheduleCommand) (*schedule.Schedule, error) {
	if cmd.Name == "" {
		return nil, fmt.Errorf("schedule name is required")
	}
	if cmd.CronExpression == "" {
		return nil, fmt.Errorf("cron expression is required")
	}

	sched := schedule.NewSchedule(cmd.WorkflowID, cmd.WorkspaceID, cmd.CreatedBy, cmd.Name, cmd.CronExpression)
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
