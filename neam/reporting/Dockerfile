# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy go modules
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o reporting-service ./reporting

# Production stage
FROM alpine:3.19

# Install required packages
RUN apk --no-cache add ca-certificates tzdata curl

# Create non-root user
RUN addgroup -g 1000 appgroup && \
    adduser -u 1000 -G appgroup -s /bin/sh -D appuser

# Set working directory
WORKDIR /home/appuser

# Copy binary from builder
COPY --from=builder /app/reporting-service .
COPY --from=builder /app/config.yaml .

# Create directories
RUN mkdir -p /var/lib/neam/reporting/exports && \
    mkdir -p /var/lib/neam/reporting/archives && \
    chown -R appuser:appgroup /var/lib/neam/reporting

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 8082

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8082/health || exit 1

# Set environment variables
ENV GIN_MODE=release
ENV SERVICE_NAME=neam-reporting

# Entry point
ENTRYPOINT ["./reporting-service"]
