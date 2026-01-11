# Build stage
FROM rust:1.73-alpine AS builder

# Install build dependencies
RUN apk add --no-cache \
    build-base \
    openssl-dev \
    pkgconfig \
    libpq-dev \
    git

# Set working directory
WORKDIR /app

# Copy dependency files first for better caching
COPY Cargo.toml Cargo.lock ./

# Download dependencies
RUN cargo fetch && cargo build --release --locked

# Copy source code
COPY . .

# Build the application
RUN cargo build --release --bin csic-cwgs

# Strip binary for smaller size
RUN strip /app/target/release/csic-cwgs

# Final stage
FROM alpine:3.19 AS runtime

# Install runtime dependencies
RUN apk add --no-cache \
    openssl \
    libpq \
    ca-certificates \
    tzdata \
    && rm -rf /var/cache/apk/*

# Create non-root user
RUN addgroup -g 1000 appgroup && \
    adduser -u 1000 -G appgroup -s /bin/sh -D appuser

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/target/release/csic-cwgs .

# Copy configuration file
COPY --from=builder /app/config.yaml ./config/

# Create directories for data
RUN mkdir -p /app/data /app/logs && \
    chown -R appuser:appgroup /app

# Switch to non-root user
USER appuser

# Expose ports
EXPOSE 8081 50052 9090

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8081/health || exit 1

# Environment variables for configuration
ENV CWGS_APP_HTTP_PORT=8081
ENV CWGS_APP_GRPC_PORT=50052
ENV CWGS_DATABASE_HOST=postgres
ENV CWGS_REDIS_HOST=redis
ENV CWGS_KAFKA_BROKERS=kafka:9092
ENV CWGS_LOGGING_LEVEL=info

# Entry point
ENTRYPOINT ["/app/csic-cwgs"]
