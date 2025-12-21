# API Documentation

LinkFlow provides a RESTful API for managing workflows, executions, credentials, and more.

## Base URL

- **Development**: `http://localhost:8090/api/v1`
- **Production**: `https://api.linkflow.ai/v1`

## Authentication

All API endpoints (except auth endpoints) require authentication using either:

### JWT Bearer Token

```http
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

### API Key

```http
X-API-Key: lf_live_xxxxxxxxxxxxxxxxxxxx
```

## Quick Start

### 1. Register/Login

```bash
# Register
curl -X POST https://api.linkflow.ai/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "securepassword123",
    "first_name": "John",
    "last_name": "Doe"
  }'

# Login
curl -X POST https://api.linkflow.ai/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "securepassword123"
  }'
```

### 2. Create a Workflow

```bash
curl -X POST https://api.linkflow.ai/v1/workflows \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "My First Workflow",
    "description": "A simple HTTP workflow",
    "nodes": [
      {
        "id": "trigger_1",
        "type": "trigger.manual",
        "name": "Manual Trigger",
        "position": {"x": 100, "y": 100},
        "config": {}
      },
      {
        "id": "http_1",
        "type": "action.http",
        "name": "HTTP Request",
        "position": {"x": 300, "y": 100},
        "config": {
          "method": "GET",
          "url": "https://api.example.com/data"
        }
      }
    ],
    "connections": [
      {
        "source": "trigger_1",
        "target": "http_1"
      }
    ]
  }'
```

### 3. Execute a Workflow

```bash
curl -X POST https://api.linkflow.ai/v1/workflows/{workflow_id}/execute \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "input_data": {
      "key": "value"
    }
  }'
```

## API Reference

### Authentication

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/auth/register` | POST | Register new user |
| `/auth/login` | POST | Login with credentials |
| `/auth/refresh` | POST | Refresh access token |
| `/auth/logout` | POST | Logout (revoke token) |
| `/auth/oauth/{provider}` | GET | Start OAuth flow |

### Users

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/users/me` | GET | Get current user |
| `/users/me` | PUT | Update current user |

### Workflows

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/workflows` | GET | List workflows |
| `/workflows` | POST | Create workflow |
| `/workflows/{id}` | GET | Get workflow |
| `/workflows/{id}` | PUT | Update workflow |
| `/workflows/{id}` | DELETE | Delete workflow |
| `/workflows/{id}/execute` | POST | Execute workflow |
| `/workflows/{id}/activate` | POST | Activate workflow |
| `/workflows/{id}/deactivate` | POST | Deactivate workflow |
| `/workflows/{id}/versions` | GET | Get version history |

### Executions

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/executions` | GET | List executions |
| `/executions/{id}` | GET | Get execution details |
| `/executions/{id}/cancel` | POST | Cancel execution |
| `/executions/{id}/retry` | POST | Retry execution |
| `/executions/{id}/logs` | GET | Get execution logs |

### Credentials

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/credentials` | GET | List credentials |
| `/credentials` | POST | Create credential |
| `/credentials/{id}` | GET | Get credential |
| `/credentials/{id}` | PUT | Update credential |
| `/credentials/{id}` | DELETE | Delete credential |

### Schedules

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/schedules` | GET | List schedules |
| `/schedules` | POST | Create schedule |
| `/schedules/{id}` | DELETE | Delete schedule |

### Webhooks

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/webhook/{endpoint_id}` | POST | Trigger webhook |

### Health

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/health` | GET | Overall health |
| `/health/live` | GET | Liveness check |
| `/health/ready` | GET | Readiness check |

## Response Format

### Success Response

```json
{
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "name": "My Workflow"
  }
}
```

### Paginated Response

```json
{
  "data": [...],
  "total": 100,
  "page": 1,
  "limit": 20
}
```

### Error Response

```json
{
  "error": "validation_error",
  "message": "Name is required",
  "details": {
    "field": "name",
    "code": "required"
  }
}
```

## Status Codes

| Code | Description |
|------|-------------|
| 200 | Success |
| 201 | Created |
| 202 | Accepted (async operation) |
| 400 | Bad Request |
| 401 | Unauthorized |
| 403 | Forbidden |
| 404 | Not Found |
| 409 | Conflict |
| 422 | Unprocessable Entity |
| 429 | Too Many Requests |
| 500 | Internal Server Error |

## Rate Limiting

| Endpoint | Limit |
|----------|-------|
| Authentication | 10 req/min |
| Workflows | 100 req/min |
| Executions | 200 req/min |
| Other | 1000 req/min |

Rate limit headers:
```http
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 95
X-RateLimit-Reset: 1640000000
```

## WebSocket

Connect to `/ws` for real-time execution updates:

```javascript
const ws = new WebSocket('wss://api.linkflow.ai/ws?token=YOUR_TOKEN');

ws.onmessage = (event) => {
  const data = JSON.parse(event.data);
  console.log('Event:', data.type, data.payload);
};
```

### Event Types

- `execution:started` - Workflow execution started
- `execution:completed` - Workflow completed
- `execution:failed` - Workflow failed
- `node:started` - Node execution started
- `node:completed` - Node completed
- `node:failed` - Node failed

## OpenAPI Specification

The full OpenAPI 3.0 specification is available at:
- [openapi.yaml](openapi.yaml)

## Postman Collection

Import our Postman collection for easy API testing:
- [LinkFlow.postman_collection.json](postman/LinkFlow.postman_collection.json)

## SDKs

- JavaScript/TypeScript: `npm install @linkflow/sdk`
- Python: `pip install linkflow`
- Go: `go get github.com/linkflow-ai/linkflow-go`
