package schedule

import "errors"

// Domain-specific errors for schedule aggregate
var (
	ErrScheduleNotFound        = errors.New("schedule not found")
	ErrScheduleNameExists      = errors.New("schedule name already exists")
	ErrInvalidCronExpression   = errors.New("invalid cron expression")
	ErrInvalidTimezone         = errors.New("invalid timezone")
	ErrScheduleAlreadyActive   = errors.New("schedule is already active")
	ErrScheduleAlreadyInactive = errors.New("schedule is already inactive")
	ErrWorkflowNotActive       = errors.New("cannot schedule inactive workflow")
	ErrScheduleLimitReached    = errors.New("schedule limit reached for plan")

	// Validation errors
	ErrScheduleNameRequired = errors.New("schedule name is required")
	ErrScheduleNameTooLong  = errors.New("schedule name must be at most 100 characters")
	ErrInvalidWorkflowID    = errors.New("valid workflow ID is required")
	ErrInvalidWorkspaceID   = errors.New("valid workspace ID is required")
	ErrInvalidCreatedBy     = errors.New("valid creator ID is required")
	ErrCronRequired         = errors.New("cron expression is required")
	ErrCronTooFrequent      = errors.New("schedule frequency exceeds plan limit")
)
