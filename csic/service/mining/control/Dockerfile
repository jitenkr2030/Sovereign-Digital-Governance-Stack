FROM golang:1.21-alpine AS builder

# Install git for go modules
RUN apk add --no-cache git

# Set working directory
WORKDIR /app

# Copy go module files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download
RUN go mod verify

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o mining-control-service ./internal/cmd/server

# Production stage
FROM alpine:3.18

# Install CA certificates, timezone data, and utils
RUN apk add --no-cache ca-certificates tzdata curl jq

# Create non-root user
RUN addgroup -g 1000 app && adduser -u 1000 -G app -s /bin/sh -D app

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/mining-control-service .

# Create necessary directories
RUN mkdir -p /var/lib/csic/mining /var/log/csic && \
    chown -R app:app /var/lib/csic /var/log/csic

# Copy default config
COPY --from=builder /app/internal/config/config.yaml ./config.yaml

# Expose ports (API: 8082, Stratum: 3333)
EXPOSE 8082 3333

# Switch to non-root user
USER app

# Health check for API
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8082/health || exit 1

# Run the service
ENTRYPOINT ["./mining-control-service"]
CMD ["-config", "internal/config/config.yaml"]
