# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install git for go modules
RUN apk add --no-cache git

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o health-monitor ./cmd/main.go

# Production stage
FROM alpine:latest

# Install ca-certificates for HTTPS connections
RUN apk --no-cache add ca-certificates

# Create non-root user
RUN addgroup -g 1000 appgroup && \
    adduser -u 1000 -G appgroup -s /bin/sh -D appuser

# Copy binary from builder
COPY --from=builder /app/health-monitor /usr/local/bin/health-monitor

# Copy configuration
COPY --from=builder /app/internal/config/config.yaml /etc/health-monitor/config.yaml

# Create directories
RUN mkdir -p /var/log/health-monitor && \
    chown -R appuser:appgroup /var/log/health-monitor

# Switch to non-root user
USER appuser

# Expose ports
EXPOSE 8081 9091

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8081/health || exit 1

# Set environment variables
ENV GIN_MODE=release

# Run the application
ENTRYPOINT ["health-monitor"]
CMD ["--config", "/etc/health-monitor/config.yaml"]
