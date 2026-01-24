# LinkFlow Dynamic RBAC Architecture (v2)

This document outlines the Database-Driven Role-Based Access Control (RBAC) system for LinkFlow. This system replaces hardcoded role strings with a flexible, enterprise-grade permission engine.

## 1. Core Concept

Permissions are the atomic units of access (e.g., `workflow:create`). Roles are collections of permissions. Users are assigned a Role within a specific Workspace.

*   **Dynamic:** Roles can be created, edited, and deleted at runtime.
*   **Scoped:** Roles belong to a specific workspace (except System Defaults).
*   **Granular:** Access control checks permissions, not role names.

## 2. Database Schema

### `permissions`
The "Menu" of all possible actions in the system.

| Column | Type | Description |
| :--- | :--- | :--- |
| `id` | `varchar` | Unique key (e.g., `workflow:execute`) |
| `scope` | `varchar` | Resource category (e.g., `workflow`, `billing`) |
| `name` | `varchar` | Human-readable name |
| `description` | `text` | Help text for UI |

### `roles`
The containers for permissions.

| Column | Type | Description |
| :--- | :--- | :--- |
| `id` | `uuid` | Primary Key |
| `workspace_id` | `uuid` | `NULL` for System Defaults, UUID for Custom Roles |
| `name` | `varchar` | Display name (e.g., "DevOps Lead") |
| `is_protected` | `bool` | If true, cannot be deleted (e.g., "Owner") |
| `is_default` | `bool` | If true, assigned to new members automatically |

### `role_permissions` (Join Table)
Links Roles to Permissions.

### `workspace_members`
Links Users to Roles.

| Column | Type | Description |
| :--- | :--- | :--- |
| `user_id` | `uuid` | The user |
| `workspace_id` | `uuid` | The context |
| `role_id` | `uuid` | FK to `roles` table |

## 3. Permission Catalog

### Workspace Scope
*   `workspace:read` - View dashboard & details
*   `workspace:update` - Rename, change icon/settings
*   `workspace:delete` - **Critical:** Delete entire workspace
*   `workspace:audit` - View audit logs

### Member Scope
*   `member:read` - View member list
*   `member:invite` - Invite new users
*   `member:update` - Change another user's role
*   `member:remove` - Remove a user

### Workflow Scope
*   `workflow:read` - View workflows
*   `workflow:create` - Create new workflows
*   `workflow:update` - Edit workflow graph/settings
*   `workflow:delete` - Delete workflows
*   `workflow:publish` - Activate/Deactivate (Production)
*   `workflow:execute` - Manually trigger runs

### Credential Scope
*   `credential:read` - View names of credentials
*   `credential:create` - Add new secrets
*   `credential:update` - Update secret values
*   `credential:delete` - Remove secrets
*   `credential:use` - Use secrets in nodes (implicit for Editors)

## 4. Default System Roles

These roles are seeded into the database and available in all workspaces.

| Role | Description | Key Permissions |
| :--- | :--- | :--- |
| **Owner** | The Creator | `*` (All Permissions) |
| **Admin** | Workspace Manager | All except `workspace:delete` |
| **Editor** | Standard User | `workflow:*`, `credential:*`, `member:read` |
| **Viewer** | Read-Only | `*:read` only |

## 5. Developer Guide

### Checking Permissions (Middleware)
Use the `RequirePermission` middleware for route protection.

```go
// Protect a route
r.With(middleware.RequirePermission(rbac.PermWorkflowCreate)).Post("/", createWorkflow)
```

### Checking Permissions (In Handler)
Use the `Member` object from context.

```go
func Handle(w http.ResponseWriter, r *http.Request) {
    wsCtx := middleware.GetWorkspaceFromContext(r.Context())
    
    if !wsCtx.Member.HasPermission(rbac.PermBillingRead) {
        common.Forbidden(w, "Cannot view invoices")
        return
    }
}
```

### Creating a New Role (Service Layer)
```go
role := rbac.NewRole(workspaceID, "Intern", "Limited access")
repo.CreateRole(ctx, role)
repo.AssignPermissions(ctx, role.ID, []string{rbac.PermWorkflowRead})
```

## 6. Migration Strategy

1.  **Schema:** Run GORM Auto-migration to create tables.
2.  **Seed:** Run the `InitSystemRoles` function on startup to populate `permissions` and default `roles`.
3.  **Data:** Existing `role` strings in `workspace_members` must be migrated to `role_id` foreign keys mapping to the new System Roles.
