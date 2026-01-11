# Data Flow Architecture

This document describes the complete data flow architecture for the CSIC Platform, covering data ingestion, processing, storage, and presentation layers.

## Architecture Overview

```mermaid
flowchart TB
    subgraph Data Sources
        subgraph Blockchain Sources
            BTC[Bitcoin Node]
            ETH[Ethereum Node]
        end
        subgraph Exchange Sources
            EX[Exchange APIs\nReal-time Stream]
        end
        subgraph Mining Sources
            MH[Mining Hardware\nTelemetry Data]
        end
        subgraph User Operations
            UA[User Actions\nAudit Logs]
        end
    end

    subgraph Ingestion Layer
        subgraph Adapters
            BA[Blockchain Adapter]
            EG[Exchange Gateway\nKafka Producer]
            DC[Device Protocol\nMQTT/Modbus]
            AC[Audit Collector\nWORM]
        end
    end

    subgraph Message Queue Layer
        MQ[Apache Kafka]
        subgraph Topics
            T1[transactions]
            T2[exchange_data]
            T3[mining_metrics]
            T4[audit]
        end
    end

    subgraph Processing Layer
        subgraph Engines
            TA[Transaction Analysis\nStream Processing]
            MM[Market Monitor\nReal-time Rules]
            RA[Risk Engine\nMachine Learning]
            AE[Audit Engine\nWORM Storage]
        end
    end

    subgraph Storage Layer
        subgraph Databases
            PG[PostgreSQL\nTransactions/Users/Config]
            TD[TimescaleDB\nTime-series Metrics]
            ES[OpenSearch\nLogs/Search]
            WS[WORM Storage\nAudit Records]
        end
    end

    subgraph Service Layer
        GW[API Gateway\nREST/gRPC]
        subgraph Services
            RP[Report Service\nPDF/CSV]
            AS[Alert Service\nNotifications]
            RA_API[Regulatory API\nQuery Interface]
        end
    end

    subgraph Access Layer
        subgraph Clients
            AC[Admin Console\nReact Dashboard]
            RE[Regulatory Reports\nPDF/Excel]
            NT[Alert Notifications\nEmail/SMS]
            API[API Clients\ngRPC]
        end
    end

    %% Data Flow Connections
    BTC --> BA
    ETH --> BA
    EX --> EG
    MH --> DC
    UA --> AC
    
    BA --> MQ
    EG --> MQ
    DC --> MQ
    AC --> MQ
    
    MQ --> T1
    MQ --> T2
    MQ --> T3
    MQ --> T4
    
    T1 --> TA
    T2 --> MM
    T3 --> RA
    T4 --> AE
    
    TA --> PG
    TA --> ES
    MM --> PG
    MM --> AS
    RA --> PG
    RA --> TD
    AE --> WS
    AE --> ES
    
    PG --> GW
    ES --> GW
    TD --> GW
    WS --> GW
    
    GW --> RP
    GW --> AS
    GW --> RA_API
    
    RP --> RE
    AS --> NT
    RA_API --> API
    GW --> AC
```

## Data Flow Details

### Blockchain Transaction Data Flow

```mermaid
sequenceDiagram
    participant BN as Blockchain Node\n(BTC/ETH)
    participant BA as Blockchain Adapter
    participant K as Kafka\n(transactions topic)
    participant TE as Transaction Engine
    participant RA as Risk Engine
    participant PG as PostgreSQL
    participant ES as OpenSearch
    participant AS as Alert Service

    BN->>BA: New Block Notification\n(ZMQ/WebSocket)
    BA->>BA: Parse Transaction Data
    BA->>K: Publish Transaction Events
    K->>TE: Consume Transactions
    TE->>TE: Stream Processing
    TE->>RA: Request Risk Assessment
    RA->>RA: Calculate Risk Score
    RA-->>TE: Risk Score Result
    TE->>PG: Store Transaction\nwith Risk Score
    TE->>ES: Index for Search
    alt High Risk Transaction
        RA->>AS: Trigger Alert
        AS->>AS: Generate Notification
    end
```

### Exchange Monitoring Data Flow

```mermaid
sequenceDiagram
    participant EX as Exchange API
    participant EG as Exchange Gateway
    participant K as Kafka\n(exchange_data topic)
    participant MM as Market Monitor
    participant PG as PostgreSQL
    participant AR as Alert Rules
    participant AS as Alert Service

    EX->>EG: Order Book Data\nReal-time Stream
    EX->>EG: Trade Data
    EG->>EG: Validate Data Format
    EG->>K: Publish Market Data
    K->>MM: Consume Market Events
    MM->>AR: Evaluate Rules
    AR-->>MM: Rule Evaluation Result
    alt Rule Violation Detected
        MM->>AS: Create Alert
        AS->>AS: Generate Report
    end
    MM->>PG: Store Market Data
    PG->>PG: Update Order Book\nState
```

### Mining Telemetry Data Flow

```mermaid
sequenceDiagram
    participant MH as Mining Hardware
    participant DC as Device Communication\n(MQTT/Modbus)
    participant K as Kafka\n(mining_metrics topic)
    participant EM as Energy Manager
    participant TD as TimescaleDB
    participant MC as Mining Control\nDashboard
    participant ES as Energy Strategy

    MH->>DC: Telemetry Data\n(Hash Rate, Power)
    DC->>DC: Parse Protocol Data
    DC->>K: Publish Metrics
    K->>EM: Consume Telemetry
    EM->>EM: Calculate Energy\nConsumption
    EM->>TD: Store Time-series\nData
    TD->>MC: Update Dashboard\nMetrics
    alt Exceeds Limits
        EM->>ES: Check Strategy
        ES->>ES: Execute Response\n(Throttle/Alert)
    end
```

### Audit Log Data Flow

```mermaid
sequenceDiagram
    participant S as Platform Services
    participant AM as Audit Middleware
    participant K as Kafka\n(audit topic)
    participant AE as Audit Engine
    participant WS as WORM Storage
    participant ES as OpenSearch
    participant Q as Query Interface

    S->>AM: Operation Event
    AM->>AM: Create Audit Record\n(Actor, Action, Resource)
    AM->>AM: Calculate Log Hash\n(Chain Structure)
    AM->>K: Publish Audit Event
    K->>AE: Consume Audit Events
    AE->>WS: Write to WORM\n(Tamper-proof)
    AE->>ES: Index for Search
    Q->>AE: Query Audit Logs
    AE->>Q: Search Results\nwith Integrity Proof
```

## Storage Architecture

```mermaid
erDiagram
    POSTGRESQL {
        uuid transaction_id PK
        string from_address
        string to_address
        decimal amount
        string currency
        int risk_score
        timestamp created_at
        string status
    }
    
    TIMESCALEDB {
        bigint time PK
        uuid machine_id FK
        decimal hash_rate
        decimal power_consumption
        decimal temperature
        string pool_url
    }
    
    OPENSEARCH {
        string log_id PK
        timestamp timestamp
        string level
        string service
        string message
        json fields
    }
    
    WORM_STORAGE {
        string audit_id PK
        timestamp timestamp
        string actor_id
        string action
        string resource
        string previous_hash
        string current_hash
        string signature
    }

    POSTGRESQL ||--o{ TIMESCALEDB : "references"
    POSTGRESQL ||--o{ OPENSEARCH : "indexes"
    WORM_STORAGE ||--o{ OPENSEARCH : "mirrors"
```

## Data Retention Policy

| Data Type | Storage Location | Retention Period | Archival Strategy |
|-----------|-----------------|------------------|-------------------|
| Transaction Records | PostgreSQL | 7 years | Archive to cold storage |
| Audit Logs | WORM Storage | 7 years | Permanent retention |
| Monitoring Metrics | TimescaleDB | 90 days | Aggregate, retain 7 years |
| System Logs | OpenSearch | 365 days | Archive to object storage |
| Alert Records | PostgreSQL | 7 years | Archive to cold storage |
| Reports | Object Storage | 7 years | Compressed archive |

## Performance Requirements

| Metric | Target | Maximum |
|--------|--------|---------|
| Transaction Processing | 10,000 TPS | 50,000 TPS |
| Exchange Data Latency | < 1 second | < 5 seconds |
| Blockchain Sync Latency | < 30 seconds | < 2 minutes |
| API Response Time | < 100ms | < 500ms |
| Alert Notification Delay | < 5 seconds | < 30 seconds |
| Query Response (Complex) | < 2 seconds | < 10 seconds |

## Message Queue Topics

```mermaid
flowchart LR
    subgraph Kafka Cluster
        subgraph Topics
            T1[t\ntransactions]
            T2[e\nexchange_data]
            T3[m\nmining_metrics]
            T4[a\naudit]
            T5[c\ncompliance]
            T6[i\ninterventions]
        end
    end
    
    subgraph Producers
        P1[Blockchain\nAdapter]
        P2[Exchange\nGateway]
        P3[Device\nCommunication]
        P4[Audit\nMiddleware]
        P5[Compliance\nModule]
        P6[Control\nLayer]
    end
    
    subgraph Consumers
        C1[Transaction\nEngine]
        C2[Market\nMonitor]
        C3[Energy\nManager]
        C4[Audit\nEngine]
        C5[Alert\nService]
        C6[Report\nService]
    end
    
    P1 --> T1
    P2 --> T2
    P3 --> T3
    P4 --> T4
    P5 --> T5
    P6 --> T6
    
    T1 --> C1
    T2 --> C2
    T3 --> C3
    T4 --> C4
    T5 --> C5
    T6 --> C6
```

## Data Transformation Pipeline

```mermaid
flowchart TB
    subgraph Raw Data
        RD1[Blockchain\nRaw Transactions]
        RD2[Exchange\nOrder Books]
        RD3[Mining\nTelemetry]
        RD4[System\nEvents]
    end
    
    subgraph Extraction
        EX1[Parse Protocol\nMessages]
        EX2[Validate\nSchema]
        EX3[Normalize\nTimestamps]
        EX4[Enrich\nMetadata]
    end
    
    subgraph Processing
        PR1[Pattern\nDetection]
        PR2[Risk\nCalculation]
        PR3[Aggre-\ngation]
        PR4[Anomaly\nScoring]
    end
    
    subgraph Enrichment
        EN1[Address\nTagging]
        EN2[Entity\nResolution]
        EN3[Risk\nLabels]
        EN4[Geo-\nlocation]
    end
    
    subgraph Storage
        ST1[Structured\nStorage]
        ST2[Time-series\nStorage]
        ST3[Full-text\nIndex]
        ST4[Immutable\nArchive]
    end
    
    RD1 --> EX1
    RD2 --> EX2
    RD3 --> EX3
    RD4 --> EX4
    
    EX1 --> PR1
    EX2 --> PR1
    EX3 --> PR2
    EX4 --> PR3
    
    PR1 --> EN1
    PR2 --> EN2
    PR3 --> EN3
    PR4 --> EN4
    
    EN1 --> ST1
    EN2 --> ST1
    EN3 --> ST2
    EN4 --> ST3
    PR4 --> ST4
```

## Quality Assurance

### Data Validation Rules

- Schema validation on all incoming messages
- Checksum verification for blockchain data
- Timestamp synchronization across sources
- Duplicate detection and handling
- Missing data interpolation

### Monitoring Metrics

- Message throughput per topic
- Processing latency per stage
- Error rates by data source
- Storage growth rates
- Query performance metrics
