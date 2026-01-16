package schedule

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/core/domain/schedule"
	"github.com/linkflow-ai/linkflow/internal/shared/types"
)

type UpdateScheduleCommand struct {
	ScheduleID     uuid.UUID
	Name           *string
	Description    *string
	CronExpression *string
	Timezone       *string
	InputData      types.JSON
}

type UpdateScheduleHandler struct {
	scheduleRepo schedule.Repository
}

func NewUpdateScheduleHandler(scheduleRepo schedule.Repository) *UpdateScheduleHandler {
	return &UpdateScheduleHandler{scheduleRepo: scheduleRepo}
}

func (h *UpdateScheduleHandler) Handle(ctx context.Context, cmd UpdateScheduleCommand) (*schedule.Schedule, error) {
	sched, err := h.scheduleRepo.FindByID(ctx, cmd.ScheduleID)
	if err != nil {
		return nil, schedule.ErrScheduleNotFound
	}

	name := sched.Name
	if cmd.Name != nil {
		name = *cmd.Name
	}
	cronExpr := sched.CronExpression
	if cmd.CronExpression != nil {
		cronExpr = *cmd.CronExpression
	}
	tz := sched.Timezone
	if cmd.Timezone != nil {
		tz = *cmd.Timezone
	}

	sched.Update(name, cmd.Description, cronExpr, tz)

	if cmd.InputData != nil {
		sched.UpdateInputData(cmd.InputData)
	}

	if cmd.CronExpression != nil {
		nextRun, err := sched.CalculateNextRun()
		if err != nil {
			return nil, fmt.Errorf("invalid cron expression: %w", err)
		}
		sched.SetNextRunAt(nextRun)
	}

	if err := h.scheduleRepo.Update(ctx, sched); err != nil {
		return nil, fmt.Errorf("failed to update schedule: %w", err)
	}

	return sched, nil
}
