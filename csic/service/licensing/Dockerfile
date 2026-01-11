# Build stage
FROM golang:1.21-alpine AS builder

# Install git for Go modules
RUN apk add --no-cache git

# Set working directory
WORKDIR /app

# Copy go.mod and go.sum files first for better caching
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download
RUN go mod verify

# Copy the rest of the application
COPY . .

# Build the application with optimizations
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo \
    -ldflags="-s -w -X main.version=$(cat VERSION 2>/dev/null || echo 'dev') -X main.buildTime=$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
    -o csic-licensing-service ./cmd/api

# Runtime stage - minimal image
FROM alpine:3.19 AS runtime

# Install necessary runtime dependencies
RUN apk add --no-cache \
    ca-certificates \
    tzdata \
    curl \
    && rm -rf /var/cache/apk/*

# Create non-root user for security
RUN addgroup -g 1000 appgroup && \
    adduser -u 1000 -G appgroup -s /bin/sh -D appuser

# Set working directory
WORKDIR /home/appuser

# Copy binary from builder
COPY --from=builder /app/csic-licensing-service .

# Copy configuration files
COPY --from=builder /app/config.yaml ./config.yaml
COPY --from=builder /app/config ./config/

# Create directories for logs and data
RUN mkdir -p logs data && \
    chown -R appuser:appgroup logs data

# Switch to non-root user
USER appuser

# Expose application port
EXPOSE 8081

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8081/health || exit 1

# Set environment variables
ENV APP_ENV=production
ENV GIN_MODE=release

# Set timezone
ENV TZ=UTC

# Entrypoint
ENTRYPOINT ["./csic-licensing-service"]

# Default command reads from config
CMD ["--config", "config.yaml"]
