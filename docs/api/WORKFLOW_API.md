# LinkFlow Workflow API - Complete Reference

## Base URL
```
/api/v1/workspaces/{workspaceID}/workflows
```

All workflow endpoints require authentication (JWT Bearer token) and workspace context.

---

## 1. LIST WORKFLOWS

**GET** `/api/v1/workspaces/{workspaceID}/workflows`

### Query Parameters (Filters)

| Parameter | Type | Description |
|-----------|------|-------------|
| `status` | string | Filter by status: `active`, `inactive`, `draft`, `archived` |
| `search` | string | Search in name/description |
| `tags` | string | Comma-separated tags (e.g., `tag1,tag2`) |
| `created_after` | string | ISO 8601 date (e.g., `2024-01-01T00:00:00Z`) |
| `created_before` | string | ISO 8601 date |
| `updated_after` | string | ISO 8601 date |
| `updated_before` | string | ISO 8601 date |
| `sort_by` | string | `name`, `created_at`, `updated_at`, `execution_count`, `last_executed_at` (default: `created_at`) |
| `order` | string | `asc` or `desc` (default: `desc`) |
| `page` | int | Page number (default: 1) |
| `per_page` | int | Items per page (default: 20, max: 100) |
| `fields` | string | Comma-separated fields to return (e.g., `id,name,status`) |
| `include` | string | Include related resources (e.g., `schedules`) |

### Response
```json
{
  "success": true,
  "data": [
    {
      "id": "uuid",
      "workspace_id": "uuid",
      "created_by": "uuid",
      "name": "My Workflow",
      "description": "Optional description",
      "status": "active",
      "version": 1,
      "nodes": [...],
      "connections": [...],
      "settings": {...},
      "tags": ["tag1", "tag2"],
      "color": "#FF5733",
      "icon": "workflow-icon",
      "category": "automation",
      "is_favorite": false,
      "folder_id": "uuid",
      "error_workflow_id": "uuid",
      "error_trigger": "on_failure",
      "execution_count": 42,
      "last_executed_at": 1704067200,
      "activated_at": 1704067200,
      "archived_at": null,
      "created_at": 1704067200,
      "updated_at": 1704067200,
      "actions": [
        {"name": "execute", "method": "POST", "href": "/api/v1/workspaces/{wsID}/workflows/{wfID}/execute"},
        {"name": "deactivate", "method": "POST", "href": "/api/v1/workspaces/{wsID}/workflows/{wfID}/deactivate"},
        {"name": "delete", "method": "DELETE", "href": "/api/v1/workspaces/{wsID}/workflows/{wfID}"}
      ]
    }
  ],
  "included": {
    "schedules": {...}
  },
  "links": {
    "self": "/api/v1/workspaces/{wsID}/workflows?page=1&per_page=20",
    "next": "/api/v1/workspaces/{wsID}/workflows?page=2&per_page=20",
    "prev": null,
    "first": "/api/v1/workspaces/{wsID}/workflows?page=1&per_page=20",
    "last": "/api/v1/workspaces/{wsID}/workflows?page=5&per_page=20"
  },
  "meta": {
    "total": 100,
    "page": 1,
    "per_page": 20,
    "total_pages": 5
  },
  "timestamp": 1704067200
}
```

---

## 2. CREATE WORKFLOW

**POST** `/api/v1/workspaces/{workspaceID}/workflows`

### Request Body
```json
{
  "name": "My Workflow",
  "description": "Optional description",
  "nodes": [
    {
      "id": "node-1",
      "type": "trigger.manual",
      "name": "Start",
      "parameters": {}
    },
    {
      "id": "node-2",
      "type": "action.http",
      "name": "HTTP Request",
      "parameters": {
        "url": "https://api.example.com",
        "method": "GET"
      }
    }
  ],
  "connections": [
    {
      "id": "conn-1",
      "source_node_id": "node-1",
      "source_handle": "main",
      "target_node_id": "node-2",
      "target_handle": "main"
    }
  ],
  "settings": {
    "timeout_seconds": 300,
    "retry_on_failure": false
  },
  "tags": ["automation", "api"],
  "color": "#FF5733",
  "icon": "api-icon",
  "category": "integrations"
}
```

### Validation Rules

| Field | Validation |
|-------|------------|
| `name` | **Required**, min=1, max=255 characters |
| `description` | Optional, max=1000 characters |
| `nodes` | **Required**, JSON array |
| `connections` | **Required**, JSON array |
| `tags` | Optional, max=10 tags, each max=50 characters |
| `color` | Optional, max=20 characters |
| `icon` | Optional, max=50 characters |
| `category` | Optional, max=50 characters |

### Response (201 Created)
```json
{
  "success": true,
  "data": {
    "id": "uuid",
    "workspace_id": "uuid",
    "created_by": "uuid",
    "name": "My Workflow",
    "status": "draft",
    "version": 1,
    ...
  }
}
```

---

## 3. GET WORKFLOW

**GET** `/api/v1/workspaces/{workspaceID}/workflows/{workflowID}`

### Response
```json
{
  "success": true,
  "data": {
    "id": "uuid",
    "workspace_id": "uuid",
    "name": "My Workflow",
    "description": "...",
    "status": "active",
    "version": 3,
    "nodes": [...],
    "connections": [...],
    "settings": {...},
    "tags": [...],
    ...
    "actions": [
      {"name": "edit", "method": "PUT", "href": "..."},
      {"name": "versions", "method": "GET", "href": "..."},
      {"name": "execute", "method": "POST", "href": "..."},
      {"name": "deactivate", "method": "POST", "href": "..."},
      {"name": "delete", "method": "DELETE", "href": "..."}
    ]
  },
  "links": {
    "self": "/api/v1/workspaces/{wsID}/workflows/{wfID}"
  }
}
```

---

## 4. UPDATE WORKFLOW

**PUT** `/api/v1/workspaces/{workspaceID}/workflows/{workflowID}`

### Request Body (all fields optional)
```json
{
  "name": "Updated Name",
  "description": "Updated description",
  "nodes": [...],
  "connections": [...],
  "settings": {...},
  "tags": ["new-tag"],
  "color": "#00FF00",
  "icon": "new-icon",
  "category": "new-category",
  "is_favorite": true,
  "folder_id": "uuid-or-null"
}
```

### Validation Rules
Same as Create, but all fields are optional. If `nodes` or `connections` are provided, workflow structure validation is performed.

### Response
Returns updated workflow object.

---

## 5. DELETE WORKFLOW

**DELETE** `/api/v1/workspaces/{workspaceID}/workflows/{workflowID}`

### Business Rules
- Cannot delete workflow with running executions

### Response (204 No Content)

---

## 6. EXECUTE WORKFLOW

**POST** `/api/v1/workspaces/{workspaceID}/workflows/{workflowID}/execute`

### Request Body (optional)
```json
{
  "input_data": {
    "key": "value"
  }
}
```

### Business Rules
- Workflow must have at least one node
- Workflow status must be `active` or `draft` (draft allowed for testing)
- Workspace must not exceed execution quota

### Response (202 Accepted)
```json
{
  "success": true,
  "data": {
    "task_id": "asynq-task-id",
    "status": "queued"
  }
}
```

---

## 7. ACTIVATE WORKFLOW

**POST** `/api/v1/workspaces/{workspaceID}/workflows/{workflowID}/activate`

### Business Rules
- Workflow must NOT already be active
- Workflow must NOT be archived
- Workflow **must have at least one trigger node** (type starting with `trigger.`)

### Response
```json
{
  "success": true,
  "data": {"status": "active"},
  "actions": [
    {"name": "execute", "method": "POST", "href": "..."},
    {"name": "deactivate", "method": "POST", "href": "..."}
  ]
}
```

---

## 8. DEACTIVATE WORKFLOW

**POST** `/api/v1/workspaces/{workspaceID}/workflows/{workflowID}/deactivate`

### Business Rules
- Workflow must be `active`
- Workflow must NOT have active schedules

### Response
```json
{
  "success": true,
  "data": {"status": "inactive"},
  "actions": [
    {"name": "activate", "method": "POST", "href": "..."}
  ]
}
```

---

## 9. CLONE WORKFLOW

**POST** `/api/v1/workspaces/{workspaceID}/workflows/{workflowID}/clone`

### Request Body
```json
{
  "name": "Cloned Workflow Name"
}
```

### Validation
| Field | Validation |
|-------|------------|
| `name` | **Required**, min=1, max=255 characters |

### Response (201 Created)
Returns the cloned workflow.

---

## 10. DUPLICATE WORKFLOW

**POST** `/api/v1/workspaces/{workspaceID}/workflows/{workflowID}/duplicate`

### Request Body (optional)
```json
{
  "name": "Custom Name",
  "variables": {
    "API_KEY": "new-value"
  }
}
```

Variables are substituted in nodes using `{{variable_name}}` syntax.

### Response (201 Created)
Returns duplicated workflow (name defaults to "Original Name (Copy)").

---

## 11. EXPORT WORKFLOW

**GET** `/api/v1/workspaces/{workspaceID}/workflows/{workflowID}/export`

### Response (JSON file download)
```json
{
  "version": "1.0",
  "exportedAt": 1704067200,
  "workflow": {
    "name": "My Workflow",
    "description": "...",
    "nodes": [...],
    "connections": [...],
    "settings": {...},
    "tags": [...]
  }
}
```

---

## 12. IMPORT WORKFLOW

**POST** `/api/v1/workspaces/{workspaceID}/workflows/import`

### Request Body
```json
{
  "version": "1.0",
  "workflow": {
    "name": "Imported Workflow",
    "description": "...",
    "nodes": [...],
    "connections": [...],
    "settings": {...},
    "tags": [...]
  }
}
```

### Response (201 Created)
Returns imported workflow (name appended with " (Imported)").

---

## 13. GET WORKFLOW VERSIONS

**GET** `/api/v1/workspaces/{workspaceID}/workflows/{workflowID}/versions`

### Response
```json
{
  "success": true,
  "data": [
    {
      "id": "uuid",
      "workflow_id": "uuid",
      "version": 3,
      "nodes": [...],
      "connections": [...],
      "settings": {...},
      "created_by": "uuid",
      "change_message": "Added HTTP node",
      "created_at": 1704067200,
      "actions": [
        {"name": "view", "method": "GET", "href": "..."},
        {"name": "diff", "method": "GET", "href": "..."},
        {"name": "rollback", "method": "POST", "href": "..."}
      ]
    }
  ],
  "meta": {"total": 3, ...}
}
```

---

## 14. GET SPECIFIC VERSION

**GET** `/api/v1/workspaces/{workspaceID}/workflows/{workflowID}/versions/{version}`

### Response
```json
{
  "success": true,
  "data": {
    "id": "uuid",
    "workflow_id": "uuid",
    "version": 2,
    "nodes": [...],
    "connections": [...],
    "settings": {...},
    "created_by": "uuid",
    "change_message": "...",
    "created_at": 1704067200,
    "actions": [
      {"name": "diff", "method": "GET", "href": "..."},
      {"name": "rollback", "method": "POST", "href": "..."}
    ]
  }
}
```

---

## 15. ROLLBACK VERSION

**POST** `/api/v1/workspaces/{workspaceID}/workflows/{workflowID}/versions/{version}/rollback`

### Response
Returns the workflow restored to the specified version.

---

## 16. COMPARE VERSIONS

**GET** `/api/v1/workspaces/{workspaceID}/workflows/{workflowID}/compare-versions`

Query parameters: `from` and `to` version numbers.

**GET** `/api/v1/workspaces/{workspaceID}/workflows/{workflowID}/versions/{version}/compare`

Compares specified version with current version.

---

## 17. VALIDATE WORKFLOW

**POST** `/api/v1/workspaces/{workspaceID}/workflows/validate`

### Request Body
```json
{
  "nodes": [...],
  "connections": [...]
}
```

### Response
```json
{
  "valid": true,
  "errors": []
}
```

Or with errors:
```json
{
  "valid": false,
  "errors": [
    {"type": "error", "node": "node-1", "message": "Unknown node type: invalid.type"},
    {"type": "warning", "message": "Workflow should have at least one trigger node"}
  ]
}
```

---

## 18. TEST NODE

**POST** `/api/v1/workspaces/{workspaceID}/workflows/test-node`

### Request Body
```json
{
  "node_type": "action.http",
  "parameters": {
    "url": "https://api.example.com",
    "method": "GET"
  },
  "input": {
    "data": "from-previous-node"
  }
}
```

### Response
```json
{
  "success": true,
  "output": {...},
  "duration": 150
}
```

Or on failure:
```json
{
  "success": false,
  "error": "connection timeout",
  "duration": 30000
}
```

---

## WORKFLOW STRUCTURE VALIDATION

When creating or updating workflows, the following validations are performed:

### Node Validation

| Code | Description |
|------|-------------|
| `EMPTY_WORKFLOW` | Workflow must have at least one node |
| `MISSING_NODE_ID` | Node ID is required |
| `DUPLICATE_NODE_ID` | Node IDs must be unique |
| `MISSING_NODE_TYPE` | Node type is required |
| `INVALID_NODE_TYPE` | Unknown node type |
| `MISSING_NODE_NAME` | Node name is recommended (warning) |

### Connection Validation

| Code | Description |
|------|-------------|
| `MISSING_CONNECTION_ID` | Connection ID is required |
| `DUPLICATE_CONNECTION_ID` | Connection IDs must be unique |
| `MISSING_SOURCE_NODE` | Source node ID is required |
| `INVALID_SOURCE_NODE` | Source node not found |
| `MISSING_TARGET_NODE` | Target node ID is required |
| `INVALID_TARGET_NODE` | Target node not found |
| `SELF_LOOP` | Node cannot connect to itself |
| `DUPLICATE_CONNECTION` | Duplicate connection between same nodes |

### Graph Structure Validation

| Code | Description |
|------|-------------|
| `NO_TRIGGER_NODE` | Workflow must have at least one trigger node |
| `MULTIPLE_TRIGGER_NODES` | Warning: multiple trigger nodes found |
| `ORPHAN_NODE` | Node not connected to workflow |
| `CYCLE_DETECTED` | Workflow contains a cycle (infinite loop) |
| `UNREACHABLE_NODE` | Node not reachable from any trigger |

---

## NODE STRUCTURE

### Node Object
```json
{
  "id": "unique-node-id",
  "type": "trigger.manual",
  "name": "Human-readable name",
  "parameters": {
    "param1": "value1"
  }
}
```

### Connection Object
```json
{
  "id": "unique-connection-id",
  "source_node_id": "node-1",
  "source_handle": "main",
  "target_node_id": "node-2",
  "target_handle": "main"
}
```

---

## NODE TYPES

### Triggers
| Type | Description | Required Parameters |
|------|-------------|---------------------|
| `trigger.manual` | Manual execution | None |
| `trigger.webhook` | HTTP webhook | `path` (optional), `method` (default: POST) |
| `trigger.schedule` | Cron schedule | `cron` (required), `timezone` (default: UTC) |
| `trigger.error` | Error trigger | None |

### Actions
| Type | Description | Required Parameters |
|------|-------------|---------------------|
| `action.http` | HTTP request | `url`, `method` |
| `action.code` | Execute JavaScript code | `code` |
| `action.function` | Execute function | `code` |
| `action.transform` | Transform data | `code` |
| `action.set` | Set variables | `name`, `value` |
| `action.respond` | Webhook response | `statusCode`, `body` |
| `action.respondWebhook` | Respond to webhook | `statusCode`, `body` |
| `action.stopError` | Stop on error | None |
| `action.sub_workflow` | Execute sub-workflow | `workflowId` |
| `action.execute_workflow` | Execute workflow | `workflowId` |

### Logic
| Type | Description | Required Parameters |
|------|-------------|---------------------|
| `logic.condition` | Conditional branch (IF) | `conditions` |
| `logic.switch` | Multi-way branch | `value`, `cases` |
| `logic.loop` | Iterate items | `items` |
| `logic.merge` | Merge branches | None |
| `logic.filter` | Filter array | `conditions` |
| `logic.sort` | Sort array | `field`, `order` |
| `logic.limit` | Limit/paginate | `limit`, `offset` |
| `logic.unique` | Remove duplicates | `field` |
| `logic.splitBatches` | Split into batches | `batchSize` |
| `logic.aggregate` | Aggregate data | `operation` |
| `logic.wait` | Pause execution | `duration` |
| `logic.noop` | Pass-through | None |
| `logic.dataFilter` | Advanced data filter | `conditions` |
| `logic.dataSort` | Advanced data sort | `field`, `order` |
| `logic.dataLimit` | Advanced data limit | `limit`, `offset` |
| `logic.remove_duplicates` | Remove duplicates | `field` |
| `logic.datetime` | Date/time operations | `operation` |
| `logic.expression` | Evaluate expression | `expression` |
| `logic.math` | Math operations | `operation` |
| `logic.crypto` | Cryptographic operations | `operation` |
| `logic.xml` | Parse/generate XML | `operation` |
| `logic.json_transform` | JSON transformation | `operation` |
| `logic.splitData` | Split data | `splitBy` |
| `logic.mergeData` | Merge data | None |
| `logic.html_extract` | Extract from HTML | `selector` |

### Error Handling
| Type | Description | Required Parameters |
|------|-------------|---------------------|
| `logic.try_catch` | Try/catch block | None |
| `logic.retry` | Retry with backoff | `maxRetries`, `delay` |
| `logic.throw_error` | Throw custom error | `message` |
| `logic.continue_on_fail` | Continue on failure | None |
| `logic.timeout` | Add timeout | `timeout` |
| `logic.fallback` | Fallback value | `fallbackValue` |

### Integrations
| Type | Description |
|------|-------------|
| `integration.slack` | Slack messages |
| `integration.discord` | Discord messages |
| `integration.telegram` | Telegram messages |
| `integration.github` | GitHub API |
| `integration.postgres` | PostgreSQL queries |
| `integration.mysql` | MySQL queries |
| `integration.mongodb` | MongoDB operations |
| `integration.redis` | Redis commands |
| `integration.email` | Send email via SMTP |
| `integration.openai` | OpenAI API |
| `integration.anthropic` | Anthropic Claude API |
| `integration.googleSheets` | Google Sheets operations |
| `integration.stripe` | Stripe payments |
| `integrations.aws_s3` | AWS S3 operations |
| `integrations.google_drive` | Google Drive operations |
| `integrations.twilio` | Twilio SMS/calls |
| `integrations.sendgrid` | SendGrid email |
| `integrations.jira` | Jira issues |
| `integrations.salesforce` | Salesforce CRM |
| `integrations.airtable` | Airtable bases |
| `integrations.notion` | Notion pages |
| `integrations.graphql` | GraphQL queries |
| `integrations.ftp` | FTP operations |
| `integrations.sftp` | SFTP operations |

---

## NODE PARAMETER VALIDATION

Parameters support **expressions** using `{{variable}}` syntax, which bypasses type validation.

### HTTP Node (`action.http`)
| Parameter | Type | Required | Values |
|-----------|------|----------|--------|
| `url` | URL | Yes | Valid URL |
| `method` | enum | Yes | GET, POST, PUT, PATCH, DELETE, HEAD, OPTIONS |
| `headers` | object | No | Key-value pairs |
| `body` | string | No | Request body |
| `timeout` | number | No | Default: 30 |
| `retry` | number | No | Default: 0 |

### Code Node (`action.code`)
| Parameter | Type | Required | Values |
|-----------|------|----------|--------|
| `code` | string | Yes | JavaScript/Python code |
| `language` | enum | No | javascript, python |

### Schedule Trigger (`trigger.schedule`)
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `cron` | cron | Yes | 5 or 6 part cron expression |
| `timezone` | string | No | IANA timezone (default: UTC) |

### Webhook Trigger (`trigger.webhook`)
| Parameter | Type | Required | Values |
|-----------|------|----------|--------|
| `path` | string | No | URL path |
| `method` | enum | No | GET, POST, PUT, DELETE, PATCH |
| `responseMode` | enum | No | onReceived, lastNode |

### Database Nodes (`integration.postgres`, `integration.mysql`)
| Parameter | Type | Required | Values |
|-----------|------|----------|--------|
| `operation` | enum | Yes | select, insert, update, delete, raw |
| `table` | string | No | Table name |
| `query` | string | No | Raw SQL query |
| `values` | object | No | Values to insert/update |
| `where` | object | No | WHERE conditions |

### Slack Node (`integration.slack`)
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `channel` | string | Yes | Slack channel |
| `text` | string | Conditional | Message text (required if no blocks) |
| `blocks` | array | Conditional | Block Kit blocks (required if no text) |
| `username` | string | No | Bot username |
| `icon_emoji` | string | No | Bot icon emoji |

### Condition Node (`logic.condition`)
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `conditions` | array | Yes | Array of condition objects |
| `combineWith` | enum | No | `and`, `or` (default: and) |

Each condition object has:
| Field | Type | Description |
|-------|------|-------------|
| `leftValue` | any | Left operand (supports `{{expression}}` syntax) |
| `operator` | string | Comparison operator (see below) |
| `rightValue` | any | Right operand (supports `{{expression}}` syntax) |

**Operators:** `equal`, `notEqual`, `greater`, `greaterEqual`, `less`, `lessEqual`, `contains`, `notContains`, `startsWith`, `endsWith`, `regex`, `isEmpty`, `isNotEmpty`, `isTrue`, `isFalse`, `isNull`, `isNotNull`, `in`, `notIn`, `between`

### Loop Node (`logic.loop`)
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `items` | expression | Yes | Items to iterate over |
| `batchSize` | number | No | Default: 1 |

---

## NODE SCHEMAS (Input/Output)

### trigger.manual
```json
{
  "inputs": [],
  "outputs": [
    {"name": "triggered", "type": "boolean"},
    {"name": "timestamp", "type": "string"}
  ]
}
```

### trigger.webhook
```json
{
  "inputs": [],
  "outputs": [
    {"name": "method", "type": "string"},
    {"name": "headers", "type": "object"},
    {"name": "body", "type": "any"},
    {"name": "query", "type": "object"}
  ]
}
```

### trigger.schedule
```json
{
  "inputs": [
    {"name": "cron", "type": "string", "required": true}
  ],
  "outputs": [
    {"name": "scheduledTime", "type": "string"}
  ]
}
```

### action.http
```json
{
  "inputs": [
    {"name": "url", "type": "string", "required": true},
    {"name": "method", "type": "select", "default": "GET", "options": ["GET", "POST", "PUT", "PATCH", "DELETE"]},
    {"name": "headers", "type": "object"},
    {"name": "body", "type": "any"}
  ],
  "outputs": [
    {"name": "status", "type": "number"},
    {"name": "headers", "type": "object"},
    {"name": "body", "type": "any"},
    {"name": "json", "type": "object"}
  ]
}
```

### action.set
```json
{
  "inputs": [
    {"name": "values", "type": "object", "required": true}
  ],
  "outputs": [
    {"name": "data", "type": "object"}
  ]
}
```

### action.code
```json
{
  "inputs": [
    {"name": "language", "type": "select", "default": "javascript", "options": ["javascript", "expr"]},
    {"name": "code", "type": "code", "required": true}
  ],
  "outputs": [
    {"name": "result", "type": "any"}
  ]
}
```

### logic.condition (IF Node)
```json
{
  "inputs": [
    {"name": "conditions", "type": "array", "required": true},
    {"name": "combineWith", "type": "string", "default": "and"}
  ],
  "outputs": [
    {"name": "result", "type": "boolean"},
    {"name": "branch", "type": "string", "values": ["true", "false"]},
    {"name": "data", "type": "any"}
  ]
}
```

**Example:**
```json
{
  "id": "if-node-1",
  "type": "logic.condition",
  "name": "Check Status",
  "parameters": {
    "conditions": [
      {
        "leftValue": "{{$json.status}}",
        "operator": "equal",
        "rightValue": "success"
      }
    ],
    "combineWith": "and"
  }
}
```

### logic.loop
```json
{
  "inputs": [
    {"name": "items", "type": "array", "required": true},
    {"name": "batchSize", "type": "number", "default": 1}
  ],
  "outputs": [
    {"name": "item", "type": "any"},
    {"name": "index", "type": "number"}
  ]
}
```

### logic.switch
```json
{
  "inputs": [
    {"name": "value", "type": "any", "required": true},
    {"name": "cases", "type": "array"}
  ],
  "outputs": [
    {"name": "matched", "type": "string"}
  ]
}
```

---

## ERROR RESPONSES

### Validation Error (400)
```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Validation failed",
    "details": [
      {"field": "name", "message": "This field is required"},
      {"field": "nodes[0].type", "message": "Invalid value"}
    ]
  },
  "timestamp": 1704067200
}
```

### Workflow Validation Error (400)
```json
{
  "success": false,
  "error": {
    "code": "WORKFLOW_VALIDATION_ERROR",
    "message": "Workflow validation failed",
    "details": [
      {"field": "node:node-1.type", "message": "Unknown node type: invalid.type"},
      {"field": "connections", "message": "Workflow contains a cycle"}
    ]
  },
  "timestamp": 1704067200
}
```

### Business Rule Error (400)
```json
{
  "success": false,
  "error": {
    "code": "BAD_REQUEST",
    "message": "workflow must have at least one trigger node to be activated"
  },
  "timestamp": 1704067200
}
```

### Not Found (404)
```json
{
  "success": false,
  "error": {
    "code": "NOT_FOUND",
    "message": "workflow not found"
  },
  "timestamp": 1704067200
}
```

### Forbidden (403)
```json
{
  "success": false,
  "error": {
    "code": "FORBIDDEN",
    "message": "execution limit reached"
  },
  "timestamp": 1704067200
}
```

---

## WORKFLOW STATUSES

| Status | Description | Can Execute | Can Activate | Can Deactivate |
|--------|-------------|-------------|--------------|----------------|
| `draft` | New workflow | Yes (testing) | Yes | No |
| `active` | Production ready | Yes | No | Yes |
| `inactive` | Deactivated | No | Yes | No |
| `archived` | Archived | No | No | No |

---

## WORKFLOW MODEL

### Database Schema
```sql
workflows (
  id              UUID PRIMARY KEY,
  workspace_id    UUID NOT NULL,
  created_by      UUID NOT NULL,
  name            VARCHAR(255) NOT NULL,
  description     TEXT,
  status          VARCHAR(20) DEFAULT 'draft',
  version         INT DEFAULT 1,
  nodes           JSONB DEFAULT '[]',
  connections     JSONB DEFAULT '[]',
  settings        JSONB DEFAULT '{}',
  tags            TEXT[],
  color           VARCHAR(20),
  icon            VARCHAR(50),
  category        VARCHAR(50),
  is_favorite     BOOLEAN DEFAULT false,
  folder_id       UUID,
  error_workflow_id UUID,
  error_trigger   VARCHAR(50),
  execution_count INT DEFAULT 0,
  last_executed_at TIMESTAMP,
  activated_at    TIMESTAMP,
  archived_at     TIMESTAMP,
  created_at      TIMESTAMP,
  updated_at      TIMESTAMP,
  deleted_at      TIMESTAMP
)
```

### Version History Schema
```sql
workflow_versions (
  id              UUID PRIMARY KEY,
  workflow_id     UUID NOT NULL,
  version         INT NOT NULL,
  nodes           JSONB NOT NULL,
  connections     JSONB NOT NULL,
  settings        JSONB,
  created_by      UUID,
  change_message  TEXT,
  created_at      TIMESTAMP
)
```

---

## RELATED ENDPOINTS

### Workflow Variables
- `GET /workspaces/{id}/workflows/{wfId}/variables` - List variables
- `POST /workspaces/{id}/workflows/{wfId}/variables` - Create variable
- `PUT /workspaces/{id}/workflows/{wfId}/variables/{varId}` - Update variable
- `DELETE /workspaces/{id}/workflows/{wfId}/variables/{varId}` - Delete variable

### Workflow Comments
- `GET /workspaces/{id}/workflows/{wfId}/comments` - List comments
- `POST /workspaces/{id}/workflows/{wfId}/comments` - Create comment
- `PUT /comments/{commentId}` - Update comment
- `DELETE /comments/{commentId}` - Delete comment
- `POST /comments/{commentId}/resolve` - Resolve comment

### Workflow Webhooks
- `POST /workspaces/{id}/workflows/{wfId}/webhooks` - Generate webhook
- `GET /workspaces/{id}/workflows/{wfId}/webhooks` - List webhooks

### Pinned Data (for testing)
- `GET /workspaces/{id}/workflows/{wfId}/pinned-data` - Get all pinned data
- `GET /workspaces/{id}/workflows/{wfId}/pinned-data/{nodeId}` - Get node pinned data
- `POST /workspaces/{id}/workflows/{wfId}/pinned-data` - Set pinned data
- `DELETE /workspaces/{id}/workflows/{wfId}/pinned-data/{nodeId}` - Delete pinned data

### Analytics
- `GET /workspaces/{id}/workflows/{wfId}/analytics` - Get workflow analytics

---

## NODE TYPES ENDPOINTS

### List All Node Types
**GET** `/api/v1/node-types`

Query: `?category=trigger|action|logic|integration`

### Get Node Type Details
**GET** `/api/v1/node-types/{nodeType}`

### Get Node Categories
**GET** `/api/v1/node-types/categories`

### Response Example
```json
{
  "success": true,
  "data": [
    {
      "type": "trigger.manual",
      "name": "Manual Trigger",
      "description": "Manually start workflow execution",
      "category": "trigger",
      "icon": "play",
      "version": "1.0.0",
      "tags": ["trigger", "manual"],
      "schema": {
        "inputs": [],
        "outputs": [...]
      }
    }
  ]
}
```
