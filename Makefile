.PHONY: all build run test clean docker

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOMOD=$(GOCMD) mod
BINARY_API=bin/api
BINARY_WORKER=bin/worker
BINARY_SCHEDULER=bin/scheduler

all: build

# Build all binaries
build:
	$(GOBUILD) -o $(BINARY_API) ./cmd/api
	$(GOBUILD) -o $(BINARY_WORKER) ./cmd/worker
	$(GOBUILD) -o $(BINARY_SCHEDULER) ./cmd/scheduler

# Build individual services
build-api:
	$(GOBUILD) -o $(BINARY_API) ./cmd/api

build-worker:
	$(GOBUILD) -o $(BINARY_WORKER) ./cmd/worker

build-scheduler:
	$(GOBUILD) -o $(BINARY_SCHEDULER) ./cmd/scheduler

# Run services
run-api:
	$(GOCMD) run ./cmd/api

# Run with hot reload (requires air: go install github.com/air-verse/air@latest)
dev:
	air

run-worker:
	$(GOCMD) run ./cmd/worker

run-scheduler:
	$(GOCMD) run ./cmd/scheduler

# Test
test:
	$(GOTEST) -v ./...

test-coverage:
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html

# Dependencies
deps:
	$(GOMOD) download
	$(GOMOD) tidy

# Clean
clean:
	rm -rf bin/
	rm -f coverage.out coverage.html

# Docker
COMPOSE_DEV=deploy/docker-compose.dev.yml
COMPOSE_PROD=deploy/docker-compose.yml

# Development Docker (simple setup)
docker-dev-up:
	docker compose -f $(COMPOSE_DEV) up -d

docker-dev-down:
	docker compose -f $(COMPOSE_DEV) down

docker-dev-logs:
	docker compose -f $(COMPOSE_DEV) logs -f

docker-dev-ps:
	docker compose -f $(COMPOSE_DEV) ps

# Production Docker
docker-prod-build:
	docker compose -f $(COMPOSE_PROD) build

docker-prod-up:
	docker compose -f $(COMPOSE_PROD) up -d

docker-prod-down:
	docker compose -f $(COMPOSE_PROD) down

docker-prod-logs:
	docker compose -f $(COMPOSE_PROD) logs -f

# Development (infra only - run services locally)
dev-infra:
	docker compose -f $(COMPOSE_DEV) up -d postgres redis

dev-api: dev-infra
	$(GOCMD) run ./cmd/api

dev-worker: dev-infra
	$(GOCMD) run ./cmd/worker

dev-scheduler: dev-infra
	$(GOCMD) run ./cmd/scheduler

dev-all: dev-infra
	@echo "Starting all services..."
	$(GOCMD) run ./cmd/api &
	$(GOCMD) run ./cmd/worker &
	$(GOCMD) run ./cmd/scheduler &

dev-stop:
	docker compose -f $(COMPOSE_DEV) down
	pkill -f "go run ./cmd" || true

# Database
migrate:
	$(GOCMD) run ./cmd/api migrate

# Seeding
seed:
	$(GOCMD) run ./cmd/seed

seed-clean:
	$(GOCMD) run ./cmd/seed -clean=true

seed-fresh: dev-infra
	$(GOCMD) run ./cmd/seed -clean=true

# Linting
lint:
	golangci-lint run ./...

# Generate
generate:
	$(GOCMD) generate ./...

# Help
help:
	@echo "LinkFlow Development Commands"
	@echo "=============================="
	@echo ""
	@echo "Quick Start:"
	@echo "  make dev-api        - Start Postgres/Redis + run API locally"
	@echo "  make docker-dev-up  - Start full stack in Docker (dev mode)"
	@echo ""
	@echo "Build:"
	@echo "  make build          - Build all binaries"
	@echo "  make build-api      - Build API service only"
	@echo "  make generate       - Generate wire dependencies"
	@echo ""
	@echo "Development (run locally, infra in Docker):"
	@echo "  make dev-infra      - Start Postgres + Redis only"
	@echo "  make dev-api        - Start infra + run API"
	@echo "  make dev-worker     - Start infra + run Worker"
	@echo "  make dev-scheduler  - Start infra + run Scheduler"
	@echo "  make dev-stop       - Stop everything"
	@echo ""
	@echo "Docker Development:"
	@echo "  make docker-dev-up   - Start full stack (dev)"
	@echo "  make docker-dev-down - Stop full stack (dev)"
	@echo "  make docker-dev-logs - View logs"
	@echo "  make docker-dev-ps   - Show status"
	@echo ""
	@echo "Docker Production:"
	@echo "  make docker-prod-build - Build images"
	@echo "  make docker-prod-up    - Start stack"
	@echo "  make docker-prod-down  - Stop stack"
	@echo ""
	@echo "Testing:"
	@echo "  make test           - Run tests"
	@echo "  make test-coverage  - Run with coverage"
	@echo "  make lint           - Run linter"
	@echo ""
	@echo "Seeding:"
	@echo "  make seed           - Seed development data"
	@echo "  make seed-clean     - Clean and re-seed dev data"
	@echo "  make seed-fresh     - Start infra + clean seed"
