package dto

import "github.com/linkflow-ai/linkflow/internal/domain/models"

// Schedule requests

type CreateScheduleRequest struct {
	WorkflowID     string      `json:"workflow_id" validate:"required,uuid"`
	Name           string      `json:"name" validate:"required,min=1,max=100"`
	Description    *string     `json:"description,omitempty" validate:"omitempty,max=500"`
	CronExpression string      `json:"cron_expression" validate:"required,cron"`
	Timezone       string      `json:"timezone" validate:"required"`
	InputData      models.JSON `json:"input_data,omitempty"`
}

type UpdateScheduleRequest struct {
	Name           *string     `json:"name,omitempty" validate:"omitempty,min=1,max=100"`
	Description    *string     `json:"description,omitempty" validate:"omitempty,max=500"`
	CronExpression *string     `json:"cron_expression,omitempty" validate:"omitempty,cron"`
	Timezone       *string     `json:"timezone,omitempty"`
	InputData      models.JSON `json:"input_data,omitempty"`
}

// Schedule responses

type ScheduleResponse struct {
	ID              string      `json:"id"`
	WorkflowID      string      `json:"workflow_id"`
	WorkspaceID     string      `json:"workspace_id"`
	CreatedBy       string      `json:"created_by"`
	Name            string      `json:"name"`
	Description     *string     `json:"description,omitempty"`
	CronExpression  string      `json:"cron_expression"`
	Timezone        string      `json:"timezone"`
	IsActive        bool        `json:"is_active"`
	InputData       interface{} `json:"input_data,omitempty"`
	NextRunAt       *int64      `json:"next_run_at,omitempty"`
	LastRunAt       *int64      `json:"last_run_at,omitempty"`
	LastExecutionID *string     `json:"last_execution_id,omitempty"`
	RunCount        int         `json:"run_count"`
	CreatedAt       int64       `json:"created_at"`
	UpdatedAt       int64       `json:"updated_at"`
}
