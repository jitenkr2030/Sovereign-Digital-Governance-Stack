# NEAM Intelligence Engine - Docker Build
# Multi-stage build with Python ML support

# ============ GO BUILDER STAGE ============
FROM golang:1.21-alpine AS go-builder

RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o intelligence-service \
    -ldflags="-s -w -X main.version=${VERSION:-dev} -X main.buildDate=${BUILD_DATE:-$(date -u +%Y-%m-%dT%H:%M:%SZ)}" \
    ./intelligence

# ============ PYTHON BUILDER STAGE ============
FROM python:3.11-slim AS python-builder

WORKDIR /app

COPY intelligence/requirements.txt .

RUN pip install --no-cache-dir -r requirements.txt && \
    python -c "import nltk; nltk.download('punkt', quiet=True)"

COPY intelligence/python/ ./python/

# ============ RUNTIME STAGE ============
FROM alpine:3.19 AS runner

# Security: Run as non-root user
RUN addgroup -g 1000 neam && \
    adduser -u 1000 -G neam -s /bin/sh -D neam

RUN apk add --no-cache ca-certificates tzdata curl libpq

# Copy Go binary
COPY --from=go-builder /app/intelligence-service /usr/local/bin/intelligence-service

# Copy Python packages
COPY --from=python-builder /usr/local/lib/python3.11/site-packages /home/neam/site-packages
COPY --from=python-builder /app/python /home/neam/python

# Create directories
RUN mkdir -p /data/intelligence /logs/intelligence /models && \
    chown -R neam:neam /data /logs /models /home/neam

USER neam

WORKDIR /home/neam

EXPOSE 8083

HEALTHCHECK --interval=30s --timeout=10s --start-period=10s --retries=3 \
    CMD curl -f http://localhost:8083/health || exit 1

ENV NEAM_ENV=production
ENV NEAM_LOG_LEVEL=info
ENV PYTHONPATH=/home/neam/site-packages

ENTRYPOINT ["intelligence-service"]

CMD ["--config", "/etc/neam/config.yaml"]
