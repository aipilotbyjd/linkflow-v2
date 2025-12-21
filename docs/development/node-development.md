# Node Development Guide

This guide explains how to create custom integration nodes for Linkflow.

## Table of Contents

1. [Architecture Overview](#architecture-overview)
2. [Node Types](#node-types)
3. [Creating a New Node](#creating-a-new-node)
4. [Best Practices](#best-practices)
5. [Testing](#testing)
6. [Examples](#examples)

## Architecture Overview

Nodes in Linkflow are the building blocks of workflows. Each node:
- Implements the `Node` interface
- Has a unique type identifier
- Receives input data and execution context
- Returns output data or an error

```
┌─────────────────────────────────────────────────────────┐
│                    Workflow Executor                     │
├─────────────────────────────────────────────────────────┤
│                                                         │
│  ┌─────────┐    ┌─────────┐    ┌─────────┐            │
│  │ Trigger │───▶│  Node   │───▶│  Node   │───▶ ...    │
│  └─────────┘    └─────────┘    └─────────┘            │
│                                                         │
│  ExecutionContext:                                      │
│  - Input data from previous node                        │
│  - Node configuration                                   │
│  - Credentials (decrypted)                              │
│  - Workflow variables                                   │
│                                                         │
└─────────────────────────────────────────────────────────┘
```

## Node Types

### 1. Trigger Nodes
Start workflow execution. Examples: Manual, Webhook, Schedule.

```go
// Location: internal/worker/nodes/triggers/
type ManualTrigger struct{}

func (t *ManualTrigger) Type() string { return "trigger.manual" }
```

### 2. Action Nodes
Perform operations. Examples: HTTP Request, Code, Set Variable.

```go
// Location: internal/worker/nodes/actions/
type HTTPRequestNode struct{}

func (n *HTTPRequestNode) Type() string { return "action.http" }
```

### 3. Logic Nodes
Control flow. Examples: IF, Switch, Loop, Merge.

```go
// Location: internal/worker/nodes/logic/
type ConditionNode struct{}

func (n *ConditionNode) Type() string { return "logic.condition" }
```

### 4. Integration Nodes
Connect to external services. Examples: Slack, GitHub, OpenAI.

```go
// Location: internal/worker/nodes/integrations/
type SlackNode struct{}

func (n *SlackNode) Type() string { return "integration.slack" }
```

## Creating a New Node

### Step 1: Define the Node Structure

Create a new file in the appropriate directory:

```go
// internal/worker/nodes/integrations/myservice.go
package integrations

import (
    "context"
    "github.com/linkflow-ai/linkflow/internal/worker/nodes"
)

type MyServiceNode struct{}

func (n *MyServiceNode) Type() string {
    return "integration.myservice"
}
```

### Step 2: Implement the Execute Method

```go
func (n *MyServiceNode) Execute(ctx context.Context, execCtx *nodes.ExecutionContext) (map[string]interface{}, error) {
    // 1. Get configuration
    config := execCtx.Config
    operation := getString(config, "operation", "")
    
    // 2. Get credentials if needed
    creds := execCtx.Credentials
    apiKey := ""
    if creds != nil {
        if key, ok := creds["api_key"].(string); ok {
            apiKey = key
        }
    }
    
    // 3. Get input data
    input := execCtx.Input
    
    // 4. Perform the operation
    switch operation {
    case "create":
        return n.create(ctx, apiKey, config, input)
    case "read":
        return n.read(ctx, apiKey, config, input)
    case "update":
        return n.update(ctx, apiKey, config, input)
    case "delete":
        return n.delete(ctx, apiKey, config, input)
    default:
        return nil, fmt.Errorf("unsupported operation: %s", operation)
    }
}
```

### Step 3: Implement Operations

```go
func (n *MyServiceNode) create(ctx context.Context, apiKey string, config, input map[string]interface{}) (map[string]interface{}, error) {
    // Build request
    client := &http.Client{Timeout: 30 * time.Second}
    
    body, _ := json.Marshal(map[string]interface{}{
        "name": getString(config, "name", ""),
        "data": input,
    })
    
    req, err := http.NewRequestWithContext(ctx, "POST", "https://api.myservice.com/items", bytes.NewReader(body))
    if err != nil {
        return nil, err
    }
    
    req.Header.Set("Authorization", "Bearer "+apiKey)
    req.Header.Set("Content-Type", "application/json")
    
    resp, err := client.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    var result map[string]interface{}
    json.NewDecoder(resp.Body).Decode(&result)
    
    return result, nil
}
```

### Step 4: Register the Node

Add your node to the registry:

```go
// internal/worker/nodes/registry.go
func NewRegistry() *Registry {
    r := &Registry{nodes: make(map[string]Node)}
    
    // ... existing nodes ...
    
    // Register your new node
    r.Register(&integrations.MyServiceNode{})
    
    return r
}
```

### Step 5: Define Node Metadata (for UI)

Create a JSON schema for the frontend:

```json
{
  "type": "integration.myservice",
  "name": "My Service",
  "description": "Connect to My Service API",
  "icon": "myservice",
  "category": "integrations",
  "credentials": {
    "type": "myservice_api",
    "required": true
  },
  "properties": [
    {
      "name": "operation",
      "type": "select",
      "label": "Operation",
      "required": true,
      "options": [
        {"value": "create", "label": "Create Item"},
        {"value": "read", "label": "Read Item"},
        {"value": "update", "label": "Update Item"},
        {"value": "delete", "label": "Delete Item"}
      ]
    },
    {
      "name": "itemId",
      "type": "string",
      "label": "Item ID",
      "displayOptions": {
        "show": {
          "operation": ["read", "update", "delete"]
        }
      }
    }
  ],
  "outputs": {
    "main": [
      {
        "name": "item",
        "type": "object"
      }
    ]
  }
}
```

## Best Practices

### 1. Error Handling

Always return meaningful errors:

```go
if resp.StatusCode >= 400 {
    body, _ := io.ReadAll(resp.Body)
    return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
}
```

### 2. Context and Timeouts

Respect context cancellation:

```go
func (n *MyNode) Execute(ctx context.Context, execCtx *nodes.ExecutionContext) (map[string]interface{}, error) {
    select {
    case <-ctx.Done():
        return nil, ctx.Err()
    default:
    }
    
    // Use context for HTTP requests
    req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
}
```

### 3. Credential Security

Never log credentials:

```go
// BAD
log.Info().Str("api_key", apiKey).Msg("Making request")

// GOOD
log.Info().Str("api_key", "***").Msg("Making request")
```

### 4. Rate Limiting

Implement backoff for rate-limited APIs:

```go
func (n *MyNode) makeRequest(ctx context.Context, req *http.Request) (*http.Response, error) {
    var resp *http.Response
    var err error
    
    for i := 0; i < 3; i++ {
        resp, err = http.DefaultClient.Do(req)
        if err != nil {
            return nil, err
        }
        
        if resp.StatusCode != 429 {
            return resp, nil
        }
        
        resp.Body.Close()
        
        // Exponential backoff
        sleepTime := time.Duration(math.Pow(2, float64(i))) * time.Second
        select {
        case <-ctx.Done():
            return nil, ctx.Err()
        case <-time.After(sleepTime):
        }
    }
    
    return nil, fmt.Errorf("rate limited after 3 retries")
}
```

### 5. Input Validation

Validate inputs early:

```go
func (n *MyNode) Execute(ctx context.Context, execCtx *nodes.ExecutionContext) (map[string]interface{}, error) {
    operation := getString(execCtx.Config, "operation", "")
    if operation == "" {
        return nil, fmt.Errorf("operation is required")
    }
    
    if operation == "read" || operation == "update" || operation == "delete" {
        itemID := getString(execCtx.Config, "itemId", "")
        if itemID == "" {
            return nil, fmt.Errorf("itemId is required for %s operation", operation)
        }
    }
    
    // ... continue with execution
}
```

### 6. Pagination Support

Handle paginated APIs:

```go
func (n *MyNode) list(ctx context.Context, apiKey string, config map[string]interface{}) (map[string]interface{}, error) {
    var allItems []interface{}
    cursor := ""
    limit := getInt(config, "limit", 100)
    
    for {
        url := fmt.Sprintf("https://api.myservice.com/items?limit=%d", limit)
        if cursor != "" {
            url += "&cursor=" + cursor
        }
        
        resp, err := n.makeRequest(ctx, "GET", url, apiKey, nil)
        if err != nil {
            return nil, err
        }
        
        items, _ := resp["items"].([]interface{})
        allItems = append(allItems, items...)
        
        nextCursor, _ := resp["next_cursor"].(string)
        if nextCursor == "" {
            break
        }
        cursor = nextCursor
    }
    
    return map[string]interface{}{
        "items": allItems,
        "count": len(allItems),
    }, nil
}
```

## Testing

### Unit Tests

```go
// internal/worker/nodes/integrations/myservice_test.go
package integrations

import (
    "context"
    "net/http"
    "net/http/httptest"
    "testing"
    
    "github.com/linkflow-ai/linkflow/internal/worker/nodes"
)

func TestMyServiceNode_Create(t *testing.T) {
    // Mock server
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if r.Method != "POST" {
            t.Errorf("expected POST, got %s", r.Method)
        }
        w.WriteHeader(http.StatusCreated)
        w.Write([]byte(`{"id": "123", "name": "test"}`))
    }))
    defer server.Close()
    
    node := &MyServiceNode{baseURL: server.URL}
    
    result, err := node.Execute(context.Background(), &nodes.ExecutionContext{
        Config: map[string]interface{}{
            "operation": "create",
            "name":      "test",
        },
        Credentials: map[string]interface{}{
            "api_key": "test-key",
        },
    })
    
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    
    if result["id"] != "123" {
        t.Errorf("expected id 123, got %v", result["id"])
    }
}
```

### Integration Tests

```go
func TestMyServiceNode_Integration(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping integration test")
    }
    
    apiKey := os.Getenv("MYSERVICE_API_KEY")
    if apiKey == "" {
        t.Skip("MYSERVICE_API_KEY not set")
    }
    
    node := &MyServiceNode{}
    
    // Create
    createResult, err := node.Execute(context.Background(), &nodes.ExecutionContext{
        Config: map[string]interface{}{
            "operation": "create",
            "name":      "integration-test-" + time.Now().Format("20060102150405"),
        },
        Credentials: map[string]interface{}{
            "api_key": apiKey,
        },
    })
    
    if err != nil {
        t.Fatalf("create failed: %v", err)
    }
    
    itemID := createResult["id"].(string)
    
    // Cleanup
    defer func() {
        node.Execute(context.Background(), &nodes.ExecutionContext{
            Config: map[string]interface{}{
                "operation": "delete",
                "itemId":    itemID,
            },
            Credentials: map[string]interface{}{
                "api_key": apiKey,
            },
        })
    }()
    
    // Read
    readResult, err := node.Execute(context.Background(), &nodes.ExecutionContext{
        Config: map[string]interface{}{
            "operation": "read",
            "itemId":    itemID,
        },
        Credentials: map[string]interface{}{
            "api_key": apiKey,
        },
    })
    
    if err != nil {
        t.Fatalf("read failed: %v", err)
    }
    
    if readResult["id"] != itemID {
        t.Errorf("expected id %s, got %v", itemID, readResult["id"])
    }
}
```

## Examples

### Simple HTTP API Integration

See `internal/worker/nodes/integrations/slack.go` for a complete example.

### OAuth Integration

See `internal/worker/nodes/integrations/github.go` for OAuth-based integration.

### Database Integration

See `internal/worker/nodes/integrations/postgresql.go` for database connectivity.

### AI/ML Integration

See `internal/worker/nodes/integrations/openai.go` for AI service integration.

## Node Categories

| Category | Prefix | Examples |
|----------|--------|----------|
| Triggers | `trigger.` | manual, webhook, schedule |
| Actions | `action.` | http, code, set_variable |
| Logic | `logic.` | condition, switch, loop, merge |
| Integrations | `integration.` | slack, github, openai |

## Helper Functions

Common helper functions available in `internal/worker/nodes/`:

```go
// Get string from map with default
func getString(m map[string]interface{}, key, defaultVal string) string

// Get int from map with default
func getInt(m map[string]interface{}, key string, defaultVal int) int

// Get bool from map with default
func getBool(m map[string]interface{}, key string, defaultVal bool) bool

// Get nested map
func getMap(m map[string]interface{}, key string) map[string]interface{}

// Get array
func getArray(m map[string]interface{}, key string) []interface{}
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Implement your node following this guide
4. Write tests (unit and integration)
5. Submit a pull request

### Checklist

- [ ] Node implements `Node` interface
- [ ] All operations have proper error handling
- [ ] Context is respected for cancellation
- [ ] Credentials are not logged
- [ ] Rate limiting is handled
- [ ] Unit tests pass
- [ ] Integration tests pass (if applicable)
- [ ] Node metadata JSON is created
- [ ] Documentation is updated
