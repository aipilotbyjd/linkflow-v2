package rbac

// System Role Names
const (
	RoleOwner  = "Owner"
	RoleAdmin  = "Admin"
	RoleEditor = "Editor"
	RoleViewer = "Viewer"
)

// GetSystemRoles returns the default system roles with their permissions
func GetSystemRoles() []*Role {
	// Create Roles
	owner := NewSystemRole(RoleOwner, "Full access to the workspace")
	admin := NewSystemRole(RoleAdmin, "Can manage everything except deleting the workspace")
	editor := NewSystemRole(RoleEditor, "Can manage workflows and credentials")
	viewer := NewSystemRole(RoleViewer, "Read-only access")

	// Assign Permissions

	// 1. Viewer (Base Layer)
	viewerPerms := []string{
		PermWorkspaceRead,
		PermMemberRead,
		PermWorkflowRead,
		PermCredentialRead,
		PermScheduleRead,
		PermBillingRead,
	}
	viewer.Permissions = mapPermissions(viewerPerms)

	// 2. Editor (Viewer + Edit Capabilities)
	editorPerms := append(viewerPerms,
		PermWorkflowWrite, PermWorkflowDelete, PermWorkflowExecute, PermWorkflowPublish,
		PermCredentialWrite, PermCredentialDelete, PermCredentialUse,
		PermScheduleWrite, PermScheduleDelete,
	)
	editor.Permissions = mapPermissions(editorPerms)

	// 3. Admin (Editor + Management)
	adminPerms := append(editorPerms,
		PermWorkspaceWrite, PermWorkspaceAudit, PermWorkspaceTransfer,
		PermMemberWrite, PermMemberDelete,
		PermBillingWrite,
	)
	admin.Permissions = mapPermissions(adminPerms)

	// 4. Owner (All Permissions)
	// Theoretically it's "everything", but explicitly:
	ownerPerms := append(adminPerms, PermWorkspaceDelete)
	owner.Permissions = mapPermissions(ownerPerms)

	return []*Role{owner, admin, editor, viewer}
}

// NewSystemRole creates a new system role
func NewSystemRole(name, description string) *Role {
	role := NewRole(nil, name, description)
	role.IsProtected = true
	role.IsDefault = (name == RoleEditor)
	// Actually, usually "Editor" or "Member" is default. Let's assume no default set here, logic handled elsewhere.
	// Re-reading plan: "is_default: If true, assigned to new members automatically"
	// Typically "Member" (Editor equivalent) is default.
	if name == RoleEditor {
		role.IsDefault = true
	}
	return role
}

// Helper to convert ID strings to Permission structs (partial)
func mapPermissions(ids []string) []Permission {
	perms := make([]Permission, len(ids))
	for i, id := range ids {
		// We only set ID here. In a real scenario, we'd look up the full Permission object
		// but for seeding/logic, ID is usually enough.
		perms[i] = Permission{ID: id}
	}
	return perms
}
