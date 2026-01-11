# =============================================================================
# Macro-Economic Modeling Service - Dockerfile
# =============================================================================
# Purpose: Containerizes the Macro-Economic Models service for GDP forecasting,
# inflation prediction, and economic simulation scenarios.
# =============================================================================

# Build Stage
FROM golang:1.21-alpine AS builder

# Install git for go mod download and build dependencies
RUN apk add --no-cache git make

# Set working directory
WORKDIR /app

# Copy go.mod and go.sum first for dependency caching
COPY macro/go.mod macro/go.sum ./

# Download dependencies
RUN cd macro && go mod download && go mod verify

# Copy source code
COPY macro/ ./macro/

# Build the binary with optimizations
RUN cd macro && CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o neam-macro main.go

# =============================================================================
# Production Stage - Minimal runtime image
# =============================================================================
FROM alpine:3.19 AS production

# Install minimal runtime dependencies
RUN apk add --no-cache ca-certificates tzdata curl

# Create non-root user for security
RUN addgroup -g 1002 neam && \
    adduser -u 1002 -G neam -s /bin/sh -D neam

# Set working directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/macro/neam-macro /app/neam-macro

# Create required directories
RUN mkdir -p /app/config /app/logs /app/data/models && \
    chown -R neam:neam /app

# Switch to non-root user
USER neam

# Expose service port
EXPOSE 8087

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget --spider -q http://localhost:8087/health || exit 1

# Environment configuration
ENV SERVICE_PORT=8087
ENV CONFIG_PATH=/app/config/config.yaml

# Entrypoint
ENTRYPOINT ["/app/neam-macro"]
