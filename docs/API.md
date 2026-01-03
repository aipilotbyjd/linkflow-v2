# LinkFlow API Documentation

**Base URL:** `http://localhost:8090/api/v1`  
**Version:** 1.0.0

## Table of Contents

- [Response Format](#response-format)
- [Authentication](#authentication)
- [Error Handling](#error-handling)
- [Rate Limiting](#rate-limiting)
- [Endpoints](#endpoints)
  - [Auth](#auth)
  - [Users](#users)
  - [Workspaces](#workspaces)
  - [Folders](#folders)
  - [Workflows](#workflows)
  - [Executions](#executions)
  - [Credentials](#credentials)
  - [Schedules](#schedules)
  - [Templates](#templates)
  - [Billing](#billing)
  - [Webhooks](#webhooks)
  - [Analytics](#analytics)
  - [Alerts](#alerts)
  - [Audit Logs](#audit-logs)
  - [Environment Variables](#environment-variables)
  - [Comments](#comments)
  - [Node Types](#node-types)
  - [Health](#health)
  - [Admin](#admin)

---

## Response Format

All API responses use a consistent enhanced response format:

```json
{
  "success": true,
  "data": { ... },
  "links": {
    "self": "/api/v1/workspaces/123/workflows",
    "next": "/api/v1/workspaces/123/workflows?page=2",
    "prev": null,
    "first": "/api/v1/workspaces/123/workflows?page=1",
    "last": "/api/v1/workspaces/123/workflows?page=5"
  },
  "meta": {
    "total": 100,
    "page": 1,
    "per_page": 20,
    "total_pages": 5
  },
  "request_id": "req_abc123",
  "timestamp": 1704067200
}
```

### Actions (HATEOAS)

List and detail responses include available actions:

```json
{
  "data": {
    "id": "wf_123",
    "name": "My Workflow",
    "status": "active",
    "actions": [
      { "name": "execute", "method": "POST", "href": "/api/v1/.../execute", "label": "Execute" },
      { "name": "deactivate", "method": "POST", "href": "/api/v1/.../deactivate", "label": "Deactivate" }
    ]
  }
}
```

### Query Parameters

| Parameter | Description | Example |
|-----------|-------------|---------|
| `page` | Page number (default: 1) | `?page=2` |
| `per_page` | Items per page (default: 20, max: 100) | `?per_page=50` |
| `fields` | Select specific fields | `?fields=id,name,status` |
| `include` | Include related data | `?include=schedules,credentials` |
| `sort` | Sort by field | `?sort=-created_at` |

---

## Authentication

All protected endpoints require a Bearer token in the Authorization header:

```
Authorization: Bearer <access_token>
```

### Token Lifecycle

1. **Access Token:** Valid for 15 minutes
2. **Refresh Token:** Valid for 7 days
3. Use `/auth/refresh` to get new tokens before expiry

---

## Error Handling

### Error Response Format

```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Validation failed",
    "details": [
      { "field": "email", "message": "Invalid email format" }
    ]
  },
  "request_id": "req_abc123",
  "timestamp": 1704067200
}
```

### HTTP Status Codes

| Code | Description |
|------|-------------|
| 200 | Success |
| 201 | Created |
| 204 | No Content |
| 400 | Bad Request |
| 401 | Unauthorized |
| 403 | Forbidden |
| 404 | Not Found |
| 409 | Conflict |
| 422 | Validation Error |
| 429 | Rate Limited |
| 500 | Internal Server Error |

---

## Rate Limiting

| Route Type | Limit |
|------------|-------|
| Public endpoints | 100 requests/minute |
| Protected endpoints | 1000 requests/minute |

Rate limit headers:
- `X-RateLimit-Limit`
- `X-RateLimit-Remaining`
- `X-RateLimit-Reset`

---

## Endpoints

---

## Auth

### Register

Create a new user account.

```
POST /auth/register
```

**Request Body:**

```json
{
  "email": "user@example.com",
  "password": "securePassword123",
  "first_name": "John",
  "last_name": "Doe"
}
```

| Field | Type | Required | Validation |
|-------|------|----------|------------|
| email | string | Yes | Valid email format |
| password | string | Yes | Min 8 characters |
| first_name | string | Yes | 1-100 characters |
| last_name | string | Yes | 1-100 characters |

**Response:** `201 Created`

```json
{
  "success": true,
  "data": {
    "user": {
      "id": "usr_abc123",
      "email": "user@example.com",
      "first_name": "John",
      "last_name": "Doe",
      "email_verified": false,
      "mfa_enabled": false,
      "created_at": 1704067200
    },
    "access_token": "eyJhbGciOiJIUzI1NiIs...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIs...",
    "expires_at": 1704068100
  }
}
```

---

### Login

Authenticate and receive tokens.

```
POST /auth/login
```

**Request Body:**

```json
{
  "email": "user@example.com",
  "password": "securePassword123",
  "mfa_code": "123456"
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| email | string | Yes | User email |
| password | string | Yes | User password |
| mfa_code | string | No | Required if MFA enabled |

**Response:** `200 OK`

```json
{
  "success": true,
  "data": {
    "user": {
      "id": "usr_abc123",
      "email": "user@example.com",
      "first_name": "John",
      "last_name": "Doe",
      "avatar_url": null,
      "email_verified": true,
      "mfa_enabled": false,
      "created_at": 1704067200
    },
    "access_token": "eyJhbGciOiJIUzI1NiIs...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIs...",
    "expires_at": 1704068100
  }
}
```

**MFA Required Response:**

```json
{
  "requires_mfa": true,
  "message": "MFA verification required"
}
```

---

### Refresh Token

Get new access token using refresh token.

```
POST /auth/refresh
```

**Request Body:**

```json
{
  "refresh_token": "eyJhbGciOiJIUzI1NiIs..."
}
```

**Response:** `200 OK`

```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIs...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIs...",
  "expires_at": 1704068100
}
```

---

### Logout

Invalidate all user tokens.

```
POST /auth/logout
```

**Headers:** `Authorization: Bearer <token>`

**Response:** `204 No Content`

---

### Forgot Password

Request password reset email.

```
POST /auth/forgot-password
```

**Request Body:**

```json
{
  "email": "user@example.com"
}
```

**Response:** `200 OK`

```json
{
  "message": "If the email exists, a password reset link has been sent"
}
```

---

### Reset Password

Reset password using token from email.

```
POST /auth/reset-password
```

**Request Body:**

```json
{
  "token": "reset_token_from_email",
  "new_password": "newSecurePassword123"
}
```

**Response:** `200 OK`

```json
{
  "message": "Password has been reset successfully"
}
```

---

### Setup MFA

Generate MFA secret and QR code.

```
POST /auth/mfa/setup
```

**Headers:** `Authorization: Bearer <token>`

**Response:** `200 OK`

```json
{
  "secret": "JBSWY3DPEHPK3PXP",
  "qr_code": "otpauth://totp/LinkFlow:user@example.com?secret=JBSWY3DPEHPK3PXP"
}
```

---

### Verify MFA

Enable MFA after scanning QR code.

```
POST /auth/mfa/verify
```

**Headers:** `Authorization: Bearer <token>`

**Request Body:**

```json
{
  "code": "123456"
}
```

**Response:** `200 OK`

```json
{
  "verified": true
}
```

---

### Disable MFA

Disable MFA (requires current MFA code).

```
DELETE /auth/mfa
```

**Headers:** `Authorization: Bearer <token>`

**Request Body:**

```json
{
  "code": "123456"
}
```

**Response:** `204 No Content`

---

### OAuth Redirect

Redirect to OAuth provider.

```
GET /auth/oauth/{provider}?purpose=login
```

| Parameter | Values |
|-----------|--------|
| provider | `google`, `github`, `microsoft` |
| purpose | `login`, `signup`, `connect` |

**Response:** `307 Temporary Redirect` to OAuth provider

---

### OAuth Callback

Handle OAuth callback (internal use).

```
GET /auth/oauth/{provider}/callback
```

---

## Users

### Get Current User

Get authenticated user's profile.

```
GET /users/me
```

**Headers:** `Authorization: Bearer <token>`

**Response:** `200 OK`

```json
{
  "success": true,
  "data": {
    "id": "usr_abc123",
    "email": "user@example.com",
    "username": "johndoe",
    "first_name": "John",
    "last_name": "Doe",
    "avatar_url": "https://...",
    "timezone": "America/New_York",
    "email_verified": true,
    "mfa_enabled": false,
    "created_at": 1704067200,
    "actions": [
      { "name": "update", "method": "PUT", "href": "/api/v1/users/me", "label": "Update Profile" },
      { "name": "setup_mfa", "method": "POST", "href": "/api/v1/auth/mfa/setup", "label": "Setup MFA" }
    ]
  }
}
```

---

### Update Current User

Update authenticated user's profile.

```
PUT /users/me
```

**Headers:** `Authorization: Bearer <token>`

**Request Body:**

```json
{
  "first_name": "John",
  "last_name": "Smith",
  "username": "johnsmith",
  "avatar_url": "https://...",
  "timezone": "America/New_York"
}
```

| Field | Type | Required | Validation |
|-------|------|----------|------------|
| first_name | string | No | 1-100 characters |
| last_name | string | No | 1-100 characters |
| username | string | No | 3-50 characters |
| avatar_url | string | No | Valid URL |
| timezone | string | No | IANA timezone (e.g., UTC, America/New_York) |

**Response:** `200 OK`

---

## Workspaces

### List Workspaces

Get all workspaces user belongs to.

```
GET /workspaces
```

**Headers:** `Authorization: Bearer <token>`

**Response:** `200 OK`

```json
{
  "success": true,
  "data": [
    {
      "id": "ws_abc123",
      "name": "My Workspace",
      "slug": "my-workspace",
      "description": "Main workspace",
      "logo_url": null,
      "timezone": "UTC",
      "plan_id": "plan_free",
      "created_at": 1704067200,
      "actions": [
        { "name": "view", "method": "GET", "href": "/api/v1/workspaces/ws_abc123", "label": "View" },
        { "name": "settings", "method": "GET", "href": "/api/v1/workspaces/ws_abc123/settings", "label": "Settings" }
      ]
    }
  ],
  "meta": { "total": 1, "page": 1, "per_page": 20, "total_pages": 1 }
}
```

---

### Create Workspace

Create a new workspace.

```
POST /workspaces
```

**Headers:** `Authorization: Bearer <token>`

**Request Body:**

```json
{
  "name": "My Workspace",
  "slug": "my-workspace",
  "description": "Optional description",
  "timezone": "America/New_York"
}
```

| Field | Type | Required | Validation |
|-------|------|----------|------------|
| name | string | Yes | 1-100 characters |
| slug | string | Yes | URL-safe slug |
| description | string | No | Max 500 characters |
| timezone | string | No | IANA timezone (defaults to UTC) |

**Response:** `201 Created`

---

### Get Workspace

Get workspace details.

```
GET /workspaces/{workspaceID}
```

**Headers:** `Authorization: Bearer <token>`

**Response:** `200 OK`

```json
{
  "success": true,
  "data": {
    "id": "ws_abc123",
    "name": "My Workspace",
    "slug": "my-workspace",
    "description": "Main workspace",
    "logo_url": null,
    "timezone": "UTC",
    "plan_id": "plan_free",
    "created_at": 1704067200,
    "actions": [
      { "name": "edit", "method": "PUT", "href": "/api/v1/workspaces/ws_abc123", "label": "Edit" },
      { "name": "members", "method": "GET", "href": "/api/v1/workspaces/ws_abc123/members", "label": "Members" },
      { "name": "billing", "method": "GET", "href": "/api/v1/workspaces/ws_abc123/billing/subscription", "label": "Billing" }
    ]
  }
}
```

---

### Update Workspace

Update workspace details.

```
PUT /workspaces/{workspaceID}
```

**Headers:** `Authorization: Bearer <token>`

**Request Body:**

```json
{
  "name": "Updated Name",
  "description": "Updated description",
  "logo_url": "https://...",
  "timezone": "Europe/London",
  "settings": { "key": "value" }
}
```

| Field | Type | Required | Validation |
|-------|------|----------|------------|
| name | string | No | 1-100 characters |
| description | string | No | Max 500 characters |
| logo_url | string | No | Valid URL |
| timezone | string | No | IANA timezone |
| settings | object | No | Workspace settings |

**Response:** `200 OK`

---

### Delete Workspace

Delete a workspace (owner only).

```
DELETE /workspaces/{workspaceID}
```

**Headers:** `Authorization: Bearer <token>`

**Response:** `204 No Content`

---

### Get Workspace Members

List workspace members.

```
GET /workspaces/{workspaceID}/members
```

**Headers:** `Authorization: Bearer <token>`

**Response:** `200 OK`

```json
{
  "success": true,
  "data": [
    {
      "id": "mem_abc123",
      "user": {
        "id": "usr_abc123",
        "email": "user@example.com",
        "first_name": "John",
        "last_name": "Doe"
      },
      "role": "owner",
      "joined_at": 1704067200,
      "actions": [
        { "name": "change_role", "method": "PUT", "href": "/api/v1/.../members/usr_abc123", "label": "Change Role" },
        { "name": "remove", "method": "DELETE", "href": "/api/v1/.../members/usr_abc123", "label": "Remove" }
      ]
    }
  ]
}
```

---

### Invite Member

Invite a new member to workspace.

```
POST /workspaces/{workspaceID}/members
```

**Headers:** `Authorization: Bearer <token>`

**Request Body:**

```json
{
  "email": "newmember@example.com",
  "role": "member"
}
```

| Field | Type | Required | Values |
|-------|------|----------|--------|
| email | string | Yes | Valid email |
| role | string | Yes | `admin`, `member`, `viewer` |

**Response:** `201 Created`

---

### Update Member Role

Change a member's role.

```
PUT /workspaces/{workspaceID}/members/{userID}
```

**Headers:** `Authorization: Bearer <token>`

**Request Body:**

```json
{
  "role": "admin"
}
```

**Response:** `200 OK`

---

### Remove Member

Remove a member from workspace.

```
DELETE /workspaces/{workspaceID}/members/{userID}
```

**Headers:** `Authorization: Bearer <token>`

**Response:** `204 No Content`

---

## Folders

Folders allow you to organize workflows into hierarchical structures (like directories).

### List Folders

Get all folders in workspace.

```
GET /workspaces/{workspaceID}/folders
```

**Headers:** `Authorization: Bearer <token>`

**Query Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| `parent_id` | uuid | Filter by parent folder (optional). If not provided, returns root folders. |

**Response:** `200 OK`

```json
{
  "success": true,
  "data": [
    {
      "id": "folder-uuid",
      "workspace_id": "workspace-uuid",
      "parent_id": null,
      "name": "Marketing Automations",
      "description": "All marketing workflows",
      "color": "#3B82F6",
      "icon": "folder",
      "created_at": "2024-01-15T10:00:00Z",
      "updated_at": "2024-01-15T10:00:00Z"
    }
  ],
  "meta": { "total": 1 }
}
```

---

### Get Folder Tree

Get folders as a hierarchical tree structure.

```
GET /workspaces/{workspaceID}/folders/tree
```

**Headers:** `Authorization: Bearer <token>`

**Response:** `200 OK`

```json
{
  "success": true,
  "data": [
    {
      "id": "folder-1",
      "name": "Marketing",
      "description": "Marketing workflows",
      "color": "#3B82F6",
      "icon": "folder",
      "parent_id": null,
      "children": [
        {
          "id": "folder-2",
          "name": "Email Campaigns",
          "parent_id": "folder-1",
          "children": []
        }
      ]
    }
  ]
}
```

---

### Create Folder

Create a new folder.

```
POST /workspaces/{workspaceID}/folders
```

**Headers:** `Authorization: Bearer <token>`

**Request Body:**

```json
{
  "name": "Marketing Automations",
  "description": "All marketing workflows",
  "color": "#3B82F6",
  "icon": "folder",
  "parent_id": null
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | Yes | Folder name |
| `description` | string | No | Folder description |
| `color` | string | No | Hex color code |
| `icon` | string | No | Icon name |
| `parent_id` | uuid | No | Parent folder ID for nesting |

**Response:** `201 Created`

---

### Get Folder

Get folder details.

```
GET /workspaces/{workspaceID}/folders/{folderID}
```

**Headers:** `Authorization: Bearer <token>`

**Response:** `200 OK`

---

### Update Folder

Update folder details.

```
PUT /workspaces/{workspaceID}/folders/{folderID}
```

**Headers:** `Authorization: Bearer <token>`

**Request Body:**

```json
{
  "name": "Updated Name",
  "description": "Updated description",
  "color": "#10B981"
}
```

**Response:** `200 OK`

---

### Delete Folder

Delete a folder. The folder must be empty (no sub-folders or workflows).

```
DELETE /workspaces/{workspaceID}/folders/{folderID}
```

**Headers:** `Authorization: Bearer <token>`

**Response:** `200 OK`

```json
{
  "success": true,
  "data": { "message": "Folder deleted" }
}
```

---

## Workflows

### List Workflows

Get all workflows in workspace.

```
GET /workspaces/{workspaceID}/workflows
```

**Headers:** `Authorization: Bearer <token>`

**Query Parameters:**

| Parameter | Type | Description | Example |
|-----------|------|-------------|---------|
| status | string | Filter by status | `?status=active` |
| search | string | Search in name/description | `?search=email` |
| tags | string | Filter by tags (comma-separated) | `?tags=production,critical` |
| created_after | string | Filter by creation date (RFC3339 or YYYY-MM-DD) | `?created_after=2024-01-01` |
| created_before | string | Filter by creation date | `?created_before=2024-12-31` |
| updated_after | string | Filter by update date | `?updated_after=2024-01-01` |
| updated_before | string | Filter by update date | `?updated_before=2024-12-31` |
| sort_by | string | Sort field: `name`, `created_at`, `updated_at`, `execution_count`, `last_executed_at` | `?sort_by=name` |
| order | string | Sort order: `asc`, `desc` (default: desc) | `?order=asc` |
| page | int | Page number (default: 1) | `?page=2` |
| per_page | int | Items per page (default: 20, max: 100) | `?per_page=50` |

**Status Values:** `draft`, `active`, `inactive`, `archived`

**Example Request:**
```
GET /workspaces/{id}/workflows?status=active&search=email&tags=production&sort_by=name&order=asc&page=1&per_page=20
```

**Response:** `200 OK`

```json
{
  "success": true,
  "data": [
    {
      "id": "wf_abc123",
      "name": "Email Automation",
      "description": "Send welcome emails",
      "status": "active",
      "version": 3,
      "tags": ["email", "automation"],
      "execution_count": 150,
      "last_executed_at": 1704067200,
      "created_at": 1704000000,
      "updated_at": 1704060000,
      "actions": [
        { "name": "execute", "method": "POST", "href": "/api/v1/.../execute", "label": "Execute" },
        { "name": "deactivate", "method": "POST", "href": "/api/v1/.../deactivate", "label": "Deactivate" },
        { "name": "delete", "method": "DELETE", "href": "/api/v1/...", "label": "Delete" }
      ]
    }
  ],
  "links": {
    "self": "/api/v1/workspaces/ws_abc123/workflows?page=1",
    "next": "/api/v1/workspaces/ws_abc123/workflows?page=2"
  },
  "meta": { "total": 25, "page": 1, "per_page": 20, "total_pages": 2 }
}
```

---

### Create Workflow

Create a new workflow.

```
POST /workspaces/{workspaceID}/workflows
```

**Headers:** `Authorization: Bearer <token>`

**Request Body:**

```json
{
  "name": "My Workflow",
  "description": "Optional description",
  "nodes": [
    {
      "id": "node_1",
      "type": "trigger.manual",
      "name": "Start",
      "position": { "x": 100, "y": 100 },
      "parameters": {}
    },
    {
      "id": "node_2",
      "type": "action.http",
      "name": "HTTP Request",
      "position": { "x": 300, "y": 100 },
      "parameters": {
        "url": "https://api.example.com",
        "method": "GET"
      }
    }
  ],
  "connections": [
    {
      "source": { "nodeId": "node_1", "output": "main" },
      "target": { "nodeId": "node_2", "input": "main" }
    }
  ],
  "settings": {
    "timeout_seconds": 300,
    "retry_on_failure": true,
    "max_retries": 3
  },
  "tags": ["automation"]
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| name | string | Yes | 1-255 characters |
| description | string | No | Max 1000 characters |
| nodes | array | Yes | Array of node definitions |
| connections | array | Yes | Array of connections |
| settings | object | No | Workflow settings |
| tags | array | No | Max 10 tags, each max 50 chars |

**Response:** `201 Created`

---

### Get Workflow

Get workflow details with nodes and connections.

```
GET /workspaces/{workspaceID}/workflows/{workflowID}
```

**Headers:** `Authorization: Bearer <token>`

**Query Parameters:**

| Parameter | Description |
|-----------|-------------|
| include | `schedules`, `credentials` |

**Response:** `200 OK`

```json
{
  "success": true,
  "data": {
    "id": "wf_abc123",
    "name": "Email Automation",
    "description": "Send welcome emails",
    "status": "active",
    "version": 3,
    "nodes": [...],
    "connections": [...],
    "settings": { "timeout_seconds": 300 },
    "tags": ["email"],
    "execution_count": 150,
    "last_executed_at": 1704067200,
    "created_at": 1704000000,
    "updated_at": 1704060000,
    "actions": [
      { "name": "edit", "method": "PUT", "href": "/api/v1/.../", "label": "Edit" },
      { "name": "execute", "method": "POST", "href": "/api/v1/.../execute", "label": "Execute" },
      { "name": "versions", "method": "GET", "href": "/api/v1/.../versions", "label": "View Versions" }
    ]
  },
  "included": {
    "schedules": [...]
  }
}
```

---

### Update Workflow

Update workflow (creates new version).

```
PUT /workspaces/{workspaceID}/workflows/{workflowID}
```

**Headers:** `Authorization: Bearer <token>`

**Request Body:**

```json
{
  "name": "Updated Name",
  "description": "Updated description",
  "nodes": [...],
  "connections": [...],
  "settings": { "timeout_seconds": 600 },
  "tags": ["updated"]
}
```

**Response:** `200 OK`

---

### Delete Workflow

Delete a workflow.

```
DELETE /workspaces/{workspaceID}/workflows/{workflowID}
```

**Headers:** `Authorization: Bearer <token>`

**Response:** `204 No Content`

---

### Execute Workflow

Execute a workflow manually.

```
POST /workspaces/{workspaceID}/workflows/{workflowID}/execute
```

**Headers:** `Authorization: Bearer <token>`

**Request Body:**

```json
{
  "input_data": {
    "key": "value"
  }
}
```

**Response:** `202 Accepted`

```json
{
  "success": true,
  "data": {
    "execution_id": "exec_abc123",
    "status": "queued"
  }
}
```

---

### Clone Workflow

Create a copy of a workflow.

```
POST /workspaces/{workspaceID}/workflows/{workflowID}/clone
```

**Headers:** `Authorization: Bearer <token>`

**Request Body:**

```json
{
  "name": "My Workflow (Copy)"
}
```

**Response:** `201 Created`

---

### Activate Workflow

Activate a workflow (enable triggers).

```
POST /workspaces/{workspaceID}/workflows/{workflowID}/activate
```

**Headers:** `Authorization: Bearer <token>`

**Response:** `200 OK`

```json
{
  "success": true,
  "data": { "status": "active" },
  "actions": [
    { "name": "execute", "method": "POST", "href": "/api/v1/.../execute", "label": "Execute" },
    { "name": "deactivate", "method": "POST", "href": "/api/v1/.../deactivate", "label": "Deactivate" }
  ]
}
```

---

### Deactivate Workflow

Deactivate a workflow (disable triggers).

```
POST /workspaces/{workspaceID}/workflows/{workflowID}/deactivate
```

**Headers:** `Authorization: Bearer <token>`

**Response:** `200 OK`

```json
{
  "success": true,
  "data": { "status": "inactive" },
  "actions": [
    { "name": "activate", "method": "POST", "href": "/api/v1/.../activate", "label": "Activate" }
  ]
}
```

---

### List Workflow Versions

Get version history of a workflow.

```
GET /workspaces/{workspaceID}/workflows/{workflowID}/versions
```

**Headers:** `Authorization: Bearer <token>`

**Response:** `200 OK`

```json
{
  "success": true,
  "data": [
    {
      "id": "ver_abc123",
      "version": 3,
      "change_message": "Added error handling",
      "created_at": 1704067200,
      "actions": [
        { "name": "view", "method": "GET", "href": "/api/v1/.../versions/3", "label": "View" },
        { "name": "rollback", "method": "POST", "href": "/api/v1/.../versions/3/rollback", "label": "Rollback" }
      ]
    }
  ]
}
```

---

### Get Workflow Version

Get specific version details.

```
GET /workspaces/{workspaceID}/workflows/{workflowID}/versions/{version}
```

**Headers:** `Authorization: Bearer <token>`

**Response:** `200 OK`

```json
{
  "success": true,
  "data": {
    "id": "ver_abc123",
    "version": 2,
    "nodes": [...],
    "connections": [...],
    "settings": {...},
    "change_message": "Initial version",
    "created_at": 1704000000,
    "actions": [
      { "name": "rollback", "method": "POST", "href": "/api/v1/.../versions/2/rollback", "label": "Rollback to this version" },
      { "name": "diff", "method": "GET", "href": "/api/v1/.../versions/2/compare", "label": "Compare with current" }
    ]
  }
}
```

---

### Rollback Workflow Version

Rollback to a previous version.

```
POST /workspaces/{workspaceID}/workflows/{workflowID}/versions/{version}/rollback
```

**Headers:** `Authorization: Bearer <token>`

**Response:** `200 OK`

---

### Compare Versions

Compare two workflow versions.

```
GET /workspaces/{workspaceID}/workflows/{workflowID}/compare-versions?from=2&to=3
```

**Headers:** `Authorization: Bearer <token>`

**Query Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| from | int | Yes | First version number |
| to | int | Yes | Second version number |

**Response:** `200 OK`

---

### Export Workflow

Export workflow as JSON.

```
GET /workspaces/{workspaceID}/workflows/{workflowID}/export
```

**Headers:** `Authorization: Bearer <token>`

**Response:** `200 OK` with downloadable JSON file

---

### Import Workflow

Import workflow from JSON.

```
POST /workspaces/{workspaceID}/workflows/import
```

**Headers:** `Authorization: Bearer <token>`  
**Content-Type:** `multipart/form-data`

**Form Fields:**

| Field | Type | Description |
|-------|------|-------------|
| file | file | JSON workflow file |

**Response:** `201 Created`

---

### Duplicate Workflow

Duplicate workflow with variable substitution.

```
POST /workspaces/{workspaceID}/workflows/{workflowID}/duplicate
```

**Headers:** `Authorization: Bearer <token>`

**Request Body:**

```json
{
  "name": "Duplicated Workflow",
  "variables": {
    "api_url": "https://new-api.example.com"
  }
}
```

**Response:** `201 Created`

---

## Executions

### List Executions

Get all executions in workspace.

```
GET /workspaces/{workspaceID}/executions
```

**Headers:** `Authorization: Bearer <token>`

**Query Parameters:**

| Parameter | Type | Description | Example |
|-----------|------|-------------|---------|
| workflow_id | uuid | Filter by workflow | `?workflow_id=uuid` |
| status | string | Filter by status | `?status=failed` |
| trigger_type | string | Filter by trigger type | `?trigger_type=webhook` |
| start_date | string | Filter by start date (RFC3339 or YYYY-MM-DD) | `?start_date=2024-01-01` |
| end_date | string | Filter by end date | `?end_date=2024-12-31` |
| search | string | Search in error messages | `?search=timeout` |
| sort_by | string | Sort field: `queued_at`, `started_at`, `completed_at`, `duration` | `?sort_by=started_at` |
| order | string | Sort order: `asc`, `desc` (default: desc) | `?order=asc` |
| page | int | Page number (default: 1) | `?page=2` |
| per_page | int | Items per page (default: 20, max: 100) | `?per_page=50` |

**Status Values:** `queued`, `running`, `completed`, `failed`, `cancelled`, `timeout`

**Trigger Types:** `manual`, `webhook`, `schedule`

**Example Request:**
```
GET /workspaces/{id}/executions?status=failed&trigger_type=webhook&start_date=2024-01-01&sort_by=started_at&order=desc
```

**Response:** `200 OK`

```json
{
  "success": true,
  "data": [
    {
      "id": "exec_abc123",
      "workflow_id": "wf_abc123",
      "workflow_version": 3,
      "status": "success",
      "trigger_type": "manual",
      "nodes_total": 5,
      "nodes_completed": 5,
      "queued_at": 1704067200,
      "started_at": 1704067201,
      "completed_at": 1704067205,
      "actions": [
        { "name": "view", "method": "GET", "href": "/api/v1/.../exec_abc123", "label": "View Details" },
        { "name": "retry", "method": "POST", "href": "/api/v1/.../exec_abc123/retry", "label": "Retry" }
      ]
    }
  ],
  "meta": { "total": 150, "page": 1, "per_page": 20, "total_pages": 8 }
}
```

---

### Search Executions

Advanced search with filters.

```
GET /workspaces/{workspaceID}/executions/search
```

**Headers:** `Authorization: Bearer <token>`

**Query Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| q | string | Search query |
| workflow_id | uuid | Filter by workflow |
| status | string | Execution status |
| trigger_type | string | Trigger type |
| from | timestamp | Start date |
| to | timestamp | End date |

**Response:** `200 OK`

---

### Get Execution Statistics

Get execution statistics for workspace.

```
GET /workspaces/{workspaceID}/executions/stats
```

**Headers:** `Authorization: Bearer <token>`

**Query Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| workflow_id | uuid | Filter by workflow |
| period | string | `day`, `week`, `month` |

**Response:** `200 OK`

```json
{
  "success": true,
  "data": {
    "total": 1500,
    "success": 1400,
    "failed": 80,
    "cancelled": 20,
    "avg_duration_ms": 2500,
    "by_day": [
      { "date": "2024-01-01", "total": 50, "success": 48, "failed": 2 }
    ]
  }
}
```

---

### Get Execution

Get execution details.

```
GET /workspaces/{workspaceID}/executions/{executionID}
```

**Headers:** `Authorization: Bearer <token>`

**Response:** `200 OK`

```json
{
  "success": true,
  "data": {
    "id": "exec_abc123",
    "workflow_id": "wf_abc123",
    "workflow_version": 3,
    "status": "success",
    "trigger_type": "manual",
    "input_data": { "key": "value" },
    "output_data": { "result": "success" },
    "nodes_total": 5,
    "nodes_completed": 5,
    "queued_at": 1704067200,
    "started_at": 1704067201,
    "completed_at": 1704067205,
    "actions": [
      { "name": "nodes", "method": "GET", "href": "/api/v1/.../nodes", "label": "View Nodes" },
      { "name": "logs", "method": "GET", "href": "/api/v1/.../logs", "label": "View Logs" },
      { "name": "retry", "method": "POST", "href": "/api/v1/.../retry", "label": "Retry" }
    ]
  }
}
```

---

### Cancel Execution

Cancel a running execution.

```
POST /workspaces/{workspaceID}/executions/{executionID}/cancel
```

**Headers:** `Authorization: Bearer <token>`

**Response:** `200 OK`

```json
{
  "status": "cancelled"
}
```

---

### Retry Execution

Retry a failed execution.

```
POST /workspaces/{workspaceID}/executions/{executionID}/retry
```

**Headers:** `Authorization: Bearer <token>`

**Response:** `202 Accepted`

```json
{
  "success": true,
  "data": {
    "execution_id": "exec_new123",
    "status": "queued"
  }
}
```

---

### Get Execution Nodes

Get node execution details.

```
GET /workspaces/{workspaceID}/executions/{executionID}/nodes
```

**Headers:** `Authorization: Bearer <token>`

**Response:** `200 OK`

```json
{
  "success": true,
  "data": [
    {
      "id": "node_exec_1",
      "node_id": "node_1",
      "node_type": "trigger.manual",
      "node_name": "Start",
      "status": "success",
      "input_data": {},
      "output_data": { "triggered": true },
      "duration_ms": 50,
      "started_at": 1704067201,
      "completed_at": 1704067201
    }
  ]
}
```

---

### Bulk Delete Executions

Delete multiple executions.

```
DELETE /workspaces/{workspaceID}/executions/bulk
```

**Headers:** `Authorization: Bearer <token>`

**Request Body:**

```json
{
  "execution_ids": ["exec_1", "exec_2", "exec_3"]
}
```

**Response:** `200 OK`

```json
{
  "deleted": 3
}
```

---

## Credentials

### List Credentials

Get all credentials in workspace.

```
GET /workspaces/{workspaceID}/credentials
```

**Headers:** `Authorization: Bearer <token>`

**Query Parameters:**

| Parameter | Type | Description | Example |
|-----------|------|-------------|---------|
| type | string | Filter by credential type | `?type=api_key` |
| search | string | Search in name/description | `?search=github` |
| sort_by | string | Sort field: `name`, `created_at`, `type`, `last_used_at` | `?sort_by=name` |
| order | string | Sort order: `asc`, `desc` (default: desc) | `?order=asc` |
| page | int | Page number (default: 1) | `?page=2` |
| per_page | int | Items per page (default: 20, max: 100) | `?per_page=50` |

**Credential Types:** `api_key`, `oauth2`, `basic`, `bearer`, `custom`

**Example Request:**
```
GET /workspaces/{id}/credentials?type=api_key&search=github&sort_by=name&order=asc
```

**Response:** `200 OK`

```json
{
  "success": true,
  "data": [
    {
      "id": "cred_abc123",
      "name": "GitHub API",
      "type": "api_key",
      "description": "GitHub personal access token",
      "last_used_at": 1704067200,
      "created_at": 1704000000,
      "actions": [
        { "name": "edit", "method": "PUT", "href": "/api/v1/.../cred_abc123", "label": "Edit" },
        { "name": "test", "method": "POST", "href": "/api/v1/.../cred_abc123/test", "label": "Test Connection" },
        { "name": "delete", "method": "DELETE", "href": "/api/v1/.../cred_abc123", "label": "Delete" }
      ]
    }
  ]
}
```

---

### Create Credential

Create a new credential.

```
POST /workspaces/{workspaceID}/credentials
```

**Headers:** `Authorization: Bearer <token>`

**Request Body:**

```json
{
  "name": "GitHub API",
  "type": "api_key",
  "description": "GitHub personal access token",
  "data": {
    "api_key": "ghp_xxxxxxxxxxxxx"
  }
}
```

| Field | Type | Required | Values |
|-------|------|----------|--------|
| name | string | Yes | 1-100 characters |
| type | string | Yes | `api_key`, `oauth2`, `basic`, `bearer`, `custom` |
| description | string | No | Max 500 characters |
| data | object | Yes | Credential data (encrypted at rest) |

**Credential Data by Type:**

**api_key:**
```json
{ "api_key": "your_api_key", "header_name": "X-API-Key" }
```

**basic:**
```json
{ "username": "user", "password": "pass" }
```

**bearer:**
```json
{ "token": "your_bearer_token" }
```

**oauth2:**
```json
{ "access_token": "...", "refresh_token": "...", "expires_at": 1704067200 }
```

**Response:** `201 Created`

---

### Get Credential

Get credential details (data is masked).

```
GET /workspaces/{workspaceID}/credentials/{credentialID}
```

**Headers:** `Authorization: Bearer <token>`

**Response:** `200 OK`

---

### Update Credential

Update credential.

```
PUT /workspaces/{workspaceID}/credentials/{credentialID}
```

**Headers:** `Authorization: Bearer <token>`

**Request Body:**

```json
{
  "name": "Updated Name",
  "data": {
    "api_key": "new_api_key"
  }
}
```

**Response:** `200 OK`

---

### Delete Credential

Delete a credential.

```
DELETE /workspaces/{workspaceID}/credentials/{credentialID}
```

**Headers:** `Authorization: Bearer <token>`

**Response:** `204 No Content`

---

### Test Credential

Test credential connection.

```
POST /workspaces/{workspaceID}/credentials/{credentialID}/test
```

**Headers:** `Authorization: Bearer <token>`

**Response:** `200 OK`

```json
{
  "success": true
}
```

---

## Schedules

### List Schedules

Get all schedules in workspace.

```
GET /workspaces/{workspaceID}/schedules
```

**Headers:** `Authorization: Bearer <token>`

**Query Parameters:**

| Parameter | Type | Description | Example |
|-----------|------|-------------|---------|
| is_active | boolean | Filter by active status | `?is_active=true` |
| workflow_id | uuid | Filter by workflow | `?workflow_id=uuid` |
| search | string | Search in name/description | `?search=daily` |
| sort_by | string | Sort field: `created_at`, `next_run_at`, `last_run_at` | `?sort_by=next_run_at` |
| order | string | Sort order: `asc`, `desc` (default: desc) | `?order=asc` |
| page | int | Page number (default: 1) | `?page=2` |
| per_page | int | Items per page (default: 20, max: 100) | `?per_page=50` |

**Example Request:**
```
GET /workspaces/{id}/schedules?is_active=true&workflow_id=uuid&sort_by=next_run_at&order=asc
```

**Response:** `200 OK`

```json
{
  "success": true,
  "data": [
    {
      "id": "sched_abc123",
      "workflow_id": "wf_abc123",
      "name": "Daily Report",
      "description": "Generate daily report at 9 AM",
      "cron_expression": "0 9 * * *",
      "timezone": "America/New_York",
      "is_active": true,
      "next_run_at": 1704110400,
      "last_run_at": 1704024000,
      "run_count": 30,
      "created_at": 1701388800,
      "actions": [
        { "name": "pause", "method": "POST", "href": "/api/v1/.../pause", "label": "Pause" },
        { "name": "trigger", "method": "POST", "href": "/api/v1/.../trigger", "label": "Trigger Now" },
        { "name": "edit", "method": "PUT", "href": "/api/v1/...", "label": "Edit" }
      ]
    }
  ]
}
```

---

### Create Schedule

Create a new schedule.

```
POST /workspaces/{workspaceID}/schedules
```

**Headers:** `Authorization: Bearer <token>`

**Request Body:**

```json
{
  "workflow_id": "wf_abc123",
  "name": "Daily Report",
  "description": "Generate daily report at 9 AM",
  "cron_expression": "0 9 * * *",
  "timezone": "America/New_York",
  "input_data": {
    "report_type": "daily"
  }
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| workflow_id | uuid | Yes | Target workflow |
| name | string | Yes | 1-100 characters |
| description | string | No | Max 500 characters |
| cron_expression | string | Yes | Valid cron expression |
| timezone | string | Yes | IANA timezone |
| input_data | object | No | Input data for execution |

**Cron Expression Examples:**

| Expression | Description |
|------------|-------------|
| `0 9 * * *` | Every day at 9 AM |
| `0 */2 * * *` | Every 2 hours |
| `0 9 * * 1-5` | Weekdays at 9 AM |
| `0 0 1 * *` | First day of month |

**Response:** `201 Created`

---

### Get Schedule

Get schedule details.

```
GET /workspaces/{workspaceID}/schedules/{scheduleID}
```

**Headers:** `Authorization: Bearer <token>`

**Response:** `200 OK`

---

### Update Schedule

Update schedule.

```
PUT /workspaces/{workspaceID}/schedules/{scheduleID}
```

**Headers:** `Authorization: Bearer <token>`

**Request Body:**

```json
{
  "name": "Updated Schedule",
  "cron_expression": "0 10 * * *",
  "timezone": "UTC"
}
```

**Response:** `200 OK`

---

### Delete Schedule

Delete a schedule.

```
DELETE /workspaces/{workspaceID}/schedules/{scheduleID}
```

**Headers:** `Authorization: Bearer <token>`

**Response:** `204 No Content`

---

### Pause Schedule

Pause a schedule.

```
POST /workspaces/{workspaceID}/schedules/{scheduleID}/pause
```

**Headers:** `Authorization: Bearer <token>`

**Response:** `200 OK`

```json
{
  "success": true,
  "data": { "is_active": false },
  "actions": [
    { "name": "resume", "method": "POST", "href": "/api/v1/.../resume", "label": "Resume" }
  ]
}
```

---

### Resume Schedule

Resume a paused schedule.

```
POST /workspaces/{workspaceID}/schedules/{scheduleID}/resume
```

**Headers:** `Authorization: Bearer <token>`

**Response:** `200 OK`

```json
{
  "success": true,
  "data": { "is_active": true },
  "actions": [
    { "name": "pause", "method": "POST", "href": "/api/v1/.../pause", "label": "Pause" }
  ]
}
```

---

## Templates

### List Templates

Get public workflow templates.

```
GET /templates
```

**Query Parameters:**

| Parameter | Type | Description | Example |
|-----------|------|-------------|---------|
| category | string | Filter by category | `?category=communication` |
| search | string | Search in name/description | `?search=slack` |
| is_featured | boolean | Filter featured templates | `?is_featured=true` |
| sort_by | string | Sort field: `name`, `created_at`, `use_count` | `?sort_by=use_count` |
| order | string | Sort order: `asc`, `desc` (default: desc) | `?order=desc` |
| page | int | Page number (default: 1) | `?page=2` |
| per_page | int | Items per page (default: 20, max: 100) | `?per_page=50` |

**Categories:** `communication`, `data`, `automation`, `integration`, `marketing`, `devops`

**Example Request:**
```
GET /templates?category=communication&is_featured=true&sort_by=use_count&order=desc
```

**Response:** `200 OK`

```json
{
  "success": true,
  "data": [
    {
      "id": "tpl_abc123",
      "name": "Slack Notification",
      "description": "Send notifications to Slack",
      "category": "communication",
      "is_featured": true,
      "use_count": 1500,
      "actions": [
        { "name": "view", "method": "GET", "href": "/api/v1/templates/tpl_abc123", "label": "View" },
        { "name": "use", "method": "POST", "href": "/api/v1/.../templates/tpl_abc123/use", "label": "Use Template" }
      ]
    }
  ]
}
```

---

### Get Featured Templates

Get featured templates.

```
GET /templates/featured
```

**Query Parameters:**

| Parameter | Type | Default |
|-----------|------|---------|
| limit | int | 10 |

**Response:** `200 OK`

---

### Get Template Categories

Get all template categories.

```
GET /templates/categories
```

**Response:** `200 OK`

```json
{
  "success": true,
  "data": [
    { "name": "communication", "actions": [{ "name": "view", "method": "GET", "href": "/api/v1/templates/categories/communication" }] },
    { "name": "data", "actions": [...] },
    { "name": "automation", "actions": [...] }
  ]
}
```

---

### Get Templates by Category

Get templates in a category.

```
GET /templates/categories/{category}
```

**Response:** `200 OK`

---

### Search Templates

Search templates.

```
GET /templates/search?q=slack
```

**Query Parameters:**

| Parameter | Type | Required |
|-----------|------|----------|
| q | string | Yes |

**Response:** `200 OK`

---

### Get Template

Get template details.

```
GET /templates/{templateID}
```

**Response:** `200 OK`

```json
{
  "success": true,
  "data": {
    "id": "tpl_abc123",
    "name": "Slack Notification",
    "description": "Send notifications to Slack channel",
    "category": "communication",
    "nodes": [...],
    "connections": [...],
    "is_featured": true,
    "use_count": 1500,
    "actions": [
      { "name": "use", "method": "POST", "href": "/api/v1/.../use", "label": "Use Template" },
      { "name": "preview", "method": "GET", "href": "/api/v1/.../preview", "label": "Preview" }
    ]
  }
}
```

---

### Use Template

Create workflow from template.

```
POST /workspaces/{workspaceID}/templates/{templateID}/use
```

**Headers:** `Authorization: Bearer <token>`

**Request Body:**

```json
{
  "name": "My Slack Workflow",
  "variables": {
    "slack_channel": "#general"
  }
}
```

**Response:** `201 Created`

```json
{
  "success": true,
  "data": {
    "message": "Workflow created from template",
    "workflow": { ... }
  },
  "actions": [
    { "name": "view", "method": "GET", "href": "/api/v1/.../workflows/wf_new", "label": "View Workflow" },
    { "name": "execute", "method": "POST", "href": "/api/v1/.../workflows/wf_new/execute", "label": "Execute" }
  ]
}
```

---

## Billing

### Get Plans

Get available billing plans.

```
GET /billing/plans
```

**Response:** `200 OK`

```json
{
  "success": true,
  "data": [
    {
      "id": "plan_free",
      "name": "Free",
      "description": "For individuals",
      "price_monthly": 0,
      "price_yearly": 0,
      "executions_limit": 1000,
      "workflows_limit": 5,
      "members_limit": 1,
      "credentials_limit": 10,
      "features": ["Basic support"],
      "actions": [
        { "name": "subscribe", "method": "POST", "href": "/api/v1/.../billing/subscription", "label": "Subscribe" }
      ]
    },
    {
      "id": "plan_pro",
      "name": "Pro",
      "price_monthly": 2900,
      "price_yearly": 29000,
      "executions_limit": 50000,
      "workflows_limit": 100,
      "members_limit": 10,
      "credentials_limit": 100,
      "features": ["Priority support", "Advanced analytics"]
    }
  ]
}
```

---

### Get Subscription

Get workspace subscription.

```
GET /workspaces/{workspaceID}/billing/subscription
```

**Headers:** `Authorization: Bearer <token>`

**Response:** `200 OK`

```json
{
  "success": true,
  "data": {
    "id": "sub_abc123",
    "plan_id": "plan_pro",
    "status": "active",
    "billing_cycle": "monthly",
    "current_period_start": 1704067200,
    "current_period_end": 1706745600,
    "actions": [
      { "name": "change_plan", "method": "PUT", "href": "/api/v1/.../billing/subscription", "label": "Change Plan" },
      { "name": "cancel", "method": "DELETE", "href": "/api/v1/.../billing/subscription", "label": "Cancel" }
    ]
  }
}
```

---

### Create Subscription

Subscribe to a plan.

```
POST /workspaces/{workspaceID}/billing/subscription
```

**Headers:** `Authorization: Bearer <token>`

**Request Body:**

```json
{
  "plan_id": "plan_pro",
  "billing_cycle": "monthly",
  "payment_token": "tok_xxx"
}
```

**Response:** `201 Created`

---

### Cancel Subscription

Cancel subscription.

```
DELETE /workspaces/{workspaceID}/billing/subscription
```

**Headers:** `Authorization: Bearer <token>`

**Response:** `200 OK`

```json
{
  "status": "cancelled"
}
```

---

### Get Usage

Get current billing period usage.

```
GET /workspaces/{workspaceID}/billing/usage
```

**Headers:** `Authorization: Bearer <token>`

**Response:** `200 OK`

```json
{
  "success": true,
  "data": {
    "executions": 5000,
    "workflows": 15,
    "members": 3,
    "credentials": 25,
    "storage_bytes": 104857600,
    "credits_used": 5000,
    "credits_included": 10000,
    "credits_remaining": 5000,
    "executions_success": 4800,
    "executions_failed": 200,
    "webhooks_called": 1500,
    "schedules_triggered": 300,
    "period_start": 1704067200,
    "period_end": 1706745600
  }
}
```

---

### Get Invoices

Get billing invoices.

```
GET /workspaces/{workspaceID}/billing/invoices
```

**Headers:** `Authorization: Bearer <token>`

**Response:** `200 OK`

```json
{
  "success": true,
  "data": [
    {
      "id": "inv_abc123",
      "number": "INV-2024-001",
      "status": "paid",
      "amount_due": 2900,
      "amount_paid": 2900,
      "currency": "usd",
      "period_start": 1701388800,
      "period_end": 1704067200,
      "paid_at": 1701388900,
      "actions": [
        { "name": "view", "method": "GET", "href": "/api/v1/.../invoices/inv_abc123", "label": "View" },
        { "name": "download", "method": "GET", "href": "/api/v1/.../invoices/inv_abc123/pdf", "label": "Download PDF" }
      ]
    }
  ]
}
```

---

## Webhooks

### Trigger Webhook

Trigger a workflow via webhook.

```
POST /webhooks/{endpointID}
GET /webhooks/{endpointID}
```

**Headers (optional):**

| Header | Description |
|--------|-------------|
| X-Webhook-Secret | HMAC signature for verification |
| Content-Type | `application/json` |

**Request Body:** Any JSON payload

**Response:** `200 OK`

```json
{
  "execution_id": "exec_abc123",
  "status": "queued"
}
```

---

### Generate Webhook

Create a webhook endpoint for a workflow.

```
POST /workspaces/{workspaceID}/workflows/{workflowID}/webhooks
```

**Headers:** `Authorization: Bearer <token>`

**Request Body:**

```json
{
  "node_id": "webhook_trigger_1",
  "method": "POST",
  "custom_path": "my-webhook"
}
```

**Response:** `201 Created`

```json
{
  "success": true,
  "data": {
    "id": "wh_abc123",
    "path": "/webhooks/wh_abc123",
    "url": "https://api.linkflow.io/webhooks/wh_abc123",
    "method": "POST",
    "secret": "whsec_xxxxx",
    "is_active": true
  },
  "actions": [
    { "name": "test", "method": "POST", "href": "/api/v1/.../test", "label": "Test" },
    { "name": "disable", "method": "POST", "href": "/api/v1/.../disable", "label": "Disable" }
  ]
}
```

---

### List Webhooks

List webhooks for a workflow.

```
GET /workspaces/{workspaceID}/workflows/{workflowID}/webhooks
```

**Headers:** `Authorization: Bearer <token>`

**Response:** `200 OK`

---

### Regenerate Webhook Secret

Generate a new secret for a webhook.

```
POST /workspaces/{workspaceID}/webhooks/{webhookID}/regenerate-secret
```

**Headers:** `Authorization: Bearer <token>`

**Response:** `200 OK`

```json
{
  "message": "Secret regenerated",
  "secret": "whsec_new_xxxxx"
}
```

---

### Activate Webhook

Activate a webhook.

```
POST /workspaces/{workspaceID}/webhooks/{webhookID}/activate
```

**Headers:** `Authorization: Bearer <token>`

**Response:** `200 OK`

---

### Deactivate Webhook

Deactivate a webhook.

```
POST /workspaces/{workspaceID}/webhooks/{webhookID}/deactivate
```

**Headers:** `Authorization: Bearer <token>`

**Response:** `200 OK`

---

## Analytics

### Get Workspace Analytics

Get workspace-level analytics.

```
GET /workspaces/{workspaceID}/analytics
```

**Headers:** `Authorization: Bearer <token>`

**Query Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| period | string | `day`, `week`, `month`, `year` |
| from | timestamp | Start date |
| to | timestamp | End date |

**Response:** `200 OK`

---

### Get Workflow Analytics

Get workflow-specific analytics.

```
GET /workspaces/{workspaceID}/workflows/{workflowID}/analytics
```

**Headers:** `Authorization: Bearer <token>`

**Response:** `200 OK`

---

## Alerts

### List Alerts

Get all alerts in workspace.

```
GET /workspaces/{workspaceID}/alerts
```

**Headers:** `Authorization: Bearer <token>`

**Query Parameters:**

| Parameter | Type | Description | Example |
|-----------|------|-------------|---------|
| type | string | Filter by alert type | `?type=workflow_failed` |
| is_active | boolean | Filter by active status | `?is_active=true` |
| workflow_id | uuid | Filter by workflow | `?workflow_id=uuid` |
| search | string | Search in name | `?search=error` |
| sort_by | string | Sort field: `name`, `created_at`, `type` | `?sort_by=name` |
| order | string | Sort order: `asc`, `desc` (default: desc) | `?order=asc` |
| page | int | Page number (default: 1) | `?page=2` |
| per_page | int | Items per page (default: 20, max: 100) | `?per_page=50` |

**Alert Types:** `workflow_failed`, `execution_timeout`, `high_error_rate`, `schedule_missed`, `credential_expiring`, `usage_limit`

**Example Request:**
```
GET /workspaces/{id}/alerts?type=workflow_failed&is_active=true&sort_by=created_at&order=desc
```

**Response:** `200 OK`

```json
{
  "success": true,
  "data": [
    {
      "id": "alert_abc123",
      "name": "Workflow Failed",
      "type": "workflow_failed",
      "is_active": true,
      "config": { "workflow_id": "wf_abc123" },
      "created_at": 1704067200,
      "actions": [
        { "name": "disable", "method": "POST", "href": "/api/v1/.../disable", "label": "Disable" },
        { "name": "edit", "method": "PUT", "href": "/api/v1/...", "label": "Edit" },
        { "name": "test", "method": "POST", "href": "/api/v1/.../test", "label": "Test" }
      ]
    }
  ]
}
```

---

### Create Alert

Create a new alert.

```
POST /workspaces/{workspaceID}/alerts
```

**Headers:** `Authorization: Bearer <token>`

**Request Body:**

```json
{
  "name": "Workflow Failed Alert",
  "type": "workflow_failed",
  "config": {
    "workflow_id": "wf_abc123",
    "notification_channels": ["email", "slack"]
  }
}
```

**Response:** `201 Created`

---

### Update Alert

Update an alert.

```
PUT /workspaces/{workspaceID}/alerts/{alertID}
```

**Headers:** `Authorization: Bearer <token>`

**Response:** `200 OK`

---

### Delete Alert

Delete an alert.

```
DELETE /workspaces/{workspaceID}/alerts/{alertID}
```

**Headers:** `Authorization: Bearer <token>`

**Response:** `204 No Content`

---

## Audit Logs

### List Audit Logs

Get audit logs for workspace.

```
GET /workspaces/{workspaceID}/audit-logs
```

**Headers:** `Authorization: Bearer <token>`

**Query Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| action | string | Filter by action type |
| user_id | uuid | Filter by user |
| resource_type | string | Filter by resource type |
| from | timestamp | Start date |
| to | timestamp | End date |

**Response:** `200 OK`

```json
{
  "success": true,
  "data": [
    {
      "id": "log_abc123",
      "action": "workflow.created",
      "user_id": "usr_abc123",
      "resource_type": "workflow",
      "resource_id": "wf_abc123",
      "metadata": { "name": "New Workflow" },
      "ip_address": "192.168.1.1",
      "created_at": 1704067200
    }
  ],
  "meta": { "total": 500, "page": 1, "per_page": 20, "total_pages": 25 }
}
```

---

### Search Audit Logs

Search audit logs.

```
GET /workspaces/{workspaceID}/audit-logs/search
```

**Headers:** `Authorization: Bearer <token>`

**Query Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| q | string | Search query |

**Response:** `200 OK`

---

## Environment Variables

### List Environment Variables

Get workspace environment variables.

```
GET /workspaces/{workspaceID}/env-vars
```

**Headers:** `Authorization: Bearer <token>`

**Response:** `200 OK`

```json
{
  "success": true,
  "data": [
    {
      "id": "var_abc123",
      "key": "API_BASE_URL",
      "value": "https://api.example.com",
      "is_secret": false,
      "created_at": 1704067200,
      "actions": [
        { "name": "edit", "method": "PUT", "href": "/api/v1/.../var_abc123", "label": "Edit" },
        { "name": "delete", "method": "DELETE", "href": "/api/v1/.../var_abc123", "label": "Delete" }
      ]
    }
  ]
}
```

---

### Create Environment Variable

Create a new environment variable.

```
POST /workspaces/{workspaceID}/env-vars
```

**Headers:** `Authorization: Bearer <token>`

**Request Body:**

```json
{
  "key": "API_BASE_URL",
  "value": "https://api.example.com",
  "is_secret": false
}
```

**Response:** `201 Created`

---

### Update Environment Variable

Update an environment variable.

```
PUT /workspaces/{workspaceID}/env-vars/{varID}
```

**Headers:** `Authorization: Bearer <token>`

**Request Body:**

```json
{
  "value": "https://new-api.example.com"
}
```

**Response:** `200 OK`

---

### Delete Environment Variable

Delete an environment variable.

```
DELETE /workspaces/{workspaceID}/env-vars/{varID}
```

**Headers:** `Authorization: Bearer <token>`

**Response:** `204 No Content`

---

## Comments

### List Workflow Comments

Get comments on a workflow.

```
GET /workspaces/{workspaceID}/workflows/{workflowID}/comments
```

**Headers:** `Authorization: Bearer <token>`

**Response:** `200 OK`

---

### Create Comment

Add a comment to a workflow.

```
POST /workspaces/{workspaceID}/workflows/{workflowID}/comments
```

**Headers:** `Authorization: Bearer <token>`

**Request Body:**

```json
{
  "content": "This workflow needs optimization",
  "node_id": "node_1"
}
```

**Response:** `201 Created`

---

### Update Comment

Update a comment.

```
PUT /workspaces/{workspaceID}/comments/{commentID}
```

**Headers:** `Authorization: Bearer <token>`

**Request Body:**

```json
{
  "content": "Updated comment content"
}
```

**Response:** `200 OK`

---

### Delete Comment

Delete a comment.

```
DELETE /workspaces/{workspaceID}/comments/{commentID}
```

**Headers:** `Authorization: Bearer <token>`

**Response:** `204 No Content`

---

### Resolve Comment

Mark a comment as resolved.

```
POST /workspaces/{workspaceID}/comments/{commentID}/resolve
```

**Headers:** `Authorization: Bearer <token>`

**Response:** `200 OK`

---

## Node Types

### List Node Types

Get all available node types.

```
GET /node-types
```

**Headers:** `Authorization: Bearer <token>`

**Query Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| category | string | Filter by category |
| search | string | Search by name |

**Response:** `200 OK`

```json
{
  "success": true,
  "data": [
    {
      "type": "action.http",
      "name": "HTTP Request",
      "description": "Make HTTP requests to any URL",
      "category": "action",
      "version": "1.0.0",
      "inputs": [{ "name": "main", "type": "any" }],
      "outputs": [{ "name": "main", "type": "any" }],
      "parameters": [
        { "name": "url", "type": "string", "required": true, "description": "Request URL" },
        { "name": "method", "type": "string", "required": true, "default": "GET" }
      ]
    }
  ]
}
```

---

### Get Node Categories

Get all node categories.

```
GET /node-types/categories
```

**Headers:** `Authorization: Bearer <token>`

**Response:** `200 OK`

```json
{
  "success": true,
  "data": [
    { "id": "trigger", "name": "Triggers", "description": "Start workflow execution" },
    { "id": "action", "name": "Actions", "description": "Perform operations" },
    { "id": "logic", "name": "Logic", "description": "Control flow" },
    { "id": "integration", "name": "Integrations", "description": "Third-party services" }
  ]
}
```

---

### Get Node Type

Get specific node type details.

```
GET /node-types/{nodeType}
```

**Headers:** `Authorization: Bearer <token>`

**Response:** `200 OK`

---

## Health

### Health Check

Full health check with dependencies.

```
GET /health
```

**Response:** `200 OK`

```json
{
  "success": true,
  "data": {
    "service": "linkflow-api",
    "status": "healthy",
    "checks": {
      "database": "ok",
      "redis": "ok"
    }
  }
}
```

---

### Liveness Check

Simple liveness probe.

```
GET /health/live
```

**Response:** `200 OK`

```json
{
  "status": "alive"
}
```

---

### Readiness Check

Readiness probe.

```
GET /health/ready
```

**Response:** `200 OK`

```json
{
  "status": "ready"
}
```

---

## Admin

Admin endpoints require admin role.

### Get Stream Stats

Get webhook stream statistics.

```
GET /admin/streams/webhooks/stats
```

**Headers:** `Authorization: Bearer <token>` (Admin)

**Response:** `200 OK`

```json
{
  "success": true,
  "data": {
    "enabled": true,
    "stats": {
      "pending": 150,
      "processing": 10,
      "dead_letter": 5,
      "total_processed": 50000
    }
  },
  "actions": [
    { "name": "replay_dlq", "method": "POST", "href": "/api/v1/admin/streams/webhooks/replay", "label": "Replay Dead Letters" },
    { "name": "trim", "method": "POST", "href": "/api/v1/admin/streams/webhooks/trim", "label": "Trim Stream" }
  ]
}
```

---

### Replay Dead Letter Queue

Replay messages from dead letter queue.

```
POST /admin/streams/webhooks/replay
```

**Headers:** `Authorization: Bearer <token>` (Admin)

**Request Body:**

```json
{
  "count": 100
}
```

**Response:** `200 OK`

---

### Trim Stream

Trim old messages from stream.

```
POST /admin/streams/webhooks/trim
```

**Headers:** `Authorization: Bearer <token>` (Admin)

**Request Body:**

```json
{
  "max_len": 100000
}
```

**Response:** `200 OK`

---

## WebSocket

### Connect to WebSocket

Real-time updates via WebSocket.

```
GET /ws
```

**Query Parameters:**

| Parameter | Type | Required |
|-----------|------|----------|
| token | string | Yes (JWT access token) |

**Events Received:**

```json
{
  "type": "execution.started",
  "data": {
    "execution_id": "exec_abc123",
    "workflow_id": "wf_abc123"
  }
}
```

**Event Types:**

| Event | Description |
|-------|-------------|
| `execution.started` | Execution started |
| `execution.completed` | Execution completed |
| `execution.failed` | Execution failed |
| `execution.node.started` | Node started |
| `execution.node.completed` | Node completed |
| `workflow.updated` | Workflow updated |

---

## Metrics

Prometheus metrics endpoint.

```
GET /metrics
```

**Response:** Prometheus format metrics

---

## Shared Executions

### Get Shared Execution

View a shared execution (public link).

```
GET /shared/executions/{token}
```

**Response:** `200 OK`

```json
{
  "success": true,
  "data": {
    "execution": {
      "id": "exec_abc123",
      "status": "success",
      "nodes": [...]
    },
    "expires_at": 1704153600
  }
}
```

---

## Pinned Data

### Get Pinned Data by Workflow

Get all pinned data for a workflow.

```
GET /workspaces/{workspaceID}/workflows/{workflowID}/pinned-data
```

**Headers:** `Authorization: Bearer <token>`

**Response:** `200 OK`

---

### Get Pinned Data by Node

Get pinned data for a specific node.

```
GET /workspaces/{workspaceID}/workflows/{workflowID}/pinned-data/{nodeID}
```

**Headers:** `Authorization: Bearer <token>`

**Response:** `200 OK`

---

### Set Pinned Data

Set pinned data for a node.

```
POST /workspaces/{workspaceID}/workflows/{workflowID}/pinned-data
```

**Headers:** `Authorization: Bearer <token>`

**Request Body:**

```json
{
  "node_id": "node_1",
  "name": "Test Data",
  "data": { "key": "value" }
}
```

**Response:** `200 OK`

---

### Delete Pinned Data

Delete pinned data for a node.

```
DELETE /workspaces/{workspaceID}/workflows/{workflowID}/pinned-data/{nodeID}
```

**Headers:** `Authorization: Bearer <token>`

**Response:** `200 OK`

---

## Wait/Resume

### Resume Waiting Execution

Resume a waiting execution.

```
POST /resume/{token}
```

**Request Body:**

```json
{
  "data": { "approved": true }
}
```

**Response:** `200 OK`

---

### Get Waiting Status

Get status of a waiting execution.

```
GET /resume/{token}/status
```

**Response:** `200 OK`

---

### List Waiting Executions

Get all waiting executions in workspace.

```
GET /workspaces/{workspaceID}/waiting-executions
```

**Headers:** `Authorization: Bearer <token>`

**Response:** `200 OK`

---

### Get Waiting by Execution

Get waiting info for an execution.

```
GET /workspaces/{workspaceID}/executions/{executionID}/waiting
```

**Headers:** `Authorization: Bearer <token>`

**Response:** `200 OK`

---

## Execution Replay

### Replay Execution

Replay an entire execution.

```
POST /workspaces/{workspaceID}/executions/{executionID}/replay
```

**Headers:** `Authorization: Bearer <token>`

**Response:** `202 Accepted`

---

### Replay from Node

Replay execution from a specific node.

```
POST /workspaces/{workspaceID}/executions/{executionID}/replay-from-node
```

**Headers:** `Authorization: Bearer <token>`

**Request Body:**

```json
{
  "node_id": "node_3"
}
```

**Response:** `202 Accepted`
