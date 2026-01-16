package template

import "errors"

var (
	ErrTemplateNotFound = errors.New("template not found")
	ErrInvalidTemplate  = errors.New("invalid template")
	ErrCategoryNotFound = errors.New("category not found")
)
