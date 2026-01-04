# LinkFlow Frontend UI Blueprint

This document describes all UI pages, components, data structures, and interactions in the LinkFlow application. Use this as a reference when building backend APIs to ensure proper data structures and endpoints.

---

## Table of Contents

1. [Application Structure](#application-structure)
2. [Data Entities](#data-entities)
3. [Pages Overview](#pages-overview)
4. [Workflows Module](#workflows-module)
5. [Credentials Module](#credentials-module)
6. [Variables Module](#variables-module)
7. [Executions Module](#executions-module)
8. [Settings Module](#settings-module)
9. [Common UI Patterns](#common-ui-patterns)
10. [Filter & Sort Options](#filter--sort-options)

---

## Application Structure

```
/app
├── /workflows          → Workflow management (list, create, edit, run)
├── /credentials        → API keys, OAuth tokens, secrets management
├── /variables          → Environment variables (plain text & secrets)
├── /executions         → Workflow execution history & logs
├── /editor/:id         → Visual workflow editor (create/edit)
├── /dashboard          → Overview & analytics
└── /settings
    ├── /general        → Workspace settings
    ├── /profile        → User profile & password
    ├── /teams          → Team members management
    └── /workspaces     → Multi-workspace management
```

---

## Data Entities

### User
```typescript
{
  id: string;
  email: string;
  username?: string;
  first_name: string;
  last_name: string;
  avatar_url?: string | null;
  email_verified: boolean;
  mfa_enabled: boolean;
  created_at: number;          // Unix timestamp
}
```

### Workspace
```typescript
{
  id: string;
  name: string;
  slug: string;                // URL-friendly identifier
  description?: string | null;
  logo_url?: string | null;
  plan_id: string;             // 'free' | 'starter' | 'pro' | 'enterprise'
  settings?: Record<string, unknown>;
  created_at: number;
  updated_at?: number;
}
```

### Workspace Member
```typescript
{
  id: string;
  user: {
    id: string;
    email: string;
    first_name: string;
    last_name: string;
    avatar_url?: string;
  };
  role: 'owner' | 'admin' | 'member' | 'viewer';
  joined_at: number;
}
```

### Workflow
```typescript
{
  id: string;
  name: string;
  description?: string | null;
  status: 'draft' | 'active' | 'inactive' | 'archived';
  version: number;
  tags: string[];
  folder_id?: string | null;
  execution_count: number;
  last_executed_at: number | null;
  nodes?: IWorkflowNode[];
  connections?: IWorkflowConnection[];
  settings?: {
    timeout_seconds?: number;
    retry_on_failure?: boolean;
    max_retries?: number;
    error_workflow_id?: string;
  };
  created_at: number;
  updated_at: number;
}
```

### Folder (for organizing workflows)
```typescript
{
  id: string;
  name: string;
  color: string;               // Tailwind class: 'bg-blue-500', 'bg-emerald-500', etc.
  icon: string;                // Icon name: 'Folder02', 'Mail01', 'UserAdd01', etc.
  workspace_id: string;
  parent_id: string | null;    // For nested folders (future)
  workflow_count: number;
  created_at: number;
  updated_at: number;
}
```

**Available Folder Colors:**
- `bg-blue-500`, `bg-emerald-500`, `bg-violet-500`, `bg-amber-500`
- `bg-red-500`, `bg-pink-500`, `bg-cyan-500`, `bg-orange-500`

**Available Folder Icons:**
- `Folder02`, `Mail01`, `UserAdd01`, `Link01`
- `Analytics01`, `Settings02`, `Code`, `Api`

### Credential
```typescript
{
  id: string;
  name: string;
  type: 'api_key' | 'oauth2' | 'basic' | 'bearer' | 'custom';
  description?: string;
  data?: Record<string, unknown>;  // Encrypted, masked in list responses
  last_used_at?: number;
  created_at: number;
  updated_at: number;
}
```

### Variable
```typescript
{
  id: string;
  key: string;
  value: string;               // Masked if is_secret=true
  is_secret: boolean;
  description?: string;
  created_at: number;
  updated_at: number;
}
```

### Execution
```typescript
{
  id: string;
  workflow_id: string;
  workflow_name?: string;
  workflow_version: number;
  status: 'queued' | 'running' | 'completed' | 'failed' | 'cancelled' | 'timeout';
  trigger_type: 'manual' | 'webhook' | 'schedule';
  nodes_total: number;
  nodes_completed: number;
  error_message?: string;
  queued_at: number;
  started_at?: number;
  completed_at?: number;
  duration_ms?: number;
}
```

---

## Pages Overview

### Page Layout Structure
Every page follows this structure:
```
┌─────────────────────────────────────────────────────────────┐
│ Header (Breadcrumb)                                         │
├─────────────────────────────────────────────────────────────┤
│ Subheader                                                   │
│ ┌─────────────────────┐  ┌─────────────────────────────────┐│
│ │ Search Input        │  │ Filters | Sort | Actions       ││
│ └─────────────────────┘  └─────────────────────────────────┘│
├─────────────────────────────────────────────────────────────┤
│ Container (Main Content)                                    │
│ ┌─────────────────────────────────────────────────────────┐ │
│ │ Stats Cards (optional)                                  │ │
│ ├─────────────────────────────────────────────────────────┤ │
│ │ Bulk Actions Bar (when items selected)                  │ │
│ ├─────────────────────────────────────────────────────────┤ │
│ │ Data Table / Grid / List                                │ │
│ │ - Card Header (Icon + Title + Count)                    │ │
│ │ - Table Body                                            │ │
│ │ - Pagination Footer                                     │ │
│ └─────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
```

---

## Workflows Module

### Page: `/app/workflows`

#### UI Components

**1. Stats Cards (4 cards in a row)**
| Card | Value | Icon |
|------|-------|------|
| Total Workflows | count | `GitMerge` |
| Active | count of status='active' | `CheckmarkCircle02` |
| Inactive | count of status='inactive' | `PauseCircle` |
| Total Executions | sum of all execution_count | `PlayCircle` |

**2. Subheader Filters**
- Search input (searches name, description)
- Status dropdown filter
- Sort dropdown (with order toggle)
- Date range picker (created_after, created_before)
- Tags multi-select dropdown
- View toggle (Table / Grid / Compact)
- "New Folder" button
- "New Workflow" button

**3. Bulk Actions Bar** (appears when items selected)
- Selected count badge
- Activate button (for inactive workflows)
- Deactivate button (for active workflows)
- Move to folder button
- Delete button
- Clear selection button

**4. Table Columns**
| Column | Data | Notes |
|--------|------|-------|
| Checkbox | - | Row selection |
| Star | - | Toggle favorite (stored in localStorage) |
| Name | name + description | With workflow icon |
| Status | status | Badge with color |
| Executions | execution_count | Number |
| Last Run | last_executed_at | Relative time ("2h ago") |
| Updated | updated_at | Relative time |
| Actions | - | Run, Edit, Duplicate, Move, Delete |

**5. Folder Tree View** (when folders exist)
- Expandable/collapsible folders
- Folder header: icon, name, workflow count, active count
- Workflows nested under folders
- "Uncategorized" section for workflows without folder

**6. Grid View** (alternative)
- Card per workflow
- Icon, name, description, status badge
- Execution count, last updated
- Quick actions on hover

**7. Compact View** (alternative)
- Single line per workflow
- Checkbox, star, icon, name, status, runs

**8. Preview Panel** (right sidebar when workflow clicked)
- Workflow details: name, status, description
- Stats: execution count, version
- Tags list
- Timestamps: last executed, created, updated
- Actions: Run Now, Edit, Activate/Deactivate, Delete

#### Modals

**Create/Edit Folder Modal**
- Name input
- Color picker (8 options)
- Icon picker (8 options)
- Preview of folder appearance

**Move to Folder Modal**
- List of folders with icons
- "Uncategorized" option
- Current folder indicator

#### API Endpoints Expected
```
GET    /workspaces/:id/workflows         → List with filters
GET    /workspaces/:id/workflows/:id     → Single workflow
POST   /workspaces/:id/workflows         → Create
PUT    /workspaces/:id/workflows/:id     → Update
DELETE /workspaces/:id/workflows/:id     → Delete
POST   /workspaces/:id/workflows/:id/execute    → Run workflow
POST   /workspaces/:id/workflows/:id/activate   → Set status=active
POST   /workspaces/:id/workflows/:id/deactivate → Set status=inactive
POST   /workspaces/:id/workflows/:id/duplicate  → Clone workflow

GET    /workspaces/:id/folders           → List folders
POST   /workspaces/:id/folders           → Create folder
PUT    /workspaces/:id/folders/:id       → Update folder
DELETE /workspaces/:id/folders/:id       → Delete folder
```

#### Filter Parameters
```typescript
{
  search?: string;           // Search in name, description
  status?: 'draft' | 'active' | 'inactive' | 'archived';
  tags?: string;             // Comma-separated tag values
  created_after?: number;    // Unix timestamp
  created_before?: number;   // Unix timestamp
  sort_by?: 'name' | 'created_at' | 'updated_at' | 'execution_count' | 'last_executed_at';
  order?: 'asc' | 'desc';
  page?: number;
  per_page?: number;
}
```

---

## Credentials Module

### Page: `/app/credentials`

#### UI Components

**1. Subheader Filters**
- Search input
- Type dropdown filter
- Sort dropdown
- "New Credential" button

**2. Table Columns**
| Column | Data | Notes |
|--------|------|-------|
| Name | name + description | |
| Type | type | Badge with color |
| Last Used | last_used_at | Date or "Never" |
| Created | created_at | Date |
| Actions | - | Edit, Test Connection, Delete |

**3. Type Colors**
| Type | Color |
|------|-------|
| api_key | Blue |
| oauth2 | Violet |
| basic | Amber |
| bearer | Emerald |
| custom | Zinc |

#### Empty State Features
- Encrypted (AES-256 encryption)
- Reusable (Use anywhere)
- Testable (Verify connection)

#### API Endpoints Expected
```
GET    /workspaces/:id/credentials       → List
GET    /workspaces/:id/credentials/:id   → Single (with masked data)
POST   /workspaces/:id/credentials       → Create
PUT    /workspaces/:id/credentials/:id   → Update
DELETE /workspaces/:id/credentials/:id   → Delete
POST   /workspaces/:id/credentials/:id/test → Test connection
```

#### Filter Parameters
```typescript
{
  search?: string;
  type?: 'api_key' | 'oauth2' | 'basic' | 'bearer' | 'custom';
  sort_by?: 'name' | 'created_at' | 'type' | 'last_used_at';
  order?: 'asc' | 'desc';
  page?: number;
  per_page?: number;
}
```

---

## Variables Module

### Page: `/app/variables`

#### UI Components

**1. Subheader Filters**
- Search input
- Secret filter (All / Plain Text / Secret)
- Sort dropdown
- "New Variable" button

**2. Table Columns**
| Column | Data | Notes |
|--------|------|-------|
| Key | key + description | With variable icon |
| Value | value | Show/hide toggle for secrets, Copy button |
| Type | is_secret | "Secret" badge if true |
| Updated | updated_at | Date |
| Actions | - | Edit, Copy, Delete |

#### Empty State Features
- Secure (Encrypted at rest)
- Flexible (Plain or secret)
- Referenced (Use in workflows)

#### API Endpoints Expected
```
GET    /workspaces/:id/variables         → List
GET    /workspaces/:id/variables/:id     → Single
POST   /workspaces/:id/variables         → Create
PUT    /workspaces/:id/variables/:id     → Update
DELETE /workspaces/:id/variables/:id     → Delete
```

#### Filter Parameters
```typescript
{
  search?: string;
  is_secret?: boolean;
  sort_by?: 'key' | 'created_at' | 'updated_at';
  order?: 'asc' | 'desc';
  page?: number;
  per_page?: number;
}
```

---

## Executions Module

### Page: `/app/executions`

#### UI Components

**1. Subheader Filters**
- Search input
- Status dropdown filter
- Trigger type dropdown filter
- Date range picker
- Sort dropdown
- Refresh button

**2. Table Columns**
| Column | Data | Notes |
|--------|------|-------|
| Workflow | workflow_name | Link to workflow |
| Status | status | Badge with color + progress bar for running |
| Trigger | trigger_type | Badge |
| Progress | nodes_completed / nodes_total | Progress bar |
| Duration | duration_ms | Formatted ("2.5s", "1m 30s") |
| Started | started_at | Relative time |
| Actions | - | View Details, Retry (if failed), Cancel (if running) |

**3. Status Colors**
| Status | Color |
|--------|-------|
| queued | Amber |
| running | Blue |
| completed | Emerald |
| failed | Red |
| cancelled | Zinc |
| timeout | Violet |

**4. Trigger Colors**
| Trigger | Color |
|---------|-------|
| manual | Blue |
| webhook | Emerald |
| schedule | Amber |

#### API Endpoints Expected
```
GET    /workspaces/:id/executions        → List
GET    /workspaces/:id/executions/:id    → Single with details
POST   /workspaces/:id/executions/:id/cancel → Cancel running
POST   /workspaces/:id/executions/:id/retry  → Retry failed
```

#### Filter Parameters
```typescript
{
  workflow_id?: string;
  status?: 'queued' | 'running' | 'completed' | 'failed' | 'cancelled' | 'timeout';
  trigger_type?: 'manual' | 'webhook' | 'schedule';
  start_date?: number;       // Unix timestamp
  end_date?: number;         // Unix timestamp
  search?: string;
  sort_by?: 'queued_at' | 'started_at' | 'completed_at' | 'duration';
  order?: 'asc' | 'desc';
  page?: number;
  per_page?: number;
}
```

---

## Settings Module

### Page: `/app/settings/general`

#### UI Components

**Form Fields**
| Field | Type | Notes |
|-------|------|-------|
| Workspace Name | Text input | Required |
| Language | Select | en, de, es, fr, ar, tr |
| Timezone | Select | List of timezones |
| Date Format | Select | MM/DD/YYYY, DD/MM/YYYY, YYYY-MM-DD |

#### API Endpoints Expected
```
GET  /workspaces/:id/settings    → Get settings
PUT  /workspaces/:id/settings    → Update settings
```

---

### Page: `/app/settings/profile`

#### UI Components

**Profile Section**
| Field | Type | Notes |
|-------|------|-------|
| Avatar | Image upload | With preview |
| First Name | Text input | Required |
| Last Name | Text input | Required |
| Username | Text input | Optional |
| Email | Text input | Read-only, verified badge |

**Password Section**
| Field | Type | Notes |
|-------|------|-------|
| Current Password | Password | Required |
| New Password | Password | With strength meter |
| Confirm Password | Password | Must match |

**Security Section**
- MFA Status (Enabled/Disabled badge)
- Enable/Disable MFA button

#### API Endpoints Expected
```
GET  /users/me                  → Get current user
PUT  /users/me                  → Update profile
POST /users/me/change-password  → Change password
POST /users/me/mfa/setup        → Get MFA QR code
POST /users/me/mfa/verify       → Verify & enable MFA
POST /users/me/mfa/disable      → Disable MFA
```

---

### Page: `/app/settings/teams`

#### UI Components

**1. Subheader**
- Search input
- Role filter dropdown
- "Invite Member" button

**2. Table Columns**
| Column | Data | Notes |
|--------|------|-------|
| Member | avatar + name + email | |
| Role | role | Badge, dropdown to change |
| Joined | joined_at | Date |
| Actions | - | Change Role, Remove |

**3. Role Colors**
| Role | Color |
|------|-------|
| owner | Violet |
| admin | Blue |
| member | Emerald |
| viewer | Zinc |

**4. Invite Modal**
- Email input
- Role select (admin, member, viewer)

#### API Endpoints Expected
```
GET    /workspaces/:id/members           → List members
POST   /workspaces/:id/members/invite    → Send invite
PUT    /workspaces/:id/members/:id       → Change role
DELETE /workspaces/:id/members/:id       → Remove member
```

---

### Page: `/app/settings/workspaces`

#### UI Components

**1. Subheader**
- Search input
- Refresh button
- "New Workspace" button

**2. Table Columns**
| Column | Data | Notes |
|--------|------|-------|
| Workspace | logo + name + slug | "Active" badge for current |
| Description | description | |
| Plan | plan_id | Badge with color |
| Created | created_at | Date |
| Actions | - | Switch, Settings, Members, Delete |

**3. Plan Colors**
| Plan | Color |
|------|-------|
| free | Zinc |
| starter | Blue |
| pro | Violet |
| enterprise | Emerald |

**4. Create Modal**
- Name input
- Slug input (auto-generated from name)
- Description textarea

#### API Endpoints Expected
```
GET    /workspaces                → List all user's workspaces
GET    /workspaces/:id            → Single workspace
POST   /workspaces                → Create
PUT    /workspaces/:id            → Update
DELETE /workspaces/:id            → Delete
```

---

## Common UI Patterns

### Table Card Structure
```
┌─────────────────────────────────────────────────────────────┐
│ Card Header                                                 │
│ ┌───────────────────────────┐  ┌───────────────────────────┐│
│ │ [Icon] Title              │  │ X items                   ││
│ └───────────────────────────┘  └───────────────────────────┘│
├─────────────────────────────────────────────────────────────┤
│ Table                                                       │
│ ┌─────────────────────────────────────────────────────────┐ │
│ │ Header Row (sortable columns)                           │ │
│ ├─────────────────────────────────────────────────────────┤ │
│ │ Data Rows                                               │ │
│ │ - Hover: show actions                                   │ │
│ │ - Click: select or preview                              │ │
│ └─────────────────────────────────────────────────────────┘ │
├─────────────────────────────────────────────────────────────┤
│ Footer                                                      │
│ ┌───────────────────────────┐  ┌───────────────────────────┐│
│ │ Show [5/10/20] ▼          │  │ ◀◀ ◀ Page X of Y ▶ ▶▶    ││
│ └───────────────────────────┘  └───────────────────────────┘│
└─────────────────────────────────────────────────────────────┘
```

### Loading States (Skeletons)
- Table skeleton: mimics header + 5-8 rows
- Card grid skeleton: 6 card placeholders
- Stats cards skeleton: 4 stat placeholders

### Error States
- Large circular icon with red background
- Title: "Failed to Load {Entity}"
- Description: helpful message
- "Try Again" button
- "Contact Support" link

### Empty States
- Animated icon container with decorative circles
- Title: "Create your first {entity}" or "No {entities} found"
- Description with feature highlights (3 cards)
- Primary CTA button + secondary action

### Filter Dropdowns
- Button shows current selection
- Solid/primary when filter active
- Checkmark icon for selected option
- "Clear" action when filter active

### Action Dropdowns
- Triggered by MoreVertical icon button
- Common actions: Edit, Delete
- Dangerous actions in red text

### Badges
- Soft variant (light background)
- Colors: emerald, amber, blue, red, violet, zinc, primary

### Time Formatting
- Relative: "Just now", "5m ago", "2h ago", "3d ago"
- Absolute: locale date string for older dates

---

## Filter & Sort Options

### Workflow Filters
| Filter | Options |
|--------|---------|
| Status | All, Active, Inactive, Draft, Archived |
| Sort By | Updated, Created, Name, Runs, Last Run |
| Order | Newest First, Oldest First |
| Tags | Production, Staging, Development, Critical, API, Scheduled, Webhook |
| Date Range | Calendar picker |

### Credential Filters
| Filter | Options |
|--------|---------|
| Type | All Types, API Key, OAuth2, Basic Auth, Bearer Token, Custom |
| Sort By | Created, Name, Type, Last Used |

### Variable Filters
| Filter | Options |
|--------|---------|
| Visibility | All, Plain Text, Secret |
| Sort By | Created, Updated, Name |

### Execution Filters
| Filter | Options |
|--------|---------|
| Status | All, Queued, Running, Completed, Failed, Cancelled, Timeout |
| Trigger | All Triggers, Manual, Webhook, Schedule |
| Sort By | Queued, Started, Completed, Duration |
| Date Range | Calendar picker |

### Team Member Filters
| Filter | Options |
|--------|---------|
| Role | All Roles, Owner, Admin, Member, Viewer |

---

## Response Formats

### Paginated List Response
```typescript
{
  data: T[];
  meta: {
    total: number;
    page: number;
    per_page: number;
    total_pages: number;
    has_next_page: boolean;
    has_prev_page: boolean;
  }
}
```

### Single Item Response
```typescript
{
  data: T;
}
```

### Error Response
```typescript
{
  error: {
    code: string;
    message: string;
    details?: Record<string, string[]>;  // Validation errors
  }
}
```

---

## Authentication

### Token Storage
- Access token: stored in memory/localStorage
- Refresh token: stored in localStorage
- Auto-refresh on 401 responses

### Required Headers
```
Authorization: Bearer <access_token>
X-Workspace-Id: <current_workspace_id>
```

### Auth Endpoints
```
POST /auth/login          → Login
POST /auth/register       → Register
POST /auth/refresh        → Refresh tokens
POST /auth/logout         → Logout
POST /auth/forgot-password → Send reset email
POST /auth/reset-password  → Reset with token
```

---

## Notes for Backend Development

1. **Timestamps**: All timestamps should be Unix timestamps (seconds)
2. **Workspace Scoping**: Most endpoints are scoped to workspace ID
3. **Soft Delete**: Consider soft delete for workflows, credentials
4. **Audit Logging**: Track who/when for all mutations
5. **Rate Limiting**: Especially for workflow execution
6. **Pagination**: Default 10-25 items, max 100
7. **Search**: Case-insensitive, partial match on name/description
8. **Sorting**: Support multiple sort fields if possible
9. **Validation**: Return field-level validation errors
10. **Secrets**: Never return actual credential data in list responses
