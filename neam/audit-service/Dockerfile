# ============================================================
# Audit, Security & Sovereignty Layer - Dockerfile
# Multi-stage build for production deployment
# ============================================================

# Stage 1: Build stage
FROM golang:1.21-alpine AS builder

# Set working directory
WORKDIR /app

# Set environment variables
ENV GO111MODULE=on
ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64

# Install git for go modules
RUN apk add --no-cache git

# Copy go mod files first for better caching
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download
RUN go mod verify

# Copy the rest of the application
COPY . .

# Build the application
RUN go build -ldflags="-s -w" -o audit-service ./main.go

# ============================================================
# Stage 2: Runtime stage
# ============================================================
FROM alpine:3.19 AS runtime

# Install required packages
RUN apk add --no-cache \
    ca-certificates \
    tzdata \
    curl \
    && rm -rf /var/cache/apk/*

# Set working directory
WORKDIR /app

# Create non-root user
RUN addgroup -g 1000 appgroup && \
    adduser -u 1000 -G appgroup -s /bin/sh -D appuser

# Copy binary from builder
COPY --from=builder /app/audit-service /app/audit-service

# Copy configuration
COPY --from=builder /app/config/ /app/config/
COPY --from=builder /app/migrations/ /app/migrations/

# Set ownership
RUN chown -R appuser:appgroup /app && \
    chmod +x /app/audit-service

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 9090

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:9090/health || exit 1

# Environment variables
ENV SERVICE_NAME=audit-service
ENV SERVICE_PORT=9090

# Entry point
ENTRYPOINT ["/app/audit-service"]

# Default command (can be overridden)
CMD ["--config", "/app/config/config.yaml"]
