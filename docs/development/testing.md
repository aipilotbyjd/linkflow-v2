# Testing Guide

This guide covers testing practices for LinkFlow.

## Running Tests

```bash
# Run all tests
make test

# Run with verbose output
go test -v ./...

# Run with coverage
make test-coverage

# Run specific package tests
go test -v ./internal/api/handlers/...

# Run specific test
go test -v -run TestCreateWorkflow ./internal/api/handlers/...
```

## Test Structure

```
internal/
├── api/
│   └── handlers/
│       ├── workflow.go
│       └── workflow_test.go     # Unit tests
├── worker/
│   └── nodes/
│       ├── integrations/
│       │   ├── slack.go
│       │   └── slack_test.go    # Unit tests
│       └── integrations_test.go # Integration tests
└── pkg/
    └── database/
        └── database_test.go     # Integration tests
```

## Unit Tests

### Handler Tests

```go
package handlers

import (
    "bytes"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
)

// Mock repository
type MockWorkflowRepo struct {
    mock.Mock
}

func (m *MockWorkflowRepo) Create(ctx context.Context, w *domain.Workflow) error {
    args := m.Called(ctx, w)
    return args.Error(0)
}

func TestWorkflowHandler_Create(t *testing.T) {
    tests := []struct {
        name       string
        body       map[string]interface{}
        setupMock  func(*MockWorkflowRepo)
        wantStatus int
    }{
        {
            name: "success",
            body: map[string]interface{}{
                "name": "Test Workflow",
            },
            setupMock: func(m *MockWorkflowRepo) {
                m.On("Create", mock.Anything, mock.Anything).Return(nil)
            },
            wantStatus: http.StatusCreated,
        },
        {
            name: "validation error - missing name",
            body: map[string]interface{}{},
            setupMock: func(m *MockWorkflowRepo) {},
            wantStatus: http.StatusBadRequest,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Setup
            mockRepo := new(MockWorkflowRepo)
            tt.setupMock(mockRepo)
            handler := NewWorkflowHandler(mockRepo)

            // Create request
            body, _ := json.Marshal(tt.body)
            req := httptest.NewRequest("POST", "/api/v1/workflows", bytes.NewReader(body))
            req.Header.Set("Content-Type", "application/json")
            rec := httptest.NewRecorder()

            // Execute
            handler.Create(rec, req)

            // Assert
            assert.Equal(t, tt.wantStatus, rec.Code)
            mockRepo.AssertExpectations(t)
        })
    }
}
```

### Node Tests

```go
package integrations

import (
    "context"
    "net/http"
    "net/http/httptest"
    "testing"

    "github.com/linkflow-ai/linkflow/internal/worker/nodes"
    "github.com/stretchr/testify/assert"
)

func TestSlackNode_SendMessage(t *testing.T) {
    // Mock Slack API
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        assert.Equal(t, "POST", r.Method)
        assert.Equal(t, "/api/chat.postMessage", r.URL.Path)
        
        w.WriteHeader(http.StatusOK)
        w.Write([]byte(`{"ok": true, "ts": "1234567890.123456"}`))
    }))
    defer server.Close()

    node := &SlackNode{baseURL: server.URL}
    
    result, err := node.Execute(context.Background(), &nodes.ExecutionContext{
        Config: map[string]interface{}{
            "operation": "send_message",
            "channel":   "#general",
            "text":      "Hello, World!",
        },
        Credentials: map[string]interface{}{
            "bot_token": "xoxb-test-token",
        },
    })

    assert.NoError(t, err)
    assert.Equal(t, true, result["ok"])
}
```

## Integration Tests

Integration tests require running infrastructure. Skip them in CI with short flag.

```go
func TestDatabase_Integration(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping integration test")
    }

    db, err := database.Connect(testConfig)
    require.NoError(t, err)
    defer db.Close()

    // Test operations
    user := &domain.User{
        Email: "test@example.com",
    }
    err = db.Users().Create(context.Background(), user)
    assert.NoError(t, err)
    assert.NotEmpty(t, user.ID)
}
```

### Running Integration Tests

```bash
# Start dependencies first
make dev-deps

# Run all tests including integration
go test -v ./...

# Skip integration tests (short mode)
go test -v -short ./...
```

## End-to-End Tests

```go
func TestE2E_WorkflowExecution(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping e2e test")
    }

    // Setup test server
    app := setupTestApp(t)
    server := httptest.NewServer(app)
    defer server.Close()

    client := &http.Client{}

    // 1. Create workflow
    createResp, err := client.Post(
        server.URL+"/api/v1/workflows",
        "application/json",
        strings.NewReader(`{"name": "E2E Test"}`),
    )
    require.NoError(t, err)
    assert.Equal(t, http.StatusCreated, createResp.StatusCode)

    var workflow map[string]interface{}
    json.NewDecoder(createResp.Body).Decode(&workflow)
    workflowID := workflow["id"].(string)

    // 2. Execute workflow
    execResp, err := client.Post(
        server.URL+"/api/v1/workflows/"+workflowID+"/execute",
        "application/json",
        nil,
    )
    require.NoError(t, err)
    assert.Equal(t, http.StatusAccepted, execResp.StatusCode)

    // 3. Wait for completion
    var execution map[string]interface{}
    json.NewDecoder(execResp.Body).Decode(&execution)
    executionID := execution["id"].(string)

    // Poll for completion
    assert.Eventually(t, func() bool {
        resp, _ := client.Get(server.URL + "/api/v1/executions/" + executionID)
        json.NewDecoder(resp.Body).Decode(&execution)
        return execution["status"] == "completed"
    }, 30*time.Second, 1*time.Second)
}
```

## Test Utilities

### Test Fixtures

```go
// testutil/fixtures.go
package testutil

func CreateTestUser(t *testing.T, db *database.DB) *domain.User {
    user := &domain.User{
        Email:     fmt.Sprintf("test-%s@example.com", uuid.New().String()),
        FirstName: "Test",
        LastName:  "User",
    }
    err := db.Users().Create(context.Background(), user)
    require.NoError(t, err)
    return user
}

func CreateTestWorkflow(t *testing.T, db *database.DB, userID uuid.UUID) *domain.Workflow {
    workflow := &domain.Workflow{
        Name:        "Test Workflow",
        WorkspaceID: userID,
        CreatedBy:   userID,
        Nodes:       []byte(`[]`),
        Connections: []byte(`[]`),
    }
    err := db.Workflows().Create(context.Background(), workflow)
    require.NoError(t, err)
    return workflow
}
```

### Test Database

```go
// testutil/database.go
package testutil

import (
    "testing"
    "github.com/linkflow-ai/linkflow/internal/pkg/database"
)

func SetupTestDB(t *testing.T) *database.DB {
    db, err := database.Connect(database.Config{
        Host:     "localhost",
        Port:     5432,
        User:     "postgres",
        Password: "postgres",
        Database: "linkflow_test",
    })
    require.NoError(t, err)

    t.Cleanup(func() {
        // Clean up test data
        db.Exec("TRUNCATE users, workflows, executions CASCADE")
        db.Close()
    })

    return db
}
```

## Mocking

### Using testify/mock

```go
type MockEmailService struct {
    mock.Mock
}

func (m *MockEmailService) Send(to, subject, body string) error {
    args := m.Called(to, subject, body)
    return args.Error(0)
}

func TestSendWelcomeEmail(t *testing.T) {
    mockEmail := new(MockEmailService)
    mockEmail.On("Send", "user@example.com", mock.Anything, mock.Anything).Return(nil)

    service := NewUserService(mockEmail)
    err := service.RegisterUser(&User{Email: "user@example.com"})

    assert.NoError(t, err)
    mockEmail.AssertCalled(t, "Send", "user@example.com", mock.Anything, mock.Anything)
}
```

### HTTP Mocking

```go
func TestExternalAPICall(t *testing.T) {
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        w.Write([]byte(`{"status": "ok"}`))
    }))
    defer server.Close()

    client := NewAPIClient(server.URL)
    result, err := client.GetStatus()

    assert.NoError(t, err)
    assert.Equal(t, "ok", result.Status)
}
```

## Coverage

```bash
# Generate coverage report
make test-coverage

# View HTML report
open coverage.html

# Coverage for specific package
go test -coverprofile=coverage.out ./internal/api/handlers/...
go tool cover -html=coverage.out
```

### Coverage Targets

| Package | Target |
|---------|--------|
| handlers | 80% |
| nodes | 70% |
| pkg | 80% |
| domain | 90% |

## CI/CD Integration

```yaml
# .github/workflows/test.yml
name: Tests

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    
    services:
      postgres:
        image: postgres:16
        env:
          POSTGRES_PASSWORD: postgres
        ports:
          - 5432:5432
      redis:
        image: redis:7
        ports:
          - 6379:6379

    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.23'
      
      - name: Run tests
        run: make test-coverage
        env:
          DATABASE_HOST: localhost
          REDIS_HOST: localhost
      
      - name: Upload coverage
        uses: codecov/codecov-action@v4
        with:
          file: coverage.out
```

## Best Practices

1. **Test file naming**: `*_test.go` alongside source files
2. **Table-driven tests**: Use for testing multiple scenarios
3. **Test isolation**: Each test should be independent
4. **Clean up**: Use `t.Cleanup()` for resource cleanup
5. **Meaningful names**: Test names should describe the scenario
6. **Fast tests**: Unit tests should run quickly
7. **Skip slow tests**: Use `-short` flag for CI
