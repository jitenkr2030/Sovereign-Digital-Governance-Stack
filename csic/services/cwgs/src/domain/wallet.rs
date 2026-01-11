use std::fmt;
use std::str::FromStr;

use chrono::{DateTime, Utc};
use rust_decimal::Decimal;
use uuid::Uuid;

/// Represents the type of blockchain wallet
#[derive(Debug, Clone, PartialEq, Eq, sqlx::Type)]
#[sqlx(type_name = "wallet_type")]
pub enum WalletType {
    /// Hot wallet - online and accessible for frequent transactions
    Hot,
    /// Cold wallet - offline storage for long-term holdings
    Cold,
    /// Operational wallet - for day-to-day operations
    Operational,
}

impl fmt::Display for WalletType {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            WalletType::Hot => write!(f, "HOT"),
            WalletType::Cold => write!(f, "COLD"),
            WalletType::Operational => write!(f, "OPERATIONAL"),
        }
    }
}

impl FromStr for WalletType {
    type Err = anyhow::Error;

    fn from_str(s: &str) -> Result<Self, Self::Err> {
        match s.to_uppercase().as_str() {
            "HOT" => Ok(WalletType::Hot),
            "COLD" => Ok(WalletType::Cold),
            "OPERATIONAL" => Ok(WalletType::Operational),
            _ => Err(anyhow::anyhow!("Invalid wallet type: {}", s)),
        }
    }
}

/// Current status of a wallet
#[derive(Debug, Clone, PartialEq, Eq, sqlx::Type)]
#[sqlx(type_name = "wallet_status")]
pub enum WalletStatus {
    /// Wallet is active and can process transactions
    Active,
    /// Wallet is temporarily frozen (regulatory hold, investigation)
    Frozen(FreezeReason),
    /// Wallet is pending key recovery procedures
    PendingRecovery,
    /// Wallet is archived and no longer in use
    Archived,
    /// Wallet is compromised and requires emergency response
    Compromised,
}

impl fmt::Display for WalletStatus {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            WalletStatus::Active => write!(f, "ACTIVE"),
            WalletStatus::Frozen(reason) => write!(f, "FROZEN: {}", reason),
            WalletStatus::PendingRecovery => write!(f, "PENDING_RECOVERY"),
            WalletStatus::Archived => write!(f, "ARCHIVED"),
            WalletStatus::Compromised => write!(f, "COMPROMISED"),
        }
    }
}

/// Reason for wallet freeze
#[derive(Debug, Clone, PartialEq, Eq, sqlx::Type)]
#[sqlx(type_name = "freeze_reason")]
pub enum FreezeReason {
    /// Regulatory investigation
    RegulatoryHold,
    /// Suspicious activity detected
    SuspiciousActivity,
    /// Court order or legal requirement
    LegalOrder,
    /// User request for security freeze
    UserRequest,
    /// System-initiated security measure
    SecurityMeasure,
    /// Maintenance or operational freeze
    Maintenance,
}

impl fmt::Display for FreezeReason {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            FreezeReason::RegulatoryHold => write!(f, "REGULATORY_HOLD"),
            FreezeReason::SuspiciousActivity => write!(f, "SUSPICIOUS_ACTIVITY"),
            FreezeReason::LegalOrder => write!(f, "LEGAL_ORDER"),
            FreezeReason::UserRequest => write!(f, "USER_REQUEST"),
            FreezeReason::SecurityMeasure => write!(f, "SECURITY_MEASURE"),
            FreezeReason::Maintenance => write!(f, "MAINTENANCE"),
        }
    }
}

impl FromStr for FreezeReason {
    type Err = anyhow::Error;

    fn from_str(s: &str) -> Result<Self, Self::Err> {
        match s.to_uppercase().as_str() {
            "REGULATORY_HOLD" => Ok(FreezeReason::RegulatoryHold),
            "SUSPICIOUS_ACTIVITY" => Ok(FreezeReason::SuspiciousActivity),
            "LEGAL_ORDER" => Ok(FreezeReason::LegalOrder),
            "USER_REQUEST" => Ok(FreezeReason::UserRequest),
            "SECURITY_MEASURE" => Ok(FreezeReason::SecurityMeasure),
            "MAINTENANCE" => Ok(FreezeReason::Maintenance),
            _ => Err(anyhow::anyhow!("Invalid freeze reason: {}", s)),
        }
    }
}

/// Supported blockchain networks
#[derive(Debug, Clone, PartialEq, Eq, Hash, sqlx::Type)]
#[sqlx(type_name = "chain_type")]
pub enum ChainType {
    /// Bitcoin network
    Bitcoin,
    /// Ethereum network
    Ethereum,
    /// Polygon network
    Polygon,
    /// Solana network
    Solana,
    /// Arbitrum network
    Arbitrum,
    /// Optimism network
    Optimism,
}

impl fmt::Display for ChainType {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            ChainType::Bitcoin => write!(f, "BTC"),
            ChainType::Ethereum => write!(f, "ETH"),
            ChainType::Polygon => write!(f, "MATIC"),
            ChainType::Solana => write!(f, "SOL"),
            ChainType::Arbitrum => write!(f, "ARB"),
            ChainType::Optimism => write!(f, "OPT"),
        }
    }
}

impl FromStr for ChainType {
    type Err = anyhow::Error;

    fn from_str(s: &str) -> Result<Self, Self::Err> {
        match s.to_uppercase().as_str() {
            "BTC" | "BITCOIN" => Ok(ChainType::Bitcoin),
            "ETH" | "ETHEREUM" => Ok(ChainType::Ethereum),
            "MATIC" | "POLYGON" => Ok(ChainType::Polygon),
            "SOL" | "SOLANA" => Ok(ChainType::Solana),
            "ARB" | "ARBITRUM" => Ok(ChainType::Arbitrum),
            "OPT" | "OPTIMISM" => Ok(ChainType::Optimism),
            _ => Err(anyhow::anyhow!("Invalid chain type: {}", s)),
        }
    }
}

/// Core wallet aggregate
#[derive(Debug, Clone)]
pub struct Wallet {
    /// Unique identifier
    pub id: Uuid,
    /// Blockchain address
    pub address: String,
    /// Type of blockchain
    pub chain_type: ChainType,
    /// Wallet classification
    pub wallet_type: WalletType,
    /// Current operational status
    pub status: WalletStatus,
    /// Associated signing policy ID
    pub signing_policy_id: Uuid,
    /// HSM key handle reference (not the actual key)
    pub hsm_key_handle: String,
    /// Public key for verification (derived from HSM)
    pub public_key: Vec<u8>,
    /// Maximum single transaction limit
    pub max_tx_limit: Decimal,
    /// Maximum daily withdrawal limit
    pub daily_limit: Decimal,
    /// Current daily usage
    pub daily_usage: Decimal,
    /// Last activity timestamp
    pub last_activity_at: Option<DateTime<Utc>>,
    /// Creation timestamp
    pub created_at: DateTime<Utc>,
    /// Last update timestamp
    pub updated_at: DateTime<Utc>,
}

impl Wallet {
    /// Creates a new active wallet
    pub fn new(
        address: String,
        chain_type: ChainType,
        wallet_type: WalletType,
        signing_policy_id: Uuid,
        hsm_key_handle: String,
        public_key: Vec<u8>,
    ) -> Self {
        let now = Utc::now();
        Self {
            id: Uuid::new_v4(),
            address,
            chain_type,
            wallet_type,
            status: WalletStatus::Active,
            signing_policy_id,
            hsm_key_handle,
            public_key,
            max_tx_limit: Decimal::from(100),
            daily_limit: Decimal::from(1000),
            daily_usage: Decimal::from(0),
            last_activity_at: None,
            created_at: now,
            updated_at: now,
        }
    }

    /// Checks if wallet can process transactions
    pub fn can_transact(&self) -> bool {
        matches!(self.status, WalletStatus::Active)
    }

    /// Freezes the wallet with a reason
    pub fn freeze(&mut self, reason: FreezeReason) {
        self.status = WalletStatus::Frozen(reason);
        self.updated_at = Utc::now();
    }

    /// Unfreezes the wallet (only valid for certain freeze reasons)
    pub fn unfreeze(&mut self) {
        if matches!(self.status, WalletStatus::Frozen(_)) {
            self.status = WalletStatus::Active;
            self.updated_at = Utc::now();
        }
    }

    /// Initiates recovery process
    pub fn initiate_recovery(&mut self) {
        self.status = WalletStatus::PendingRecovery;
        self.updated_at = Utc::now();
    }

    /// Checks if transaction amount is within limits
    pub fn check_limits(&self, amount: Decimal) -> bool {
        amount <= self.max_tx_limit && (self.daily_usage + amount) <= self.daily_limit
    }

    /// Records transaction activity
    pub fn record_activity(&mut self, amount: Decimal) {
        self.daily_usage += amount;
        self.last_activity_at = Some(Utc::now());
        self.updated_at = Utc::now();
    }

    /// Resets daily usage (called at start of new day)
    pub fn reset_daily_usage(&mut self) {
        self.daily_usage = Decimal::from(0);
        self.updated_at = Utc::now();
    }
}

/// Wallet creation request
#[derive(Debug, Clone, Validate)]
pub struct CreateWalletRequest {
    /// Blockchain address
    #[validate(length(min = 26, max = 70))]
    pub address: String,
    
    /// Type of blockchain
    pub chain_type: ChainType,
    
    /// Wallet classification
    pub wallet_type: WalletType,
    
    /// Maximum single transaction limit
    #[validate(decimal_places(equals = 8))]
    pub max_tx_limit: Option<Decimal>,
    
    /// Maximum daily withdrawal limit
    #[validate(decimal_places(equals = 8))]
    pub daily_limit: Option<Decimal>,
}

/// Wallet freeze request
#[derive(Debug, Clone, Validate)]
pub struct FreezeWalletRequest {
    /// Wallet ID to freeze
    #[validate(uuid)]
    pub wallet_id: String,
    
    /// Reason for freezing
    pub reason: FreezeReason,
    
    /// Optional notes
    #[validate(length(max = 500))]
    pub notes: Option<String>,
}

/// Wallet unfreeze request
#[derive(Debug, Clone, Validate)]
pub struct UnfreezeWalletRequest {
    /// Wallet ID to unfreeze
    #[validate(uuid)]
    pub wallet_id: String,
    
    /// Authorization reference (e.g., case number, approval ID)
    #[validate(length(min = 10, max = 100))]
    pub authorization_ref: String,
}

/// Wallet recovery request
#[derive(Debug, Clone, Validate)]
pub struct WalletRecoveryRequest {
    /// Wallet ID to recover
    #[validate(uuid)]
    pub wallet_id: String,
    
    /// Recovery method
    pub recovery_method: RecoveryMethod,
    
    /// New HSM key handle for recovered wallet
    pub new_hsm_key_handle: String,
    
    /// New public key
    pub new_public_key: Vec<u8>,
}

/// Methods for wallet recovery
#[derive(Debug, Clone, PartialEq, Eq)]
pub enum RecoveryMethod {
    /// Multi-signature quorum recovery
    MultiSigQuorum,
    /// HSM backup key recovery
    HsmBackup,
    /// Emergency admin recovery (requires high-level authorization)
    EmergencyAdmin,
    /// Key ceremony with key shards
    KeyShards,
}

impl fmt::Display for RecoveryMethod {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            RecoveryMethod::MultiSigQuorum => write!(f, "MULTI_SIG_QUORUM"),
            RecoveryMethod::HsmBackup => write!(f, "HSM_BACKUP"),
            RecoveryMethod::EmergencyAdmin => write!(f, "EMERGENCY_ADMIN"),
            RecoveryMethod::KeyShards => write!(f, "KEY_SHARDS"),
        }
    }
}
