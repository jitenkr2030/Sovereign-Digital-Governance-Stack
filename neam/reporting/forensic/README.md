# Forensic Evidence Service

This package provides tamper-proof evidence collection, chain of custody management, and cryptographic verification capabilities essential for regulatory compliance and legal proceedings.

## Overview

The Forensic Evidence Service creates an immutable audit trail of all economic interventions and significant events within the NEAM platform. Every action is cryptographically logged with complete chain of custody information, ensuring that evidence can be verified and authenticated even years after the original event. This service is critical for regulatory compliance, dispute resolution, and legal proceedings.

## Core Concepts

### Evidence Records
Evidence records capture complete information about economic events:
- **Intervention Evidence**: Records of policy interventions including rationale, implementation, and outcomes
- **Transaction Evidence**: Detailed records of significant financial transactions
- **Anomaly Evidence**: Documentation of detected anomalies including analysis and response
- **Manual Evidence**: Evidence created manually for events not automatically captured

### Evidence Chain
Each piece of evidence is linked to its predecessors through a hash chain:
- **Chain Hash**: Cryptographic hash linking to previous evidence
- **Parent Hash**: Reference to the immediate predecessor
- **Merkle Proof**: Cryptographic proof enabling efficient verification
- **Timestamp**: Trusted timestamp from a certified time source

### Chain Sealing
Evidence chains can be sealed to prevent further modifications:
- Once sealed, no new evidence can be added
- The final root hash is recorded as an immutable timestamp
- Sealed chains can be independently verified by third parties

## Key Features

### Tamper Detection
The service implements comprehensive tamper detection:
- Content Hash Verification: Confirms evidence content matches recorded hash
- Chain Integrity Check: Validates the complete hash chain
- Timestamp Verification: Ensures timestamps are consistent
- Merkle Proof Validation: Verifies cryptographic proofs

### Evidence Verification
Multiple verification levels are supported:
- **Quick Verification**: Checks content hash against stored value
- **Full Verification**: Validates complete chain integrity
- **Third-Party Verification**: Generates proof packages for external verification

### Chain Management
The service provides complete chain management capabilities:
- Create new evidence chains
- Add evidence to existing chains
- Seal chains to prevent modifications
- Retrieve complete chain history

## Architecture

### Evidence Creation Flow
1. Event detected or manual evidence requested
2. Evidence content collected from source systems
3. Content hash computed using SHA-256
4. Parent hash retrieved from chain head
5. Merkle proof generated
6. Evidence stored with cryptographic metadata
7. Chain entry created linking to previous evidence

### Verification Flow
1. Verification request received
2. Content hash recomputed from stored content
3. Chain integrity verified through hash links
4. Merkle proof validated
5. Verification record created with all check results
6. Result returned with details of any discrepancies

### Storage Strategy
Evidence is stored using a tiered approach:
- **Recent Evidence**: PostgreSQL for rapid access
- **Historical Evidence**: ClickHouse for efficient queries
- **Archived Evidence**: Cold storage with encryption

## Usage

### Generate Evidence
```go
metadata := forensic.EvidenceMetadata{
    SourceSystem:   "intervention-service",
    SourceID:       "int-12345",
    EventType:      "intervention_executed",
    Jurisdiction:   "NATIONAL",
    Classification: "OFFICIAL",
    Tags:           []string{"price-control", "food-sector"},
}
evidence, err := service.GenerateEvidence(ctx, "intervention", content, metadata)
```

### Check for Tampering
```go
result, err := service.CheckTamper(ctx, evidenceID)
if result.Tampered {
    // Evidence has been modified
}
```

### Verify Evidence
```go
record, err := service.VerifyEvidence(ctx, evidenceID, "regulator@agency.gov")
```

### Seal Evidence Chain
```go
err := service.SealChain(ctx, "chain-intervention")
```

## Dependencies
- PostgreSQL: Evidence metadata and chain storage
- ClickHouse: Historical evidence queries
- Redis: Caching for frequent evidence access
- Kafka: Event distribution for new evidence

## Cryptographic Standards

### Hash Algorithm
All evidence uses SHA-256 for content hashing, providing:
- Collision resistance
- Pre-image resistance
- Second pre-image resistance

### Timestamp Format
Timestamps follow RFC 3339 format with UTC timezone:
```go
2006-01-02T15:04:05Z07:00
```

### Merkle Tree Structure
Merkle proofs are constructed as binary trees:
- Leaf nodes contain evidence hashes
- Parent nodes contain hashes of children
- Root node represents complete chain state
