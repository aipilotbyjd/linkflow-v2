# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Install dependencies
RUN apk add --no-cache git ca-certificates

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build binaries with version info
ARG VERSION=dev
ARG COMMIT=unknown
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w -X main.Version=${VERSION} -X main.Commit=${COMMIT}" -o /bin/api ./cmd/api
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w -X main.Version=${VERSION} -X main.Commit=${COMMIT}" -o /bin/worker ./cmd/worker
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w -X main.Version=${VERSION} -X main.Commit=${COMMIT}" -o /bin/scheduler ./cmd/scheduler

# API image
FROM alpine:3.19 AS api

LABEL org.opencontainers.image.source="https://github.com/linkflow-ai/linkflow-v2"
LABEL org.opencontainers.image.description="LinkFlow API Service"
LABEL org.opencontainers.image.licenses="MIT"

RUN apk add --no-cache ca-certificates tzdata curl \
    && addgroup -g 1000 linkflow \
    && adduser -u 1000 -G linkflow -s /bin/sh -D linkflow

COPY --from=builder /bin/api /bin/api
COPY --from=builder /app/configs /configs

RUN chown -R linkflow:linkflow /configs

USER linkflow

EXPOSE 8090

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8090/api/v1/health || exit 1

ENTRYPOINT ["/bin/api"]

# Worker image
FROM alpine:3.19 AS worker

LABEL org.opencontainers.image.source="https://github.com/linkflow-ai/linkflow-v2"
LABEL org.opencontainers.image.description="LinkFlow Worker Service"
LABEL org.opencontainers.image.licenses="MIT"

RUN apk add --no-cache ca-certificates tzdata \
    && addgroup -g 1000 linkflow \
    && adduser -u 1000 -G linkflow -s /bin/sh -D linkflow

COPY --from=builder /bin/worker /bin/worker
COPY --from=builder /app/configs /configs

RUN chown -R linkflow:linkflow /configs

USER linkflow

ENTRYPOINT ["/bin/worker"]

# Scheduler image
FROM alpine:3.19 AS scheduler

LABEL org.opencontainers.image.source="https://github.com/linkflow-ai/linkflow-v2"
LABEL org.opencontainers.image.description="LinkFlow Scheduler Service"
LABEL org.opencontainers.image.licenses="MIT"

RUN apk add --no-cache ca-certificates tzdata \
    && addgroup -g 1000 linkflow \
    && adduser -u 1000 -G linkflow -s /bin/sh -D linkflow

COPY --from=builder /bin/scheduler /bin/scheduler
COPY --from=builder /app/configs /configs

RUN chown -R linkflow:linkflow /configs

USER linkflow

ENTRYPOINT ["/bin/scheduler"]
