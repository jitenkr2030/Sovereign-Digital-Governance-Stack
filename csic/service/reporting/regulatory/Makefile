# ============================================
# Regulatory Reporting Service - Makefile
# ============================================
# Common deployment and operational tasks
# ============================================

# Variables
DOCKER_COMPOSE = docker-compose
COMPOSE_FILE = internal/deploy/docker-compose.yml
ENV_FILE = .env
SERVICE_NAME = csic-reporting

# Colors for output
RED = \033[0;31m
GREEN = \033[0;32m
YELLOW = \033[1;33m
NC = \033[0m # No Color

# Default target
.PHONY: help
help:
	@echo ""
	@echo "=========================================="
	@echo "  CSIC Regulatory Reporting Service"
	@echo "=========================================="
	@echo ""
	@echo "Usage: make <command>"
	@echo ""
	@echo "Development Commands:"
	@echo "  install          Install dependencies"
	@echo "  dev              Start development environment"
	@echo "  dev-down         Stop development environment"
	@echo ""
	@echo "Production Commands:"
	@echo "  up               Start production services"
	@echo "  down             Stop all services"
	@echo "  restart          Restart all services"
	@echo ""
	@echo "Database Commands:"
	@echo "  db-migrate       Run database migrations"
	@echo "  db-seed          Seed database with initial data"
	@echo "  db-backup        Create database backup"
	@echo "  db-restore       Restore database from backup"
	@echo "  db-console       Connect to database console"
	@echo ""
	@echo "Maintenance Commands:"
	@echo "  logs             View application logs"
	@echo "  logs-tail        Tail application logs"
	@echo "  status           Check service status"
	@echo "  health           Check service health"
	@echo "  metrics          View service metrics"
	@echo ""
	@echo "Build Commands:"
	@echo "  build            Build Docker image"
	@echo "  rebuild          Rebuild Docker image (no cache)"
	@echo "  push             Push Docker image to registry"
	@echo ""
	@echo "Monitoring Commands:"
	@echo "  grafana          Open Grafana dashboard"
	@echo "  prometheus       Open Prometheus interface"
	@echo ""
	@echo "Cleanup Commands:"
	@echo "  clean            Stop and remove all containers"
	@echo "  clean-volumes    Remove all data volumes"
	@echo "  clean-images     Remove all service images"
	@echo ""
	@echo "=========================================="

# ============================================
# Development Commands
# ============================================

.PHONY: install
install:
	@echo "$(GREEN)Installing dependencies...$(NC)"
	npm ci
	@echo "$(GREEN)Dependencies installed successfully!$(NC)"

.PHONY: dev
dev: check-env
	@echo "$(GREEN)Starting development environment...$(NC)"
	$(DOCKER_COMPOSE) -f $(COMPOSE_FILE) up -d
	@echo "$(GREEN)Development environment started!$(NC)"
	@echo "$(YELLOW)API: http://localhost:3001$(NC)"
	@echo "$(YELLOW)Grafana: http://localhost:3002$(NC)"

.PHONY: dev-down
dev-down:
	@echo "$(YELLOW)Stopping development environment...$(NC)"
	$(DOCKER_COMPOSE) -f $(COMPOSE_FILE) down
	@echo "$(GREEN)Development environment stopped!$(NC)"

# ============================================
# Production Commands
# ============================================

.PHONY: up
up: check-env build
	@echo "$(GREEN)Starting production services...$(NC)"
	$(DOCKER_COMPOSE) -f $(COMPOSE_FILE) up -d
	@echo "$(GREEN)Production services started!$(NC)"
	@make status

.PHONY: down
down:
	@echo "$(YELLOW)Stopping all services...$(NC)"
	$(DOCKER_COMPOSE) -f $(COMPOSE_FILE) down
	@echo "$(GREEN)All services stopped!$(NC)"

.PHONY: restart
restart: down up
	@echo "$(GREEN)All services restarted!$(NC)"

# ============================================
# Database Commands
# ============================================

.PHONY: db-migrate
db-migrate:
	@echo "$(GREEN)Running database migrations...$(NC)"
	docker exec csic-reporting-api npm run migration:run
	@echo "$(GREEN)Migrations completed!$(NC)"

.PHONY: db-seed
db-seed:
	@echo "$(GREEN)Seeding database...$(NC)"
	docker exec csic-reporting-api npm run seed
	@echo "$(GREEN)Database seeded!$(NC)"

.PHONY: db-backup
db-backup:
	@echo "$(GREEN)Creating database backup...$(NC)"
	mkdir -p backups
	docker exec csic-reporting-db pg_dump -U $(POSTGRES_USER) $(POSTGRES_DB) | gzip > backups/report_$$(date +%Y%m%d_%H%M%S).sql.gz
	@echo "$(GREEN)Backup created in backups/ directory!$(NC)"

.PHONY: db-restore
db-restore:
	@echo "$(YELLOW)Restoring database...$(NC)"
	@read -p "Enter backup file path: " FILE; \
	if [ -f "$$FILE" ]; then \
		docker exec -i csic-reporting-db psql -U $(POSTGRES_USER) $(POSTGRES_DB) < $$FILE; \
		echo "$(GREEN)Database restored successfully!$(NC)"; \
	else \
		echo "$(RED)File not found!$(NC)"; \
	fi

.PHONY: db-console
db-console:
	@echo "$(GREEN)Connecting to database console...$(NC)"
	docker exec -it csic-reporting-db psql -U $(POSTGRES_USER) -d $(POSTGRES_DB)

# ============================================
# Maintenance Commands
# ============================================

.PHONY: logs
logs:
	@echo "$(GREEN)Viewing application logs...$(NC)"
	$(DOCKER_COMPOSE) -f $(COMPOSE_FILE) logs $(SERVICE_NAME)

.PHONY: logs-tail
logs-tail:
	@echo "$(GREEN)Tailing application logs...$(NC)"
	$(DOCKER_COMPOSE) -f $(COMPOSE_FILE) logs -f $(SERVICE_NAME)

.PHONY: status
status:
	@echo "$(GREEN)Checking service status...$(NC)"
	$(DOCKER_COMPOSE) -f $(COMPOSE_FILE) ps
	@echo ""
	@echo "Health Status:"
	@curl -s http://localhost:3001/health | jq . 2>/dev/null || curl -s http://localhost:3001/health

.PHONY: health
health:
	@echo "$(GREEN)Checking service health...$(NC)"
	@echo "API: $$(curl -s http://localhost:3001/health)"
	@echo "Database: $$(docker exec csic-reporting-db pg_isready -U $(POSTGRES_USER) -d $(POSTGRES_DB))"
	@echo "Redis: $$(docker exec csic-reporting-redis redis-cli ping)"
	@echo "Kafka: $$(docker exec csic-reporting-kafka kafka-broker-api-versions --bootstrap-server localhost:9092 > /dev/null 2>&1 && echo 'OK' || echo 'FAILED')"

.PHONY: metrics
metrics:
	@echo "$(GREEN)Fetching service metrics...$(NC)"
	curl -s http://localhost:9091/metrics | head -50

# ============================================
# Build Commands
# ============================================

.PHONY: build
build:
	@echo "$(GREEN)Building Docker image...$(NC)"
	$(DOCKER_COMPOSE) -f $(COMPOSE_FILE) build

.PHONY: rebuild
rebuild:
	@echo "$(GREEN)Rebuilding Docker image (no cache)...$(NC)"
	$(DOCKER_COMPOSE) -f $(COMPOSE_FILE) build --no-cache

.PHONY: push
push:
	@echo "$(GREEN)Pushing Docker image to registry...$(NC)"
	@read -p "Enter registry URL: " REGISTRY; \
	$(DOCKER_COMPOSE) -f $(COMPOSE_FILE) build; \
	docker tag csic-reporting-api:latest $$REGISTRY/csic-reporting-api:latest; \
	docker push $$REGISTRY/csic-reporting-api:latest; \
	echo "$(GREEN)Image pushed successfully!$(NC)"

# ============================================
# Monitoring Commands
# ============================================

.PHONY: grafana
grafana:
	@echo "$(GREEN)Opening Grafana dashboard...$(NC)"
	@if command -v xdg-open > /dev/null; then \
		xdg-open http://localhost:3002; \
	elif command -v open > /dev/null; then \
		open http://localhost:3002; \
	else \
		echo "Open http://localhost:3002 in your browser"; \
	fi

.PHONY: prometheus
prometheus:
	@echo "$(GREEN)Opening Prometheus interface...$(NC)"
	@if command -v xdg-open > /dev/null; then \
		xdg-open http://localhost:9090; \
	elif command -v open > /dev/null; then \
		open http://localhost:9090; \
	else \
		echo "Open http://localhost:9090 in your browser"; \
	fi

# ============================================
# Cleanup Commands
# ============================================

.PHONY: clean
clean:
	@echo "$(YELLOW)Stopping and removing all containers...$(NC)"
	$(DOCKER_COMPOSE) -f $(COMPOSE_FILE) down --remove-orphans
	@echo "$(GREEN)Containers removed!$(NC)"

.PHONY: clean-volumes
clean-volumes: clean
	@echo "$(YELLOW)Removing all data volumes...$(NC)"
	$(DOCKER_COMPOSE) -f $(COMPOSE_FILE) down -v --remove-orphans
	@echo "$(GREEN)All volumes removed!$(NC)"
	@echo "$(YELLOW)WARNING: All data will be lost!$(NC)"

.PHONY: clean-images
clean-images:
	@echo "$(YELLOW)Removing all service images...$(NC)"
	docker rmi $$(docker images -q $(SERVICE_NAME)*) 2>/dev/null || true
	docker rmi $$(docker images -q *reporting* -f reference='*reporting*') 2>/dev/null || true
	@echo "$(GREEN)Images removed!$(NC)"

.PHONY: clean-all
clean-all: clean clean-volumes clean-images
	@echo "$(GREEN)All cleanup complete!$(NC)"

# ============================================
# Utility Commands
# ============================================

.PHONY: check-env
check-env:
	@if [ ! -f $(ENV_FILE) ]; then \
		echo "$(RED)Error: $(ENV_FILE) not found!$(NC)"; \
		echo "$(YELLOW)Copy .env.example to .env and configure values.$(NC)"; \
		exit 1; \
	fi

.PHONY: validate
validate: check-env
	@echo "$(GREEN)Validating configuration...$(NC)"
	$(DOCKER_COMPOSE) -f $(COMPOSE_FILE) config
	@echo "$(GREEN)Configuration is valid!$(NC)"

.PHONY: shell
shell:
	@echo "$(GREEN)Opening shell in service container...$(NC)"
	docker exec -it csic-reporting-api sh

.PHONY: reload
reload:
	@echo "$(GREEN)Reloading service configuration...$(NC)"
	$(DOCKER_COMPOSE) -f $(COMPOSE_FILE) kill -s HUP reporting-api
	@echo "$(GREEN)Service reloaded!$(NC)"

.PHONY: scale
scale:
	@echo "$(GREEN)Scaling service...$(NC)"
	@read -p "Enter number of instances: " COUNT; \
	$(DOCKER_COMPOSE) -f $(COMPOSE_FILE) up -d --scale $(SERVICE_NAME)=$$COUNT; \
	@echo "$(GREEN)Scaled to $$COUNT instances!$(NC)"

# ============================================
# Test Commands
# ============================================

.PHONY: test
test:
	@echo "$(GREEN)Running tests...$(NC)"
	docker exec csic-reporting-api npm test

.PHONY: test-coverage
test-coverage:
	@echo "$(GREEN)Running tests with coverage...$(NC)"
	docker exec csic-reporting-api npm run test:coverage

.PHONY: lint
lint:
	@echo "$(GREEN)Running linter...$(NC)"
	npm run lint

.PHONY: format
format:
	@echo "$(GREEN)Formatting code...$(NC)"
	npm run format
