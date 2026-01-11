# =============================================================================
# Black Economy Intelligence Service - Dockerfile
# =============================================================================
# Purpose: Containerizes the Black Economy Monitoring service for detecting
# shadow economy activities, anomaly detection, and hidden network analysis.
# =============================================================================

# Build Stage
FROM golang:1.21-alpine AS builder

# Install git for go mod download and build dependencies
RUN apk add --no-cache git make

# Set working directory
WORKDIR /app

# Copy go.mod and go.sum first for dependency caching
COPY black-economy/go.mod black-economy/go.sum ./

# Download dependencies
RUN cd black-economy && go mod download && go mod verify

# Copy source code
COPY black-economy/ ./black-economy/

# Build the binary with optimizations
RUN cd black-economy && CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o neam-black-economy main.go

# =============================================================================
# Production Stage - Minimal runtime image
# =============================================================================
FROM alpine:3.19 AS production

# Install minimal runtime dependencies
RUN apk add --no-cache ca-certificates tzdata curl

# Create non-root user for security
RUN addgroup -g 1000 neam && \
    adduser -u 1000 -G neam -s /bin/sh -D neam

# Set working directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/black-economy/neam-black-economy /app/neam-black-economy

# Create required directories
RUN mkdir -p /app/config /app/logs && \
    chown -R neam:neam /app

# Switch to non-root user
USER neam

# Expose service port
EXPOSE 8085

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget --spider -q http://localhost:8085/health || exit 1

# Environment configuration
ENV SERVICE_PORT=8085
ENV CONFIG_PATH=/app/config/config.yaml

# Entrypoint
ENTRYPOINT ["/app/neam-black-economy"]
