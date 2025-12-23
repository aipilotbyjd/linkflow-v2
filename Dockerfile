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

# Build binaries
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/api ./cmd/api
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/worker ./cmd/worker
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/scheduler ./cmd/scheduler

# API image
FROM alpine:3.19 AS api
RUN apk add --no-cache ca-certificates tzdata
COPY --from=builder /bin/api /bin/api
COPY --from=builder /app/configs /configs
EXPOSE 8090
ENTRYPOINT ["/bin/api"]

# Worker image
FROM alpine:3.19 AS worker
RUN apk add --no-cache ca-certificates tzdata
COPY --from=builder /bin/worker /bin/worker
COPY --from=builder /app/configs /configs
ENTRYPOINT ["/bin/worker"]

# Scheduler image
FROM alpine:3.19 AS scheduler
RUN apk add --no-cache ca-certificates tzdata
COPY --from=builder /bin/scheduler /bin/scheduler
COPY --from=builder /app/configs /configs
ENTRYPOINT ["/bin/scheduler"]
