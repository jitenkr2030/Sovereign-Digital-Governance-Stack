# Build stage
FROM golang:1.21-alpine AS builder

# Install git for go mod download
RUN apk add --no-cache git

# Set working directory
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download && go mod verify

# Copy source code
COPY . .

# Build the application with optimizations
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o csic-tx-monitor ./cmd/server

# Production stage
FROM gcr.io/distroless/static-debian11 AS production

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/csic-tx-monitor .

# Copy configuration file
COPY --from=builder /app/config.yaml .

# Create non-root user
RUN addgroup -g 1000 app && adduser -u 1000 -G app -s /bin/sh -D app

# Change ownership
RUN chown -R app:app /app

# Switch to non-root user
USER app

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD ["/app/csic-tx-monitor", "--health"] || exit 1

# Set environment variables
ENV GIN_MODE=release

# Entry point
ENTRYPOINT ["/app/csic-tx-monitor"]
