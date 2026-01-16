package schedule

import (
	"context"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/core/domain/schedule"
)

type PauseScheduleCommand struct {
	ScheduleID uuid.UUID
}

type PauseScheduleHandler struct {
	scheduleRepo schedule.Repository
}

func NewPauseScheduleHandler(scheduleRepo schedule.Repository) *PauseScheduleHandler {
	return &PauseScheduleHandler{scheduleRepo: scheduleRepo}
}

func (h *PauseScheduleHandler) Handle(ctx context.Context, cmd PauseScheduleCommand) error {
	sched, err := h.scheduleRepo.FindByID(ctx, cmd.ScheduleID)
	if err != nil {
		return schedule.ErrScheduleNotFound
	}

	sched.Deactivate()

	return h.scheduleRepo.Update(ctx, sched)
}
