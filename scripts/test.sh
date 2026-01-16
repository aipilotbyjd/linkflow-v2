#!/bin/bash
# Run tests

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

# Parse arguments
TEST_TYPE="${1:-unit}"
COVERAGE="${2:-}"

show_help() {
    echo "Usage: $0 [type] [options]"
    echo ""
    echo "Types:"
    echo "  unit          Run unit tests (default)"
    echo "  integration   Run integration tests"
    echo "  e2e           Run end-to-end tests"
    echo "  all           Run all tests"
    echo ""
    echo "Options:"
    echo "  coverage      Generate coverage report"
    echo ""
    echo "Examples:"
    echo "  $0 unit"
    echo "  $0 unit coverage"
    echo "  $0 integration"
    echo "  $0 all coverage"
}

run_unit_tests() {
    echo -e "${YELLOW}Running unit tests...${NC}"
    if [ "${COVERAGE}" == "coverage" ]; then
        go test -v -race -coverprofile=coverage.out ./internal/...
        go tool cover -html=coverage.out -o coverage.html
        echo -e "${GREEN}Coverage report generated: coverage.html${NC}"
    else
        go test -v -race ./internal/...
    fi
    echo -e "${GREEN}✓ Unit tests passed${NC}"
}

run_integration_tests() {
    echo -e "${YELLOW}Running integration tests...${NC}"
    go test -v -race -tags=integration ./test/integration/...
    echo -e "${GREEN}✓ Integration tests passed${NC}"
}

run_e2e_tests() {
    echo -e "${YELLOW}Running e2e tests...${NC}"
    go test -v -race -tags=e2e ./test/e2e/...
    echo -e "${GREEN}✓ E2E tests passed${NC}"
}

case "${TEST_TYPE}" in
    unit)
        run_unit_tests
        ;;
    integration)
        run_integration_tests
        ;;
    e2e)
        run_e2e_tests
        ;;
    all)
        run_unit_tests
        run_integration_tests
        run_e2e_tests
        ;;
    help|--help|-h)
        show_help
        ;;
    *)
        echo -e "${RED}Unknown test type: ${TEST_TYPE}${NC}"
        show_help
        exit 1
        ;;
esac
