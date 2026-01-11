//! Domain models for the Custodial Wallet Governance System
//!
//! This module contains the core domain entities, value objects, and aggregates
//! that represent the business logic of wallet governance, policy management,
//! cryptographic operations, and audit logging.

pub mod wallet;
pub mod policy;
pub mod crypto;
pub mod audit;

pub use wallet::{
    Wallet, WalletType, WalletStatus, FreezeReason, ChainType,
    CreateWalletRequest, FreezeWalletRequest, UnfreezeWalletRequest,
    WalletRecoveryRequest, RecoveryMethod,
};

pub use policy::{
    Policy, PolicyRule, PolicySeverity, PolicyEvaluationResult, 
    PolicyRequirement, PolicyEvaluationContext, CreatePolicyRequest,
    BlacklistEntry, WhitelistEntry, BlacklistSource, ReviewerType,
};

pub use crypto::{
    KeyMetadata, KeyAlgorithm, KeyUsage, KeyStatus,
    KeyRotationPolicy, MultiSigConfig, Signer, SignerType,
    Signature, SignatureType, SigningRequest, SigningStatus,
    SigningResult, CreateSigningRequest,
    EmergencyAccessConfig, EmergencyAction, EmergencyAccessRequest,
    AuthorizationLevel,
};

pub use audit::{
    AuditEvent, AuditCategory, AuditEventType, AuditSeverity,
    AuditActor, ActorType, AuditQuery, AuditQueryResult,
    AuditSortField, SortDirection,
    ComplianceReportRequest, ComplianceReportType, ReportFormat,
};

use std::fmt;

/// Result type for domain operations
pub type DomainResult<T> = Result<T, DomainError>;

/// Domain errors
#[derive(Debug, thiserror::Error)]
pub enum DomainError {
    #[error("Wallet not found: {0}")]
    WalletNotFound(String),
    
    #[error("Wallet is not active: {0}")]
    WalletNotActive(String),
    
    #[error("Wallet is frozen: {0}")]
    WalletFrozen(String),
    
    #[error("Transaction limit exceeded: {max} attempted {attempted}")]
    TransactionLimitExceeded { max: String, attempted: String },
    
    #[error("Policy violation: {0}")]
    PolicyViolation(String),
    
    #[error("Blacklisted address: {0}")]
    BlacklistedAddress(String),
    
    #[error("Unauthorized: {0}")]
    Unauthorized(String),
    
    #[error("Invalid signature: {0}")]
    InvalidSignature(String),
    
    #[error("Insufficient signatures: required {required}, have {have}")]
    InsufficientSignatures { required: u8, have: u8 },
    
    #[error("Signing request expired")]
    SigningRequestExpired,
    
    #[error("Signing request not found: {0}")]
    SigningRequestNotFound(String),
    
    #[error("Key not found: {0}")]
    KeyNotFound(String),
    
    #[error("Key is not active")]
    KeyNotActive,
    
    #[error("HSM operation failed: {0}")]
    HsmOperationFailed(String),
    
    #[error("Emergency access denied: {0}")]
    EmergencyAccessDenied(String),
    
    #[error("Emergency access requires authorization level {required}, have {have}")]
    InsufficientAuthorization { required: String, have: String },
    
    #[error("Recovery already in progress")]
    RecoveryInProgress,
    
    #[error("Policy not found: {0}")]
    PolicyNotFound(String),
    
    #[error("Policy is not active")]
    PolicyNotActive,
    
    #[error("Configuration error: {0}")]
    ConfigurationError(String),
    
    #[error("Validation error: {0}")]
    ValidationError(String),
    
    #[error("Operation not permitted: {0}")]
    OperationNotPermitted(String),
}

/// Extension trait for converting domain errors to audit severity
pub trait ToAuditSeverity {
    fn to_audit_severity(&self) -> audit::AuditSeverity;
}

impl ToAuditSeverity for DomainError {
    fn to_audit_severity(&self) -> audit::AuditSeverity {
        match self {
            DomainError::WalletNotFound(_) => audit::AuditSeverity::Low,
            DomainError::WalletNotActive(_) => audit::AuditSeverity::Medium,
            DomainError::WalletFrozen(_) => audit::AuditSeverity::High,
            DomainError::TransactionLimitExceeded { .. } => audit::AuditSeverity::Medium,
            DomainError::PolicyViolation(_) => audit::AuditSeverity::Medium,
            DomainError::BlacklistedAddress(_) => audit::AuditSeverity::Critical,
            DomainError::Unauthorized(_) => audit::AuditSeverity::Medium,
            DomainError::InvalidSignature(_) => audit::AuditSeverity::High,
            DomainError::InsufficientSignatures { .. } => audit::AuditSeverity::Medium,
            DomainError::SigningRequestExpired => audit::AuditSeverity::Medium,
            DomainError::SigningRequestNotFound(_) => audit::AuditSeverity::Low,
            DomainError::KeyNotFound(_) => audit::AuditSeverity::Medium,
            DomainError::KeyNotActive => audit::AuditSeverity::High,
            DomainError::HsmOperationFailed(_) => audit::AuditSeverity::Critical,
            DomainError::EmergencyAccessDenied(_) => audit::AuditSeverity::Critical,
            DomainError::InsufficientAuthorization { .. } => audit::AuditSeverity::High,
            DomainError::RecoveryInProgress => audit::AuditSeverity::Medium,
            DomainError::PolicyNotFound(_) => audit::AuditSeverity::Low,
            DomainError::PolicyNotActive => audit::AuditSeverity::Medium,
            DomainError::ConfigurationError(_) => audit::AuditSeverity::High,
            DomainError::ValidationError(_) => audit::AuditSeverity::Low,
            DomainError::OperationNotPermitted(_) => audit::AuditSeverity::High,
        }
    }
}
