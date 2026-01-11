# Blockchain Node Manager Service

A comprehensive microservice for managing and monitoring blockchain nodes, built with Go and designed for enterprise-grade blockchain infrastructure operations.

## Overview

The Blockchain Node Manager Service is a critical component of the CSIC Platform that provides centralized management, monitoring, and health checking for blockchain nodes across multiple networks. This service enables operators to register blockchain nodes, track their operational status, collect performance metrics, and receive alerts when issues arise. It serves as the foundational infrastructure layer that ensures reliable connectivity to blockchain networks and provides the visibility needed for effective network operations.

Modern blockchain ecosystems often involve running multiple nodes across different networks, each with their own connection requirements, performance characteristics, and monitoring needs. The Node Manager Service addresses the operational complexity of managing such infrastructure by providing a unified interface for all node-related operations. Whether you need to add a new Ethereum node, monitor Polygon validator performance, or troubleshoot BSC connection issues, this service provides the tools and visibility required for effective blockchain operations.

The service integrates seamlessly with other CSIC Platform components, providing node status information to the Compliance Module for transaction verification, metrics data to the Dashboard for visualization, and event notifications to the Reporting System for audit trails. This integration ensures that node health directly impacts the reliability of all platform operations that depend on blockchain connectivity.

## Features

The Blockchain Node Manager Service offers a comprehensive set of features designed to meet the demanding requirements of blockchain infrastructure operations. Multi-network support enables management of nodes across Ethereum, Polygon, BNB Smart Chain, Arbitrum, and other networks from a single service instance. Each network configuration includes optimized defaults for connection parameters, health check intervals, and recovery procedures, reducing the configuration burden while ensuring best practices are followed.

Real-time health monitoring provides continuous visibility into node operational status. The service performs configurable health checks at regular intervals, verifying RPC connectivity, measuring response latency, and tracking synchronization progress. Status changes are immediately detected and published to the event system, enabling rapid response to connectivity issues or synchronization problems. The health check system supports both active probing and passive event monitoring, providing comprehensive coverage of potential issues.

Performance metrics collection captures detailed information about node resource utilization and network performance. The service tracks block height, peer count, transaction pool size, gas prices, memory and CPU usage, network I/O, and RPC latency. This data is stored for historical analysis and made available through the metrics API for integration with monitoring and alerting systems. The metrics system supports both real-time dashboards and long-term trend analysis.

Automated recovery capabilities help maintain high availability by automatically attempting to restore node connectivity when issues are detected. The service can be configured to attempt automatic restarts, connection retries, and peer rebalancing when nodes go offline or performance degrades. Recovery attempts are logged and can trigger alerts if issues persist, ensuring operators are informed of persistent problems even when automatic recovery is enabled.

The management API provides complete CRUD operations for node registration, configuration updates, and removal. Nodes can be grouped by network and type, with filtering and pagination support for large deployments. The API supports both individual node operations and bulk actions, enabling efficient management of multi-node configurations.

## Architecture

The Node Manager Service follows a clean architecture pattern with clear separation between domain logic, application services, and infrastructure adapters. This design ensures maintainability, testability, and extensibility while keeping the codebase organized and understandable. The architecture enables the service to evolve independently of infrastructure concerns, supporting future enhancements without requiring fundamental changes to the core logic.

The domain layer defines the core business entities including nodes, metrics, health checks, and events. These entities capture the essential properties and behaviors of blockchain node management, providing a stable foundation for the application logic. The domain models are designed to be technology-agnostic, representing concepts that are universally applicable to blockchain node operations regardless of the specific networks being managed.

The application layer contains the service implementations that orchestrate business operations. The Node Service handles all node-related operations including registration, updates, and lifecycle management. The Metrics Service manages collection, storage, and retrieval of performance data. The Health Service coordinates continuous monitoring and status tracking. Each service is designed to be independently testable and can operate without knowledge of the specific infrastructure implementations being used.

The infrastructure layer implements the adapters that connect the application to external systems. Repository implementations provide database access using PostgreSQL for persistent storage and Redis for caching and real-time data. The messaging adapter integrates with Kafka for event streaming to downstream systems. HTTP handlers implement the REST API using the Gin framework, providing clean integration with standard HTTP clients and API gateways.

## Getting Started

### Prerequisites

Before deploying the Node Manager Service, ensure your environment meets the following requirements. The service is tested primarily on Linux systems but should run on any platform supported by Go 1.21 or later. Docker and Docker Compose are required for containerized deployments, while manual deployments require Go 1.21 or later, PostgreSQL 14 or later, Redis 7 or later, and access to a Kafka instance.

You will need network access to the blockchain nodes you intend to manage. For production deployments, dedicate adequate resources including storage for metrics history, network bandwidth for health check traffic, and sufficient memory for caching. The specific resource requirements depend on the number of nodes being managed and the metrics collection frequency.

### Installation

Clone the repository and navigate to the service directory. The service uses Go modules for dependency management, so no additional installation steps are required beyond having Go installed. Dependencies are automatically resolved when building the service.

For containerized deployment, the included Dockerfile builds an optimized image that can be deployed to any container runtime. Build the image using standard Docker commands and tag it appropriately for your container registry. The image is based on a minimal Alpine Linux image with the Go runtime, ensuring a small attack surface and fast startup times.

### Configuration

The service is configured through the `config.yaml` file, which provides a centralized location for all operational parameters. Configuration can be overridden through environment variables for container-friendly deployments where secrets and sensitive values are injected at runtime.

The configuration file is organized into logical sections covering application settings, database connections, Redis configuration, Kafka settings, node manager behavior, and blockchain network definitions. Each blockchain network requires its own configuration block specifying connection details, network parameters, and operational preferences. The default configuration includes settings for major networks including Ethereum Mainnet, Polygon, BNB Smart Chain, Arbitrum, and Optimism.

Database configuration requires connection details for the PostgreSQL instance. The schema is automatically applied on startup through migration files, ensuring the database is properly initialized. Connection pooling parameters can be tuned to match the expected load characteristics of your deployment.

### Running the Service

The service can be started using Docker Compose for development and testing environments. The included `docker-compose.yml` file defines a complete stack including PostgreSQL, Redis, Kafka, and the Node Manager Service. This provides a quick way to get a functioning environment for evaluation purposes.

For production deployments, the service binary should be started directly or managed through your organization's standard service management approach. The service accepts command-line flags for controlling log verbosity and configuration file paths. Health check endpoints are exposed on the API port, enabling integration with container orchestrators and load balancers.

## API Reference

The Node Manager Service exposes a RESTful API for node management, metrics retrieval, and health monitoring. All API endpoints require authentication via API key, which should be included in the `X-API-Key` header of requests. The base URL for API requests is determined by the server configuration, typically `http://localhost:8081` for local deployments.

### Nodes

The nodes endpoints provide access to node management operations. Create a new node by sending a POST request with the node configuration including name, network, RPC URL, and optional metadata. List all nodes with optional filtering by network, status, or type. Retrieve, update, or delete individual nodes using their unique identifiers.

### Networks

The networks endpoints provide information about configured blockchain networks. List all networks to see configured networks along with their active and total node counts. Get detailed information about nodes on a specific network including their status and performance metrics.

### Metrics

The metrics endpoints provide access to performance data. Get the latest metrics for a specific node including block height, peer count, and resource utilization. Retrieve historical metrics for analysis over specified time ranges. Get aggregated metrics for entire networks to understand overall network performance.

### Health

The health endpoints provide system and node health information. Get overall system health including counts of online, offline, and syncing nodes. Get detailed health information for individual nodes including latency measurements and error details. Use the liveness and readiness endpoints for integration with container orchestrators and health monitoring systems.

## Development

### Project Structure

The project follows Go project layout conventions with clear separation between layers. The `cmd` directory contains the entry point for the service. The `internal` directory houses all application code organized by package responsibility. The `migrations` directory contains database migration files.

Domain models reside in `internal/domain`, defining the data structures and interfaces that represent core business concepts. Repository implementations in `internal/repository` handle database operations. Service implementations in `internal/service` contain the business logic for node management and monitoring. HTTP handlers in `internal/handler` implement the REST API. The messaging package in `internal/messaging` handles Kafka integration.

### Building

The project uses Go modules for dependency management. Build the service by running `go build -o bin/nodes-service ./cmd` from the project root. This produces a statically linked binary suitable for deployment in containerized environments. The build process supports several compile-time options for features that may require additional dependencies.

Run tests using `go test ./...` to execute the test suite. The project maintains test coverage for critical paths including business logic, repository operations, and API handlers. Integration tests verify database and messaging interactions in test environments.

### Extending

Adding support for new blockchain networks requires adding configuration to `config.yaml`. No code changes are typically required for new networks as the service uses a generic node management approach that works with any JSON-RPC compatible blockchain node.

Custom health checks can be implemented by extending the Health Service with specialized check logic. This is useful for networks with non-standard health verification requirements or for adding application-specific monitoring beyond basic RPC connectivity.

API extensions should follow the established patterns in the handler layer. New endpoints are added by implementing handler methods and registering routes in the application setup code. Authentication and validation middleware can be applied to protect sensitive endpoints.

## Contributing

Contributions to the Node Manager Service are welcome and encouraged. Before beginning work on a significant feature or bug fix, please open an issue to discuss the proposed changes. This helps ensure that efforts align with the project's direction and prevents duplication of work.

Code contributions should follow the project's coding standards emphasizing clarity, testability, and documentation. All contributions must pass the automated test suite and lint checks before being merged. The project uses conventional commit messages for changelog generation.

## License

The Blockchain Node Manager Service is proprietary software. All rights are reserved. Use of this software is subject to the terms and conditions established in your licensing agreement. Contact the project maintainers for licensing inquiries.
