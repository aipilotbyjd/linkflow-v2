package template

import "errors"

var (
	ErrTemplateNotFound = errors.New("template not found")
	ErrInvalidTemplate  = errors.New("invalid template")
	ErrCategoryNotFound = errors.New("category not found")

	// Validation errors
	ErrNameRequired        = errors.New("template name is required")
	ErrNameTooLong         = errors.New("template name must be at most 100 characters")
	ErrCategoryRequired    = errors.New("category is required")
	ErrNodesRequired       = errors.New("nodes are required")
	ErrConnectionsRequired = errors.New("connections are required")
)
