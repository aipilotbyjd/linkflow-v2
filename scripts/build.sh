#!/bin/bash
# Build all LinkFlow services

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Build output directory
BUILD_DIR="${BUILD_DIR:-bin}"
VERSION="${VERSION:-$(git describe --tags --always --dirty 2>/dev/null || echo 'dev')}"
BUILD_TIME=$(date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS="-s -w -X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME}"

echo -e "${YELLOW}Building LinkFlow services...${NC}"
echo "Version: ${VERSION}"
echo "Build time: ${BUILD_TIME}"
echo ""

# Create build directory
mkdir -p "${BUILD_DIR}"

# Build API
echo -e "${YELLOW}Building API...${NC}"
go build -ldflags "${LDFLAGS}" -o "${BUILD_DIR}/api" ./cmd/api
echo -e "${GREEN}✓ API built successfully${NC}"

# Build Worker
echo -e "${YELLOW}Building Worker...${NC}"
go build -ldflags "${LDFLAGS}" -o "${BUILD_DIR}/worker" ./cmd/worker
echo -e "${GREEN}✓ Worker built successfully${NC}"

# Build Scheduler
echo -e "${YELLOW}Building Scheduler...${NC}"
go build -ldflags "${LDFLAGS}" -o "${BUILD_DIR}/scheduler" ./cmd/scheduler
echo -e "${GREEN}✓ Scheduler built successfully${NC}"

# Build Migrate
echo -e "${YELLOW}Building Migrate...${NC}"
go build -ldflags "${LDFLAGS}" -o "${BUILD_DIR}/migrate" ./cmd/migrate
echo -e "${GREEN}✓ Migrate built successfully${NC}"

echo ""
echo -e "${GREEN}All services built successfully!${NC}"
echo "Binaries are in ${BUILD_DIR}/"
ls -la "${BUILD_DIR}/"
