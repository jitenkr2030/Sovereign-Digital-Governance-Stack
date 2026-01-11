# NEAM - National Economic Activity Monitor
# Build and Operations Makefile

# Version and Configuration
VERSION := 1.0.0
BUILD_DATE := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
GIT_COMMIT := $(shell git rev-parse HEAD 2>/dev/null || echo "unknown")
GIT_TAG := $(shell git describe --tags --always 2>/dev/null || echo "unknown")

# Directories
BINARY_DIR := bin
BUILD_DIR := build
DIST_DIR := dist
COVERAGE_DIR := coverage
TEST_REPORTS_DIR := test-reports

# Go Configuration
GO := go
GOFLAGS := -mod=mod
LDFLAGS := -X main.version=$(VERSION) -X main.buildDate=$(BUILD_DATE) -X main.gitCommit=$(GIT_COMMIT)
CGO_ENABLED := 0

# Module directories
SENSING_MODULES := $(wildcard sensing/*/)
INTELLIGENCE_MODULES := $(wildcard intelligence/*/)
INTERVENTION_MODULES := $(wildcard intervention/*/)
REPORTING_MODULES := $(wildcard reporting/*/)
SECURITY_MODULES := $(wildcard security/*/)

# Docker Configuration
DOCKER_REGISTRY ?= neamacr.azurecr.io
DOCKER_IMAGE_PREFIX := neam
DOCKER_TAG := $(VERSION)-$(GIT_COMMIT)

# Kubernetes Configuration
K8S_NAMESPACE := neam-system
K8S_CONTEXT ?= $(shell kubectl config current-context 2>/dev/null || echo "")

# Terraform Configuration
TERRAFORM_DIR := deployments/terraform
TERRAFORM_STATE ?= tfstate/neam.tfstate

# Default Target
.PHONY: help
help:
	@echo "NEAM - National Economic Activity Monitor"
	@echo "=========================================="
	@echo ""
	@echo "Build Targets:"
	@echo "  all              - Build all components (default)"
	@echo "  clean            - Clean build artifacts"
	@echo "  binaries         - Build all binary executables"
	@echo "  docker           - Build all Docker images"
	@echo "  helm             - Package Helm charts"
	@echo ""
	@echo "Test Targets:"
	@echo "  test             - Run all tests"
	@echo "  test-unit        - Run unit tests"
	@echo "  test-integration - Run integration tests"
	@echo "  test-load        - Run load tests"
	@echo "  test-security    - Run security tests"
	@echo "  coverage         - Generate code coverage reports"
	@echo ""
	@echo "Deployment Targets:"
	@echo "  deploy           - Deploy to Kubernetes"
	@echo "  deploy-dry-run   - Dry-run deployment"
	@echo "  rollback         - Rollback deployment"
	@echo "  status           - Show deployment status"
	@echo ""
	@echo "Infrastructure Targets:"
	@echo "  terraform-plan   - Plan Terraform changes"
	@echo "  terraform-apply  - Apply Terraform changes"
	@echo "  terraform-destroy- Destroy Terraform infrastructure"
	@echo ""
	@echo "Maintenance Targets:"
	@echo "  lint             - Run linters"
	@echo "  fmt              - Format code"
	@echo "  security-scan    - Run security scanning"
	@echo "  dependencies     - Update dependencies"
	@echo "  generate         - Generate code from specs"
	@echo ""

# Build Targets
.PHONY: all
all: clean binaries docker helm

.PHONY: clean
clean:
	rm -rf $(BINARY_DIR) $(BUILD_DIR) $(DIST_DIR) $(COVERAGE_DIR) $(TEST_REPORTS_DIR)
	find . -type d -name vendor -exec rm -rf {} + 2>/dev/null || true
	find . -type f -name "*.test" -delete 2>/dev/null || true
	find . -type f -name "*.cover" -delete 2>/dev/null || true
	find . -type f -name "*.tmp" -delete 2>/dev/null || true

.PHONY: binaries
binaries: $(BINARY_DIR)
	$(MAKE) -C services all
	$(MAKE) -C intelligence all
	$(MAKE) -C intervention all
	$(MAKE) -C sensing all
	$(MAKE) -C reporting all
	$(MAKE) -C security all

$(BINARY_DIR):
	mkdir -p $(BINARY_DIR)

# Docker Targets
.PHONY: docker
docker: docker-build-all

.PHONY: docker-build-all
docker-build-all:
	$(MAKE) -C services docker
	$(MAKE) -C intelligence docker
	$(MAKE) -C intervention docker
	$(MAKE) -C sensing docker
	$(MAKE) -C reporting docker
	$(MAKE) -C security docker
	$(MAKE) -C frontend docker

.PHONY: docker-push-all
docker-push-all:
	$(MAKE) -C services docker-push
	$(MAKE) -C intelligence docker-push
	$(MAKE) -C intervention docker-push
	$(MAKE) -C sensing docker-push
	$(MAKE) -C reporting docker-push
	$(MAKE) -C security docker-push
	$(MAKE) -C frontend docker-push

.PHONY: docker-login
docker-login:
	@if [ -n "$(DOCKER_REGISTRY)" ]; then \
		docker login $(DOCKER_REGISTRY); \
	fi

# Helm Targets
.PHONY: helm
helm:
	$(MAKE) -C deployments/helm package

# Test Targets
.PHONY: test
test: test-unit test-integration test-load test-security

.PHONY: test-unit
test-unit:
	$(MAKE) -C tests/unit all

.PHONY: test-integration
test-integration:
	$(MAKE) -C tests/integration all

.PHONY: test-load
test-load:
	$(MAKE) -C tests/load all

.PHONY: test-security
test-security:
	$(MAKE) -C tests/security all

.PHONY: coverage
coverage:
	mkdir -p $(COVERAGE_DIR)
	$(MAKE) -C tests/coverage all

# Deployment Targets
.PHONY: deploy
deploy: $(K8S_NAMESPACE)
	@echo "Deploying NEAM to Kubernetes..."
	kubectl apply -k deployments/k8s/overlays/$(ENVIRONMENT) --context $(K8S_CONTEXT)
	@echo "Deployment initiated. Use 'make status' to check progress."

$(K8S_NAMESPACE):
	kubectl create namespace $(K8S_NAMESPACE) --dry-run=client -o yaml | kubectl apply -f -

.PHONY: deploy-dry-run
deploy-dry-run:
	kubectl apply -k deployments/k8s/overlays/$(ENVIRONMENT) --dry-run=client

.PHONY: rollback
rollback:
	@echo "Rolling back NEAM deployment..."
	kubectl rollout undo deployment/$(DEPLOYMENT_NAME) -n $(K8S_NAMESPACE) --context $(K8S_CONTEXT)

.PHONY: status
status:
	@echo "NEAM Deployment Status"
	@echo "======================"
	kubectl get all -n $(K8S_NAMESPACE) --context $(K8S_CONTEXT)
	kubectl get pods -n $(K8S_NAMESPACE) --context $(K8S_CONTEXT) -o wide

# Terraform Targets
.PHONY: terraform-init
terraform-init:
	cd $(TERRAFORM_DIR) && terraform init -upgrade

.PHONY: terraform-plan
terraform-plan: terraform-init
	cd $(TERRAFORM_DIR) && terraform plan -out=$(TERRAFORM_STATE).plan

.PHONY: terraform-apply
terraform-apply: terraform-init
	cd $(TERRAFORM_DIR) && terraform apply $(TERRAFORM_STATE).plan

.PHONY: terraform-destroy
terraform-destroy:
	cd $(TERRAFORM_DIR) && terraform destroy -auto-approve

.PHONY: terraform-output
terraform-output:
	cd $(TERRAFORM_DIR) && terraform output

# Maintenance Targets
.PHONY: lint
lint:
	@echo "Running linters..."
	golangci-lint run ./...
	eslint frontend/src/
	helm lint deployments/helm/

.PHONY: fmt
fmt:
	@echo "Formatting code..."
	gofmt -w services/ intelligence/ intervention/ sensing/ reporting/ security/
	npx prettier --write frontend/src/

.PHONY: security-scan
security-scan:
	@echo "Running security scans..."
	trivy fs --exit-code 1 .
	snyk test --severity-threshold=high
	syft packages . -o table

.PHONY: dependencies
dependencies:
	@echo "Updating dependencies..."
	go mod tidy
	go mod verify
	npm update --prefix frontend/

.PHONY: generate
generate:
	@echo "Generating code from specs..."
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative api/**/*.proto
	openapi-generator-cli generate -i docs/api/openapi.yaml -g go -o services/api-client
	typescript-json-schema --ref --id schema frontend/src/types/*.ts --out shared/types/schema.json

# Development Targets
.PHONY: dev-up
dev-up:
	docker-compose -f deployments/docker-compose.yml up -d

.PHONY: dev-down
dev-down:
	docker-compose -f deployments/docker-compose.yml down

.PHONY: dev-logs
dev-logs:
	docker-compose -f deployments/docker-compose.yml logs -f

.PHONY: mock-data
mock-data:
	@echo "Generating mock economic data..."
	python scripts/mocks/generate_mock_data.py --days 90 --output data/mock/

.PHONY: benchmark
benchmark:
	@echo "Running performance benchmarks..."
	k6 run tests/load/benchmark.js
	hey -c 100 -n 10000 http://localhost:8080/health

# Documentation Targets
.PHONY: docs
docs:
	@echo "Generating documentation..."
	godoc -http=:6060 &
	glow README.md
	@echo "Documentation server started on port 6060"

# Release Targets
.PHONY: release
release: tag-and-release
	$(MAKE) docker-push-all
	$(MAKE) helm
	$(MAKE) -C scripts/release create-release

.PHONY: tag-and-release
tag-and-release:
	git tag v$(VERSION)
	git push origin v$(VERSION)

.PHONY: changelog
changelog:
	git-changelog -o CHANGELOG.md --issue-pattern '#\d+' --commitpattern '(\w+):\s*(.*)'

# Monitoring Targets
.PHONY: metrics
metrics:
	@echo "Collecting metrics..."
	curl -s http://localhost:9090/api/v1/query?query=up | jq .
	prometheus metrics

.PHONY: dashboards
dashboards:
	@echo "Importing Grafana dashboards..."
	kubectl apply -f monitoring/dashboards/ -n monitoring

# Backup and Recovery Targets
.PHONY: backup
backup:
	@echo "Creating backup..."
	kubectl exec -n $(K8S_NAMESPACE) neam-postgresql-0 -- pg_dump -U postgres neam > backups/neam_$(shell date +%Y%m%d_%H%M%S).sql
	kubectl exec -n $(K8S_NAMESPACE) neam-redis-0 -- save > /dev/null 2>&1
	kubectl cp $(K8S_NAMESPACE)/neam-redis-0:data/dump.rdb backups/redis_$(shell date +%Y%m%d_%H%M%S).rdb

.PHONY: restore
restore:
	@echo "Restoring from backup..."
	@read -p "Enter backup file: " BACKUP_FILE; \
	kubectl exec -n $(K8S_NAMESPACE) neam-postgresql-0 -- psql -U postgres neam < $$BACKUP_FILE

# Utility Targets
.PHONY: version
version:
	@echo "NEAM Version: $(VERSION)"
	@echo "Build Date: $(BUILD_DATE)"
	@echo "Git Commit: $(GIT_COMMIT)"
	@echo "Git Tag: $(GIT_TAG)"

.PHONY: env
env:
	@echo "Current Environment Configuration"
	@echo "=================================="
	@echo "DOCKER_REGISTRY: $(DOCKER_REGISTRY)"
	@echo "DOCKER_TAG: $(DOCKER_TAG)"
	@echo "K8S_NAMESPACE: $(K8S_NAMESPACE)"
	@echo "K8S_CONTEXT: $(K8S_CONTEXT)"
	@echo "ENVIRONMENT: $(ENVIRONMENT)"

.PHONY: health
health:
	@echo "Health Check"
	@echo "============"
	@echo "Checking service health..."
	curl -s http://localhost:8080/health | jq .
	curl -s http://localhost:9090/-/healthy
	kubectl get pods -n $(K8S_NAMESPACE) --context $(K8S_CONTEXT) | grep -v "Running" || echo "All pods running"

.PHONY: logs
logs:
	@echo "Collecting logs..."
	@read -p "Service name: " SERVICE; \
	kubectl logs -n $(K8S_NAMESPACE) -l app=$$SERVICE --tail=100 --timestamps

.PHONY: shell
shell:
	@echo "Opening shell in pod..."
	@read -p "Pod name: " POD; \
	kubectl exec -it -n $(K8S_NAMESPACE) $$POD -- /bin/sh

.PHONY: debug
debug:
	@echo "Debug information"
	@echo "================="
	kubectl describe pods -n $(K8S_NAMESPACE) --context $(K8S_CONTEXT)
	kubectl get events -n $(K8S_NAMESPACE) --sort-by='.lastTimestamp' | tail -20

# Include environment-specific configuration
-include .env
-include Makefile.local

# Color output for terminal
RED := \033[0;31m
GREEN := \033[0;32m
YELLOW := \033[1;33m
NC := \033[0m # No Color

print-success:
	@echo -e "$(GREEN)✓ $1$(NC)"

print-warning:
	@echo -e "$(YELLOW)! $1$(NC)"

print-error:
	@echo -e "$(RED)✗ $1$(NC)"
