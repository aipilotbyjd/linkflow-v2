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
)
