package mappers

import (
	"github.com/linkflow-ai/linkflow/internal/api/dto"
	"github.com/linkflow-ai/linkflow/internal/domain/models"
)

// ScheduleToResponse converts a Schedule model to ScheduleResponse DTO
func ScheduleToResponse(s *models.Schedule) dto.ScheduleResponse {
	var nextRunAt, lastRunAt *int64
	if s.NextRunAt != nil {
		ts := s.NextRunAt.Unix()
		nextRunAt = &ts
	}
	if s.LastRunAt != nil {
		ts := s.LastRunAt.Unix()
		lastRunAt = &ts
	}

	return dto.ScheduleResponse{
		ID:             s.ID.String(),
		WorkflowID:     s.WorkflowID.String(),
		Name:           s.Name,
		Description:    s.Description,
		CronExpression: s.CronExpression,
		Timezone:       s.Timezone,
		IsActive:       s.IsActive,
		InputData:      s.InputData,
		NextRunAt:      nextRunAt,
		LastRunAt:      lastRunAt,
		RunCount:       s.RunCount,
		CreatedAt:      s.CreatedAt.Unix(),
	}
}

// SchedulesToResponse converts a slice of Schedule models to ScheduleResponse DTOs
func SchedulesToResponse(schedules []models.Schedule) []dto.ScheduleResponse {
	result := make([]dto.ScheduleResponse, len(schedules))
	for i := range schedules {
		result[i] = ScheduleToResponse(&schedules[i])
	}
	return result
}
