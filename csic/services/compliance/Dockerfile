FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install git for go modules
RUN apk add --no-cache git

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o compliance-api ./cmd/main.go

# Runtime stage
FROM alpine:latest

# Install necessary packages
RUN apk --no-cache add ca-certificates tzdata

# Create non-root user
RUN addgroup -g 1000 appgroup && \
    adduser -u 1000 -G appgroup -s /bin/sh -D appuser

# Copy the built binary
COPY --from=builder /app/compliance-api /usr/local/bin/compliance-api

# Copy configuration file
COPY --from=builder /app/config.yaml /etc/compliance-api/config.yaml

# Create directories for data and logs
RUN mkdir -p /var/log/compliance-api && \
    chown -R appuser:appgroup /var/log/compliance-api

# Switch to non-root user
USER appuser

# Set working directory
WORKDIR /home/appuser

# Expose application port
EXPOSE 8081

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8081/health || exit 1

# Run the application
ENTRYPOINT ["compliance-api"]
CMD ["-config", "/etc/compliance-api/config.yaml"]
