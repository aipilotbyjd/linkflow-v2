# Developer Guide

This guide will help you set up your development environment and understand the codebase.

## Prerequisites

- **Go 1.23+** - [Download Go](https://golang.org/dl/)
- **Docker & Docker Compose** - [Install Docker](https://docs.docker.com/get-docker/)
- **Make** - Usually pre-installed on macOS/Linux
- **Git** - [Install Git](https://git-scm.com/downloads)

### Recommended Tools

- **GoLand** or **VS Code** with Go extension
- **TablePlus** or **pgAdmin** for database management
- **Postman** for API testing

## Quick Setup

```bash
# Clone the repository
git clone https://github.com/linkflow-ai/linkflow.git
cd linkflow

# Copy environment file
cp .env.example .env

# Install dependencies
make deps

# Start infrastructure (PostgreSQL, Redis, MinIO)
make dev-deps

# Run API server with hot reload
make dev
```

The API will be available at `http://localhost:8090`.

## Project Structure

```
linkflow/
├── cmd/                          # Application entry points
│   ├── api/main.go              # API server
│   ├── worker/main.go           # Background worker
│   └── scheduler/main.go        # Job scheduler
│
├── internal/                     # Private application code
│   ├── api/                     # HTTP layer
│   │   ├── handlers/           # Request handlers
│   │   ├── middleware/         # HTTP middleware
│   │   ├── routes.go           # Route definitions
│   │   └── server.go           # Server setup
│   │
│   ├── domain/                  # Business entities
│   │   ├── user.go
│   │   ├── workflow.go
│   │   └── execution.go
│   │
│   ├── pkg/                     # Shared packages
│   │   ├── config/             # Configuration loading
│   │   ├── database/           # Database connection
│   │   ├── queue/              # Redis queue client
│   │   └── logger/             # Logging
│   │
│   ├── worker/                  # Worker implementation
│   │   ├── executor/           # Workflow executor
│   │   ├── nodes/              # Node implementations
│   │   └── processor/          # Job processor
│   │
│   └── scheduler/               # Scheduler implementation
│       ├── jobs/               # Scheduled jobs
│       └── leader.go           # Leader election
│
├── configs/                      # Configuration files
│   └── config.yaml              # Default configuration
│
├── docs/                         # Documentation
├── Dockerfile                    # Multi-stage Docker build
├── docker-compose.yaml           # Development composition
├── Makefile                      # Build commands
└── go.mod                        # Go module definition
```

## Running Services

### Development Mode (Hot Reload)

```bash
# Install air for hot reload
go install github.com/air-verse/air@latest

# Run API with hot reload
make dev
```

### Running Individual Services

```bash
# API Server
make run-api

# Worker (in separate terminal)
make run-worker

# Scheduler (in separate terminal)
make run-scheduler
```

### Using Docker

```bash
# Start all services
docker-compose up -d

# View logs
docker-compose logs -f api

# Stop services
docker-compose down
```

## Configuration

Configuration can be set via:
1. Environment variables (highest priority)
2. `configs/config.yaml` file
3. `.env` file (for development)

See [Configuration Guide](configuration.md) for complete reference.

### Key Configuration

```yaml
# configs/config.yaml
server:
  port: 8090

database:
  host: localhost
  port: 5432
  name: linkflow

redis:
  host: localhost
  port: 6379
```

## Database

### Connecting to PostgreSQL

```bash
# Via docker
docker exec -it linkflow-postgres psql -U postgres -d linkflow

# Using connection string
psql "postgresql://postgres:postgres@localhost:5432/linkflow"
```

### Migrations

Migrations are embedded in the application and run automatically on startup.

## Testing

```bash
# Run all tests
make test

# Run with coverage
make test-coverage

# Run specific tests
go test -v ./internal/api/handlers/...
```

### Writing Tests

```go
func TestHandler_CreateWorkflow(t *testing.T) {
    // Setup
    app := setupTestApp(t)
    
    // Test
    req := httptest.NewRequest("POST", "/api/v1/workflows", body)
    resp := httptest.NewRecorder()
    app.ServeHTTP(resp, req)
    
    // Assert
    assert.Equal(t, http.StatusCreated, resp.Code)
}
```

## Code Style

- Follow standard Go conventions
- Use `gofmt` and `goimports`
- Run `make lint` before committing

See [Code Style Guide](../contributing/code-style.md) for details.

## Common Tasks

### Adding a New API Endpoint

1. Define handler in `internal/api/handlers/`
2. Add route in `internal/api/routes.go`
3. Add tests

### Adding a New Node

See [Node Development Guide](node-development.md).

### Debugging

```bash
# Enable debug logging
APP_DEBUG=true make run-api

# View structured logs
make run-api 2>&1 | jq .
```

## Make Commands Reference

| Command | Description |
|---------|-------------|
| `make build` | Build all binaries |
| `make test` | Run tests |
| `make dev` | Run API with hot reload |
| `make dev-deps` | Start PostgreSQL, Redis, MinIO |
| `make docker-up` | Start all Docker services |
| `make lint` | Run linter |
| `make clean` | Remove build artifacts |

## Troubleshooting

### Database Connection Issues

```bash
# Check if PostgreSQL is running
docker ps | grep postgres

# Test connection
pg_isready -h localhost -p 5432
```

### Redis Connection Issues

```bash
# Check if Redis is running
docker ps | grep redis

# Test connection
redis-cli ping
```

### Port Already in Use

```bash
# Find process using port 8090
lsof -i :8090

# Kill process
kill -9 <PID>
```

## Further Reading

- [Configuration Guide](configuration.md)
- [Node Development](node-development.md)
- [Testing Guide](testing.md)
- [Architecture Overview](../architecture/README.md)
