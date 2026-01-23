package folder

import "github.com/linkflow-ai/linkflow/internal/shared/errors"

var (
	ErrFolderNotFound    = errors.NewNotFoundError("folder")
	ErrFolderNotEmpty    = errors.NewValidationError("folder is not empty")
	ErrFolderNameExists  = errors.NewConflictError("folder with this name already exists")
	ErrInvalidParent     = errors.NewValidationError("invalid parent folder")
	ErrCircularReference = errors.NewValidationError("cannot create circular folder reference")

	// Validation errors
	ErrNameRequired       = errors.NewValidationError("folder name is required")
	ErrNameTooLong        = errors.NewValidationError("folder name must be at most 100 characters")
	ErrInvalidWorkspaceID = errors.NewValidationError("valid workspace ID is required")
	ErrInvalidCreatedBy   = errors.NewValidationError("valid creator ID is required")
	ErrDescriptionTooLong = errors.NewValidationError("description must be at most 500 characters")
	ErrInvalidColorFormat = errors.NewValidationError("color must be a valid hex code")
)
