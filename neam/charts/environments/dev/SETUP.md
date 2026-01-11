# NEAM Platform Development Environment Setup Guide

This guide provides step-by-step instructions for setting up the NEAM Platform development environment.

## Prerequisites

- Docker and Docker Compose
- kubectl configured
- Access to development container registry
- At least 8GB of available RAM
- 50GB of free disk space

## Quick Start

### 1. Clone Repository

```bash
git clone https://github.com/neam-platform/neam-platform.git
cd neam-platform
```

### 2. Start Local Infrastructure

```bash
# Start all development services
docker-compose -f docker-compose.dev.yml up -d

# Verify services are running
docker-compose ps
```

### 3. Configure Local Environment

```bash
# Copy environment template
cp .env.example .env

# Configure environment variables
# Edit .env with your local settings
```

### 4. Deploy Development Charts

```bash
# Deploy to development namespace
./charts/deploy-dev.sh
```

### 5. Verify Deployment

```bash
# Check pod status
kubectl get pods -n neam-dev

# Access services
# Use port-forward to access services
kubectl port-forward -n neam-dev svc/neam-sensing 8080:8080
```

## Local Infrastructure Components

### PostgreSQL

```yaml
# Connection parameters
Host: localhost
Port: 5432
Database: neam_dev
Username: neam_dev
Password: dev_password
SSL: Disabled
```

**Start PostgreSQL:**
```bash
docker run -d \
  --name neam-postgres-dev \
  -e POSTGRES_USER=neam_dev \
  -e POSTGRES_PASSWORD=dev_password \
  -e POSTGRES_DB=neam_dev \
  -p 5432:5432 \
  postgres:15
```

### Redis

```yaml
# Connection parameters
Host: localhost
Port: 6379
Password: (none)
Database: 0
Key Prefix: dev:
```

**Start Redis:**
```bash
docker run -d \
  --name neam-redis-dev \
  -p 6379:6379 \
  redis:7-alpine
```

### Kafka

```yaml
# Connection parameters
Brokers: localhost:9092
Consumer Group: neam-dev
Auto Offset Reset: earliest
```

**Start Kafka:**
```bash
docker run -d \
  --name neam-kafka-dev \
  -p 9092:9092 \
  -e KAFKA_ZOOKEEPER_CONNECT=localhost:2181 \
  -e KAFKA_ADVERTISED_LISTENERS=PLAINTEXT://localhost:9092 \
  -e KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR=1 \
  confluentinc/cp-kafka:7.0.0
```

### MongoDB

```yaml
# Connection parameters
Host: localhost
Port: 27017
Database: neam_dev
Username: neam_dev
Password: dev_password
Auth Source: admin
```

**Start MongoDB:**
```bash
docker run -d \
  --name neam-mongodb-dev \
  -e MONGO_INITDB_DATABASE=neam_dev \
  -e MONGO_INITDB_ROOT_USERNAME=neam_dev \
  -e MONGO_INITDB_ROOT_PASSWORD=dev_password \
  -p 27017:27017 \
  mongo:6
```

### Elasticsearch

```yaml
# Connection parameters
Host: localhost
Port: 9200
Index: neam-dev
SSL: Disabled
```

**Start Elasticsearch:**
```bash
docker run -d \
  --name neam-elasticsearch-dev \
  -p 9200:9200 \
  -e "discovery.type=single-node" \
  -e "xpack.security.enabled=false" \
  docker.elastic.co/elasticsearch/elasticsearch:8.8.0
```

### Vault (Optional for Development)

```yaml
# Connection parameters
Address: http://localhost:8200
Token: dev-token
Namespace: dev
```

**Start Vault:**
```bash
docker run -d \
  --name neam-vault-dev \
  -p 8200:8200 \
  -e VAULT_ADDR=http://localhost:8200 \
  -e VAULT_DEV_ROOT_TOKEN=dev-token \
  vault:1.13
```

**Initialize Vault for Development:**
```bash
export VAULT_ADDR=http://localhost:8200
export VAULT_TOKEN=dev-token

# Enable KV secrets engine
vault secrets enable -path=secret kv-v2

# Create development secrets
vault kv put secret/dev/database/postgresql \
  username=neam_dev \
  password=dev_password \
  connection_string="postgresql://neam_dev:dev_password@localhost:5432/neam_dev"

vault kv put secret/dev/application \
  jwt_secret="development-jwt-secret-change-in-production" \
  encryption_key="development-encryption-key-change-in-production"
```

## Development Database Schema

### PostgreSQL Schema

```sql
-- Create development database schema
-- Connect to PostgreSQL: psql -h localhost -U neam_dev -d neam_dev

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Create tables
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS entities (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    type VARCHAR(50) NOT NULL,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS transactions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    entity_id UUID REFERENCES entities(id),
    amount DECIMAL(15, 2) NOT NULL,
    currency VARCHAR(3) DEFAULT 'USD',
    status VARCHAR(20) DEFAULT 'pending',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_entities_type ON entities(type);
CREATE INDEX IF NOT EXISTS idx_transactions_entity ON transactions(entity_id);
CREATE INDEX IF NOT EXISTS idx_transactions_status ON transactions(status);
```

## Feature Flags for Development

All feature flags are enabled for development to allow full testing:

```yaml
featureFlags:
  advancedAnalytics: true
  aiPrediction: true
  realTimeProcessing: true
  customRules: true
  auditLogging: true
  debugMode: true
  performanceMonitoring: true
  betaFeatures: true
```

## Logging Configuration

Development environment uses debug logging:

```yaml
env:
  - name: LOG_LEVEL
    value: "debug"
  - name: LOG_FORMAT
    value: "json"
```

**View Logs:**
```bash
# View all logs
kubectl logs -n neam-dev -l app.kubernetes.io/part-of=neam-platform -f

# View specific service logs
kubectl logs -n neam-dev deployment/neam-sensing -f

# View with timestamps
kubectl logs -n neam-dev deployment/neam-sensing --timestamps -f
```

## Troubleshooting

### Service Won't Start

```bash
# Check pod status
kubectl get pods -n neam-dev

# Check events
kubectl describe pod <pod-name> -n neam-dev

# Check logs
kubectl logs <pod-name> -n neam-dev
```

### Database Connection Issues

```bash
# Verify PostgreSQL is running
docker ps | grep postgres

# Test connection
PGPASSWORD=dev_password psql -h localhost -U neam_dev -d neam_dev -c "SELECT 1"

# Check service
kubectl exec -it -n neam-dev <pod-name> -- nc -zv localhost 5432
```

### Kafka Connection Issues

```bash
# Verify Kafka is running
docker ps | grep kafka

# Test connection
kubectl exec -n neam-dev <pod-name> -- kafkacat -b localhost:9092 -L
```

### Memory Issues

```bash
# Check resource usage
kubectl top pods -n neam-dev

# Increase resource limits
# Edit values.yaml and increase memory/cpu limits
```

## Development Tips

### Hot Reload

For Go applications, use:
```bash
# Build with hot reload support
make run-dev

# Or use Air for auto-reload
air -c .air.conf
```

### Debug Mode

Enable debug mode for detailed logging:
```yaml
env:
  - name: DEBUG
    value: "true"
```

### Database Seeding

```bash
# Seed development database
kubectl exec -n neam-dev deployment/neam-sensing -- ./sensing seed --environment dev
```

## Access Points

| Service | URL | Local Port |
|---------|-----|------------|
| Sensing API | http://localhost:8080 | 8080 |
| PostgreSQL | localhost:5432 | 5432 |
| Redis | localhost:6379 | 6379 |
| Kafka | localhost:9092 | 9092 |
| MongoDB | localhost:27017 | 27017 |
| Elasticsearch | http://localhost:9200 | 9200 |
| Vault UI | http://localhost:8200 | 8200 |

## Next Steps

1. Review [API Documentation](../api/README.md)
2. Set up [IDE Configuration](../.vscode/extensions.json)
3. Learn about [Development Workflow](../CONTRIBUTING.md)
4. Explore [Architecture Documentation](../docs/architecture.md)
