package schedule

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/linkflow-ai/linkflow/internal/shared/types"
	"github.com/robfig/cron/v3"
	"gorm.io/gorm"
)

// Schedule entity (aggregate root)
type Schedule struct {
	ID              uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	WorkflowID      uuid.UUID      `gorm:"type:uuid;index;not null" json:"workflow_id"`
	WorkspaceID     uuid.UUID      `gorm:"type:uuid;index;not null" json:"workspace_id"`
	CreatedBy       uuid.UUID      `gorm:"type:uuid;not null" json:"created_by"`
	Name            string         `gorm:"size:100;not null" json:"name"`
	Description     *string        `gorm:"type:text" json:"description,omitempty"`
	CronExpression  string         `gorm:"size:100;not null" json:"cron_expression"`
	Timezone        string         `gorm:"size:50;default:UTC" json:"timezone"`
	IsActive        bool           `gorm:"default:true" json:"is_active"`
	InputData       types.JSON     `gorm:"type:jsonb" json:"input_data,omitempty"`
	NextRunAt       *time.Time     `gorm:"index" json:"next_run_at,omitempty"`
	LastRunAt       *time.Time     `json:"last_run_at,omitempty"`
	LastExecutionID *uuid.UUID     `gorm:"type:uuid" json:"last_execution_id,omitempty"`
	RunCount        int            `gorm:"default:0" json:"run_count"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Schedule) TableName() string {
	return "schedules"
}

// NewSchedule creates a new schedule
func NewSchedule(workflowID, workspaceID, createdBy uuid.UUID, name, cronExpression string) *Schedule {
	return &Schedule{
		ID:             uuid.New(),
		WorkflowID:     workflowID,
		WorkspaceID:    workspaceID,
		CreatedBy:      createdBy,
		Name:           name,
		CronExpression: cronExpression,
		Timezone:       "UTC",
		IsActive:       true,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
}

// GetWorkspaceID implements the WorkspaceOwned interface
func (s *Schedule) GetWorkspaceID() uuid.UUID {
	return s.WorkspaceID
}

// WithTimezone sets the timezone
func (s *Schedule) WithTimezone(tz string) *Schedule {
	s.Timezone = tz
	return s
}

// WithDescription sets the description
func (s *Schedule) WithDescription(desc string) *Schedule {
	s.Description = &desc
	return s
}

// WithInputData sets the input data
func (s *Schedule) WithInputData(data types.JSON) *Schedule {
	s.InputData = data
	return s
}

// Update updates schedule details
func (s *Schedule) Update(name string, description *string, cronExpression, timezone string) {
	s.Name = name
	s.Description = description
	s.CronExpression = cronExpression
	if timezone != "" {
		s.Timezone = timezone
	}
	s.UpdatedAt = time.Now()
}

// UpdateInputData updates the input data
func (s *Schedule) UpdateInputData(data types.JSON) {
	s.InputData = data
	s.UpdatedAt = time.Now()
}

// Activate activates the schedule
func (s *Schedule) Activate() {
	s.IsActive = true
	s.UpdatedAt = time.Now()
}

// Deactivate deactivates the schedule
func (s *Schedule) Deactivate() {
	s.IsActive = false
	s.UpdatedAt = time.Now()
}

// SetNextRunAt sets the next run time
func (s *Schedule) SetNextRunAt(nextRun time.Time) {
	s.NextRunAt = &nextRun
	s.UpdatedAt = time.Now()
}

// RecordRun records a schedule run
func (s *Schedule) RecordRun(executionID uuid.UUID) {
	now := time.Now()
	s.LastRunAt = &now
	s.LastExecutionID = &executionID
	s.RunCount++
	s.UpdatedAt = now
}

// IsDue checks if the schedule is due to run
func (s *Schedule) IsDue() bool {
	if !s.IsActive {
		return false
	}
	if s.NextRunAt == nil {
		return false
	}
	return s.NextRunAt.Before(time.Now()) || s.NextRunAt.Equal(time.Now())
}

// CalculateNextRun calculates the next run time based on cron expression
func (s *Schedule) CalculateNextRun() (time.Time, error) {
	parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
	schedule, err := parser.Parse(s.CronExpression)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid cron expression: %w", err)
	}

	loc, err := time.LoadLocation(s.Timezone)
	if err != nil {
		loc = time.UTC
	}

	now := time.Now().In(loc)
	return schedule.Next(now), nil
}

// MarkExecuted marks the schedule as executed and calculates next run time
func (s *Schedule) MarkExecuted() {
	now := time.Now()
	s.LastRunAt = &now
	s.RunCount++
	s.UpdatedAt = now

	// Calculate next run time
	if nextRun, err := s.CalculateNextRun(); err == nil {
		s.NextRunAt = &nextRun
	}
}

// Pause pauses the schedule (alias for Deactivate)
func (s *Schedule) Pause() {
	s.Deactivate()
}

// Resume resumes the schedule (alias for Activate)
func (s *Schedule) Resume() {
	s.Activate()
	// Recalculate next run time when resuming
	if nextRun, err := s.CalculateNextRun(); err == nil {
		s.NextRunAt = &nextRun
	}
}

// Enabled returns true if the schedule is active
func (s *Schedule) Enabled() bool {
	return s.IsActive
}

// ValidateCronMinInterval validates that a cron expression doesn't run more frequently than minIntervalMinutes
func ValidateCronMinInterval(cronExpr string, minIntervalMinutes int) error {
	if minIntervalMinutes <= 0 {
		return nil // No restriction
	}

	parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
	schedule, err := parser.Parse(cronExpr)
	if err != nil {
		return fmt.Errorf("invalid cron expression: %w", err)
	}

	// Calculate the interval between two consecutive runs
	now := time.Now()
	next1 := schedule.Next(now)
	next2 := schedule.Next(next1)

	intervalMinutes := int(next2.Sub(next1).Minutes())

	if intervalMinutes < minIntervalMinutes {
		return fmt.Errorf("cron interval %d minutes is less than minimum %d minutes", intervalMinutes, minIntervalMinutes)
	}

	return nil
}
