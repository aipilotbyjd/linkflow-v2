package execution

import "errors"

// Domain-specific errors for execution aggregate
var (
	ErrExecutionNotFound      = errors.New("execution not found")
	ErrExecutionAlreadyDone   = errors.New("execution already completed")
	ErrExecutionNotRunning    = errors.New("execution is not running")
	ErrExecutionCancelled     = errors.New("execution was canceled")
	ErrExecutionTimeout       = errors.New("execution timed out")
	ErrCannotCancel           = errors.New("cannot cancel execution in current state")
	ErrCannotRetry            = errors.New("cannot retry execution")
	ErrRetryLimitExceeded     = errors.New("retry limit exceeded")
	ErrNodeExecutionNotFound  = errors.New("node execution not found")
	ErrNodeExecutionFailed    = errors.New("node execution failed")
	ErrInvalidTriggerData     = errors.New("invalid trigger data")
	ErrInvalidInputData       = errors.New("invalid input data")
	ErrQueueFull              = errors.New("execution queue is full")
	ErrWorkflowNotActive      = errors.New("workflow is not active")
	ErrInvalidStateTransition = errors.New("invalid execution state transition")
)
