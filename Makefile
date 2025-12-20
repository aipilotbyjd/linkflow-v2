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
docker-build:
	docker-compose build

docker-up:
	docker-compose up -d

docker-down:
	docker-compose down

docker-logs:
	docker-compose logs -f

# Development
dev-deps:
	docker-compose up -d postgres redis minio

dev-api: dev-deps
	$(GOCMD) run ./cmd/api

dev-worker: dev-deps
	$(GOCMD) run ./cmd/worker

dev-scheduler: dev-deps
	$(GOCMD) run ./cmd/scheduler

# Database
migrate:
	$(GOCMD) run ./cmd/api migrate

# Linting
lint:
	golangci-lint run ./...

# Generate
generate:
	$(GOCMD) generate ./...

# Help
help:
	@echo "Available targets:"
	@echo "  build          - Build all binaries"
	@echo "  build-api      - Build API service"
	@echo "  build-worker   - Build Worker service"
	@echo "  build-scheduler- Build Scheduler service"
	@echo "  run-api        - Run API service"
	@echo "  run-worker     - Run Worker service"
	@echo "  run-scheduler  - Run Scheduler service"
	@echo "  test           - Run tests"
	@echo "  test-coverage  - Run tests with coverage"
	@echo "  deps           - Download dependencies"
	@echo "  clean          - Clean build artifacts"
	@echo "  docker-build   - Build Docker images"
	@echo "  docker-up      - Start Docker containers"
	@echo "  docker-down    - Stop Docker containers"
	@echo "  dev-deps       - Start development dependencies"
	@echo "  dev-api        - Run API in development mode"
	@echo "  lint           - Run linter"
