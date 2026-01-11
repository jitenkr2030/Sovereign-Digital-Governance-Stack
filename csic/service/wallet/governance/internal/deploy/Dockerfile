# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w" \
    -o wallet-governance \
    ./internal/cmd/server/main.go

# Production stage
FROM alpine:3.19 AS production

RUN apk add --no-cache ca-certificates tzdata && rm -rf /var/cache/apk/*

RUN addgroup -g 1000 appgroup && \
    adduser -u 1000 -G appgroup -s /bin/sh -D appuser

WORKDIR /home/appuser

COPY --from=builder /app/wallet-governance .
COPY --from=builder /app/internal/config/config.yaml ./internal/config/

RUN mkdir -p logs data && chown -R appuser:appgroup /home/appuser

USER appuser

EXPOSE 8084 9094

ENTRYPOINT ["./wallet-governance"]

LABEL maintainer="CSIC Platform Team"
LABEL version="1.0.0"
LABEL description="Wallet Governance Service for CSIC Platform"
