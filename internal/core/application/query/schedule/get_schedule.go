package schedule

import (
	"context"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/core/domain/schedule"
)

type GetScheduleQuery struct {
	ScheduleID uuid.UUID
}

type GetScheduleHandler struct {
	scheduleRepo schedule.Repository
}

func NewGetScheduleHandler(scheduleRepo schedule.Repository) *GetScheduleHandler {
	return &GetScheduleHandler{scheduleRepo: scheduleRepo}
}

func (h *GetScheduleHandler) Handle(ctx context.Context, q GetScheduleQuery) (*schedule.Schedule, error) {
	return h.scheduleRepo.FindByID(ctx, q.ScheduleID)
}
