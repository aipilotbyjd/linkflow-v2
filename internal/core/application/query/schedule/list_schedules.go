package schedule

import (
	"context"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/core/domain/schedule"
	"github.com/linkflow-ai/linkflow/internal/shared/types"
)

type ListSchedulesQuery struct {
	WorkspaceID uuid.UUID
	WorkflowID  *uuid.UUID
	Page        int
	PageSize    int
}

type ListSchedulesResult struct {
	Schedules  []schedule.Schedule
	Total      int64
	Page       int
	PageSize   int
	TotalPages int
}

type ListSchedulesHandler struct {
	scheduleRepo schedule.Repository
}

func NewListSchedulesHandler(scheduleRepo schedule.Repository) *ListSchedulesHandler {
	return &ListSchedulesHandler{scheduleRepo: scheduleRepo}
}

func (h *ListSchedulesHandler) Handle(ctx context.Context, query ListSchedulesQuery) (*ListSchedulesResult, error) {
	opts := types.NewListOptions(query.Page, query.PageSize)

	schedules, total, err := h.scheduleRepo.FindByWorkspaceID(ctx, query.WorkspaceID, opts)
	if err != nil {
		return nil, err
	}

	pageSize := query.PageSize
	if pageSize <= 0 {
		pageSize = types.DefaultPageSize
	}
	totalPages := int(total) / pageSize
	if int(total)%pageSize != 0 {
		totalPages++
	}

	return &ListSchedulesResult{
		Schedules:  schedules,
		Total:      total,
		Page:       query.Page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}
