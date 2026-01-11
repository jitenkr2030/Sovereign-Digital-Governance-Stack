# NEAM Simulation Service Dockerfile
# What-If Analysis and Scenario Simulation Engine

# Use Python 3.11 slim image for optimal size and performance
FROM python:3.11-slim-bookworm

# Set environment variables
ENV PYTHONUNBUFFERED=1 \
    PYTHONDONTWRITEBYTECODE=1 \
    PIP_NO_CACHE_DIR=1 \
    PIP_DISABLE_PIP_VERSION_CHECK=1 \
    NUMEXPR_MAX_THREADS=4 \
    OMP_NUM_THREADS=4

# Set working directory
WORKDIR /workspace

# Install system dependencies
RUN apt-get update && apt-get install -y --no-install-recommends \
    gcc \
    g++ \
    build-essential \
    libpq-dev \
    && rm -rf /var/lib/apt/lists/*

# Copy requirements first for better caching
COPY simulation/requirements.txt .

# Install Python dependencies
RUN pip install --no-cache-dir -r requirements.txt

# Copy application code
COPY simulation/ ./simulation/

# Create log directory
RUN mkdir -p /var/log/neam && chown -R nobody:nogroup /var/log/neam

# Create non-root user for security
RUN groupadd -r neam && useradd -r -g neam neam && \
    chown -R neam:neam /workspace

# Switch to non-root user
USER neam

# Expose the service port
EXPOSE 8083

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD python -c "import urllib.request; urllib.request.urlopen('http://localhost:8083/api/v1/health/live')" || exit 1

# Set environment variables for runtime
ENV REDIS_HOST=10.112.2.4 \
    REDIS_PORT=6379 \
    DATABASE_HOST=10.112.2.4 \
    DATABASE_PORT=5432

# Run the application
CMD ["uvicorn", "simulation.main:app", "--host", "0.0.0.0", "--port", "8083", "--workers", "4"]
