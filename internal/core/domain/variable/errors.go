package variable

import "errors"

var (
	ErrVariableNotFound      = errors.New("variable not found")
	ErrVariableKeyExists     = errors.New("variable key already exists")
	ErrEnvironmentNotFound   = errors.New("environment not found")
	ErrEnvironmentNameExists = errors.New("environment name already exists")
	ErrCannotDeleteDefault   = errors.New("cannot delete default environment")
	ErrInvalidScope          = errors.New("invalid variable scope")

	// Validation errors
	ErrKeyRequired        = errors.New("variable key is required")
	ErrKeyTooLong         = errors.New("variable key must be at most 100 characters")
	ErrValueRequired      = errors.New("variable value is required")
	ErrInvalidWorkspaceID = errors.New("valid workspace ID is required")
	ErrInvalidCreatedBy   = errors.New("valid creator ID is required")
)
