# Getting Started with LinkFlow

This guide will help you set up LinkFlow and create your first workflow.

## Prerequisites

- Docker and Docker Compose
- `curl` or a REST client (Postman)

## Installation

### 1. Clone the Repository

```bash
git clone https://github.com/linkflow-ai/linkflow.git
cd linkflow
```

### 2. Configure Environment

```bash
cp .env.example .env
```

Edit `.env` and set at minimum:
```bash
JWT_SECRET=your-32-character-secret-key-here
```

### 3. Start Services

```bash
docker-compose up -d
```

Wait for services to start:
```bash
docker-compose logs -f api
# Wait for "Server started on :8090"
```

### 4. Verify Installation

```bash
curl http://localhost:8090/health
```

Expected response:
```json
{
  "status": "healthy"
}
```

## Create Your First Workflow

### 1. Register an Account

```bash
curl -X POST http://localhost:8090/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "securepassword123",
    "first_name": "John",
    "last_name": "Doe"
  }'
```

Save the `access_token` from the response.

### 2. Create a Simple Workflow

This workflow will make an HTTP request when triggered manually:

```bash
export TOKEN="your-access-token"

curl -X POST http://localhost:8090/api/v1/workflows \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Hello World Workflow",
    "description": "My first workflow",
    "nodes": [
      {
        "id": "trigger",
        "type": "trigger.manual",
        "name": "Manual Trigger",
        "position": {"x": 100, "y": 100},
        "config": {}
      },
      {
        "id": "http",
        "type": "action.http",
        "name": "HTTP Request",
        "position": {"x": 300, "y": 100},
        "config": {
          "method": "GET",
          "url": "https://httpbin.org/get"
        }
      }
    ],
    "connections": [
      {
        "source": "trigger",
        "target": "http"
      }
    ]
  }'
```

Note the workflow `id` from the response.

### 3. Execute the Workflow

```bash
export WORKFLOW_ID="your-workflow-id"

curl -X POST http://localhost:8090/api/v1/workflows/$WORKFLOW_ID/execute \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{}'
```

The response includes an `execution_id`.

### 4. Check Execution Status

```bash
export EXECUTION_ID="your-execution-id"

curl http://localhost:8090/api/v1/executions/$EXECUTION_ID \
  -H "Authorization: Bearer $TOKEN"
```

## Understanding Workflows

### Workflow Structure

```json
{
  "name": "Workflow Name",
  "nodes": [
    {
      "id": "unique-id",
      "type": "node-type",
      "name": "Display Name",
      "position": {"x": 0, "y": 0},
      "config": {}
    }
  ],
  "connections": [
    {
      "source": "source-node-id",
      "target": "target-node-id"
    }
  ]
}
```

### Node Types

| Type | Category | Description |
|------|----------|-------------|
| `trigger.manual` | Trigger | Start workflow manually |
| `trigger.webhook` | Trigger | Start via webhook |
| `trigger.schedule` | Trigger | Start on schedule |
| `action.http` | Action | Make HTTP request |
| `action.code` | Action | Execute code |
| `action.set` | Action | Set variables |
| `logic.condition` | Logic | If/then branching |
| `logic.loop` | Logic | Loop over items |
| `integration.slack` | Integration | Send Slack messages |
| `integration.email` | Integration | Send emails |

## Next Steps

### Add a Condition

```json
{
  "nodes": [
    {
      "id": "trigger",
      "type": "trigger.manual",
      "config": {}
    },
    {
      "id": "condition",
      "type": "logic.condition",
      "config": {
        "conditions": [
          {
            "left": "{{ $input.value }}",
            "operator": "greater_than",
            "right": 10
          }
        ]
      }
    },
    {
      "id": "true_branch",
      "type": "action.http",
      "config": {
        "method": "POST",
        "url": "https://example.com/high"
      }
    },
    {
      "id": "false_branch",
      "type": "action.http",
      "config": {
        "method": "POST",
        "url": "https://example.com/low"
      }
    }
  ],
  "connections": [
    {"source": "trigger", "target": "condition"},
    {"source": "condition", "target": "true_branch", "sourceHandle": "true"},
    {"source": "condition", "target": "false_branch", "sourceHandle": "false"}
  ]
}
```

### Create a Webhook Trigger

```bash
# Create workflow with webhook trigger
curl -X POST http://localhost:8090/api/v1/workflows \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Webhook Workflow",
    "nodes": [
      {
        "id": "webhook",
        "type": "trigger.webhook",
        "config": {
          "method": "POST",
          "path": "/my-webhook"
        }
      }
    ]
  }'

# Activate the workflow
curl -X POST http://localhost:8090/api/v1/workflows/$WORKFLOW_ID/activate \
  -H "Authorization: Bearer $TOKEN"

# Trigger via webhook
curl -X POST http://localhost:8090/webhooks/endpoint-id \
  -H "Content-Type: application/json" \
  -d '{"data": "test"}'
```

### Schedule a Workflow

```bash
# Create a schedule (run every hour)
curl -X POST http://localhost:8090/api/v1/schedules \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "workflow_id": "'$WORKFLOW_ID'",
    "cron": "0 * * * *",
    "timezone": "UTC"
  }'
```

## Using Credentials

For integrations that require authentication:

```bash
# Create a credential
curl -X POST http://localhost:8090/api/v1/credentials \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "My Slack Bot",
    "type": "slack",
    "data": {
      "bot_token": "xoxb-your-token"
    }
  }'

# Use in workflow node
{
  "id": "slack",
  "type": "integration.slack",
  "config": {
    "operation": "send_message",
    "channel": "#general",
    "text": "Hello from LinkFlow!"
  },
  "credentials": "credential-id"
}
```

## Expressions

Use expressions to reference data from previous nodes:

```json
{
  "config": {
    "text": "Hello {{ $trigger.body.name }}!",
    "value": "{{ $http.response.data.count * 2 }}"
  }
}
```

### Expression Syntax

- `{{ $trigger }}` - Trigger node output
- `{{ $nodeId.field }}` - Specific node output
- `{{ $input }}` - Input data passed to execution
- `{{ $env.VAR }}` - Environment variable

### Functions

```
{{ toUpper($trigger.name) }}
{{ formatDate($trigger.date, "YYYY-MM-DD") }}
{{ length($http.response.items) }}
```

## Real-time Updates

Connect via WebSocket for live execution updates:

```javascript
const ws = new WebSocket('ws://localhost:8090/ws?token=YOUR_TOKEN');

ws.onmessage = (event) => {
  const data = JSON.parse(event.data);
  console.log('Event:', data.type);
  console.log('Data:', data.payload);
};
```

## Troubleshooting

### Workflow Won't Execute

1. Check workflow is active:
```bash
curl http://localhost:8090/api/v1/workflows/$WORKFLOW_ID \
  -H "Authorization: Bearer $TOKEN"
```

2. Check execution logs:
```bash
docker-compose logs worker
```

### Credential Issues

1. Verify credential exists:
```bash
curl http://localhost:8090/api/v1/credentials \
  -H "Authorization: Bearer $TOKEN"
```

2. Test credential connection from node config

## Resources

- [API Reference](../api/README.md)
- [Node Development](../development/node-development.md)
- [Configuration Guide](../development/configuration.md)
