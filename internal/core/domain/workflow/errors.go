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
)
