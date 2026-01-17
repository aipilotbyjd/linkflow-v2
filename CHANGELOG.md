# Changelog

All notable changes to LinkFlow will be documented in this file.

## [Unreleased]

### Added
- Clean architecture v2 implementation
- 22 repository implementations
- Handler types.go for shared types
- Analytics handlers and routes
- Config files for local/dev/staging/prod

### Changed
- Restructured handlers to one-file-per-handler pattern
- Reorganized config file naming

### Fixed
- IPv6 compatibility in SMTP

## [0.1.0] - 2025-01-17

### Added
- Initial project setup
- Core domain models (user, workspace, workflow, execution, credential, schedule, webhook)
- PostgreSQL repositories
- Redis caching
- JWT authentication
- Asynq worker queue
- WebSocket support
