# ============================================
# Regulatory Reporting Service - Dockerfile
# ============================================
# Multi-stage build for optimized image size
# ============================================

# Stage 1: Build stage
FROM node:20-alpine AS builder

# Set working directory
WORKDIR /app

# Install build dependencies
RUN apk add --no-cache python3 make g++

# Copy package files
COPY package*.json ./

# Install dependencies
RUN npm ci --only=production

# Copy source code
COPY . .

# Build TypeScript
RUN npm run build

# Stage 2: Production stage
FROM node:20-alpine AS production

# Set NODE_ENV
ENV NODE_ENV=production

# Create non-root user
RUN addgroup -g 1001 -S nodejs && \
    adduser -S reporting -u 1001 -G nodejs

# Set working directory
WORKDIR /app

# Copy package files
COPY package*.json ./

# Install production dependencies only
RUN npm ci --only=production && \
    npm cache clean --force

# Copy built application from builder stage
COPY --from=builder --chown=reporting:nodejs /app/dist ./dist
COPY --from=builder --chown=reporting:nodejs /app/config.yaml ./

# Create reports directory
RUN mkdir -p /app/reports && \
    chown -R reporting:nodejs /app

# Switch to non-root user
USER reporting

# Expose application port
EXPOSE 3001

# Expose metrics port
EXPOSE 9091

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=30s --retries=3 \
    CMD node -e "require('http').get('http://localhost:3001/health', (r) => process.exit(r.statusCode === 200 ? 0 : 1))"

# Default command
CMD ["node", "dist/main.js"]
