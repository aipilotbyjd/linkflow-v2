package workspace

// Role represents a workspace member role
type Role string

const (
	RoleOwner  Role = "owner"
	RoleAdmin  Role = "admin"
	RoleMember Role = "member"
	RoleViewer Role = "viewer"
)

func (r Role) String() string {
	return string(r)
}

func (r Role) IsValid() bool {
	switch r {
	case RoleOwner, RoleAdmin, RoleMember, RoleViewer:
		return true
	default:
		return false
	}
}

func ParseRole(s string) (Role, bool) {
	role := Role(s)
	return role, role.IsValid()
}

// Level returns the permission level (higher = more permissions)
func (r Role) Level() int {
	switch r {
	case RoleOwner:
		return 100
	case RoleAdmin:
		return 80
	case RoleMember:
		return 50
	case RoleViewer:
		return 10
	default:
		return 0
	}
}

// CanManage checks if this role can manage the target role
func (r Role) CanManage(target Role) bool {
	// Owner can manage everyone except other owners
	if r == RoleOwner {
		return target != RoleOwner
	}
	// Admin can manage members and viewers
	if r == RoleAdmin {
		return target == RoleMember || target == RoleViewer
	}
	return false
}

// Permissions returns the permissions for this role
func (r Role) Permissions() []string {
	base := []string{"workspace:read", "workflow:read", "execution:read"}
	
	switch r {
	case RoleOwner:
		return append(base,
			"workspace:write", "workspace:delete", "workspace:transfer",
			"member:write", "member:delete",
			"workflow:write", "workflow:delete", "workflow:execute",
			"credential:write", "credential:delete",
			"schedule:write", "schedule:delete",
			"billing:read", "billing:write",
		)
	case RoleAdmin:
		return append(base,
			"workspace:write",
			"member:write", "member:delete",
			"workflow:write", "workflow:delete", "workflow:execute",
			"credential:write", "credential:delete",
			"schedule:write", "schedule:delete",
		)
	case RoleMember:
		return append(base,
			"workflow:write", "workflow:execute",
			"credential:write",
			"schedule:write",
		)
	case RoleViewer:
		return base
	default:
		return nil
	}
}

// HasPermission checks if this role has a specific permission
func (r Role) HasPermission(permission string) bool {
	for _, p := range r.Permissions() {
		if p == permission {
			return true
		}
	}
	return false
}
