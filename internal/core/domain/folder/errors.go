package folder

import "github.com/linkflow-ai/linkflow/internal/shared/errors"

var (
	ErrFolderNotFound    = errors.NewNotFoundError("folder")
	ErrFolderNotEmpty    = errors.NewValidationError("folder is not empty")
	ErrFolderNameExists  = errors.NewConflictError("folder with this name already exists")
	ErrInvalidParent     = errors.NewValidationError("invalid parent folder")
	ErrCircularReference = errors.NewValidationError("cannot create circular folder reference")
)
