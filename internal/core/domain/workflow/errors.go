package workflow

import "errors"

// Domain-specific errors for workflow aggregate
var (
	ErrWorkflowNotFound        = errors.New("workflow not found")
	ErrWorkflowNameExists      = errors.New("workflow name already exists")
	ErrWorkflowNotActive       = errors.New("workflow is not active")
	ErrWorkflowAlreadyActive   = errors.New("workflow is already active")
	ErrEmptyWorkflow           = errors.New("workflow has no nodes")
	ErrNoTriggerNode           = errors.New("workflow requires a trigger node")
	ErrInvalidWorkflow         = errors.New("invalid workflow configuration")
	ErrCircularDependency      = errors.New("circular dependency detected")
	ErrInvalidNode             = errors.New("invalid node configuration")
	ErrNodeNotFound            = errors.New("node not found")
	ErrInvalidConnection       = errors.New("invalid connection")
	ErrVersionNotFound         = errors.New("version not found")
	ErrCannotModifyActive      = errors.New("cannot modify active workflow")
	ErrCannotDeleteActive      = errors.New("cannot delete active workflow")
	ErrFolderNotFound          = errors.New("folder not found")
	ErrFolderNameExists        = errors.New("folder name already exists")
	ErrFolderHasChildren       = errors.New("folder has children")
	ErrFolderHasWorkflows      = errors.New("folder contains workflows")
	ErrCannotArchiveActive     = errors.New("cannot archive active workflow")
	ErrInvalidStatusTransition = errors.New("invalid status transition")
	ErrInsufficientVariants    = errors.New("experiment must have at least 2 variants")
	ErrExperimentNotRunning    = errors.New("experiment is not running")
	ErrWorkflowHasCycle        = errors.New("workflow has infinite loop (cycle) detected")

	// Validation errors
	ErrWorkflowNameRequired   = errors.New("workflow name is required")
	ErrWorkflowNameTooLong    = errors.New("workflow name must be at most 255 characters")
	ErrInvalidWorkspaceID     = errors.New("valid workspace ID is required")
	ErrInvalidCreatedBy       = errors.New("valid creator ID is required")
	ErrDescriptionTooLong     = errors.New("description must be at most 2000 characters")
	ErrSelfReferencingError   = errors.New("error workflow cannot reference itself")
	ErrInvalidTimeoutValue    = errors.New("timeout must be between 1 and 86400 seconds")
	ErrInvalidMaxRetriesValue = errors.New("max retries must be between 0 and 10")
)
