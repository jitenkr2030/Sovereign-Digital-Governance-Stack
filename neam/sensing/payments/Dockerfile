# Payment Adapter Dockerfile
# Multi-stage build for optimized production deployment

# =============================================================================
# Build Stage
# =============================================================================
FROM golang:1.21-alpine AS builder

# Install git for go mod download and build dependencies
RUN apk add --no-cache git make ca-certificates

# Set working directory
WORKDIR /app

# Copy go.mod and go.sum first for dependency caching
COPY sensing/payments/go.mod sensing/payments/go.sum ./

# Download dependencies
RUN go mod download && go mod verify

# Copy source code
COPY sensing/payments/ ./sensing/payments/

# Build the binary with optimizations
RUN cd sensing/payments && \
    CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o payment-adapter main.go

# =============================================================================
# Runtime Stage - Minimal production image
# =============================================================================
FROM alpine:3.19 AS production

# Install minimal runtime dependencies
RUN apk add --no-cache ca-certificates tzdata curl wget redis redis-tools && \
    rm -rf /var/cache/apk/*

# Create non-root user for security
RUN addgroup -g 1000 neam && \
    adduser -u 1000 -G neam -s /bin/sh -D neam && \
    mkdir -p /app/data /app/logs && \
    chown -R neam:neam /app

# Set working directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/sensing/payments/payment-adapter /app/payment-adapter
COPY --from=builder /app/sensing/payments/config.yaml /app/config.yaml

# Create symlink for config
RUN ln -sf /app/config.yaml /app/sensing/payments/config.yaml 2>/dev/null || true

# Switch to non-root user
USER neam

# Expose service ports
EXPOSE 8084 9090

# Health checks
HEALTHCHECK --interval=30s --timeout=10s --start-period=10s --retries=3 \
    CMD wget --spider -q http://localhost:8084/health || exit 1

# Metrics health check
METRICS_HEALTHCHECK --interval=30s --timeout=5s --retries=1 \
    CMD wget --spider -q http://localhost:9090/metrics || exit 1

# Environment configuration
ENV SERVICE_NAME=payment-adapter
ENV SERVICE_PORT=8084
ENV METRICS_PORT=9090
ENV CONFIG_PATH=/app/config.yaml

# Set timezone
ENV TZ=UTC

# Entrypoint
ENTRYPOINT ["/app/payment-adapter"]

# Default command (can be overridden)
CMD ["--config", "/app/config.yaml"]

# =============================================================================
# Development Stage - For local testing and debugging
# =============================================================================
FROM golang:1.21-alpine AS development

RUN apk add --no-cache git make delv

WORKDIR /app

COPY sensing/payments/ ./sensing/payments/

RUN cd sensing/payments && \
    go mod download && \
    go install github.com/go-delve/delve/cmd/dlv@latest

EXPOSE 8084 9090 40000

ENV DELVE_PORT=40000
ENV CONFIG_PATH=/app/config.yaml

# Development entrypoint with debug support
ENTRYPOINT ["dlv", "debug", "--accept-multiclient", "--listen", ":40000", "--headless", "--", "--config", "/app/config.yaml"]

# =============================================================================
# Build Instructions
# =============================================================================
#
# Production build:
#   docker build -t neam/payment-adapter:latest -f sensing/payments/Dockerfile .
#
# Development build:
#   docker build -t neam/payment-adapter:dev --target development -f sensing/payments/Dockerfile .
#
# Multi-platform build:
#   docker buildx build --platform linux/amd64,linux/arm64 -t neam/payment-adapter:latest -f sensing/payments/Dockerfile --push .
#
# Run production container:
#   docker run -d --name payment-adapter \
#     -p 8084:8084 -p 9090:9090 \
#     -v $(pwd)/sensing/payments/config.yaml:/app/config.yaml \
#     neam/payment-adapter:latest
#
# Run development container with debug:
#   docker run -d --name payment-adapter-dev \
#     -p 8084:8084 -p 9090:9090 -p 40000:40000 \
#     -v $(pwd)/sensing/payments/config.yaml:/app/config.yaml \
#     neam/payment-adapter:dev
#
# =============================================================================
