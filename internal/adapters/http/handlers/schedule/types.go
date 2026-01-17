package schedule

import (
	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/core/domain/schedule"
	"github.com/linkflow-ai/linkflow/internal/shared/types"
)

// CreateRequest represents schedule creation request
type CreateRequest struct {
	WorkflowID     uuid.UUID  `json:"workflow_id" validate:"required"`
	Name           string     `json:"name" validate:"required"`
	Description    *string    `json:"description,omitempty"`
	CronExpression string     `json:"cron_expression" validate:"required"`
	Timezone       string     `json:"timezone"`
	InputData      types.JSON `json:"input_data,omitempty"`
}

// UpdateRequest represents schedule update request
type UpdateRequest struct {
	Name           *string    `json:"name,omitempty"`
	Description    *string    `json:"description,omitempty"`
	CronExpression *string    `json:"cron_expression,omitempty"`
	Timezone       *string    `json:"timezone,omitempty"`
	InputData      types.JSON `json:"input_data,omitempty"`
}

// ScheduleResponse represents schedule in responses
type ScheduleResponse struct {
	ID              string     `json:"id"`
	WorkflowID      string     `json:"workflow_id"`
	WorkspaceID     string     `json:"workspace_id"`
	Name            string     `json:"name"`
	Description     *string    `json:"description,omitempty"`
	CronExpression  string     `json:"cron_expression"`
	Timezone        string     `json:"timezone"`
	IsActive        bool       `json:"is_active"`
	InputData       types.JSON `json:"input_data,omitempty"`
	NextRunAt       *string    `json:"next_run_at,omitempty"`
	LastRunAt       *string    `json:"last_run_at,omitempty"`
	LastExecutionID *string    `json:"last_execution_id,omitempty"`
	RunCount        int        `json:"run_count"`
	CreatedAt       string     `json:"created_at"`
	UpdatedAt       string     `json:"updated_at"`
}

// ToScheduleResponse converts domain schedule to response
func ToScheduleResponse(s *schedule.Schedule) ScheduleResponse {
	resp := ScheduleResponse{
		ID:             s.ID.String(),
		WorkflowID:     s.WorkflowID.String(),
		WorkspaceID:    s.WorkspaceID.String(),
		Name:           s.Name,
		Description:    s.Description,
		CronExpression: s.CronExpression,
		Timezone:       s.Timezone,
		IsActive:       s.IsActive,
		InputData:      s.InputData,
		RunCount:       s.RunCount,
		CreatedAt:      s.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:      s.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}

	if s.NextRunAt != nil {
		str := s.NextRunAt.Format("2006-01-02T15:04:05Z")
		resp.NextRunAt = &str
	}
	if s.LastRunAt != nil {
		str := s.LastRunAt.Format("2006-01-02T15:04:05Z")
		resp.LastRunAt = &str
	}
	if s.LastExecutionID != nil {
		str := s.LastExecutionID.String()
		resp.LastExecutionID = &str
	}

	return resp
}
