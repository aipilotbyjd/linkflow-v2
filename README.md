# LinkFlow

> A production-ready workflow automation platform built with Go

[![Go Version](https://img.shields.io/badge/Go-1.23-blue.svg)](https://golang.org/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

## Overview

LinkFlow is a workflow automation platform similar to n8n, designed to help you automate tasks by connecting various services and executing custom logic. Built with a three-tier architecture (API, Worker, Scheduler), it's designed for horizontal scalability and high availability.

## Features

- **Visual Workflow Builder** - Create workflows with a drag-and-drop interface
- **50+ Integrations** - Connect to Slack, GitHub, OpenAI, databases, and more
- **Real-time Execution** - Monitor workflow execution via WebSocket
- **Schedule & Webhooks** - Trigger workflows on schedule or via webhooks
- **Multi-tenant** - Support for workspaces and team collaboration
- **Secure Credentials** - AES-256-GCM encrypted credential storage
- **Expression Engine** - Powerful expression evaluation with 50+ functions
- **Horizontal Scaling** - Scale workers independently based on load

## Architecture

```
┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
│   API Service   │     │  Worker Service │     │   Scheduler     │
│   (Stateless)   │────▶│   (Scalable)    │◀────│  (Leader/HA)    │
└────────┬────────┘     └────────┬────────┘     └────────┬────────┘
         │                       │                       │
         ▼                       ▼                       ▼
┌─────────────────────────────────────────────────────────────────┐
│                  PostgreSQL  │  Redis  │  MinIO                 │
└─────────────────────────────────────────────────────────────────┘
```

## Quick Start

### Prerequisites

- Go 1.23+
- Docker & Docker Compose
- Make

### Development Setup

```bash
# Clone the repository
git clone https://github.com/linkflow-ai/linkflow.git
cd linkflow

# Copy environment file
cp .env.example .env

# Start dependencies (PostgreSQL, Redis, MinIO)
make dev-deps

# Run the API server with hot reload
make dev

# In separate terminals, run worker and scheduler
make run-worker
make run-scheduler
```

### Using Docker

```bash
# Build and start all services
docker-compose up -d

# View logs
docker-compose logs -f

# Stop services
docker-compose down
```

## Project Structure

```
linkflow/
├── cmd/                    # Application entry points
│   ├── api/               # API server
│   ├── worker/            # Background worker
│   └── scheduler/         # Job scheduler
├── internal/              # Private application code
│   ├── api/              # HTTP handlers, middleware
│   ├── domain/           # Business entities
│   ├── pkg/              # Shared packages
│   ├── worker/           # Worker implementation
│   └── scheduler/        # Scheduler implementation
├── configs/               # Configuration files
├── docs/                  # Documentation
│   ├── architecture/     # Architecture docs
│   ├── development/      # Development guides
│   ├── deployment/       # Deployment guides
│   ├── api/              # API documentation
│   ├── operations/       # Operations guides
│   └── guides/           # How-to guides
└── docker-compose.yaml    # Docker composition
```

## Documentation

| Document | Description |
|----------|-------------|
| [Architecture Overview](docs/architecture/README.md) | System architecture and design |
| [Developer Guide](docs/development/README.md) | Setup and development workflow |
| [API Reference](docs/api/README.md) | REST API documentation |
| [Deployment Guide](docs/deployment/README.md) | Production deployment |
| [Operations Guide](docs/operations/README.md) | Monitoring and maintenance |
| [Contributing](docs/contributing/README.md) | Contribution guidelines |

## Configuration

LinkFlow can be configured via environment variables or `configs/config.yaml`. See [Configuration Guide](docs/development/configuration.md) for details.

### Key Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `DATABASE_HOST` | PostgreSQL host | `localhost` |
| `REDIS_HOST` | Redis host | `localhost` |
| `JWT_SECRET` | JWT signing secret | Required |
| `SERVER_PORT` | API server port | `8090` |

## Make Commands

```bash
make build          # Build all binaries
make test           # Run tests
make dev            # Run API with hot reload
make docker-up      # Start Docker services
make lint           # Run linter
make help           # Show all commands
```

## Contributing

We welcome contributions! Please see our [Contributing Guide](docs/contributing/README.md) for details.

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Submit a pull request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Support

- **Documentation**: [docs/](docs/)
- **Issues**: [GitHub Issues](https://github.com/linkflow-ai/linkflow/issues)
- **Email**: support@linkflow.ai
