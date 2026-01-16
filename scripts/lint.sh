#!/bin/bash
# Run linters

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

FIX="${1:-}"

echo -e "${YELLOW}Running linters...${NC}"

# Check if golangci-lint is installed
if ! command -v golangci-lint &> /dev/null; then
    echo -e "${YELLOW}golangci-lint not found, installing...${NC}"
    go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
fi

# Run golangci-lint
if [ "${FIX}" == "fix" ] || [ "${FIX}" == "--fix" ]; then
    echo -e "${YELLOW}Running with auto-fix...${NC}"
    golangci-lint run --fix ./...
else
    golangci-lint run ./...
fi

echo -e "${GREEN}✓ Linting passed${NC}"

# Run go vet
echo -e "${YELLOW}Running go vet...${NC}"
go vet ./...
echo -e "${GREEN}✓ go vet passed${NC}"

# Check formatting
echo -e "${YELLOW}Checking formatting...${NC}"
UNFORMATTED=$(gofmt -l .)
if [ -n "${UNFORMATTED}" ]; then
    echo -e "${RED}The following files are not formatted:${NC}"
    echo "${UNFORMATTED}"
    if [ "${FIX}" == "fix" ] || [ "${FIX}" == "--fix" ]; then
        echo -e "${YELLOW}Formatting...${NC}"
        gofmt -w .
        echo -e "${GREEN}✓ Files formatted${NC}"
    else
        exit 1
    fi
else
    echo -e "${GREEN}✓ All files formatted${NC}"
fi

echo ""
echo -e "${GREEN}All checks passed!${NC}"
