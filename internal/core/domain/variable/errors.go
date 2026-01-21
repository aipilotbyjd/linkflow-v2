package variable

import "errors"

var (
	ErrVariableNotFound      = errors.New("variable not found")
	ErrVariableKeyExists     = errors.New("variable key already exists")
	ErrEnvironmentNotFound   = errors.New("environment not found")
	ErrEnvironmentNameExists = errors.New("environment name already exists")
	ErrCannotDeleteDefault   = errors.New("cannot delete default environment")
	ErrInvalidScope          = errors.New("invalid variable scope")
)
