# ============================================================================
# LinkFlow Multi-Target Dockerfile
# Build: docker build --build-arg SERVICE=api -t linkflow-api .
# ============================================================================

# ============ BUILD STAGE ============
FROM golang:1.23-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git ca-certificates tzdata

# Download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Build arguments
ARG SERVICE=api
ARG VERSION=dev
ARG COMMIT=unknown

# Build binary
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w -X main.Version=${VERSION} -X main.Commit=${COMMIT}" \
    -o /bin/app ./cmd/${SERVICE}

# ============ RUNTIME ============
FROM alpine:3.19

LABEL org.opencontainers.image.source="https://github.com/linkflow-ai/linkflow-v2"
LABEL org.opencontainers.image.description="LinkFlow Service"

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata curl \
    && addgroup -g 1000 linkflow \
    && adduser -u 1000 -G linkflow -s /bin/sh -D linkflow

WORKDIR /app

# Copy binary
COPY --from=builder /bin/app /bin/app

# Copy configs (if needed)
COPY --from=builder /app/configs ./configs

# Set ownership
RUN chown -R linkflow:linkflow /app

USER linkflow

# Default port (API)
EXPOSE 8090

# Health check
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD curl -sf http://localhost:8090/api/v1/health || exit 1

ENTRYPOINT ["/bin/app"]
