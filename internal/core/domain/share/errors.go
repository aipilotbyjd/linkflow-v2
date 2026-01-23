package share

import "errors"

var (
	ErrShareNotFound       = errors.New("share not found")
	ErrShareAlreadyExists  = errors.New("resource is already shared with this user")
	ErrPermissionDenied    = errors.New("permission denied")
	ErrInvalidResourceType = errors.New("invalid resource type")
	ErrInvalidPermission   = errors.New("invalid permission")

	// Validation errors
	ErrResourceTypeRequired    = errors.New("resource type is required")
	ErrResourceIDRequired      = errors.New("resource ID is required")
	ErrResourceNameRequired    = errors.New("resource name is required")
	ErrSharedByRequired        = errors.New("shared by ID is required")
	ErrSharedByEmailRequired   = errors.New("shared by email is required")
	ErrSharedWithRequired      = errors.New("shared with ID is required")
	ErrSharedWithEmailRequired = errors.New("shared with email is required")
)
