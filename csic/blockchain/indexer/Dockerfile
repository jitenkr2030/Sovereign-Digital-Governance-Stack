# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o indexer ./cmd

# Production stage
FROM alpine:latest

# Install ca-certificates for HTTPS connections
RUN apk --no-cache add ca-certificates

# Create non-root user
RUN addgroup -g 1000 app && adduser -u 1000 -G app -s /bin/sh -D app

# Copy binary from builder
COPY --from=builder /app/indexer /usr/local/bin/

# Copy configuration
COPY --from=builder /app/config.yaml /etc/csic/indexer/config.yaml

# Create data directory
RUN mkdir -p /var/lib/csic/indexer && chown -R app:app /var/lib/csic/indexer

# Switch to non-root user
USER app

# Expose port
EXPOSE 8087

# Set environment variables
ENV CSIC_CONFIG_PATH=/etc/csic/indexer/config.yaml

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8087/health || exit 1

# Entrypoint
ENTRYPOINT ["indexer"]
