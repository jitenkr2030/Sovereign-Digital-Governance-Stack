# NEAM Sensing Layer - Docker Build
# Multi-stage build for optimized image

# ============ BUILDER STAGE ============
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy go modules
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# Copy source code
COPY . .

# Build binary with optimizations
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o sensing-service \
    -ldflags="-s -w -X main.version=${VERSION:-dev} -X main.buildDate=${BUILD_DATE:-$(date -u +%Y-%m-%dT%H:%M:%SZ)}" \
    ./sensing

# ============ RUNNER STAGE ============
FROM alpine:3.19 AS runner

# Security: Run as non-root user
RUN addgroup -g 1000 neam && \
    adduser -u 1000 -G neam -s /bin/sh -D neam

# Install minimal runtime dependencies
RUN apk add --no-cache ca-certificates tzdata curl

# Create non-root home directory
RUN mkdir /home/neam && chown neam:neam /home/neam

# Copy binary from builder
COPY --from=builder /app/sensing-service /usr/local/bin/sensing-service

# Create directories for data and logs
RUN mkdir -p /data/sensing /logs/sensing && \
    chown -R neam:neam /data /logs

# Switch to non-root user
USER neam

# Set working directory
WORKDIR /home/neam

# Expose service port
EXPOSE 8082

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8082/health || exit 1

# Set environment variables
ENV NEAM_ENV=production
ENV NEAM_LOG_LEVEL=info

# Run the service
ENTRYPOINT ["sensing-service"]

# Default command arguments
CMD ["--config", "/etc/neam/config.yaml"]
