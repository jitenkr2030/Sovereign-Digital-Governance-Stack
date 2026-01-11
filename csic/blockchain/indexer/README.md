# Blockchain Indexer Service

A high-performance microservice for indexing and querying blockchain data, built with Go and designed for scalability and reliability in enterprise environments.

## Overview

The Blockchain Indexer Service is a critical component of the CSIC Platform that continuously monitors blockchain networks, extracts relevant transaction and state data, and makes it readily available through a performant API. This service is essential for maintaining real-time visibility into blockchain activities, supporting compliance operations, and enabling sophisticated analytics and reporting capabilities.

This service is designed to handle high-throughput blockchain data ingestion while maintaining low-latency query responses. It supports multiple blockchain networks and provides flexible indexing strategies that can be configured to suit various use cases, from simple transaction tracking to complex smart contract event analysis. The architecture follows clean design principles, ensuring maintainability and extensibility as blockchain ecosystems evolve.

The indexer operates by connecting to blockchain nodes, either through direct RPC connections or by subscribing to node event streams, depending on the network's capabilities. It processes blocks and transactions in real-time, extracting and transforming data according to predefined schemas before persisting it to the database. This data then becomes available for querying through the service's RESTful API, enabling downstream applications to build rich user experiences on top of blockchain activity.

## Features

The Blockchain Indexer Service offers a comprehensive set of features designed to meet the demanding requirements of enterprise blockchain applications. Real-time block processing ensures that data is available within seconds of on-chain activity, enabling applications to react immediately to important events. The service supports multiple blockchain networks simultaneously, allowing organizations to monitor diverse ecosystems from a single deployment.

Smart contract event indexing provides deep visibility into decentralized applications and token activities. The service can parse arbitrary contract ABIs and extract specific events, filtering noise to focus on the data that matters most to your application. This capability is particularly valuable for DeFi protocols, NFT platforms, and any application that relies on understanding complex smart contract interactions.

Transaction tracking and tracing capabilities allow for detailed analysis of fund flows and contract interactions. The service maintains comprehensive transaction metadata, including gas consumption, status, and linked events, enabling forensic analysis and compliance verification. Address monitoring features support watchlist functionality, alerting applications when monitored addresses interact with the blockchain.

A powerful query API provides flexible access to all indexed data. Complex filters, pagination, and aggregation capabilities enable efficient data retrieval without overwhelming network resources. The API supports both synchronous queries for interactive applications and webhook notifications for event-driven architectures.

## Architecture

The Blockchain Indexer Service is built using a layered architecture that separates concerns and enables independent scaling of components. This design choice facilitates maintenance and allows the service to evolve without requiring wholesale changes to the codebase.

At the foundation lies the configuration layer, which loads settings from `config.yaml` and environment variables. This approach supports containerized deployments and enables the same binary to be configured differently across environments. The configuration system supports hot reload for certain parameters, minimizing the need for restarts during operational adjustments.

The domain layer defines the core business entities and interfaces that govern the service's behavior. This includes block and transaction representations, event models, and repository interfaces. By isolating domain logic, the service remains focused on its core purpose and is protected from infrastructure concerns bleeding into business code.

The application layer contains the services and use cases that orchestrate business operations. Indexer services manage the blockchain data ingestion pipelines, processing blocks and events according to configured strategies. Query services handle data retrieval operations, applying caching and optimization techniques to ensure responsive performance.

The infrastructure layer implements the adapters and repositories that connect the service to external systems. Database repositories handle persistence using PostgreSQL with TimescaleDB extensions for time-series optimization. Blockchain adapters connect to node APIs using configurable providers that abstract network-specific details. Message publishers integrate with Kafka for event distribution to downstream systems.

## Getting Started

### Prerequisites

Before deploying the Blockchain Indexer Service, ensure that your environment meets the following requirements. The service is tested primarily on Linux systems, though it should run on any platform supported by the Go runtime. Docker and Docker Compose are required for containerized deployments, while manual deployments require Go 1.21 or later, PostgreSQL 14 or later, and access to a Kafka instance.

You will need network access to the blockchain nodes you intend to index. For production deployments, dedicated infrastructure is recommended, including sufficient storage for historical data and network bandwidth to maintain synchronization with network tip. The specific resource requirements depend heavily on the blockchains being indexed and the indexing depth required.

### Installation

To set up the Blockchain Indexer Service, begin by cloning the repository and navigating to the service directory. The service uses Go modules for dependency management, so no additional installation steps are required beyond having Go installed. Dependencies are automatically resolved when building the service.

For containerized deployment, the included Dockerfile builds an optimized image that can be deployed to any container runtime. The image is based on a minimal Go runtime image and includes all necessary components for operation. Build the image using Docker's standard build commands, tagging it appropriately for your container registry.

Manual deployment requires compiling the binary and ensuring all configuration files are in place. The service is a single binary with no external runtime dependencies beyond the database and message queue infrastructure. This simplicity makes it straightforward to deploy using configuration management tools or container orchestration systems.

### Configuration

The service is configured through the `config.yaml` file, which provides a centralized location for all operational parameters. Configuration can be overridden through environment variables, enabling container-friendly deployments where secrets and sensitive values are injected at runtime.

The configuration file is organized into logical sections covering database connections, blockchain network definitions, indexing behavior, and API server settings. Each blockchain network requires its own configuration block specifying connection details, network parameters, and indexing preferences. The service supports both HTTP and WebSocket connections to blockchain nodes, automatically selecting the appropriate protocol based on network capabilities.

Database configuration requires connection details for the PostgreSQL instance, including host, port, credentials, and database name. The schema is automatically applied on startup through migration files, ensuring the database is properly initialized. Connection pooling parameters can be tuned to match the expected load characteristics of your deployment.

### Running the Service

The service can be started using Docker Compose for development and testing environments. The included `docker-compose.yml` file defines a complete stack including PostgreSQL, Kafka, and the indexer service. This provides a quick way to get a functioning environment for evaluation purposes.

For production deployments, the service binary should be started directly or managed through your organization's standard service management approach. The service accepts command-line flags for controlling log verbosity and configuration file paths. Health check endpoints are exposed on the API port, enabling integration with container orchestrators and load balancers.

Monitor the service logs during initial startup to verify successful database migration and blockchain connection establishment. The indexer will begin processing blocks from its configured starting point, which can be a specific block height or the genesis block for full historical indexing. Progress is logged regularly, providing visibility into the synchronization status.

## API Reference

The Blockchain Indexer Service exposes a RESTful API for querying indexed data and managing indexing operations. All API endpoints require authentication via API key, which should be included in the `X-API-Key` header of requests. The base URL for API requests is determined by the server configuration, typically `http://localhost:8080` for local deployments.

### Blocks

The blocks endpoints provide access to indexed block data. The primary endpoint for retrieving blocks accepts a block height or hash parameter, returning comprehensive block information including transactions, gas consumption, and miner details. The response structure is normalized across supported blockchains, abstracting network-specific variations in block format.

Block listings support pagination and filtering by time range. This is particularly useful for generating reports or analyzing blockchain activity over specific periods. The response includes pagination metadata enabling efficient traversal of large result sets without memory overhead.

### Transactions

Transaction endpoints expose detailed transaction data including inputs, outputs, value transfers, and execution status. Transactions can be retrieved by their hash identifier, with responses including all available metadata such as gas consumption, nonce, and linked events. This granular data supports both display requirements and analytical use cases.

Transaction search capabilities enable filtering by address, time range, or value characteristics. Complex queries can be constructed using the query parameter syntax, supporting combinations of filters for precise data retrieval. Results are returned in a consistent format regardless of the source blockchain network.

### Events

Event endpoints provide access to smart contract event logs that have been indexed by the service. Events can be filtered by contract address, event signature, or block range. This capability is essential for applications that need to track specific on-chain activities such as token transfers, governance actions, or protocol interactions.

The event indexing system supports dynamic ABI loading, enabling the service to decode event parameters without requiring preconfiguration of contract schemas. Decoded events include both the raw log data and the human-readable parameter values, facilitating both programmatic processing and user interface display.

### Monitoring

Health and status endpoints provide operational visibility into the service's state. The health endpoint returns the service's current operational status and connectivity to dependencies. Metrics endpoints expose Prometheus-compatible metrics for integration with monitoring infrastructure.

Blockchain synchronization status is available through dedicated endpoints, showing current block height, indexed height, and synchronization progress. This information is valuable for determining data freshness and identifying synchronization issues before they impact downstream applications.

## Development

### Project Structure

The project follows Go project layout conventions with clear separation between layers. The `cmd` directory contains the entry point for the service, while `internal` houses all application code organized by package responsibility. This structure prevents accidental import cycles and clarifies the public API surface of each package.

Domain models reside in `internal/domain`, defining the data structures and interfaces that represent core business concepts. The repository interfaces defined here provide abstraction over database implementations, enabling different storage backends without changes to application code. Service implementations in `internal/service` contain the business logic for indexing and querying operations.

Infrastructure code lives in `internal/repository` and `internal/indexer`, implementing the adapters that connect to external systems. Database repositories use Go's standard database interface, supporting any SQL database with appropriate drivers. Indexer implementations encapsulate blockchain-specific logic, making it straightforward to add support for new networks.

### Building

The project uses Go modules for dependency management, with versioned dependencies declared in `go.mod`. To build the service, run `go build -o bin/indexer ./cmd` from the project root. This produces a statically linked binary suitable for deployment in containerized environments.

The build process supports several compile-time options through Go build tags. These include features that may require additional dependencies or are specific to certain deployment scenarios. Review the build command in the Dockerfile for the recommended configuration options.

Testing is integrated into the build process, with unit tests covering core functionality and integration tests verifying database and blockchain interactions. Run `go test ./...` to execute the test suite. The project maintains high test coverage for critical paths, with coverage reports generated during continuous integration builds.

### Extending

Adding support for new blockchain networks requires implementing the `BlockReader` and `EventSubscriber` interfaces defined in the domain package. Each blockchain implementation should be placed in its own package under `internal/blockchain`, following the pattern established by existing implementations. The configuration system supports registration of new blockchain types through package initialization.

Custom indexing strategies can be implemented by extending the `Indexer` service with specialized processing logic. This is useful for applications that need to extract application-specific data from transactions or events. The clean architecture makes it straightforward to add new processing stages without modifying core indexing infrastructure.

API extensions should follow the established patterns in the handler layer. New endpoints are added by implementing the `Handler` interface and registering routes in the application setup code. Authentication and validation middleware can be applied to protect sensitive endpoints or enforce rate limiting.

## Contributing

Contributions to the Blockchain Indexer Service are welcome and encouraged. Before beginning work on a significant feature or bug fix, please open an issue to discuss the proposed changes. This helps ensure that efforts align with the project's direction and prevents duplication of work.

Code contributions should follow the project's coding standards, which emphasize clarity, testability, and documentation. All contributions must pass the automated test suite and lint checks before being merged. The project uses conventional commit messages for changelog generation, following the Conventional Commits specification.

## Code Examples

The following examples demonstrate common usage patterns for the Blockchain Indexer Service, including client initialization, querying blockchain data, and processing real-time events.

### Basic Client Usage

The client library provides a convenient interface for interacting with the indexer service. The following example shows how to initialize the client and perform basic queries.

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"
    
    "github.com/csic-platform/blockchain-indexer/internal/client"
)

func main() {
    // Create client with custom configuration
    cfg := client.Config{
        BaseURL: "http://localhost:8080",
        APIKey:  "your-api-key-here",
        Timeout: 30 * time.Second,
    }
    
    indexerClient, err := client.NewClient(cfg)
    if err != nil {
        log.Fatalf("Failed to create indexer client: %v", err)
    }
    
    ctx := context.Background()
    
    // Check service health
    health, err := indexerClient.Health(ctx)
    if err != nil {
        log.Fatalf("Health check failed: %v", err)
    }
    fmt.Printf("Service status: %s\n", health.Status)
    
    // Get network status for Bitcoin
    status, err := indexerClient.GetNetworkStatus(ctx, "bitcoin")
    if err != nil {
        log.Fatalf("Failed to get network status: %v", err)
    }
    
    fmt.Printf("Bitcoin height: %d\n", status.BlockHeight)
    fmt.Printf("Connected peers: %d\n", status.Peers)
    fmt.Printf("Sync progress: %.2f%%\n", status.SyncProgress)
}
```

### Querying Blocks

The following example demonstrates how to retrieve block data and navigate through blockchain history using pagination and filtering.

```go
package main

import (
    "context"
    "encoding/json"
    "fmt"
    "log"
    
    "github.com/csic-platform/blockchain-indexer/internal/client"
)

func queryBlocks() {
    client, _ := client.NewClient(client.Config{
        BaseURL: "http://localhost:8080",
    })
    
    ctx := context.Background()
    
    // Get a specific block by height
    block, err := client.GetBlock(ctx, "bitcoin", 820000)
    if err != nil {
        log.Fatalf("Failed to get block: %v", err)
    }
    
    data, _ := json.MarshalIndent(block, "", "  ")
    fmt.Printf("Block data:\n%s\n", data)
    
    // Get a block by hash
    blockByHash, err := client.GetBlockByHash(ctx, "bitcoin", block.Hash)
    if err != nil {
        log.Fatalf("Failed to get block by hash: %v", err)
    }
    fmt.Printf("Block verification: %s\n", blockByHash.Hash)
    
    // List recent blocks with pagination
    blocks, err := client.ListBlocks(ctx, "bitbitcoin", &client.PaginationRequest{
        Limit:  10,
        Offset: 0,
    })
    if err != nil {
        log.Fatalf("Failed to list blocks: %v", err)
    }
    
    fmt.Printf("\nRecent blocks (%d):\n", len(blocks))
    for _, b := range blocks {
        fmt.Printf("  Height: %d, Transactions: %d, Time: %s\n",
            b.Height, len(b.Transactions), b.Timestamp.Format(time.RFC3339))
    }
}
```

### Transaction Queries

Transaction queries support filtering by address, time range, value thresholds, and other criteria. The following examples show various query patterns.

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"
    
    "github.com/csic-platform/blockchain-indexer/internal/client"
)

func queryTransactions() {
    client, _ := client.NewClient(client.Config{
        BaseURL: "http://localhost:8080",
    })
    
    ctx := context.Background()
    
    // Get a specific transaction by hash
    tx, err := client.GetTransaction(ctx, "bitcoin", 
        "a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2")
    if err != nil {
        log.Fatalf("Failed to get transaction: %v", err)
    }
    
    fmt.Printf("Transaction: %s\n", tx.Hash)
    fmt.Printf("Block height: %d\n", tx.BlockHeight)
    fmt.Printf("Confirmations: %d\n", tx.Confirmations)
    fmt.Printf("Fee: %f BTC\n", tx.Fee)
    fmt.Printf("Risk score: %d\n", tx.RiskScore)
    
    // Print transaction inputs and outputs
    fmt.Println("\nInputs:")
    for i, input := range tx.Inputs {
        fmt.Printf("  %d: %s = %f BTC\n", i, input.Address, input.Value)
    }
    
    fmt.Println("\nOutputs:")
    for i, output := range tx.Outputs {
        fmt.Printf("  %d: %s = %f BTC\n", i, output.Address, output.Value)
    }
    
    // Search for transactions by address
    address := "1A2B3C4D5E6F7890ABCDEF1234567890"
    txs, err := client.GetAddressTransactions(ctx, "bitcoin", address, 50)
    if err != nil {
        log.Fatalf("Failed to get address transactions: %v", err)
    }
    
    fmt.Printf("\nTransactions for %s (%d):\n", address, len(txs))
    for _, tx := range txs {
        fmt.Printf("  - %s: %f BTC @ height %d\n", 
            tx.Hash[:16], tx.Value, tx.BlockHeight)
    }
    
    // Advanced search with filters
    searchReq := &client.TransactionSearchRequest{
        Network:  "bitcoin",
        From:     time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
        To:       time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
        MinValue: 10.0, // Minimum 10 BTC
        Limit:    100,
    }
    
    results, err := client.SearchTransactions(ctx, searchReq)
    if err != nil {
        log.Fatalf("Search failed: %v", err)
    }
    
    fmt.Printf("\nHigh-value transactions found: %d\n", len(results))
    for _, tx := range results {
        fmt.Printf("  - %s: %f BTC\n", tx.Hash[:16], tx.Value)
    }
}
```

### Address Monitoring

Address monitoring enables watchlist functionality and alerting on blockchain activity for specific addresses.

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"
    
    "github.com/csic-platform/blockchain-indexer/internal/client"
)

func monitorAddresses() {
    client, _ := client.NewClient(client.Config{
        BaseURL: "http://localhost:8080",
    })
    
    ctx := context.Background()
    
    // Get comprehensive address information
    address := "1A2B3C4D5E6F7890ABCDEF1234567890"
    info, err := client.GetAddressInfo(ctx, "bitcoin", address)
    if err != nil {
        log.Fatalf("Failed to get address info: %v", err)
    }
    
    fmt.Printf("Address: %s\n", address)
    fmt.Printf("Network: %s\n", info.Network)
    fmt.Printf("First seen: %s\n", info.FirstSeen.Format(time.RFC3339))
    fmt.Printf("Last activity: %s\n", info.LastActivity.Format(time.RFC3339))
    fmt.Printf("Total received: %f BTC\n", info.TotalReceived)
    fmt.Printf("Total sent: %f BTC\n", info.TotalSent)
    fmt.Printf("Current balance: %f BTC\n", info.Balance)
    fmt.Printf("Transaction count: %d\n", info.TransactionCount)
    fmt.Printf("Risk score: %d\n", info.RiskScore)
    fmt.Printf("Labels: %v\n", info.Labels)
    
    // Get balance changes over time
    balanceHistory, err := client.GetAddressBalanceHistory(ctx, "bitcoin", address,
        time.Now().AddDate(0, -1, 0), time.Now())
    if err != nil {
        log.Fatalf("Failed to get balance history: %v", err)
    }
    
    fmt.Printf("\nBalance history (%d points):\n", len(balanceHistory))
    for _, point := range balanceHistory {
        fmt.Printf("  %s: %f BTC\n", point.Timestamp.Format(time.RFC3339), point.Balance)
    }
    
    // Get related addresses (cluster analysis)
    related, err := client.GetRelatedAddresses(ctx, "bitcoin", address)
    if err != nil {
        log.Fatalf("Failed to get related addresses: %v", err)
    }
    
    fmt.Printf("\nRelated addresses (%d):\n", len(related))
    for _, addr := range related {
        fmt.Printf("  - %s (relationship: %s)\n", addr.Address, addr.Relationship)
    }
}
```

### Event and Smart Contract Indexing

The indexer provides comprehensive smart contract event indexing with ABI decoding support.

```go
package main

import (
    "context"
    "fmt"
    "log"
    
    "github.com/csic-platform/blockchain-indexer/internal/client"
)

func queryEvents() {
    client, _ := client.NewClient(client.Config{
        BaseURL: "http://localhost:8080",
    })
    
    ctx := context.Background()
    
    // Get events for a specific contract
    contractAddress := "0x742d35Cc6634C0532925a3b844Bc9e7595f7547b"
    events, err := client.GetContractEvents(ctx, "ethereum", contractAddress, 100)
    if err != nil {
        log.Fatalf("Failed to get contract events: %v", err)
    }
    
    fmt.Printf("Events for contract %s (%d):\n", contractAddress, len(events))
    for _, event := range events {
        fmt.Printf("  - %s: %s\n", event.Name, event.Data)
    }
    
    // Query by event signature
    transferEvents, err := client.QueryEvents(ctx, &client.EventQuery{
        Network:     "ethereum",
        EventSig:    "Transfer(address,address,uint256)",
        FromBlock:   18000000,
        ToBlock:     18000100,
    })
    if err != nil {
        log.Fatalf("Failed to query events: %v", err)
    }
    
    fmt.Printf("\nTransfer events: %d\n", len(transferEvents))
    for _, event := range transferEvents {
        fmt.Printf("  - Block %d: %s\n", event.BlockNumber, event.DecodedData)
    }
    
    // Track specific token transfers
    tokenEvents, err := client.TrackTokenTransfers(ctx, "ethereum",
        "0xTokenAddress", "0xHolderAddress", 100)
    if err != nil {
        log.Fatalf("Failed to track transfers: %v", err)
    }
    
    fmt.Printf("\nToken transfers for holder: %d\n", len(tokenEvents))
    for _, event := range tokenEvents {
        fmt.Printf("  - %s: %s -> %s, value: %s\n",
            event.Timestamp.Format("2006-01-02"),
            event.From[:8]+"...",
            event.To[:8]+"...",
            event.Value)
    }
}
```

### Real-Time Event Streaming

The indexer supports real-time event streaming through Kafka for building reactive applications.

```go
package main

import (
    "context"
    "encoding/json"
    "fmt"
    "log"
    "time"
    
    "github.com/csic-platform/blockchain-indexer/internal/publisher"
    "github.com/csic-platform/blockchain-indexer/internal/stream"
)

func eventStreaming() {
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()
    
    // Subscribe to new block events
    subscriber, err := stream.NewSubscriber(stream.Config{
        Brokers: []string{"localhost:9092"},
        Topic:   "csic-blocks",
        GroupID: "my-app-consumer",
    })
    if err != nil {
        log.Fatalf("Failed to create subscriber: %v", err)
    }
    
    // Start consuming blocks
    blockChan, err := subscriber.SubscribeBlocks(ctx, "bitcoin")
    if err != nil {
        log.Fatalf("Failed to subscribe to blocks: %v", err)
    }
    
    fmt.Println("Monitoring Bitcoin blocks...")
    for block := range blockChan {
        fmt.Printf("New block: #%d, transactions: %d, value: %f BTC\n",
            block.Height, len(block.Transactions), block.TotalValue)
        
        // Process high-value transactions in the block
        for _, tx := range block.Transactions {
            if tx.Value > 1000.0 {
                fmt.Printf("  High-value tx: %s = %f BTC\n", tx.Hash[:16], tx.Value)
            }
        }
    }
}

func publishEvents() {
    pub, err := publisher.NewKafkaPublisher(publisher.Config{
        Brokers: []string{"localhost:9092"},
        Topic:   "csic-blockchain-events",
    })
    if err != nil {
        log.Fatalf("Failed to create publisher: %v", err)
    }
    defer pub.Close()
    
    // Publish a custom event
    event := map[string]interface{}{
        "type":      "ADDRESS_ALERT",
        "network":   "bitcoin",
        "address":   "1A2B3C4D5E6F7890ABCDEF1234567890",
        "alert_type": "HIGH_VALUE_DEPOSIT",
        "value":     500.0,
        "timestamp": time.Now().UTC(),
    }
    
    data, _ := json.Marshal(event)
    if err := pub.Publish(context.Background(), data); err != nil {
        log.Fatalf("Failed to publish event: %v", err)
    }
    
    fmt.Println("Alert event published successfully")
}
```

### Risk Assessment Integration

The indexer provides built-in risk scoring and can be integrated with external risk assessment engines.

```go
package main

import (
    "context"
    "fmt"
    "log"
    
    "github.com/csic-platform/blockchain-indexer/internal/risk"
)

func riskAssessment() {
    // Initialize risk engine with custom thresholds
    engine, err := risk.NewEngine(risk.Config{
        HighValueThreshold:   10.0,      // 10 BTC
        MixingThreshold:      50.0,      // Mixing pattern score
        SanctionsThreshold:   100.0,     // Sanctions list match
        VelocityWindow:       24 * time.Hour,
    })
    if err != nil {
        log.Fatalf("Failed to create risk engine: %v", err)
    }
    
    ctx := context.Background()
    
    // Assess a single transaction
    tx := &risk.Transaction{
        Hash:              "a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6",
        Network:           "bitcoin",
        BlockHeight:       820000,
        Value:             25.0,
        InputCount:        5,
        OutputCount:       3,
        IsCoinJoin:        false,
        InvolvesKnownBad:  false,
        VelocityScore:     30.0,
    }
    
    result, err := engine.Assess(ctx, tx)
    if err != nil {
        log.Fatalf("Risk assessment failed: %v", err)
    }
    
    fmt.Printf("Transaction: %s\n", tx.Hash)
    fmt.Printf("Risk Score: %d/100\n", result.Score)
    fmt.Printf("Risk Level: %s\n", result.Level)
    fmt.Printf("Risk Factors: %v\n", result.Factors)
    fmt.Printf("Recommended Action: %s\n", result.RecommendedAction)
    
    // Batch assess multiple transactions
    transactions := []*risk.Transaction{
        {Hash: "tx1", Value: 50.0, IsCoinJoin: true},
        {Hash: "tx2", Value: 0.5, InvolvesKnownBad: false},
        {Hash: "tx3", Value: 100.0, InputCount: 100},
    }
    
    batchResults, err := engine.AssessBatch(ctx, transactions)
    if err != nil {
        log.Fatalf("Batch assessment failed: %v", err)
    }
    
    fmt.Printf("\nBatch assessment results:\n")
    for i, result := range batchResults {
        fmt.Printf("  %s: score=%d, level=%s\n", 
            transactions[i].Hash, result.Score, result.Level)
    }
    
    // Get flagged transactions for review
    flagged, err := engine.GetFlaggedTransactions(ctx, 70)
    if err != nil {
        log.Fatalf("Failed to get flagged transactions: %v", err)
    }
    
    fmt.Printf("\nFlagged for review (%d):\n", len(flagged))
    for _, tx := range flagged {
        fmt.Printf("  - %s: score=%d, factors=%v\n", 
            tx.Hash, tx.RiskScore, tx.RiskFactors)
    }
}
```

### Batch Operations

For high-throughput scenarios, the indexer supports batch operations to reduce network overhead.

```go
package main

import (
    "context"
    "fmt"
    "log"
    
    "github.com/csic-platform/blockchain-indexer/internal/client"
)

func batchOperations() {
    client, _ := client.NewClient(client.Config{
        BaseURL: "http://localhost:8080",
    })
    
    ctx := context.Background()
    
    // Batch get blocks
    heights := []int64{819999, 819998, 819997, 819996, 819995}
    blocks, err := client.GetBlocksBatch(ctx, "bitcoin", heights)
    if err != nil {
        log.Fatalf("Failed to get blocks batch: %v", err)
    }
    
    fmt.Printf("Retrieved %d blocks:\n", len(blocks))
    for _, block := range blocks {
        fmt.Printf("  #%d: %d transactions\n", block.Height, len(block.Transactions))
    }
    
    // Batch get transactions
    txHashes := []string{
        "a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6",
        "b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7",
        "c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8",
    }
    
    transactions, err := client.GetTransactionsBatch(ctx, "bitcoin", txHashes)
    if err != nil {
        log.Fatalf("Failed to get transactions batch: %v", err)
    }
    
    fmt.Printf("\nRetrieved %d transactions:\n", len(transactions))
    for _, tx := range transactions {
        fmt.Printf("  %s: %f BTC\n", tx.Hash[:16], tx.Value)
    }
    
    // Batch check addresses against watchlist
    addresses := []string{
        "1A2B3C4D5E6F7890ABCDEF1234567890",
        "1XYZ2ABC3DEF456GHI789JKL012MNO345",
        "1ABC2DEF3GHI456JKL789MNO012PQR345",
    }
    
    results, err := client.CheckWatchlistBatch(ctx, "bitcoin", addresses)
    if err != nil {
        log.Fatalf("Failed to check watchlist batch: %v", err)
    }
    
    fmt.Printf("\nWatchlist check results:\n")
    for i, result := range results {
        status := "clean"
        if result.Flagged {
            status = "FLAGGED - " + result.Reason
        }
        fmt.Printf("  %s: %s\n", addresses[i][:16]+"...", status)
    }
}
```

## License

The Blockchain Indexer Service is proprietary software. All rights are reserved. Use of this software is subject to the terms and conditions established in your licensing agreement. Contact the project maintainers for licensing inquiries.
