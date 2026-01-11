# Sovereign Digital Governance Stack (SDGS)

## Overview

The Sovereign Digital Governance Stack (SDGS) is a comprehensive framework for state-grade digital infrastructure, combining two independent but aligned platforms:

- **CSIC** (Crypto-State-Infrastructure-Contractor): Digital money control, financial surveillance, crypto/CBDC oversight
- **NEAM** (National Economic Activity Monitor): Economic sensing, payments monitoring, industrial output analysis

SDGS is **not a monolithic application**. It is a **governance + orchestration layer** that provides the constitutional framework for both platforms to operate in alignment under sovereign control.

## Architecture Principles

1. **Independent Operation**: Each platform maintains its own lifecycle, deployment, and operational procedures
2. **Loose Coupling**: Systems communicate through well-defined signals and events, not shared databases
3. **Data Sovereignty**: All data remains under sovereign control with clear classification boundaries
4. **Auditability**: Complete traceability of all operations and inter-system communications

## Directory Structure

```
SDGS/
├── csic/                          # Crypto-State-Infrastructure-Contractor
├── neam/                          # NEAM Platform
├── integration/                   # Signal bridges (future)
│   └── signal-gateway/
│       ├── main.go
│       ├── contracts/
│       └── adapters/
├── shared/                        # Shared schemas & standards
│   ├── schemas/
│   └── standards/
├── infra/                         # Kubernetes / Terraform / Helm
│   ├── kubernetes/
│   ├── terraform/
│   └── helm/
├── docs/                          # Sovereign governance docs
└── README.md
```

## Quick Start

### Development Setup

```bash
# Clone the SDGS workspace
git clone <sdgs-repo> SDGS
cd SDGS

# Initialize submodules (if any)
git submodule update --init --recursive

# Set up development environment
./scripts/setup-dev.sh
```

### Platform-Specific Development

**CSIC Development:**
```bash
cd SDGS/csic
git checkout -b feature/<feature-name>
# develop
git commit -m "Enhancement description"
git push origin feature/<feature-name>
```

**NEAM Development:**
```bash
cd SDGS/neam
git checkout -b feature/<feature-name>
# develop
git commit -m "Enhancement description"
git push origin feature/<feature-name>
```

## Documentation

- [SDGS Overview](docs/sdgs-overview.md)
- [Governance Model](docs/governance-model.md)
- [Security Architecture](docs/security-architecture.md)
- [Data Classification](docs/data-classification.md)
- [Audit Framework](docs/audit-framework.md)
- [Inter-System Signals](docs/inter-system-signals.md)
- [Compliance Mapping](docs/compliance-mapping.md)

## Integration

SDGS defines signal contracts for inter-system communication:

- **Financial Risk Signals**: CSIC → NEAM (risk indicators, compliance flags)
- **Macro-Economic Signals**: NEAM → CSIC (activity indices, sector health)
- **Policy Signals**: Bidirectional (regulatory updates, intervention requests)

See [Inter-System Signals](docs/inter-system-signals.md) for detailed specifications.

## Infrastructure

### Development/Testing
- Separate docker-compose stacks per platform
- Isolated Kafka topics
- Independent databases per platform

### Sovereign/Production
- Kubernetes namespaces per system
- Shared services:
  - IAM (Identity and Access Management)
  - HSM/Vault (Key Management)
  - Centralized Audit Logging
- Air-gapped capable deployments

## Contributing

1. Work within the appropriate platform directory (csic or neam)
2. Follow platform-specific contribution guidelines
3. Update SDGS documentation if changes affect governance
4. Submit PRs through standard review process

## Security

This is **state-grade digital infrastructure**. All contributions must:
- Maintain security boundaries between platforms
- Follow data classification requirements
- Ensure complete audit trail coverage
- Support sovereign deployment requirements

## License

All components of SDGS are proprietary state infrastructure. See individual platform licenses for details.
