FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy go modules files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o energy-adapter .

# Final stage - minimal image
FROM alpine:3.19

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata curl

# Create non-root user
RUN addgroup -g 1000 appgroup && \
    adduser -u 1000 -G appgroup -s /bin/sh -D appuser

# Set working directory
WORKDIR /home/appuser

# Copy binary from builder
COPY --from=builder /app/energy-adapter .
COPY --from=builder /app/config.yaml.example /etc/neam/config.yaml

# Create directories
RUN mkdir -p /var/log/neam /etc/neam

# Set ownership
RUN chown -R appuser:appgroup /home/appuser /var/log/neam /etc/neam

# Switch to non-root user
USER appuser

# Expose ports
EXPOSE 8085 9091

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:9091/health || exit 1

# Run the adapter
ENTRYPOINT ["./energy-adapter"]
CMD ["--config=/etc/neam/config.yaml"]
