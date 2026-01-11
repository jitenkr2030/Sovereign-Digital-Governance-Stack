# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy go modules files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o smart-cities ./cmd/server

# Production stage
FROM alpine:latest

# Install ca-certificates for HTTPS connections
RUN apk --no-cache add ca-certificates

# Create non-root user
RUN addgroup -g 1000 app && adduser -u 1000 -G app -s /bin/sh -D app

# Copy binary from builder
COPY --from=builder /app/smart-cities /usr/local/bin/

# Copy configuration
COPY --from=builder /app/config.yaml /etc/smart-cities/config.yaml

# Create directories for data and logs
RUN mkdir -p /var/log/smart-cities && chown -R app:app /var/log/smart-cities

# Switch to non-root user
USER app

# Expose port
EXPOSE 8090

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8090/health || exit 1

# Set environment variables
ENV GIN_MODE=release
ENV CONFIG_PATH=/etc/smart-cities/config.yaml

# Entry point
ENTRYPOINT ["smart-cities"]
CMD ["--config", "/etc/smart-cities/config.yaml"]
