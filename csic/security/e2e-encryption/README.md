# End-to-End Encryption Service

A comprehensive end-to-end encryption service for securing sensitive data in the CSIC Platform.

## Features

- **Symmetric Encryption**: AES-256-GCM, ChaCha20-Poly1305
- **Asymmetric Encryption**: RSA-2048, RSA-4096
- **Key Management**: Generation, rotation, and lifecycle management
- **Key Storage**: HSM, Vault, Database, or File-based storage
- **Session Keys**: ECDH-based key exchange
- **Key Derivation**: PBKDF2, Argon2, HKDF
- **Audit Logging**: Comprehensive audit trail

## Quick Start

### Using Docker

```bash
# Build and start the service
docker-compose up -d

# Access the API
curl https://localhost:8086/health
```

### From Source

```bash
# Build the service
cd security/e2e-encryption
go build -o e2e-encrypt ./cmd/server

# Start the server
./e2e-encrypt serve

# Generate a symmetric key
./e2e-encrypt generate-key my-key --type symmetric --algorithm AES-256-GCM

# Generate an asymmetric key pair
./e2e-encrypt generate-key my-rsa-key --type asymmetric --algorithm RSA-4096

# Encrypt data
./e2e-encrypt encrypt my-key --input plaintext.txt --output encrypted.dat

# Decrypt data
./e2e-encrypt decrypt my-key --input encrypted.dat --output decrypted.txt
```

## Configuration

All configuration is managed via `config.yaml`:

```yaml
# Encryption settings
encryption:
  algorithm: "AES-256-GCM"
  key_derivation: "Argon2"
  key_rotation_period: 90

# Key management
key_management:
  storage_backend: "database"
  hsm:
    enabled: false
  vault:
    enabled: false

# Session keys
session_keys:
  lifetime: 60
  protocol: "ECDH"
  curve: "P-256"

# Network settings
network:
  host: "0.0.0.0"
  port: 8086
  tls:
    enabled: true
    cert_path: "/etc/csic/certs/server.crt"
    key_path: "/etc/csic/certs/server.key"
```

## API Endpoints

### Health Check
```
GET /health
```

### Encryption
```
POST /api/v1/encrypt   - Encrypt data
POST /api/v1/decrypt   - Decrypt data
```

### Key Management
```
GET    /api/v1/keys              - List all keys
POST   /api/v1/keys              - Create a new key
GET    /api/v1/keys/{id}         - Get key details
PUT    /api/v1/keys/{id}         - Update key
DELETE /api/v1/keys/{id}         - Delete a key
POST   /api/v1/keys/{id}/rotate  - Rotate a key
```

### Key Exchange
```
POST /api/v1/key-exchange        - Initiate key exchange
GET  /api/v1/key-exchange/{id}   - Get exchange session
POST /api/v1/key-exchange/{id}/complete - Complete exchange
```

### Audit
```
GET /api/v1/audit               - Get audit logs
GET /api/v1/audit?key_id=xxx    - Get audit logs for key
```

## Key Types

### Symmetric Keys
- AES-256-GCM (recommended)
- ChaCha20-Poly1305

### Asymmetric Keys
- RSA-2048
- RSA-4096
- ECDSA-P256
- ECDSA-P384
- ECDSA-P521

### Session Keys
- Ephemeral keys for secure communication
- Auto-expire after configured lifetime
- Used with ECDH key exchange

## Key Lifecycle

### Generation
```go
key, err := keyService.GenerateKey(ctx, KeyGenerationConfig{
    Type:       KeyTypeSymmetric,
    Algorithm:  AlgorithmAES256GCM,
    Expiration: 90 * 24 * time.Hour,
    Usage:      []KeyUsage{KeyUsageEncrypt, KeyUsageDecrypt},
    Metadata: KeyMetadata{
        Description: "Production encryption key",
        Labels:      map[string]string{"env": "production"},
    },
})
```

### Rotation
```go
newKey, err := keyService.RotateKey(ctx, keyID, KeyRotationConfig{
    PreserveOldKey:   true,
    MigrationPeriod:  24 * time.Hour,
})
```

### Revocation
```go
err := keyService.RevokeKey(ctx, keyID, "Compromised key")
```

### Destruction
```go
err := keyService.DestroyKey(ctx, keyID)
```

## Encryption Usage

### Basic Encryption
```go
result, err := encryptionService.Encrypt(ctx, plaintext, keyID)
```

### With Additional Authenticated Data
```go
result, err := encryptionService.EncryptWithAAD(ctx, plaintext, keyID, additionalData)
```

### Stream Encryption
```go
err := encryptionService.EncryptStream(ctx, reader, writer, keyID)
```

### String Encryption
```go
encrypted, err := encryptionService.EncryptString(ctx, plaintext, keyID)
decrypted, err := encryptionService.DecryptString(ctx, encrypted, keyID)
```

## Key Derivation

### PBKDF2
```go
key, err := cryptoEngine.DeriveKey(password, salt, 100000, 32)
```

### HKDF
```go
key, err := cryptoEngine.DeriveKeyHKDF(secret, salt, info, 32)
```

## Session Key Exchange

### ECDH Key Exchange
```go
// Generate ephemeral key pair
ephemeralPriv, ephemeralPub, err := cryptoEngine.GenerateECKeyPair(elliptic.P256())

// Get peer's public key
peerPub, _ := keyService.GetKey(ctx, peerKeyID)

// Perform key exchange
sharedSecret, err := cryptoEngine.ECDHKeyExchange(ephemeralPriv, peerPub)
```

## Key Storage Backends

### Database (Default)
```yaml
key_management:
  storage_backend: "database"
  database:
    enabled: true
    key_table: "encryption_keys"
```

### HashiCorp Vault
```yaml
key_management:
  storage_backend: "vault"
  vault:
    enabled: true
    address: "https://vault.example.com"
    token: "your-token"
```

### HSM (PKCS#11)
```yaml
key_management:
  storage_backend: "hsm"
  hsm:
    enabled: true
    provider: "pkcs11"
    slot: 0
    pin: "your-pin"
```

## Security Considerations

1. **Key Storage**: Use HSM or Vault for production key storage
2. **Key Rotation**: Automatic rotation every 90 days
3. **Audit Logging**: All key operations are logged
4. **TLS**: Enable TLS for all API communications
5. **Rate Limiting**: Prevent brute-force attacks
6. **Access Control**: Implement proper API key management

## Integration Points

The E2E Encryption service integrates with:

- **Forensic Tools**: Encrypt sensitive forensic data
- **Incident Response**: Secure incident data handling
- **Audit Log**: Encrypted audit trails
- **User Data**: End-to-end encryption for user data

## Architecture

```
security/e2e-encryption/
├── cmd/
│   └── server/           # Main entry point
├── encryption/           # Encryption service
├── key-management/       # Key management service
├── crypto/               # Cryptographic utilities
├── config.yaml           # Configuration
├── Dockerfile            # Docker build
└── docker-compose.yml    # Docker orchestration
```

## License

Part of the CSIC Platform.
