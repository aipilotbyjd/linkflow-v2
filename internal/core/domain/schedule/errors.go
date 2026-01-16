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
)
