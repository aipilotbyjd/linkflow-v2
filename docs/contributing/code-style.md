# Code Style Guide

This guide documents the coding conventions used in LinkFlow.

## Go Code Style

### Formatting

All Go code must be formatted with `gofmt` and `goimports`:

```bash
# Format code
gofmt -w .
goimports -w .

# Or use the linter
make lint
```

### Import Organization

Imports should be grouped in the following order:

```go
import (
    // Standard library
    "context"
    "encoding/json"
    "fmt"
    "net/http"

    // Third-party packages
    "github.com/google/uuid"
    "github.com/rs/zerolog"

    // Internal packages
    "github.com/linkflow-ai/linkflow/internal/domain"
    "github.com/linkflow-ai/linkflow/internal/pkg/logger"
)
```

### Package Structure

```go
// Package declaration with comment
// Package handlers provides HTTP request handlers for the API.
package handlers

// Constants
const (
    MaxPageSize     = 100
    DefaultPageSize = 20
)

// Variables
var (
    ErrNotFound = errors.New("resource not found")
)

// Types (interfaces first, then structs)
type WorkflowService interface {
    Create(ctx context.Context, w *domain.Workflow) error
    Get(ctx context.Context, id uuid.UUID) (*domain.Workflow, error)
}

type WorkflowHandler struct {
    service WorkflowService
    logger  zerolog.Logger
}

// Constructor functions
func NewWorkflowHandler(svc WorkflowService, log zerolog.Logger) *WorkflowHandler {
    return &WorkflowHandler{
        service: svc,
        logger:  log,
    }
}

// Methods (public first, then private)
func (h *WorkflowHandler) Create(w http.ResponseWriter, r *http.Request) {
    // ...
}

func (h *WorkflowHandler) validateRequest(r *http.Request) error {
    // ...
}
```

### Naming Conventions

#### Packages
```go
// Good: lowercase, short, no underscores
package handlers
package workflow
package auth

// Bad
package WorkflowHandlers
package workflow_handlers
```

#### Variables
```go
// Good: camelCase, descriptive
userID := r.Context().Value("user_id")
workflowCount := len(workflows)

// Bad
UserID := ...    // Unexported should be camelCase
wc := ...        // Too abbreviated
```

#### Functions
```go
// Good: PascalCase for exported, camelCase for unexported
func CreateWorkflow() {}   // Exported
func validateInput() {}    // Unexported

// Method receivers: short, consistent
func (h *Handler) Create() {}    // Good
func (handler *Handler) Create() {}  // Avoid
```

#### Constants
```go
// Good: PascalCase for exported
const MaxRetries = 3
const DefaultTimeout = 30 * time.Second

// For unexported, either is acceptable
const maxRetries = 3
const MAX_RETRIES = 3
```

#### Interfaces
```go
// Good: -er suffix for single method interfaces
type Reader interface {
    Read(p []byte) (n int, err error)
}

type WorkflowService interface {
    Create(ctx context.Context, w *Workflow) error
    Get(ctx context.Context, id uuid.UUID) (*Workflow, error)
}
```

### Error Handling

#### Basic Pattern
```go
result, err := doSomething()
if err != nil {
    return err
}
```

#### Error Wrapping
```go
// Wrap errors with context
result, err := db.Query(ctx, query)
if err != nil {
    return fmt.Errorf("query workflows: %w", err)
}
```

#### Custom Errors
```go
// Define custom error types for specific cases
type NotFoundError struct {
    Resource string
    ID       uuid.UUID
}

func (e *NotFoundError) Error() string {
    return fmt.Sprintf("%s %s not found", e.Resource, e.ID)
}

// Check error type
var notFound *NotFoundError
if errors.As(err, &notFound) {
    // Handle not found
}
```

#### Sentinel Errors
```go
var (
    ErrNotFound     = errors.New("not found")
    ErrUnauthorized = errors.New("unauthorized")
    ErrInvalidInput = errors.New("invalid input")
)

// Check sentinel error
if errors.Is(err, ErrNotFound) {
    // Handle not found
}
```

### Context Usage

```go
// Always pass context as first parameter
func (s *Service) GetWorkflow(ctx context.Context, id uuid.UUID) (*Workflow, error) {
    // Use context for cancellation
    select {
    case <-ctx.Done():
        return nil, ctx.Err()
    default:
    }
    
    // Pass context to downstream calls
    return s.repo.Get(ctx, id)
}
```

### Structs

```go
// Field alignment (optional but recommended for large structs)
type Workflow struct {
    ID          uuid.UUID `json:"id" db:"id"`
    Name        string    `json:"name" db:"name"`
    Description string    `json:"description" db:"description"`
    Status      string    `json:"status" db:"status"`
    CreatedAt   time.Time `json:"created_at" db:"created_at"`
    UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// Use struct literals with field names
workflow := &Workflow{
    ID:     uuid.New(),
    Name:   "My Workflow",
    Status: "draft",
}
```

### Comments

```go
// Package comment (required for main packages)
// Package main provides the entry point for the API server.
package main

// Exported function comment (required)
// CreateWorkflow creates a new workflow in the database.
// It validates the input and returns the created workflow with its ID.
func CreateWorkflow(ctx context.Context, w *Workflow) (*Workflow, error) {
    // Implementation comments for complex logic
    // First, validate the workflow structure...
}

// TODO/FIXME comments
// TODO(username): Implement retry logic
// FIXME(username): This breaks when X happens
```

### Testing

```go
// Test function naming
func TestWorkflowHandler_Create(t *testing.T) {}
func TestWorkflowHandler_Create_InvalidInput(t *testing.T) {}

// Table-driven tests
func TestValidateEmail(t *testing.T) {
    tests := []struct {
        name    string
        email   string
        wantErr bool
    }{
        {"valid email", "user@example.com", false},
        {"missing @", "userexample.com", true},
        {"empty", "", true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := ValidateEmail(tt.email)
            if (err != nil) != tt.wantErr {
                t.Errorf("ValidateEmail(%q) error = %v, wantErr %v", tt.email, err, tt.wantErr)
            }
        })
    }
}
```

### HTTP Handlers

```go
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()

    // 1. Parse request
    var req CreateRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        h.respondError(w, http.StatusBadRequest, "invalid request body")
        return
    }

    // 2. Validate
    if err := req.Validate(); err != nil {
        h.respondError(w, http.StatusBadRequest, err.Error())
        return
    }

    // 3. Call service
    result, err := h.service.Create(ctx, &req)
    if err != nil {
        h.handleError(w, err)
        return
    }

    // 4. Respond
    h.respondJSON(w, http.StatusCreated, result)
}
```

### Logging

```go
// Use structured logging
log.Info().
    Str("workflow_id", workflow.ID.String()).
    Str("status", workflow.Status).
    Msg("workflow created")

// Include context
log.Error().
    Err(err).
    Str("request_id", requestID).
    Str("user_id", userID).
    Msg("failed to create workflow")

// Never log sensitive data
// Bad: log.Info().Str("api_key", apiKey)...
// Good: log.Info().Str("api_key", "***")...
```

## Linting

We use `golangci-lint` with the following configuration:

```yaml
# .golangci.yml
linters:
  enable:
    - errcheck
    - govet
    - ineffassign
    - staticcheck
    - unused
    - gosimple
    - structcheck
    - deadcode
    - goimports
    - misspell

linters-settings:
  govet:
    check-shadowing: true
  goimports:
    local-prefixes: github.com/linkflow-ai/linkflow
```

Run the linter:
```bash
make lint
```

## IDE Configuration

### VS Code

```json
{
  "go.formatTool": "goimports",
  "go.lintTool": "golangci-lint",
  "go.lintOnSave": "package",
  "[go]": {
    "editor.formatOnSave": true,
    "editor.codeActionsOnSave": {
      "source.organizeImports": true
    }
  }
}
```

### GoLand

1. Enable `goimports` formatter
2. Enable `golangci-lint` inspections
3. Enable "Optimize imports on the fly"
