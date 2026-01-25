# LinkFlow Editor API Documentation

**Base URL:** `http://localhost:8090/api/v1`  
**Version:** 1.0.0

This documentation specifically covers the API endpoints used by the Workflow Editor for creating, managing, debugging, and executing workflows. It provides detailed schemas and realistic examples to aid in frontend development.

## Table of Contents

- [Node Types](#node-types)
- [Workflows](#workflows)
  - [Management](#workflow-management)
  - [Validation](#workflow-validation)
  - [Testing](#node-testing)
  - [Versioning](#workflow-versioning)
  - [Pinned Data](#pinned-data)
- [Executions](#executions)
- [Credentials](#credentials)
- [Variables](#variables)

---

## Node Types

Endpoints for retrieving available nodes for the editor palette. The editor uses these definitions to render nodes and their configuration forms.

### List Node Types

Get all available node types with their full definitions, including parameters, inputs, and outputs.

```
GET /node-types
```

**Response:** `200 OK`

```json
{
  "success": true,
  "data": [
    {
      "type": "trigger.webhook",
      "name": "Webhook Trigger",
      "displayName": "Webhook",
      "category": "trigger",
      "icon": "webhook",
      "color": "#6366F1",
      "description": "Trigger workflow when an HTTP request is received at the webhook URL",
      "inputs": [],
      "outputs": [
        {
          "name": "main",
          "type": "object",
          "description": "Webhook request data including headers, body, and query parameters"
        }
      ],
      "parameters": [
        {
          "name": "http_method",
          "displayName": "HTTP Method",
          "type": "options",
          "required": false,
          "default": "POST",
          "options": [
            { "name": "GET", "value": "GET" },
            { "name": "POST", "value": "POST" }
          ]
        },
        {
          "name": "path",
          "displayName": "Webhook Path",
          "type": "string",
          "placeholder": "/my-webhook"
        },
        {
          "name": "authentication",
          "displayName": "Authentication",
          "type": "options",
          "default": "none",
          "options": [
            { "name": "None", "value": "none" },
            { "name": "Basic Auth", "value": "basic" }
          ]
        }
      ]
    }
  ]
}
```

### Get Node Categories

Get a list of node categories to organize the palette in the UI.

```
GET /node-types/categories
```

**Response:** `200 OK`

```json
{
  "success": true,
  "data": ["trigger", "action", "logic", "transform"]
}
```

---

## Workflows

### Workflow Management

Standard CRUD operations for workflows. The editor primarily interacts with these for saving and loading the canvas state.

#### Create Workflow

```
POST /workflows
```

**Request Body:**

```json
{
  "name": "Payment Webhook Handler",
  "description": "Handles incoming Stripe webhooks",
  "nodes": [
    {
      "id": "node_1",
      "type": "trigger.webhook",
      "name": "Webhook",
      "position": { "x": 100, "y": 200 },
      "parameters": {
        "http_method": "POST",
        "path": "/stripe-hook",
        "authentication": "none"
      }
    },
    {
      "id": "node_2",
      "type": "action.webhook_response",
      "name": "Send 200 OK",
      "position": { "x": 400, "y": 200 },
      "parameters": {
        "status_code": 200,
        "content_type": "application/json",
        "body": { "received": true }
      }
    }
  ],
  "connections": [
    {
      "source": "node_1",
      "target": "node_2"
    }
  ],
  "settings": {
    "error_workflow": "wf_error_handler_123"
  },
  "tags": ["payment", "webhook"]
}
```

#### Get Workflow

```
GET /workflows/{workflowId}
```

#### Update Workflow

```
PUT /workflows/{workflowId}
```

**Request Body:**

```json
{
  "name": "Payment Webhook Handler (Updated)",
  "nodes": [ ... ],
  "connections": [ ... ],
  "settings": { ... }
}
```

### Workflow Validation

Validate a workflow graph before saving or execution. This checks for disconnected nodes, missing required parameters, and invalid connections.

```
POST /workflows/validate
```

**Request Body:**

```json
{
  "nodes": [
    {
      "id": "node_1",
      "type": "trigger.webhook",
      "name": "Webhook",
      "parameters": {}
    }
  ],
  "connections": []
}
```

**Response:** `200 OK`

```json
{
  "success": true,
  "data": {
    "valid": false,
    "errors": [],
    "warnings": [
       {
        "node_id": "node_1",
        "message": "Node is not connected to any other node",
        "code": "ORPHAN_NODE"
      }
    ]
  }
}
```

### Node Testing

Execute a single node individually using provided input data in a "mock" runtime environment. This is used for "Test Node" functionality in the editor.

```
POST /workflows/test-node
```

**Request Body:**

```json
{
  "node_type": "action.webhook_response",
  "parameters": {
    "status_code": 201,
    "body": { "message": "Created" }
  },
  "input": {
    "json": { "original_request_id": "req_123" }
  }
}
```

**Response:** `200 OK`

```json
{
  "success": true,
  "data": {
    "success": true,
    "output": {
      "_webhook_response": true,
      "status_code": 201,
      "body": { "message": "Created" },
      "headers": { "Content-Type": "application/json" }
    },
    "duration_ms": 5
  }
}
```

### Workflow Versioning

Manage workflow history and versions.

#### Get Versions

List all versions of a workflow.

```
GET /workflows/{workflowId}/versions
```

#### Get Specific Version

```
GET /workflows/{workflowId}/versions/{versionNumber}
```

#### Rollback Version

Restore a previous version.

```
POST /workflows/{workflowId}/versions/{versionNumber}/rollback
```

#### Compare Versions

Compare the current workflow state with a specific version or two versions.

```
GET /workflows/{workflowId}/compare-versions?v1=1&v2=2
```

**Response:** `200 OK`

```json
{
  "success": true,
  "data": {
    "summary": {
      "nodesAdded": 1,
      "nodesRemoved": 0,
      "nodesModified": 2,
      "connectionsAdded": 1,
      "connectionsRemoved": 0,
       "settingsChanged": false
    },
    "differences": [
      {
        "type": "modified",
        "path": "nodes[0].parameters.url",
        "oldValue": "http://api.com",
        "newValue": "http://api.v2.com",
         "description": "Changed 'url' parameter in 'HTTP Request' node"
      }
    ]
  }
}
```

### Pinned Data

Manage "pinned" data (mock/test data) for nodes in the editor. This allows developers to save the output of a node execution to use as input for subsequent nodes during design time.

#### Get Pinned Data

Get all pinned data for a workflow.

```
GET /workflows/{workflowId}/pinned-data
```

#### Get Pinned Data for Node

```
GET /workflows/{workflowId}/pinned-data/{nodeId}
```

#### Set Pinned Data

Save pinned data for a specific node (usually the output of a test run).

```
POST /workflows/{workflowId}/pinned-data
```

**Request Body:**

```json
{
  "node_id": "node_1",
  "data": [ 
    { 
      "json": { 
        "id": 1, 
        "name": "Test User", 
        "email": "test@example.com" 
      } 
    } 
  ]
}
```

#### Delete Pinned Data

Remove pinned data for a node.

```
DELETE /workflows/{workflowId}/pinned-data/{nodeId}
```

---

## Executions

Endpoints for viewing and managing workflow runs.

### List Executions

```
GET /executions?workflowId={id}
```

### Get Execution Details

Get full details of a specific execution, including step-by-step data.

```
GET /executions/{executionId}
```

### Get Execution Node Data

Get detailed data (input/output/error) for a specific node in an execution.

```
GET /executions/{executionId}/nodes/{nodeId}
```

### Retry Execution

Retry a failed execution.

```
POST /executions/{executionId}/retry
```

---

## Credentials

Manage authentication credentials for nodes.

### List Credentials

```
GET /credentials
```

### Create Credential

```
POST /credentials
```

**Request Body:**

```json
{
  "name": "Production Stripe API Key",
  "type": "stripe_api",
  "data": { 
    "api_key": "sk_live_..." 
  }
}
```

### Test Credential

Verify if a credential works.

```
POST /credentials/{credentialId}/test
```

---

## Variables

Access environment variables and secrets available to the workspace.

### List Variables

```
GET /variables
```

### Resolve Variable

Resolve a variable value (requires permission).

```
GET /variables/resolve?name=MY_VAR
```
