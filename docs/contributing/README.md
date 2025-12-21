# Contributing to LinkFlow

Thank you for your interest in contributing to LinkFlow! This document provides guidelines and instructions for contributing.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Workflow](#development-workflow)
- [Pull Request Process](#pull-request-process)
- [Coding Standards](#coding-standards)
- [Testing Guidelines](#testing-guidelines)
- [Documentation](#documentation)

## Code of Conduct

By participating in this project, you agree to abide by our Code of Conduct:

- Be respectful and inclusive
- Welcome newcomers and help them learn
- Focus on constructive feedback
- Accept responsibility for mistakes

## Getting Started

### Prerequisites

- Go 1.23+
- Docker & Docker Compose
- Git

### Setup

```bash
# Fork the repository on GitHub
# Clone your fork
git clone https://github.com/YOUR_USERNAME/linkflow.git
cd linkflow

# Add upstream remote
git remote add upstream https://github.com/linkflow-ai/linkflow.git

# Install dependencies
make deps

# Start development environment
make dev-deps
make dev
```

### Running Tests

```bash
# Run all tests
make test

# Run with coverage
make test-coverage

# Run linter
make lint
```

## Development Workflow

### 1. Create a Branch

```bash
# Update main branch
git checkout main
git pull upstream main

# Create feature branch
git checkout -b feature/your-feature-name
# or
git checkout -b fix/your-bug-fix
```

### Branch Naming

- `feature/` - New features
- `fix/` - Bug fixes
- `docs/` - Documentation changes
- `refactor/` - Code refactoring
- `test/` - Test additions/changes

### 2. Make Changes

- Write clean, readable code
- Follow existing patterns
- Add tests for new functionality
- Update documentation if needed

### 3. Commit Changes

Write clear, concise commit messages:

```bash
# Format
<type>(<scope>): <description>

# Examples
feat(api): add workflow versioning endpoint
fix(worker): handle timeout in HTTP node
docs(readme): update installation instructions
refactor(scheduler): simplify leader election
test(handlers): add tests for credential handler
```

Types:
- `feat` - New feature
- `fix` - Bug fix
- `docs` - Documentation
- `refactor` - Code refactoring
- `test` - Tests
- `chore` - Maintenance

### 4. Push Changes

```bash
git push origin feature/your-feature-name
```

## Pull Request Process

### Before Submitting

- [ ] Tests pass locally (`make test`)
- [ ] Linter passes (`make lint`)
- [ ] Documentation updated (if needed)
- [ ] Commit messages follow convention
- [ ] Branch is up to date with main

### Submitting a PR

1. Go to GitHub and create a Pull Request
2. Fill out the PR template
3. Link any related issues
4. Request review from maintainers

### PR Template

```markdown
## Description
Brief description of changes

## Type of Change
- [ ] Bug fix
- [ ] New feature
- [ ] Breaking change
- [ ] Documentation update

## Testing
How was this tested?

## Checklist
- [ ] Tests added/updated
- [ ] Documentation updated
- [ ] No breaking changes (or documented)
```

### Review Process

1. Maintainers will review your PR
2. Address any feedback
3. Once approved, a maintainer will merge

## Coding Standards

### Go Code Style

- Follow [Effective Go](https://golang.org/doc/effective_go)
- Use `gofmt` for formatting
- Use `goimports` for imports
- Run `golangci-lint` before committing

### File Organization

```go
package handlers

import (
    // Standard library
    "context"
    "net/http"

    // External packages
    "github.com/google/uuid"

    // Internal packages
    "github.com/linkflow-ai/linkflow/internal/domain"
)
```

### Naming Conventions

```go
// Package names: lowercase, single word
package handlers

// Exported functions: PascalCase
func CreateWorkflow(w http.ResponseWriter, r *http.Request) {}

// Unexported functions: camelCase
func validateRequest(r *http.Request) error {}

// Constants: PascalCase or UPPER_SNAKE_CASE
const MaxRetries = 3
const DEFAULT_TIMEOUT = 30

// Variables: camelCase
var workflowService *WorkflowService
```

### Error Handling

```go
// Always check errors
result, err := doSomething()
if err != nil {
    return fmt.Errorf("failed to do something: %w", err)
}

// Use custom error types when appropriate
type NotFoundError struct {
    Resource string
    ID       string
}

func (e *NotFoundError) Error() string {
    return fmt.Sprintf("%s with ID %s not found", e.Resource, e.ID)
}
```

### Comments

```go
// Package handlers provides HTTP request handlers.
package handlers

// CreateWorkflow creates a new workflow.
// It validates the request body and persists the workflow to the database.
func CreateWorkflow(w http.ResponseWriter, r *http.Request) {
    // ...
}
```

See [Code Style Guide](code-style.md) for more details.

## Testing Guidelines

### Test File Location

```
internal/api/handlers/
├── workflow.go
└── workflow_test.go  # Tests go alongside source
```

### Writing Tests

```go
func TestCreateWorkflow(t *testing.T) {
    tests := []struct {
        name       string
        input      CreateWorkflowRequest
        wantStatus int
        wantErr    bool
    }{
        {
            name:       "valid request",
            input:      CreateWorkflowRequest{Name: "Test"},
            wantStatus: http.StatusCreated,
        },
        {
            name:       "missing name",
            input:      CreateWorkflowRequest{},
            wantStatus: http.StatusBadRequest,
            wantErr:    true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

### Coverage Requirements

- Aim for 80%+ coverage on new code
- Critical paths should have 90%+ coverage
- All bug fixes should include regression tests

## Documentation

### Code Documentation

- All exported functions should have comments
- Complex logic should be explained
- Update README if behavior changes

### Doc Updates

Update documentation when:
- Adding new features
- Changing API behavior
- Modifying configuration
- Fixing documentation bugs

## Questions?

- Open a GitHub Issue
- Join our Discord server
- Email: support@linkflow.ai

## Recognition

Contributors are recognized in:
- CONTRIBUTORS.md file
- Release notes
- Project documentation

Thank you for contributing!
