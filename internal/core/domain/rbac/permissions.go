package rbac

// Scopes
const (
	ScopeWorkspace  = "workspace"
	ScopeMember     = "member"
	ScopeWorkflow   = "workflow"
	ScopeCredential = "credential"
	ScopeSchedule   = "schedule"
	ScopeBilling    = "billing"
	ScopeSystem     = "system"
)

// Workspace Permissions
const (
	PermWorkspaceRead     = "workspace:read"
	PermWorkspaceWrite    = "workspace:write" // Update settings
	PermWorkspaceDelete   = "workspace:delete"
	PermWorkspaceAudit    = "workspace:audit"
	PermWorkspaceTransfer = "workspace:transfer"
)

// Member Permissions
const (
	PermMemberRead   = "member:read"
	PermMemberWrite  = "member:write" // Invite/Update role
	PermMemberDelete = "member:delete"
)

// Workflow Permissions
const (
	PermWorkflowRead    = "workflow:read"
	PermWorkflowWrite   = "workflow:write" // Create/Update
	PermWorkflowDelete  = "workflow:delete"
	PermWorkflowExecute = "workflow:execute"
	PermWorkflowPublish = "workflow:publish"
)

// Credential Permissions
const (
	PermCredentialRead   = "credential:read"
	PermCredentialWrite  = "credential:write"
	PermCredentialDelete = "credential:delete"
	PermCredentialUse    = "credential:use" // Implicit for editors usually
)

// Schedule Permissions
const (
	PermScheduleRead   = "schedule:read"
	PermScheduleWrite  = "schedule:write"
	PermScheduleDelete = "schedule:delete"
)

// Billing Permissions
const (
	PermBillingRead  = "billing:read"
	PermBillingWrite = "billing:write"
)

// List of all permissions for seeding/validation
var AllPermissions = []Permission{
	// Workspace
	{ID: PermWorkspaceRead, Scope: ScopeWorkspace, Name: "View Workspace", Description: "View workspace details and dashboard"},
	{ID: PermWorkspaceWrite, Scope: ScopeWorkspace, Name: "Update Workspace", Description: "Update workspace settings"},
	{ID: PermWorkspaceDelete, Scope: ScopeWorkspace, Name: "Delete Workspace", Description: "Delete the entire workspace"},
	{ID: PermWorkspaceAudit, Scope: ScopeWorkspace, Name: "View Audit Logs", Description: "View workspace audit logs"},
	{ID: PermWorkspaceTransfer, Scope: ScopeWorkspace, Name: "Transfer Ownership", Description: "Transfer workspace ownership"},

	// Member
	{ID: PermMemberRead, Scope: ScopeMember, Name: "View Members", Description: "View workspace members"},
	{ID: PermMemberWrite, Scope: ScopeMember, Name: "Manage Members", Description: "Invite members and update roles"},
	{ID: PermMemberDelete, Scope: ScopeMember, Name: "Remove Members", Description: "Remove members from workspace"},

	// Workflow
	{ID: PermWorkflowRead, Scope: ScopeWorkflow, Name: "View Workflows", Description: "View workflows"},
	{ID: PermWorkflowWrite, Scope: ScopeWorkflow, Name: "Edit Workflows", Description: "Create and update workflows"},
	{ID: PermWorkflowDelete, Scope: ScopeWorkflow, Name: "Delete Workflows", Description: "Delete workflows"},
	{ID: PermWorkflowExecute, Scope: ScopeWorkflow, Name: "Execute Workflows", Description: "Manually execute workflows"},
	{ID: PermWorkflowPublish, Scope: ScopeWorkflow, Name: "Publish Workflows", Description: "Activate or deactivate workflows"},

	// Credential
	{ID: PermCredentialRead, Scope: ScopeCredential, Name: "View Credentials", Description: "View credential names (not secrets)"},
	{ID: PermCredentialWrite, Scope: ScopeCredential, Name: "Manage Credentials", Description: "Create and update credentials"},
	{ID: PermCredentialDelete, Scope: ScopeCredential, Name: "Delete Credentials", Description: "Delete credentials"},

	// Schedule
	{ID: PermScheduleRead, Scope: ScopeSchedule, Name: "View Schedules", Description: "View schedules"},
	{ID: PermScheduleWrite, Scope: ScopeSchedule, Name: "Manage Schedules", Description: "Create and update schedules"},
	{ID: PermScheduleDelete, Scope: ScopeSchedule, Name: "Delete Schedules", Description: "Delete schedules"},

	// Billing
	{ID: PermBillingRead, Scope: ScopeBilling, Name: "View Billing", Description: "View subscription and invoices"},
	{ID: PermBillingWrite, Scope: ScopeBilling, Name: "Manage Billing", Description: "Update subscription and payment methods"},
}
