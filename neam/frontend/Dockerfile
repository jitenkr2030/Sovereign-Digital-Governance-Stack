# NEAM Frontend - Docker Build
# Production-ready React frontend with Nginx

# ============ BUILD STAGE ============
FROM node:20-alpine AS builder

WORKDIR /app

# Copy package files
COPY frontend/package*.json ./

# Install dependencies
RUN npm ci

# Copy source code
COPY frontend/ ./

# Build production bundle
ENV NODE_ENV=production
RUN npm run build

# ============ PRODUCTION STAGE ============
FROM nginx:alpine AS runner

# Security: Run as non-root user
RUN addgroup -g 1000 neam && \
    adduser -u 1000 -G neam -s /bin/sh -D neam

# Install security utilities
RUN apk add --no-cache curl ca-certificates bash && \
    rm -rf /var/cache/apk/*

# Copy nginx configuration
COPY frontend/nginx.conf /etc/nginx/conf.d/default.conf

# Copy built assets
COPY --from=builder /app/dist /usr/share/nginx/html

# Set permissions
RUN chown -R neam:neam /usr/share/nginx/html /var/cache/nginx /var/run && \
    chmod -R 755 /usr/share/nginx/html

USER neam

EXPOSE 80

HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:80/health || exit 1

# Start nginx
CMD ["nginx", "-g", "daemon off;"]
