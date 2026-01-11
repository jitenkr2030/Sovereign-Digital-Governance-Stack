# Transaction Monitoring & Risk Intelligence Service

A comprehensive blockchain transaction monitoring and risk analysis system for the CSIC Platform. This service ingests real-time blockchain data, analyzes transactions for suspicious patterns, calculates wallet risk scores, and provides entity clustering through graph analysis.

## Overview

The Transaction Monitoring Service enables regulatory authorities to detect and investigate cryptocurrency-related illicit activities including money laundering, sanctions violations, and darknet market transactions. The system provides real-time monitoring capabilities, automated risk scoring, and comprehensive case management tools for compliance officers and law enforcement investigators.

Core capabilities include blockchain transaction ingestion from multiple networks, deterministic rule-based risk scoring with configurable thresholds, graph-based entity clustering to identify related wallet addresses, and complete investigation case management with audit trails. The service integrates with other CSIC platform components including the Wallet Governance Service for wallet status information and the Audit Log Service for comprehensive activity logging.

## Features

### Blockchain Ingestion

The service connects to multiple blockchain networks to ingest transactions as they occur. It normalizes transaction data from different blockchain formats into a unified structure that can be processed by downstream services. The ingestion system supports both real-time block streaming via WebSocket connections and historical backfilling for catching up on missed blocks. Transaction parsing includes extraction of sender and receiver addresses, token transfers for ERC-20 and similar tokens, gas consumption metrics, and block metadata for verification purposes.

The service supports real-time blockchain data ingestion from major cryptocurrency networks. Bitcoin network monitoring captures all transactions including UTXO details, while Ethereum monitoring tracks all transaction types including internal calls. The system normalizes blockchain-specific data formats into a unified schema regardless of source network, enabling consistent analysis across all monitored cryptocurrencies. Ingestion uses efficient polling and subscription mechanisms to achieve near-real-time block processing with configurable confirmation requirements.

### Wallet Risk Scoring

Every wallet address that interacts with the platform is assigned a dynamic risk score between 0 and 100. The risk scoring engine evaluates multiple factors including historical transaction patterns, velocity of recent activity, links to known risky entities, transaction pattern analysis for suspicious behaviors like structuring or layering, and sanctions screening results. The scoring algorithm uses weighted combinations of these factors to produce actionable risk assessments that enable compliance teams to focus their attention on the highest-risk activity.

A deterministic rule-based risk scoring engine evaluates each transaction against configurable risk factors. Direct exposure scoring identifies transactions involving known sanctioned addresses or darknet market wallets. Behavioral analysis detects patterns such as structuring, velocity anomalies, and round number transactions. The system calculates composite risk scores on a 0-100 scale with automatic critical alert generation when thresholds are exceeded. Risk scores persist with full calculation transparency for audit purposes.

The algorithm produces scores between 0 and 100, with thresholds defined in the configuration for critical (80+), high (60+), medium (40+), and low (below 40) risk levels. The risk scoring algorithm calculates a composite score from multiple weighted factors. Historical factors including wallet age and transaction history contribute 25% of the total score. Velocity factors including recent transaction volume and frequency contribute 25% of the total score. Pattern factors for detecting structuring and layering contribute 20% of the total score. Link analysis factors for connections to risky entities contribute 15% of the total score. Sanctions status contributes 10% of the total score, with a value of 100 if the wallet is sanctioned. Behavioral factors for detecting anomalies contribute the remaining 5%.

### Entity Clustering

The clustering service identifies groups of wallet addresses that likely belong to the same entity. It employs several heuristics to make these determinations including common input detection for Bitcoin transactions where multiple inputs typically belong to the same wallet owner, deposit address clustering for identifying wallets that deposit to common exchange hot wallets, change address linking for recognizing addresses that receive change from transactions, and behavioral clustering for grouping wallets with similar transaction patterns. The resulting clusters enable investigators to see the full scope of an entity's activity across many addresses.

Advanced graph analysis algorithms cluster related wallet addresses to identify likely common owners. The common input heuristic groups Bitcoin addresses that appear as transaction inputs together, indicating common ownership. Deposit address clustering links multiple external addresses that send funds to the same exchange deposit address. Transaction flow analysis traces fund movements across multiple hops to identify complex money laundering patterns. Cluster relationships are stored in Neo4j for efficient graph traversal queries.

### Sanctions Screening

The service maintains cached copies of major sanctions lists including OFAC SDN, UN consolidated, and EU sanctions lists. Every wallet address is screened against these lists with exact matching and fuzzy matching capabilities. The system also detects indirect links by checking if a wallet has transacted with a sanctioned entity within configurable relationship depth limits. When sanctions matches are detected, critical priority alerts are generated immediately for compliance review.

The screening process includes exact matching for direct address matches against the sanctions lists, fuzzy matching for addresses with slight variations, and indirect link detection for wallets that have transacted with sanctioned entities. The sanctions screening cache is refreshed daily to ensure compliance with the latest sanctions designations.

### Case Management

Comprehensive investigation case management supports the complete analyst workflow from alert triage through case resolution. Alert queuing presents new risk alerts with filtering and prioritization capabilities. Investigation workspaces capture analyst notes, evidence documents, and timeline events. Case escalation pathways route high-priority cases to senior investigators or law enforcement. Report generation creates compliance reports and law enforcement briefs with exportable formats.

## Architecture

The service follows Clean Architecture principles with clear separation between layers. The domain layer contains all business entities and validation rules. The service layer implements business logic including risk calculation algorithms and clustering heuristics. The repository layer abstracts data access for both relational and graph databases. The handler layer provides RESTful API endpoints for external integration.

The service implements a hexagonal architecture (also known as Ports and Adapters) to achieve clean separation between core business logic and external dependencies. The domain layer contains pure business logic including entity models, value objects, domain services, and business rules without any dependencies on frameworks or external systems. The application layer coordinates domain objects to fulfill use cases, implementing service interfaces defined by the domain layer and handling transaction boundaries and security context. The infrastructure layer provides concrete adapters for database access, message queue communication, REST API exposure, and blockchain node connectivity.

The technology stack includes Go 1.21, chosen for its strong performance characteristics, excellent concurrency model, and mature ecosystem for building high-throughput services. PostgreSQL serves as the primary data store for persistent state including transactions, wallets, clusters, and alerts. Redis provides caching for risk scores, velocity tracking, and rate limiting operations. Apache Kafka enables asynchronous event streaming between services for scalability and loose coupling. Gin provides the HTTP framework for the REST API with middleware support for logging, CORS, and authentication.

```
transaction/monitoring/
├── cmd/
│   └── server/main.go                # Application entry point
├── config/                           # Configuration management
├── internal/
│   ├── config/                       # Configuration types
│   ├── db/migrations/                # Database schema
│   ├── domain/models/                # Business entities
│   │   ├── transaction.go            # Transaction models
│   │   ├── wallet.go                 # Wallet models
│   │   ├── cluster.go                # Cluster models
│   │   └── alert.go                  # Alert models
│   ├── handler/                      # HTTP API handlers
│   │   ├── http/                     # REST API handlers
│   │   └── kafka/                    # Kafka consumers
│   ├── repository/                   # Data access layer
│   │   ├── repository.go             # PostgreSQL repositories
│   │   └── cache_repository.go       # Redis cache
│   ├── service/                      # Business logic
│   │   ├── ingest/                   # Blockchain ingestion
│   │   ├── risk/                     # Risk scoring
│   │   ├── graph/                    # Entity clustering
│   │   └── sanctions/                # Sanctions screening
│   └── monitoring/                   # Prometheus metrics
├── deploy/
│   ├── prometheus/                   # Prometheus configuration
│   │   ├── prometheus.yml
│   │   └── rules/
│   │       └── tx-monitor-alerts.yml
│   └── grafana/                      # Grafana dashboards
├── Dockerfile                        # Container image
├── docker-compose.yml                # Local environment
├── go.mod                            # Go module
└── README.md                         # Documentation
└── deploy/                        # Container configuration
```

## Getting Started

### Prerequisites

Deployment requires Go 1.21 or higher for local development. PostgreSQL 16 serves as the primary relational database. Neo4j 5.x provides graph database capabilities for entity clustering. Bitcoin Core node and/or Geth node are required for blockchain data ingestion. Redis enables high-speed caching of risk lists and recent transactions.

### Configuration

The service reads configuration from `internal/config/config.yaml`. Environment variables override file-based settings in containerized deployments. Critical configuration includes database connection strings for PostgreSQL and Neo4j, blockchain node RPC endpoints, risk scoring thresholds, and alert notification settings.

### Running the Service

Development mode launches the service with hot reload capabilities using Docker Compose:

```bash
docker-compose -f internal/deploy/docker-compose.yml up -d
```

The service exposes several endpoints after startup. The health check at `/health` confirms service availability. Metrics endpoints at `/metrics` provide Prometheus-formatted monitoring data. API documentation is available at `/swagger` when enabled.

### Quick Start Checklist

Verify PostgreSQL connection and ensure the monitoring database exists. Start Neo4j and confirm the audit database accepts connections. Configure blockchain node endpoints in the configuration file. Run database migrations to create required tables and indexes. Start the service and verify health check passes. Monitor ingestion pipeline to confirm blockchain data is being processed.

## API Reference

### Transaction Endpoints

The transaction API provides endpoints for transaction queries and analysis. Retrieve transaction details by hash using GET `/api/v1/transactions/:tx_hash`. List transactions with filtering using GET `/api/v1/transactions` with query parameters for address, chain, time range, and risk score. Retrieve related transactions for graph analysis using GET `/api/v1/transactions/:tx_hash/related`.

### Risk Endpoints

Risk-related endpoints provide access to scoring and alerts. Get wallet risk profile including score breakdown using GET `/api/v1/risk/wallets/:address`. List active risk alerts with filtering using GET `/api/v1/risk/alerts`. Acknowledge or escalate alerts using POST `/api/v1/risk/alerts/:alert_id/acknowledge`.

### Graph Endpoints

Graph analysis endpoints support entity clustering and relationship queries. Get entity cluster details including all related addresses using GET `/api/v1/graph/clusters/:cluster_id`. Query transaction flow between addresses using POST `/api/v1/graph/flow` with source and target addresses. Retrieve graph statistics for monitoring using GET `/api/v1/graph/stats`.

### Case Endpoints

Case management endpoints support investigation workflows. List cases with filtering by status and priority using GET `/api/v1/cases`. Create new investigation case using POST `/api/v1/cases`. Add evidence to case using POST `/api/v1/cases/:case_id/evidence`. Generate case report using GET `/api/v1/cases/:case_id/report`.

## Risk Scoring

The risk scoring engine evaluates transactions against multiple risk factors to calculate a composite risk score. Scores range from 0 (no risk) to 100 (critical risk). The default critical threshold is 75, generating automatic alerts for transactions exceeding this level.

Risk factors are organized into categories with configurable weights. Direct exposure factors include sanctions list matches scoring 100 points, darknet market interaction scoring 80 points, and known exchange hacks scoring 60 points. Indirect exposure factors include first-hop connection to bad addresses scoring 40 points and second-hop connection scoring 20 points. Behavioral factors include high transaction velocity scoring 20 points, structuring patterns scoring 30 points, and round number transactions scoring 10 points.

The calculation uses a weighted sum model: Score = Sum of applicable factor scores, capped at 100. Scores are stored with full factor breakdown for audit purposes and can be recalculated when risk configurations change.

## Entity Clustering

The graph analysis service clusters related wallet addresses using multiple heuristics. Common input clustering identifies addresses appearing together as transaction inputs, indicating common ownership based on the assumption that a single entity controls all inputs to a transaction. Deposit address clustering links external addresses sending to the same exchange deposit address, indicating they may belong to the same user. Change address linking tracks change outputs to identify wallet software behavior and link related addresses.

Clustering runs as scheduled batch jobs against the Neo4j database. The common input heuristic runs hourly for Bitcoin data. Deposit clustering runs daily for exchange monitoring. Results persist with cluster membership timestamps for audit purposes.

## Integration

The service integrates with other CSIC platform components for comprehensive monitoring. Wallet Governance integration provides wallet status including freeze and blacklist information for risk scoring. Exchange Surveillance integration identifies exchange-related transactions for enhanced monitoring. Audit Log integration records all significant actions for compliance and forensic purposes.

External integrations support standard protocols. Webhook notifications deliver real-time alerts to external systems. SIEM integration exports security events in standard formats. Law enforcement API provides secure case data sharing with authorized agencies.

## Monitoring and Observability

The service exposes comprehensive metrics for operational monitoring. Ingestion metrics track block processing latency, transaction throughput, and queue depths. Risk scoring metrics monitor calculation latency, score distributions, and alert volumes. Graph metrics track cluster sizes, traversal latency, and pattern detections.

Grafana dashboards provide visualization for operational monitoring. Real-time transaction flow shows ingestion rates and blockchain latency. Risk score distribution displays alert volumes by severity. Graph topology visualization shows entity cluster sizes and relationships.

## Security Considerations

All API endpoints require authentication and authorization. Role-based access control limits sensitive operations to authorized personnel. Audit logging captures all API calls with user identification. Data encryption protects sensitive information at rest and in transit.

## License

Part of CSIC Platform - All rights reserved.
