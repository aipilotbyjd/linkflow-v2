package schedule

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/core/domain/schedule"
)

type ResumeScheduleCommand struct {
	ScheduleID uuid.UUID
}

type ResumeScheduleHandler struct {
	scheduleRepo schedule.Repository
}

func NewResumeScheduleHandler(scheduleRepo schedule.Repository) *ResumeScheduleHandler {
	return &ResumeScheduleHandler{scheduleRepo: scheduleRepo}
}

func (h *ResumeScheduleHandler) Handle(ctx context.Context, cmd ResumeScheduleCommand) error {
	sched, err := h.scheduleRepo.FindByID(ctx, cmd.ScheduleID)
	if err != nil {
		return schedule.ErrScheduleNotFound
	}

	sched.Activate()

	nextRun, err := sched.CalculateNextRun()
	if err != nil {
		return fmt.Errorf("failed to calculate next run: %w", err)
	}
	sched.SetNextRunAt(nextRun)

	return h.scheduleRepo.Update(ctx, sched)
}
