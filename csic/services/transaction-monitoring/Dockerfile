FROM golang:1.21-alpine AS builder

# Install git for go modules
RUN apk add --no-cache git

# Set working directory
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o transaction-monitoring ./cmd/server

# Production stage
FROM alpine:latest

# Install ca-certificates for HTTPS connections
RUN apk --no-cache add ca-certificates curl

# Create non-root user
RUN addgroup -g 1000 app && adduser -u 1000 -G app -s /bin/sh -D app

# Copy binary from builder
COPY --from=builder /app/transaction-monitoring /usr/local/bin/

# Copy configuration
COPY --from=builder /app/config.yaml /etc/transaction-monitoring/config.yaml

# Create directories for logs and data
RUN mkdir -p /var/log/transaction-monitoring && chown -R app:app /var/log/transaction-monitoring

# Switch to non-root user
USER app

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8080/health || exit 1

# Set environment variables
ENV TZ=UTC
ENV GOGC=100

# Entry point
ENTRYPOINT ["transaction-monitoring"]
CMD ["-config", "/etc/transaction-monitoring/config.yaml"]
