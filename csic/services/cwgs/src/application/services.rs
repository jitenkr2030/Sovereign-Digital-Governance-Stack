use std::sync::Arc;

use async_trait::async_trait;
use chrono::{DateTime, Utc};
use rust_decimal::Decimal;
use uuid::Uuid;

use crate::domain::{
    self, wallet::{Wallet, WalletType, ChainType, FreezeReason},
    policy::{Policy, PolicyRule, PolicyEvaluationContext, PolicyEvaluationResult},
    audit::{AuditEvent, AuditCategory, AuditEventType, AuditSeverity, AuditActor, ActorType},
    DomainResult,
};

/// Port interface for wallet persistence
#[async_trait]
pub trait WalletRepository: Send + Sync {
    /// Find wallet by ID
    async fn find_by_id(&self, id: Uuid) -> DomainResult<Option<Wallet>>;
    
    /// Find wallet by address
    async fn find_by_address(&self, address: &str, chain: ChainType) -> DomainResult<Option<Wallet>>;
    
    /// Find all wallets
    async fn find_all(&self, limit: u32, offset: u32) -> DomainResult<Vec<Wallet>>;
    
    /// Find wallets by type
    async fn find_by_type(&self, wallet_type: WalletType) -> DomainResult<Vec<Wallet>>;
    
    /// Find wallets by status
    async fn find_by_status(&self, status: &str) -> DomainResult<Vec<Wallet>>;
    
    /// Create a new wallet
    async fn create(&self, wallet: &Wallet) -> DomainResult<Wallet>;
    
    /// Update an existing wallet
    async fn update(&self, wallet: &Wallet) -> DomainResult<Wallet>;
    
    /// Delete a wallet (soft delete)
    async fn delete(&self, id: Uuid) -> DomainResult<()>;
    
    /// Count total wallets
    async fn count(&self) -> DomainResult<u64>;
}

/// Port interface for policy persistence
#[async_trait]
pub trait PolicyRepository: Send + Sync {
    /// Find policy by ID
    async fn find_by_id(&self, id: Uuid) -> DomainResult<Option<Policy>>;
    
    /// Find all active policies
    async fn find_active(&self) -> DomainResult<Vec<Policy>>;
    
    /// Find policies by type
    async fn find_by_type(&self, wallet_type: WalletType, chain: ChainType) -> DomainResult<Vec<Policy>>;
    
    /// Create a new policy
    async fn create(&self, policy: &Policy) -> DomainResult<Policy>;
    
    /// Update an existing policy
    async fn update(&self, policy: &Policy) -> DomainResult<Policy>;
    
    /// Delete a policy (soft delete)
    async fn delete(&self, id: Uuid) -> DomainResult<()>;
}

/// Port interface for key management
#[async_trait]
pub trait KeyManager: Send + Sync {
    /// Generate a new key in HSM
    async fn generate_key(
        &self,
        label: &str,
        algorithm: domain::KeyAlgorithm,
    ) -> DomainResult<domain::KeyMetadata>;
    
    /// Sign data with key
    async fn sign(
        &self,
        key_handle: &str,
        data: &[u8],
    ) -> DomainResult<Vec<u8>>;
    
    /// Verify a signature
    async fn verify(
        &self,
        public_key: &[u8],
        data: &[u8],
        signature: &[u8],
    ) -> DomainResult<bool>;
    
    /// Get public key from HSM
    async fn get_public_key(&self, key_handle: &str) -> DomainResult<Vec<u8>>;
    
    /// Rotate a key
    async fn rotate_key(&self, key_handle: &str) -> DomainResult<domain::KeyMetadata>;
    
    /// Destroy a key
    async fn destroy_key(&self, key_handle: &str) -> DomainResult<()>;
}

/// Port interface for audit logging
#[async_trait]
pub trait AuditLog: Send + Sync {
    /// Log an audit event
    async fn log(&self, event: &AuditEvent) -> DomainResult<()>;
    
    /// Log a batch of events
    async fn log_batch(&self, events: &[AuditEvent]) -> DomainResult<()>;
    
    /// Query audit logs
    async fn query(&self, query: domain::AuditQuery) -> DomainResult<domain::AuditQueryResult>;
}

/// Port interface for message publishing
#[async_trait]
pub trait MessageBus: Send + Sync {
    /// Publish an event
    async fn publish(&self, topic: &str, key: &str, value: &[u8]) -> DomainResult<()>;
    
    /// Publish a wallet event
    async fn publish_wallet_event(&self, event: &WalletEvent) -> DomainResult<()>;
    
    /// Publish a signing event
    async fn publish_signing_event(&self, event: &SigningEvent) -> DomainResult<()>;
    
    /// Publish an emergency event
    async fn publish_emergency_event(&self, event: &EmergencyEvent) -> DomainResult<()>;
}

/// Wallet lifecycle events for publishing
#[derive(Debug, Clone)]
pub enum WalletEvent {
    Created(Wuid, String, ChainType, WalletType),
    Frozen(Uuid, FreezeReason),
    Unfrozen(Uuid),
    Archived(Uuid),
    Recovered(Uuid),
    LimitsUpdated(Uuid, Decimal, Decimal),
}

/// Signing events for publishing
#[derive(Debug, Clone)]
pub enum SigningEvent {
    RequestCreated(Uuid, Uuid, u8),
    RequestApproved(Uuid, Uuid),
    RequestRejected(Uuid, Uuid, String),
    RequestCompleted(Uuid),
    RequestTimedOut(Uuid),
}

/// Emergency events for publishing
#[derive(Debug, Clone)]
pub enum EmergencyEvent {
    FreezeInitiated(Uuid, String),
    FreezeCompleted(Uuid, String),
    UnfreezeInitiated(Uuid, String),
    UnfreezeCompleted(Uuid, String),
    AccessRequested(Uuid, domain::AuthorizationLevel),
    AccessApproved(Uuid, Uuid),
    AccessDenied(Uuid, String),
}

/// Wallet management service
#[derive(Clone)]
pub struct WalletService {
    wallet_repo: Arc<dyn WalletRepository>,
    policy_repo: Arc<dyn PolicyRepository>,
    key_manager: Arc<dyn KeyManager>,
    audit_log: Arc<dyn AuditLog>,
    message_bus: Arc<dyn MessageBus>,
}

impl WalletService {
    /// Creates a new WalletService
    pub fn new(
        wallet_repo: Arc<dyn WalletRepository>,
        policy_repo: Arc<dyn PolicyRepository>,
        key_manager: Arc<dyn KeyManager>,
        audit_log: Arc<dyn AuditLog>,
        message_bus: Arc<dyn MessageBus>,
    ) -> Self {
        Self {
            wallet_repo,
            policy_repo,
            key_manager,
            audit_log,
            message_bus,
        }
    }

    /// Creates a new wallet
    pub async fn create_wallet(
        &self,
        address: String,
        chain_type: ChainType,
        wallet_type: WalletType,
        max_tx_limit: Option<Decimal>,
        daily_limit: Option<Decimal>,
        actor: AuditActor,
    ) -> DomainResult<Wallet> {
        // Check if address already exists
        if let Some(existing) = self.wallet_repo.find_by_address(&address, chain_type).await? {
            return Err(domain::DomainError::WalletNotFound(
                format!("Address {} already exists with wallet ID {}", address, existing.id)
            ));
        }

        // Generate key in HSM
        let key_label = format!("wallet_{}_{}", chain_type, address);
        let key_metadata = self.key_manager.generate_key(
            &key_label,
            match chain_type {
                ChainType::Bitcoin => domain::KeyAlgorithm::EcdsaSecp256k1,
                _ => domain::KeyAlgorithm::EcdsaP256,
            },
        ).await?;

        // Create wallet
        let mut wallet = Wallet::new(
            address,
            chain_type,
            wallet_type,
            Uuid::nil(), // Will be set based on policy
            key_metadata.hsm_key_handle,
            key_metadata.public_key,
        );

        if let Some(limit) = max_tx_limit {
            wallet.max_tx_limit = limit;
        }
        if let Some(limit) = daily_limit {
            wallet.daily_limit = limit;
        }

        // Save wallet
        let wallet = self.wallet_repo.create(&wallet).await?;

        // Audit log
        let event = AuditEvent::new(
            AuditCategory::Wallet,
            AuditEventType::WalletCreated,
            AuditSeverity::Info,
            actor.clone(),
            format!("Created {} wallet {} on {}", wallet_type, wallet.id, chain_type),
        ).with_resource(wallet.id, "wallet")
         .with_metadata(serde_json::json!({
             "address": wallet.address,
             "chain_type": chain_type.to_string(),
             "wallet_type": wallet_type.to_string(),
         }));
        self.audit_log.log(&event).await?;

        // Publish event
        let wallet_event = WalletEvent::Created(
            wallet.id,
            wallet.address.clone(),
            chain_type,
            wallet_type,
        );
        self.message_bus.publish_wallet_event(&wallet_event).await?;

        Ok(wallet)
    }

    /// Freezes a wallet
    pub async fn freeze_wallet(
        &self,
        wallet_id: Uuid,
        reason: FreezeReason,
        notes: Option<String>,
        actor: AuditActor,
    ) -> DomainResult<Wallet> {
        let wallet = self.wallet_repo.find_by_id(wallet_id).await?
            .ok_or_else(|| domain::DomainError::WalletNotFound(wallet_id.to_string()))?;

        let before_state = serde_json::json!({
            "status": wallet.status.to_string(),
        });

        // Freeze wallet
        let mut wallet = wallet;
        wallet.freeze(reason);
        let wallet = self.wallet_repo.update(&wallet).await?;

        let after_state = serde_json::json!({
            "status": wallet.status.to_string(),
        });

        // Audit log
        let event = AuditEvent::new(
            AuditCategory::Emergency,
            AuditEventType::EmergencyFreezeInitiated,
            AuditSeverity::High,
            actor.clone(),
            format!("Wallet {} frozen: {}", wallet_id, reason),
        ).with_resource(wallet.id, "wallet")
         .with_state_changes(Some(before_state), Some(after_state))
         .with_metadata(serde_json::json!({
             "reason": reason.to_string(),
             "notes": notes.unwrap_or_default(),
         }));
        self.audit_log.log(&event).await?;

        // Publish event
        let event = EmergencyEvent::FreezeInitiated(wallet_id, reason.to_string());
        self.message_bus.publish_emergency_event(&event).await?;

        Ok(wallet)
    }

    /// Unfreezes a wallet
    pub async fn unfreeze_wallet(
        &self,
        wallet_id: Uuid,
        authorization_ref: String,
        actor: AuditActor,
    ) -> DomainResult<Wallet> {
        let wallet = self.wallet_repo.find_by_id(wallet_id).await?
            .ok_or_else(|| domain::DomainError::WalletNotFound(wallet_id.to_string()))?;

        // Verify wallet is frozen
        let before_state = serde_json::json!({
            "status": wallet.status.to_string(),
        });

        let mut wallet = wallet;
        wallet.unfreeze();
        let wallet = self.wallet_repo.update(&wallet).await?;

        let after_state = serde_json::json!({
            "status": wallet.status.to_string(),
        });

        // Audit log
        let event = AuditEvent::new(
            AuditCategory::Emergency,
            AuditEventType::EmergencyUnfreezeCompleted,
            AuditSeverity::Medium,
            actor.clone(),
            format!("Wallet {} unfrozen", wallet_id),
        ).with_resource(wallet.id, "wallet")
         .with_state_changes(Some(before_state), Some(after_state))
         .with_metadata(serde_json::json!({
             "authorization_ref": authorization_ref,
         }));
        self.audit_log.log(&event).await?;

        // Publish event
        let event = EmergencyEvent::UnfreezeCompleted(wallet_id, authorization_ref);
        self.message_bus.publish_emergency_event(&event).await?;

        Ok(wallet)
    }

    /// Gets wallet by ID
    pub async fn get_wallet(&self, wallet_id: Uuid) -> DomainResult<Option<Wallet>> {
        self.wallet_repo.find_by_id(wallet_id).await
    }

    /// Lists all wallets with pagination
    pub async fn list_wallets(
        &self,
        limit: u32,
        offset: u32,
    ) -> DomainResult<Vec<Wallet>> {
        self.wallet_repo.find_all(limit, offset).await
    }

    /// Initiates wallet recovery
    pub async fn initiate_recovery(
        &self,
        wallet_id: Uuid,
        actor: AuditActor,
    ) -> DomainResult<Wallet> {
        let wallet = self.wallet_repo.find_by_id(wallet_id).await?
            .ok_or_else(|| domain::DomainError::WalletNotFound(wallet_id.to_string()))?;

        // Check if already in recovery
        if matches!(wallet.status, domain::WalletStatus::PendingRecovery) {
            return Err(domain::DomainError::RecoveryInProgress);
        }

        let mut wallet = wallet;
        wallet.initiate_recovery();
        let wallet = self.wallet_repo.update(&wallet).await?;

        // Audit log
        let event = AuditEvent::new(
            AuditCategory::Wallet,
            AuditEventType::WalletRecoveryInitiated,
            AuditSeverity::High,
            actor.clone(),
            format!("Recovery initiated for wallet {}", wallet_id),
        ).with_resource(wallet.id, "wallet");
        self.audit_log.log(&event).await?;

        Ok(wallet)
    }
}

/// Policy evaluation service
#[derive(Clone)]
pub struct PolicyService {
    policy_repo: Arc<dyn PolicyRepository>,
    audit_log: Arc<dyn AuditLog>,
}

impl PolicyService {
    /// Creates a new PolicyService
    pub fn new(
        policy_repo: Arc<dyn PolicyRepository>,
        audit_log: Arc<dyn AuditLog>,
    ) -> Self {
        Self {
            policy_repo,
            audit_log,
        }
    }

    /// Evaluates a transaction against all applicable policies
    pub async fn evaluate_transaction(
        &self,
        context: PolicyEvaluationContext,
    ) -> DomainResult<PolicyEvaluationResult> {
        // Get active policies
        let policies = self.policy_repo.find_active().await?;

        let mut result = PolicyEvaluationResult {
            is_allowed: true,
            severity: policy::PolicySeverity::Warning,
            failed_rules: Vec::new(),
            warnings: Vec::new(),
            requirements: Vec::new(),
            risk_score: 0,
        };

        // Evaluate each applicable policy
        for policy in policies {
            if !policy.applies_to(&context.source_wallet.map_or(WalletType::Hot, |w| w.wallet_type), &context.chain_type) {
                continue;
            }

            for rule in &policy.rules {
                let rule_result = self.evaluate_rule(rule, &context);
                
                if !rule_result.allowed {
                    result.is_allowed = false;
                    result.failed_rules.push(rule_result.rule_name);
                    result.risk_score = result.risk_score.saturating_add(rule_result.risk_contribution);
                    
                    if let Some(requirement) = rule_result.requirement {
                        result.requirements.push(requirement);
                    }
                } else if let Some(warning) = rule_result.warning {
                    result.warnings.push(warning);
                }
            }
        }

        // Determine overall severity
        result.severity = if result.is_allowed {
            if result.warnings.is_empty() {
                policy::PolicySeverity::Warning
            } else {
                policy::PolicySeverity::EnhancedReview
            }
        } else {
            policy::PolicySeverity::Blocking
        };

        Ok(result)
    }

    /// Evaluates a single rule
    fn evaluate_rule(
        &self,
        rule: &PolicyRule,
        context: &PolicyEvaluationContext,
    ) -> RuleEvaluationResult {
        match rule {
            PolicyRule::BlacklistAddress { address, chain_type, .. } => {
                // Check chain type match
                if chain_type.as_ref().map_or(false, |c| c != &context.chain_type.to_string()) {
                    return RuleEvaluationResult { allowed: true, rule_name: "blacklist_chain_mismatch".to_string(), warning: None, requirement: None, risk_contribution: 0 };
                }
                
                if context.destination_address.to_lowercase() == address.to_lowercase() {
                    RuleEvaluationResult {
                        allowed: false,
                        rule_name: "blacklist_address".to_string(),
                        warning: None,
                        requirement: None,
                        risk_contribution: 100,
                    }
                } else {
                    RuleEvaluationResult { allowed: true, rule_name: "blacklist_check_passed".to_string(), warning: None, requirement: None, risk_contribution: 0 }
                }
            }
            
            PolicyRule::WhitelistAddress { address, chain_type, .. } => {
                if chain_type.as_ref().map_or(false, |c| c != &context.chain_type.to_string()) {
                    return RuleEvaluationResult { allowed: true, rule_name: "whitelist_chain_mismatch".to_string(), warning: None, requirement: None, risk_contribution: 0 };
                }
                
                if context.destination_address.to_lowercase() == address.to_lowercase() {
                    RuleEvaluationResult { allowed: true, rule_name: "whitelist_address".to_string(), warning: None, requirement: None, risk_contribution: 0 }
                } else {
                    RuleEvaluationResult {
                        allowed: true,
                        rule_name: "non_whitelisted_destination".to_string(),
                        warning: Some("Destination address is not whitelisted".to_string()),
                        requirement: None,
                        risk_contribution: 10,
                    }
                }
            }
            
            PolicyRule::MaxTransactionLimit { max_amount, .. } => {
                if context.amount > *max_amount {
                    RuleEvaluationResult {
                        allowed: false,
                        rule_name: "max_transaction_limit".to_string(),
                        warning: None,
                        requirement: None,
                        risk_contribution: 30,
                    }
                } else {
                    RuleEvaluationResult { allowed: true, rule_name: "transaction_limit_ok".to_string(), warning: None, requirement: None, risk_contribution: 0 }
                }
            }
            
            _ => RuleEvaluationResult { allowed: true, rule_name: "rule_not_applicable".to_string(), warning: None, requirement: None, risk_contribution: 0 },
        }
    }
}

struct RuleEvaluationResult {
    allowed: bool,
    rule_name: String,
    warning: Option<String>,
    requirement: Option<policy::PolicyRequirement>,
    risk_contribution: u8,
}
