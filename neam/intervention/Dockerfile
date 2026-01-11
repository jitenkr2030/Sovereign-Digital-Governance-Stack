FROM golang:1.21-alpine AS builder

# Install dependencies
RUN apk add --no-cache git ca-certificates tzdata

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
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o intervention-service \
    -ldflags "-X main.version=$(git describe --tags --always 2>/dev/null || echo 'dev')" \
    -ldflags "-s -w" \
    ./main.go

# Create non-root user
RUN addgroup -g 1000 appgroup && \
    adduser -u 1000 -G appgroup -s /bin/sh -D appuser

# Final stage - runtime image
FROM scratch

# Copy ca-certificates for HTTPS support
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy application
COPY --from=builder /app/intervention-service /app/intervention-service

# Copy configuration
COPY --from=builder /app/config.yaml /app/config.yaml

# Create non-root user
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/group /etc/group

# Set ownership
RUN chown -R appuser:appgroup /app

# Switch to non-root user
USER appuser

# Expose ports
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Set working directory
WORKDIR /app

# Entry point
ENTRYPOINT ["/app/intervention-service"]

# Default command
CMD ["serve"]
