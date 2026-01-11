# Multi-stage build for CBDC Monitoring Service
# Stage 1: Builder
FROM golang:1.21-alpine AS builder

# Install git for go modules
RUN apk add --no-cache git

# Set working directory
WORKDIR /app

# Copy go mod and sum files first for better caching
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download && go mod verify

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o cbdc-monitoring-service \
    ./cmd/server/main.go

# Stage 2: Runtime
FROM gcr.io/distroless/static:nonroot AS runtime

# Create non-root user
RUN addgroup -g 1000 appgroup && \
    adduser -u 1000 -G appgroup -s /bin/sh -D appuser

# Set working directory
WORKDIR /home/appuser

# Copy binary from builder
COPY --from=builder /app/cbdc-monitoring-service .

# Copy configuration
COPY --from=builder /app/config.yaml ./config.yaml

# Create logs directory
RUN mkdir -p /home/appuser/logs && chown -R appuser:appgroup /home/appuser

# Change to non-root user
USER appuser

# Expose application port
EXPOSE 8085

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8085/health || exit 1

# Set environment variables
ENV GIN_MODE=release
ENV PORT=8085

# Entry point
ENTRYPOINT ["./cbdc-monitoring-service"]
