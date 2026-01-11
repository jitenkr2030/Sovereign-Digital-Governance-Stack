# =============================================================================
# Feature Store Service - Dockerfile
# =============================================================================
# Purpose: Containerizes the ML Feature Store for managing online and offline
# feature retrieval, feature registry, and data ingestion pipelines.
# =============================================================================

# Build Stage
FROM golang:1.21-alpine AS builder

# Install git for go mod download and build dependencies
RUN apk add --no-cache git make

# Set working directory
WORKDIR /app

# Copy go.mod and go.sum first for dependency caching
COPY feature-store/go.mod feature-store/go.sum ./

# Download dependencies
RUN cd feature-store && go mod download && go mod verify

# Copy source code
COPY feature-store/ ./feature-store/

# Build the binary with optimizations
RUN cd feature-store && CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o neam-feature-store main.go

# =============================================================================
# Production Stage - Minimal runtime image
# =============================================================================
FROM alpine:3.19 AS production

# Install minimal runtime dependencies
RUN apk add --no-cache ca-certificates tzdata curl

# Create non-root user for security
RUN addgroup -g 1001 neam && \
    adduser -u 1001 -G neam -s /bin/sh -D neam

# Set working directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/feature-store/neam-feature-store /app/neam-feature-store

# Create required directories for feature data and model storage
RUN mkdir -p /app/config /app/logs /app/data/features /app/data/models && \
    chown -R neam:neam /app

# Switch to non-root user
USER neam

# Expose service port
EXPOSE 8086

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget --spider -q http://localhost:8086/health || exit 1

# Environment configuration
ENV SERVICE_PORT=8086
ENV CONFIG_PATH=/app/config/config.yaml

# Entrypoint
ENTRYPOINT ["/app/neam-feature-store"]
