use std::collections::HashSet;
use std::fmt;

use chrono::{DateTime, Utc};
use rust_decimal::Decimal;
use serde::{Deserialize, Serialize};
use uuid::Uuid;

/// Severity level of a policy rule
#[derive(Debug, Clone, PartialEq, Eq, sqlx::Type)]
#[sqlx(type_name = "policy_severity")]
pub enum PolicySeverity {
    /// Transaction is blocked immediately
    Blocking,
    /// Transaction is flagged for review but allowed
    Warning,
    /// Transaction requires additional approval
    EnhancedReview,
}

impl fmt::Display for PolicySeverity {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            PolicySeverity::Blocking => write!(f, "BLOCKING"),
            PolicySeverity::Warning => write!(f, "WARNING"),
            PolicySeverity::EnhancedReview => write!(f, "ENHANCED_REVIEW"),
        }
    }
}

/// Types of policy rules
#[derive(Debug, Clone, PartialEq, Eq, Serialize, Deserialize)]
pub enum PolicyRule {
    /// Whitelist specific addresses (always allowed)
    WhitelistAddress {
        address: String,
        chain_type: Option<String>,
        description: String,
    },
    /// Blacklist specific addresses (always blocked)
    BlacklistAddress {
        address: String,
        chain_type: Option<String>,
        reason: String,
        source: BlacklistSource,
    },
    /// Maximum transaction amount limit
    MaxTransactionLimit {
        max_amount: Decimal,
        chain_type: Option<String>,
    },
    /// Maximum daily withdrawal limit
    DailyWithdrawalLimit {
        max_amount: Decimal,
        window_hours: u32,
    },
    /// Velocity limit - minimum time between transactions
    VelocityLimit {
        min_interval_seconds: u64,
        per_wallet: bool,
    },
    /// Require multi-signature for transactions
    MultiSigRequirement {
        required_signatures: u8,
        total_signers: u8,
        description: String,
    },
    /// Geographic restrictions
    GeographicRestriction {
        allowed_countries: HashSet<String>,
        blocked_countries: HashSet<String>,
    },
    /// Require identity verification level
    KycRequirement {
        min_kyc_level: u8,
    },
    /// Time-based restrictions
    TimeRestriction {
        allowed_hours_utc: Vec<u8>, // Hours 0-23
        blocked_hours_utc: Vec<u8>,
    },
    /// Whitelist for withdrawal destinations
    DestinationWhitelist {
        require_destination_whitelist: bool,
        allowed_destinations: HashSet<String>,
    },
}

/// Source of blacklist entry
#[derive(Debug, Clone, PartialEq, Eq, sqlx::Type)]
#[sqlx(type_name = "blacklist_source")]
pub enum BlacklistSource {
    /// Government regulatory mandate
    Regulatory,
    /// Law enforcement request
    LawEnforcement,
    /// Internal risk assessment
    InternalRisk,
    /// Exchange self-reported
    ExchangeReported,
    /// Community reported
    CommunityReported,
    /// Automated detection system
    AutomatedDetection,
}

impl fmt::Display for BlacklistSource {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            BlacklistSource::Regulatory => write!(f, "REGULATORY"),
            BlacklistSource::LawEnforcement => write!(f, "LAW_ENFORCEMENT"),
            BlacklistSource::InternalRisk => write!(f, "INTERNAL_RISK"),
            BlacklistSource::ExchangeReported => write!(f, "EXCHANGE_REPORTED"),
            BlacklistSource::CommunityReported => write!(f, "COMMUNITY_REPORTED"),
            BlacklistSource::AutomatedDetection => write!(f, "AUTOMATED_DETECTION"),
        }
    }
}

/// Result of policy evaluation
#[derive(Debug, Clone)]
pub struct PolicyEvaluationResult {
    /// Whether the transaction is allowed
    pub is_allowed: bool,
    /// Severity level if not allowed
    pub severity: PolicySeverity,
    /// List of failed rules
    pub failed_rules: Vec<String>,
    /// List of warnings
    pub warnings: Vec<String>,
    /// Additional requirements triggered
    pub requirements: Vec<PolicyRequirement>,
    /// Risk score (0-100)
    pub risk_score: u8,
}

/// Additional requirements triggered by policy
#[derive(Debug, Clone, PartialEq, Eq)]
pub enum PolicyRequirement {
    /// Requires multi-signature approval
    RequiresMultiSig {
        required: u8,
        total: u8,
    },
    /// Requires manual review
    RequiresManualReview {
        reviewer_type: ReviewerType,
    },
    /// Requires enhanced KYC verification
    RequiresEnhancedKyc,
    /// Transaction is flagged for reporting
    RequiresRegulatoryReport,
    /// Requires additional time delay
    RequiresDelay {
        seconds: u64,
    },
}

/// Types of reviewers for manual review
#[derive(Debug, Clone, PartialEq, Eq)]
pub enum ReviewerType {
    /// Compliance officer review required
    ComplianceOfficer,
    /// Risk team review required
    RiskTeam,
    /// Senior management review required
    SeniorManagement,
    /// Legal team review required
    LegalTeam,
}

/// Governance policy aggregate
#[derive(Debug, Clone)]
pub struct Policy {
    /// Unique identifier
    pub id: Uuid,
    /// Policy name
    pub name: String,
    /// Policy description
    pub description: String,
    /// Policy version
    pub version: u32,
    /// Whether the policy is currently active
    pub is_active: bool,
    /// Priority (higher = evaluated first)
    pub priority: i32,
    /// List of rules in this policy
    pub rules: Vec<PolicyRule>,
    /// Applies to these wallet types (empty = all)
    pub applies_to_wallet_types: Vec<wallet::WalletType>,
    /// Applies to these chains (empty = all)
    pub applies_to_chains: Vec<wallet::ChainType>,
    /// Effective from timestamp
    pub effective_from: DateTime<Utc>,
    /// Effective until timestamp (None = no expiration)
    pub effective_until: Option<DateTime<Utc>>,
    /// Created by user ID
    pub created_by: String,
    /// Created timestamp
    pub created_at: DateTime<Utc>,
    /// Last updated timestamp
    pub updated_at: DateTime<Utc>,
}

impl Policy {
    /// Creates a new policy
    pub fn new(
        name: String,
        description: String,
        created_by: String,
    ) -> Self {
        let now = Utc::now();
        Self {
            id: Uuid::new_v4(),
            name,
            description,
            version: 1,
            is_active: true,
            priority: 0,
            rules: Vec::new(),
            applies_to_wallet_types: Vec::new(),
            applies_to_chains: Vec::new(),
            effective_from: now,
            effective_until: None,
            created_by,
            created_at: now,
            updated_at: now,
        }
    }

    /// Adds a rule to the policy
    pub fn add_rule(&mut self, rule: PolicyRule) {
        self.rules.push(rule);
        self.updated_at = Utc::now();
    }

    /// Checks if policy is currently effective
    pub fn is_effective(&self) -> bool {
        let now = Utc::now();
        self.is_active && now >= self.effective_from && 
            self.effective_until.map_or(true, |until| now < until)
    }

    /// Checks if policy applies to a specific wallet
    pub fn applies_to(&self, wallet_type: &wallet::WalletType, chain: &wallet::ChainType) -> bool {
        // Check wallet type filter
        if !self.applies_to_wallet_types.is_empty() {
            if !self.applies_to_wallet_types.contains(wallet_type) {
                return false;
            }
        }
        
        // Check chain filter
        if !self.applies_to_chains.is_empty() {
            if !self.applies_to_chains.contains(chain) {
                return false;
            }
        }
        
        true
    }

    /// Increments version
    pub fn increment_version(&mut self) {
        self.version += 1;
        self.updated_at = Utc::now();
    }
}

/// Policy evaluation context
#[derive(Debug, Clone)]
pub struct PolicyEvaluationContext {
    /// Source wallet
    pub source_wallet: Option<&wallet::Wallet>,
    /// Destination address
    pub destination_address: String,
    /// Transaction amount
    pub amount: Decimal,
    /// Chain type
    pub chain_type: wallet::ChainType,
    /// Origin country (ISO code)
    pub origin_country: Option<String>,
    /// User's KYC level
    pub user_kyc_level: u8,
    /// Transaction timestamp
    pub timestamp: DateTime<Utc>,
    /// Additional metadata
    pub metadata: serde_json::Value,
}

/// Policy creation request
#[derive(Debug, Clone, Validate)]
pub struct CreatePolicyRequest {
    #[validate(length(min = 3, max = 100))]
    pub name: String,
    
    #[validate(length(max = 500))]
    pub description: Option<String>,
    
    pub severity: PolicySeverity,
    
    pub rules: Vec<PolicyRule>,
    
    #[validate(length(max = 10))]
    pub applies_to_wallet_types: Option<Vec<wallet::WalletType>>,
    
    #[validate(length(max = 10))]
    pub applies_to_chains: Option<Vec<wallet::ChainType>>,
}

/// Blacklist entry
#[derive(Debug, Clone)]
pub struct BlacklistEntry {
    /// Unique identifier
    pub id: Uuid,
    /// Blacklisted address
    pub address: String,
    /// Chain type (None = all chains)
    pub chain_type: Option<wallet::ChainType>,
    /// Reason for blacklisting
    pub reason: String,
    /// Source of the blacklist entry
    pub source: BlacklistSource,
    /// Risk level (1-5)
    pub risk_level: u8,
    /// Whether entry is currently active
    pub is_active: bool,
    /// Expiration date (None = permanent)
    pub expires_at: Option<DateTime<Utc>>,
    /// Created by
    pub created_by: String,
    /// Created timestamp
    pub created_at: DateTime<Utc>,
    /// Updated timestamp
    pub updated_at: DateTime<Utc>,
}

impl BlacklistEntry {
    /// Creates a new blacklist entry
    pub fn new(
        address: String,
        chain_type: Option<wallet::ChainType>,
        reason: String,
        source: BlacklistSource,
        risk_level: u8,
        created_by: String,
    ) -> Self {
        let now = Utc::now();
        Self {
            id: Uuid::new_v4(),
            address,
            chain_type,
            reason,
            source,
            risk_level: risk_level.clamp(1, 5),
            is_active: true,
            expires_at: None,
            created_by,
            created_at: now,
            updated_at: now,
        }
    }

    /// Checks if entry is currently valid
    pub fn is_valid(&self) -> bool {
        self.is_active && 
            self.expires_at.map_or(true, |exp| Utc::now() < exp)
    }

    /// Expires the entry
    pub fn expire(&mut self) {
        self.is_active = false;
        self.updated_at = Utc::now();
    }
}

/// Whitelist entry
#[derive(Debug, Clone)]
pub struct WhitelistEntry {
    /// Unique identifier
    pub id: Uuid,
    /// Whitelisted address
    pub address: String,
    /// Chain type (None = all chains)
    pub chain_type: Option<wallet::ChainType>,
    /// Description/label
    pub label: String,
    /// Associated entity (if any)
    pub entity_name: Option<String>,
    /// Verification level
    pub verification_level: u8,
    /// Whether entry is currently active
    pub is_active: bool,
    /// Created by
    pub created_by: String,
    /// Created timestamp
    pub created_at: DateTime<Utc>,
    /// Updated timestamp
    pub updated_at: DateTime<Utc>,
}

impl WhitelistEntry {
    /// Creates a new whitelist entry
    pub fn new(
        address: String,
        chain_type: Option<wallet::ChainType>,
        label: String,
        entity_name: Option<String>,
        created_by: String,
    ) -> Self {
        let now = Utc::now();
        Self {
            id: Uuid::new_v4(),
            address,
            chain_type,
            label,
            entity_name,
            verification_level: 1,
            is_active: true,
            created_by,
            created_at: now,
            updated_at: now,
        }
    }
}
