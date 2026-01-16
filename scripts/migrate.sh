#!/bin/bash
# Database migration script

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

# Default values
MIGRATIONS_DIR="${MIGRATIONS_DIR:-migrations/postgres}"
DATABASE_URL="${DATABASE_URL:-postgres://postgres:postgres@localhost:5432/linkflow?sslmode=disable}"

# Parse arguments
ACTION="${1:-up}"
STEPS="${2:-}"

show_help() {
    echo "Usage: $0 [action] [steps]"
    echo ""
    echo "Actions:"
    echo "  up        Run all pending migrations (default)"
    echo "  down      Rollback the last migration"
    echo "  drop      Drop all tables"
    echo "  force N   Set migration version to N"
    echo "  version   Show current migration version"
    echo "  create    Create a new migration file"
    echo ""
    echo "Environment variables:"
    echo "  DATABASE_URL    Database connection string"
    echo "  MIGRATIONS_DIR  Path to migrations directory"
    echo ""
    echo "Examples:"
    echo "  $0 up              # Run all pending migrations"
    echo "  $0 down 1          # Rollback 1 migration"
    echo "  $0 create add_users_table"
}

# Check if golang-migrate is installed
if ! command -v migrate &> /dev/null; then
    echo -e "${RED}Error: golang-migrate is not installed${NC}"
    echo "Install it with: go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest"
    exit 1
fi

case "${ACTION}" in
    up)
        echo -e "${YELLOW}Running migrations...${NC}"
        migrate -path "${MIGRATIONS_DIR}" -database "${DATABASE_URL}" up ${STEPS}
        echo -e "${GREEN}✓ Migrations completed${NC}"
        ;;
    down)
        STEPS="${STEPS:-1}"
        echo -e "${YELLOW}Rolling back ${STEPS} migration(s)...${NC}"
        migrate -path "${MIGRATIONS_DIR}" -database "${DATABASE_URL}" down ${STEPS}
        echo -e "${GREEN}✓ Rollback completed${NC}"
        ;;
    drop)
        echo -e "${RED}WARNING: This will drop all tables!${NC}"
        read -p "Are you sure? (y/N) " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            migrate -path "${MIGRATIONS_DIR}" -database "${DATABASE_URL}" drop -f
            echo -e "${GREEN}✓ All tables dropped${NC}"
        else
            echo "Cancelled"
        fi
        ;;
    force)
        if [ -z "${STEPS}" ]; then
            echo -e "${RED}Error: Version number required${NC}"
            exit 1
        fi
        echo -e "${YELLOW}Forcing version to ${STEPS}...${NC}"
        migrate -path "${MIGRATIONS_DIR}" -database "${DATABASE_URL}" force ${STEPS}
        echo -e "${GREEN}✓ Version set to ${STEPS}${NC}"
        ;;
    version)
        migrate -path "${MIGRATIONS_DIR}" -database "${DATABASE_URL}" version
        ;;
    create)
        if [ -z "${STEPS}" ]; then
            echo -e "${RED}Error: Migration name required${NC}"
            echo "Usage: $0 create <migration_name>"
            exit 1
        fi
        echo -e "${YELLOW}Creating migration: ${STEPS}${NC}"
        migrate create -ext sql -dir "${MIGRATIONS_DIR}" -seq "${STEPS}"
        echo -e "${GREEN}✓ Migration files created${NC}"
        ;;
    help|--help|-h)
        show_help
        ;;
    *)
        echo -e "${RED}Unknown action: ${ACTION}${NC}"
        show_help
        exit 1
        ;;
esac
