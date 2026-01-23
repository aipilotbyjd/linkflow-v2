package workspace

import "errors"

// Domain-specific errors for workspace aggregate
var (
	ErrWorkspaceNotFound       = errors.New("workspace not found")
	ErrSlugAlreadyExists       = errors.New("workspace slug already exists")
	ErrMemberNotFound          = errors.New("member not found")
	ErrMemberAlreadyExists     = errors.New("user is already a member")
	ErrCannotRemoveOwner       = errors.New("cannot remove workspace owner")
	ErrCannotDemoteOwner       = errors.New("cannot demote workspace owner")
	ErrCannotChangeOwnerRole   = errors.New("cannot change owner role")
	ErrInvitationNotFound      = errors.New("invitation not found")
	ErrInvitationExpired       = errors.New("invitation has expired")
	ErrInvitationAlreadyUsed   = errors.New("invitation already used")
	ErrInvitationAlreadyExists = errors.New("pending invitation already exists")
	ErrInsufficientPermissions = errors.New("insufficient permissions")
	ErrPermissionDenied        = errors.New("permission denied")
	ErrNotAMember              = errors.New("user is not a member of workspace")
	ErrCannotLeaveAsOwner      = errors.New("owner cannot leave workspace")
	ErrLastAdmin               = errors.New("cannot remove the last admin")
	ErrSelfRoleChange          = errors.New("cannot change own role")
	ErrInvalidRole             = errors.New("invalid role")
	ErrPlanLimitReached        = errors.New("plan member limit reached")

	// Validation errors
	ErrWorkspaceNameRequired = errors.New("workspace name is required")
	ErrWorkspaceNameTooLong  = errors.New("workspace name must be at most 100 characters")
	ErrSlugRequired          = errors.New("workspace slug is required")
	ErrSlugTooLong           = errors.New("workspace slug must be at most 100 characters")
	ErrInvalidSlugFormat     = errors.New("workspace slug must be lowercase alphanumeric with hyphens")
	ErrInvalidOwnerID        = errors.New("valid owner ID is required")
	ErrDescriptionTooLong    = errors.New("description must be at most 500 characters")

	// Member validation errors
	ErrInvalidWorkspaceID = errors.New("valid workspace ID is required")
	ErrInvalidUserID      = errors.New("valid user ID is required")
	ErrInvalidInviterID   = errors.New("valid inviter ID is required")

	// Invitation validation errors
	ErrEmailRequired      = errors.New("email is required")
	ErrInvalidEmailFormat = errors.New("invalid email format")
	ErrTokenRequired      = errors.New("invitation token is required")
	ErrInvalidExpiresIn   = errors.New("expiration duration must be positive")
)
