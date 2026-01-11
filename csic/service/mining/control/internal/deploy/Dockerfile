# Build stage
# Using a multi-stage build to create a lightweight production image
FROM golang:1.21-alpine AS builder

# Install git for Go module dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy go mod and sum files first for better layer caching
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download && \
    go mod verify

# Copy the source code
COPY . .

# Build the application with optimizations
# -ldflags: -s -w removes debug symbols to reduce binary size
# -buildvcs=false: don't embed VCS information
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w -X=main.version=$(cat VERSION 2>/dev/null || echo 'dev')" \
    -o mining-control \
    -buildvcs=false \
    ./main.go

# Production stage
# Lightweight image with only the compiled binary
FROM alpine:3.19 AS production

# Install runtime dependencies
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

# Copy the compiled binary from builder stage
COPY --from=builder /app/mining-control .

# Create necessary directories
RUN mkdir -p logs config data && \
    chown -R appuser:appgroup /home/appuser

# Switch to non-root user
USER appuser

# Expose application ports
# 8080: HTTP API port
# 9090: Metrics port
EXPOSE 8080 9090

# Set environment variables
ENV APP_ENV=production
ENV GIN_MODE=release

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget --spider -q http://localhost:8080/health || exit 1

# Set default command
ENTRYPOINT ["./mining-control"]

# Labels for container metadata
LABEL maintainer="CSIC Platform Team"
LABEL version="1.0.0"
LABEL description="Mining Control Microservice for Energy Compliance and Monitoring"
