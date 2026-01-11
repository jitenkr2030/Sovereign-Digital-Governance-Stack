use std::fmt;
use std::sync::Arc;

use chrono::{DateTime, Utc};
use uuid::Uuid;

/// Algorithm type for key generation
#[derive(Debug, Clone, PartialEq, Eq)]
pub enum KeyAlgorithm {
    /// ECDSA over secp256k1 (Bitcoin, etc.)
    EcdsaSecp256k1,
    /// ECDSA over P-256 (Ethereum, etc.)
    EcdsaP256,
    /// EdDSA (Solana, etc.)
    EdDSA,
    /// RSA for证书 operations
    Rsa2048,
    /// RSA for证书 operations
    Rsa4096,
}

impl fmt::Display for KeyAlgorithm {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            KeyAlgorithm::EcdsaSecp256k1 => write!(f, "ECDSA_SECP256K1"),
            KeyAlgorithm::EcdsaP256 => write!(f, "ECDSA_P256"),
            KeyAlgorithm::EdDSA => write!(f, "ED25519"),
            KeyAlgorithm::Rsa2048 => write!(f, "RSA_2048"),
            KeyAlgorithm::Rsa4096 => write!(f, "RSA_4096"),
        }
    }
}

/// Key usage purpose
#[derive(Debug, Clone, PartialEq, Eq)]
pub enum KeyUsage {
    /// Signing transactions
    Signing,
    /// Encrypting data
    Encryption,
    /// Authentication
    Authentication,
    /// Certificate signing
    CertificateSigning,
    /// Key derivation
    KeyDerivation,
}

/// Key status
#[derive(Debug, Clone, PartialEq, Eq)]
pub enum KeyStatus {
    /// Key is active and can be used
    Active,
    /// Key is rotated, backup still valid
    Rotated,
    /// Key is compromised, immediate action required
    Compromised,
    /// Key is retired, no longer valid
    Retired,
    /// Key is pending activation
    Pending,
    /// Key is destroyed
    Destroyed,
}

/// Key metadata (does not contain actual key material)
#[derive(Debug, Clone)]
pub struct KeyMetadata {
    /// Unique key identifier
    pub id: Uuid,
    /// Human-readable key label
    pub label: String,
    /// Key algorithm
    pub algorithm: KeyAlgorithm,
    /// Key usage
    pub usage: KeyUsage,
    /// HSM slot/token handle (not key material)
    pub hsm_slot_id: String,
    /// HSM key handle reference
    pub hsm_key_handle: String,
    /// Public key bytes (can be stored externally)
    pub public_key: Vec<u8>,
    /// Key status
    pub status: KeyStatus,
    /// Associated wallet ID (if applicable)
    pub wallet_id: Option<Uuid>,
    /// Key version (for rotation tracking)
    pub version: u32,
    /// Creation timestamp
    pub created_at: DateTime<Utc>,
    /// Activation timestamp
    pub activated_at: Option<DateTime<Utc>>,
    /// Expiration timestamp (None = no expiration)
    pub expires_at: Option<DateTime<Utc>>,
    /// Last used timestamp
    pub last_used_at: Option<DateTime<Utc>>,
    /// Rotation policy ID
    pub rotation_policy_id: Option<Uuid>,
    /// Number of times used
    pub usage_count: u64,
}

/// Key rotation policy
#[derive(Debug, Clone)]
pub struct KeyRotationPolicy {
    /// Unique identifier
    pub id: Uuid,
    /// Policy name
    pub name: String,
    /// Description
    pub description: String,
    /// Key algorithm this applies to
    pub algorithm: KeyAlgorithm,
    /// Maximum age in days before rotation required
    pub max_age_days: u32,
    /// Maximum usage count before rotation
    pub max_usage_count: u64,
    /// Whether automatic rotation is enabled
    pub auto_rotate: bool,
    /// Grace period in days before key becomes unusable
    pub grace_period_days: u32,
    /// Number of backup keys to maintain
    pub backup_keys_count: u8,
    /// Created timestamp
    pub created_at: DateTime<Utc>,
}

/// Multi-signature configuration
#[derive(Debug, Clone)]
pub struct MultiSigConfig {
    /// Unique identifier
    pub id: Uuid,
    /// Configuration name
    pub name: String,
    /// Description
    pub description: String,
    /// Required number of signatures
    pub required_signatures: u8,
    /// Total number of signers
    pub total_signers: u8,
    /// List of authorized signers
    pub signers: Vec<Signer>,
    /// Whether this is a rolling quorum
    pub is_rolling: bool,
    /// Time limit for signature collection (seconds)
    pub signature_timeout_seconds: u32,
    /// Whether to allow partial signing with escalation
    pub allow_partial_with_escalation: bool,
    /// Minimum signers for escalation level
    pub escalation_threshold: Option<u8>,
}

/// Signer in a multi-signature setup
#[derive(Debug, Clone)]
pub struct Signer {
    /// Unique signer ID
    pub id: Uuid,
    /// Signer type
    pub signer_type: SignerType,
    /// Signer identifier
    pub identifier: String,
    /// Signer label/name
    pub label: String,
    /// HSM key handle
    pub hsm_key_handle: String,
    /// Whether signer is currently active
    pub is_active: bool,
    /// Order in signing sequence (if ordered)
    pub order: Option<u8>,
}

/// Types of signers
#[derive(Debug, Clone, PartialEq, Eq)]
pub enum SignerType {
    /// HSM-managed key
    HsmKey,
    /// Hardware wallet
    HardwareWallet,
    /// Admin user with MFA
    AdminUser,
    /// Threshold signature scheme
    Threshold,
    /// MPC (Multi-Party Computation) node
    MpcNode,
    /// External custodian
    ExternalCustodian,
}

impl fmt::Display for SignerType {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            SignerType::HsmKey => write!(f, "HSM_KEY"),
            SignerType::HardwareWallet => write!(f, "HARDWARE_WALLET"),
            SignerType::AdminUser => write!(f, "ADMIN_USER"),
            SignerType::Threshold => write!(f, "THRESHOLD"),
            SignerType::MpcNode => write!(f, "MPC_NODE"),
            SignerType::ExternalCustodian => write!(f, "EXTERNAL_CUSTODIAN"),
        }
    }
}

/// Signature data
#[derive(Debug, Clone)]
pub struct Signature {
    /// Signature ID
    pub id: Uuid,
    /// Associated signing request ID
    pub request_id: Uuid,
    /// Signer ID who created this signature
    pub signer_id: Uuid,
    /// Signature bytes (DER or raw)
    pub signature_bytes: Vec<u8>,
    /// Signature type/format
    pub signature_type: SignatureType,
    /// Timestamp of signature
    pub signed_at: DateTime<Utc>,
    /// Whether signature is valid
    pub is_valid: bool,
}

/// Types of signature formats
#[derive(Debug, Clone, PartialEq, Eq)]
pub enum SignatureType {
    /// DER-encoded ECDSA signature
    EcdsaDer,
    /// Raw ECDSA signature (r || s)
    EcdsaRaw,
    /// EdDSA signature
    EdDSA,
    /// Bitcoin-style signature
    Bitcoin,
    /// Multi-signature combined
    MultiSig,
    /// Threshold signature
    Threshold,
}

impl fmt::Display for SignatureType {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            SignatureType::EcdsaDer => write!(f, "ECDSA_DER"),
            SignatureType::EcdsaRaw => write!(f, "ECDSA_RAW"),
            SignatureType::EdDSA => write!(f, "ED25519"),
            SignatureType::Bitcoin => write!(f, "BITCOIN"),
            SignatureType::MultiSig => write!(f, "MULTI_SIG"),
            SignatureType::Threshold => write!(f, "THRESHOLD"),
        }
    }
}

/// Signing request status
#[derive(Debug, Clone, PartialEq, Eq)]
pub enum SigningStatus {
    /// Request created, awaiting signatures
    Pending,
    /// Some signatures collected
    PartiallySigned,
    /// All required signatures collected
    Complete,
    /// Request was rejected
    Rejected,
    /// Request timed out
    TimedOut,
    /// Request was cancelled
    Cancelled,
}

/// Signing request for multi-signature transactions
#[derive(Debug, Clone)]
pub struct SigningRequest {
    /// Unique request ID
    pub id: Uuid,
    /// Associated wallet ID
    pub wallet_id: Uuid,
    /// Multi-sig configuration ID
    pub multisig_config_id: Uuid,
    /// Transaction payload to sign
    pub payload: Vec<u8>,
    /// Payload hash
    pub payload_hash: Vec<u8>,
    /// Required number of signatures
    pub required_signatures: u8,
    /// Current signatures collected
    pub signatures: Vec<Signature>,
    /// Current status
    pub status: SigningStatus,
    /// Reason for rejection (if applicable)
    pub rejection_reason: Option<String>,
    /// Created timestamp
    pub created_at: DateTime<Utc>,
    /// Expiration timestamp
    pub expires_at: DateTime<Utc>,
    /// Completed timestamp
    pub completed_at: Option<DateTime<Utc>>,
}

/// Signing request creation request
#[derive(Debug, Clone, Validate)]
pub struct CreateSigningRequest {
    #[validate(uuid)]
    pub wallet_id: String,
    
    #[validate(uuid)]
    pub multisig_config_id: String,
    
    #[validate(length(min = 1, max = 1000))]
    pub payload: Vec<u8>,
}

/// Emergency access configuration
#[derive(Debug, Clone)]
pub struct EmergencyAccessConfig {
    /// Unique identifier
    pub id: Uuid,
    /// Configuration name
    pub name: String,
    /// Description
    pub description: String,
    /// Minimum number of admin approvals required
    pub required_approvals: u8,
    /// List of authorized admin IDs
    pub admin_ids: Vec<Uuid>,
    /// Whether emergency access is enabled
    pub is_enabled: bool,
    /// Time limit for emergency operations (seconds)
    pub operation_timeout_seconds: u32,
    /// Actions allowed in emergency mode
    pub allowed_actions: Vec<EmergencyAction>,
    /// Cooldown period between emergency activations (seconds)
    pub cooldown_seconds: u32,
    /// Created timestamp
    pub created_at: DateTime<Utc>,
}

/// Types of emergency actions
#[derive(Debug, Clone, PartialEq, Eq)]
pub enum EmergencyAction {
    /// Freeze all wallets
    FreezeAllWallets,
    /// Suspend transactions
    SuspendTransactions,
    /// Emergency key rotation
    EmergencyKeyRotation,
    /// Force transaction signing
    ForceSign,
    /// Export wallet data
    ExportWalletData,
    /// Revoke admin access
    RevokeAdminAccess,
    /// Enable read-only mode
    EnableReadOnly,
}

impl fmt::Display for EmergencyAction {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            EmergencyAction::FreezeAllWallets => write!(f, "FREEZE_ALL_WALLETS"),
            EmergencyAction::SuspendTransactions => write!(f, "SUSPEND_TRANSACTIONS"),
            EmergencyAction::EmergencyKeyRotation => write!(f, "EMERGENCY_KEY_ROTATION"),
            EmergencyAction::ForceSign => write!(f, "FORCE_SIGN"),
            EmergencyAction::ExportWalletData => write!(f, "EXPORT_WALLET_DATA"),
            EmergencyAction::RevokeAdminAccess => write!(f, "REVOKE_ADMIN_ACCESS"),
            EmergencyAction::EnableReadOnly => write!(f, "ENABLE_READ_ONLY"),
        }
    }
}

/// Emergency access request
#[derive(Debug, Clone, Validate)]
pub struct EmergencyAccessRequest {
    /// Type of emergency action
    pub action: EmergencyAction,
    
    /// Reason for emergency access
    #[validate(length(min = 10, max = 500))]
    pub reason: String,
    
    /// Authorization level
    pub authorization_level: AuthorizationLevel,
    
    /// Target wallet IDs (if applicable)
    #[validate]
    pub target_wallet_ids: Option<Vec<Uuid>>,
    
    /// Duration in seconds (if applicable)
    pub duration_seconds: Option<u32>,
}

/// Authorization levels for emergency access
#[derive(Debug, Clone, PartialEq, Eq)]
pub enum AuthorizationLevel {
    /// Level 1: Auto-approved (for system-initiated)
    System,
    /// Level 2: Single admin approval
    SingleAdmin,
    /// Level 3: Dual admin approval
    DualAdmin,
    /// Level 4: Multi-admin with time delay
    MultiAdminDelay,
    /// Level 5: Emergency council required
    EmergencyCouncil,
}

impl fmt::Display for AuthorizationLevel {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            AuthorizationLevel::System => write!(f, "SYSTEM"),
            AuthorizationLevel::SingleAdmin => write!(f, "SINGLE_ADMIN"),
            AuthorizationLevel::DualAdmin => write!(f, "DUAL_ADMIN"),
            AuthorizationLevel::MultiAdminDelay => write!(f, "MULTI_ADMIN_DELAY"),
            AuthorizationLevel::EmergencyCouncil => write!(f, "EMERGENCY_COUNCIL"),
        }
    }
}

/// Result of a signing operation
#[derive(Debug, Clone)]
pub struct SigningResult {
    /// Whether signing was successful
    pub success: bool,
    /// Signature if successful
    pub signature: Option<Signature>,
    /// Error message if failed
    pub error: Option<String>,
    /// Whether additional signatures are required
    pub requires_more_signatures: bool,
    /// Current signature count
    pub current_signature_count: u8,
    /// Required signature count
    pub required_signature_count: u8,
}
