# NEAM API Gateway Service - Docker Build

FROM golang:1.21-alpine AS builder

RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY services/ ./

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o api-gateway \
    -ldflags="-s -w -X main.version=${VERSION:-dev} -X main.buildDate=${BUILD_DATE:-$(date -u +%Y-%m-%dT%H:%M:%SZ)}" \
    ./services

FROM alpine:3.19 AS runner

RUN addgroup -g 1000 neam && \
    adduser -u 1000 -G neam -s /bin/sh -D neam

RUN apk add --no-cache ca-certificates tzdata curl

COPY --from=builder /app/api-gateway /usr/local/bin/api-gateway

RUN mkdir -p /data/api /logs/api && \
    chown -R neam:neam /data /logs

USER neam

WORKDIR /home/neam

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8080/health || exit 1

ENV NEAM_ENV=production
ENV NEAM_LOG_LEVEL=info

ENTRYPOINT ["api-gateway"]

CMD ["--config", "/etc/neam/config.yaml"]
