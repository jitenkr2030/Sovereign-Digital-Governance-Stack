# Archives Service

This package manages long-term data archival, retention policy enforcement, and historical data retrieval for the NEAM platform.

## Overview

The Archives Service ensures compliance with regulatory data retention requirements while optimizing storage costs through intelligent tiering. The service automatically archives historical data at defined intervals, enforces retention policies, and provides efficient retrieval mechanisms for historical investigations and audits. All archived data is compressed and encrypted to ensure security and minimize storage costs.

## Archive Types

### Daily Archives
Daily archives capture complete daily snapshots of economic data:
- Aggregated payment statistics
- Daily energy consumption totals
- Transport throughput summaries
- Industrial output indices
- Inflation readings

**Retention**: 90 days in warm storage, then migrated to cold storage

### Monthly Archives
Monthly archives provide consolidated monthly views:
- Monthly payment aggregates by category and region
- Energy consumption patterns and trends
- Transport volume analysis
- Sector performance indices
- Intervention summaries

**Retention**: 10 years in cold storage

### Quarterly Archives
Quarterly archives support fiscal and policy analysis:
- Quarterly economic performance summaries
- Budget tracking and expenditure analysis
- Intervention effectiveness assessments
- Regional economic comparisons

**Retention**: 20 years in cold storage

### Annual Archives
Annual archives create permanent records for historical analysis:
- Annual economic review
- Fiscal year summary
- Long-term trend analysis
- Policy impact assessments

**Retention**: Indefinite in cold storage

## Key Features

### Tiered Storage Management
The service implements a sophisticated tiered storage strategy:

- **Hot Storage**: Recent data (0-7 days) on high-performance storage
- **Warm Storage**: Recent historical data (7-90 days) on cost-effective storage
- **Cold Storage**: Long-term archives on archival-grade storage

Data is automatically migrated between tiers based on age and access patterns.

### Retention Policy Enforcement
Automatic enforcement of retention policies:
- Daily policy checks identify expired archives
- Expired archives are permanently deleted
- Retention compliance reports generated
- Policy violations trigger alerts

### Retrieval Workflows
Multiple retrieval options based on urgency:
- **Standard**: Standard priority for routine requests (4 hour ETA)
- **Expedited**: Faster retrieval for time-sensitive needs (1 hour ETA)
- **Urgent**: Priority retrieval for critical investigations (15 minute ETA)

### Compression and Encryption
All archived data is protected:
- **Compression**: Zstandard algorithm for optimal ratio
- **Encryption**: AES-256 encryption for all archives
- **Integrity Checks**: SHA-256 checksums for verification

## Architecture

### Archive Creation Flow
1. Scheduler triggers archive creation at scheduled interval
2. Data collector queries source systems for specified period
3. Data validated for completeness and quality
4. Archive assembled with metadata
5. Archive compressed using Zstandard
6. Archive encrypted with AES-256
7. Checksum computed for integrity verification
8. Archive uploaded to object storage
9. Archive metadata stored in database

### Retrieval Flow
1. Retrieval request received with priority level
2. Archive location determined from metadata
3. Storage tier identified for ETA calculation
4. Retrieval task queued based on priority
5. Archive retrieved from storage
6. Archive decrypted if encrypted
7. Archive decompressed if compressed
8. Data extracted for requested tables
9. Data delivered to requesting system

### Retention Enforcement Flow
1. Daily scan identifies expired archives
2. Expired archives marked for deletion
3. Deletion scheduled during low-activity period
4. Storage cleanup after successful deletion
5. Compliance audit trail recorded

## Usage

### Create Monthly Archive
```go
archive, err := service.CreateMonthlyArchive(ctx, 2024, 3)
```

### List Archives
```go
archives, err := service.ListArchives(ctx, map[string]string{
    "type":        "monthly",
    "storage_tier": "cold",
})
```

### Request Retrieval
```go
request, err := service.RequestRetrieval(ctx, 
    archiveID, 
    "analyst@agency.gov", 
    "Q3 economic analysis",
    "expedited",
)
```

### List Retention Policies
```go
policies, err := service.ListRetentionPolicies(ctx)
```

## Storage Configuration

### MinIO/S3 Configuration
The archives service uses MinIO or S3-compatible storage:
```yaml
storage:
  endpoint: minio:9000
  bucket: neam-archives
  access_key: ${MINIO_ACCESS_KEY}
  secret_key: ${MINIO_SECRET_KEY}
```

### Retention Policy Configuration
```yaml
retention:
  daily:
    retention_days: 90
    compression: true
    encryption: true
  monthly:
    retention_years: 10
    compression: true
    encryption: true
  quarterly:
    retention_years: 20
    compression: true
    encryption: true
```

## Dependencies
- PostgreSQL: Archive metadata and retrieval queue
- ClickHouse: Historical data queries
- Redis: Caching for archive metadata
- MinIO/S3: Object storage for archives
