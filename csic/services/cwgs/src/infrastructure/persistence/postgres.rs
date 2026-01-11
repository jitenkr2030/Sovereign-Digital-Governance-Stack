use std::str::FromStr;

use async_trait::async_trait;
use chrono::{DateTime, Utc};
use rust_decimal::Decimal;
use sqlx::{postgres::PgPool, Row};
use uuid::Uuid;

use crate::domain::{self, wallet::{Wallet, WalletType, WalletStatus, FreezeReason, ChainType}};
use crate::application::{WalletRepository, PolicyRepository, AuditLog, MessageBus};
use crate::domain::audit::{AuditEvent, AuditCategory, AuditEventType, AuditSeverity, AuditActor};
use crate::domain::policy::{Policy, PolicyRule, PolicySeverity};

/// PostgreSQL wallet repository implementation
pub struct PostgresWalletRepository {
    pool: PgPool,
}

impl PostgresWalletRepository {
    /// Creates a new PostgreSQL wallet repository
    pub fn new(pool: PgPool) -> Self {
        Self { pool }
    }
}

#[async_trait]
impl WalletRepository for PostgresWalletRepository {
    async fn find_by_id(&self, id: Uuid) -> domain::DomainResult<Option<Wallet>> {
        let row = sqlx::query!(
            r#"
            SELECT id, address, chain_type, wallet_type, status, signing_policy_id,
                   hsm_key_handle, public_key, max_tx_limit, daily_limit, daily_usage,
                   last_activity_at, created_at, updated_at
            FROM wallets
            WHERE id = $1 AND status != 'ARCHIVED'
            "#,
            id
        )
        .fetch_optional(&self.pool)
        .await?;

        match row {
            Some(r) => Ok(Some(self.row_to_wallet(&r)?)),
            None => Ok(None),
        }
    }

    async fn find_by_address(&self, address: &str, chain: ChainType) -> domain::DomainResult<Option<Wallet>> {
        let row = sqlx::query!(
            r#"
            SELECT id, address, chain_type, wallet_type, status, signing_policy_id,
                   hsm_key_handle, public_key, max_tx_limit, daily_limit, daily_usage,
                   last_activity_at, created_at, updated_at
            FROM wallets
            WHERE address = $1 AND chain_type = $2 AND status != 'ARCHIVED'
            "#,
            address,
            chain.to_string()
        )
        .fetch_optional(&self.pool)
        .await?;

        match row {
            Some(r) => Ok(Some(self.row_to_wallet(&r)?)),
            None => Ok(None),
        }
    }

    async fn find_all(&self, limit: u32, offset: u32) -> domain::DomainResult<Vec<Wallet>> {
        let rows = sqlx::query!(
            r#"
            SELECT id, address, chain_type, wallet_type, status, signing_policy_id,
                   hsm_key_handle, public_key, max_tx_limit, daily_limit, daily_usage,
                   last_activity_at, created_at, updated_at
            FROM wallets
            WHERE status != 'ARCHIVED'
            ORDER BY created_at DESC
            LIMIT $1 OFFSET $2
            "#,
            limit as i64,
            offset as i64
        )
        .fetch_all(&self.pool)
        .await?;

        rows.iter().map(|r| self.row_to_wallet(r)).collect()
    }

    async fn find_by_type(&self, wallet_type: WalletType) -> domain::DomainResult<Vec<Wallet>> {
        let rows = sqlx::query!(
            r#"
            SELECT id, address, chain_type, wallet_type, status, signing_policy_id,
                   hsm_key_handle, public_key, max_tx_limit, daily_limit, daily_usage,
                   last_activity_at, created_at, updated_at
            FROM wallets
            WHERE wallet_type = $1 AND status != 'ARCHIVED'
            ORDER BY created_at DESC
            "#,
            wallet_type.to_string()
        )
        .fetch_all(&self.pool)
        .await?;

        rows.iter().map(|r| self.row_to_wallet(r)).collect()
    }

    async fn find_by_status(&self, status: &str) -> domain::DomainResult<Vec<Wallet>> {
        let rows = sqlx::query!(
            r#"
            SELECT id, address, chain_type, wallet_type, status, signing_policy_id,
                   hsm_key_handle, public_key, max_tx_limit, daily_limit, daily_usage,
                   last_activity_at, created_at, updated_at
            FROM wallets
            WHERE status = $1 AND status != 'ARCHIVED'
            ORDER BY created_at DESC
            "#,
            status
        )
        .fetch_all(&self.pool)
        .await?;

        rows.iter().map(|r| self.row_to_wallet(r)).collect()
    }

    async fn create(&self, wallet: &Wallet) -> domain::DomainResult<Wallet> {
        sqlx::query!(
            r#"
            INSERT INTO wallets (
                id, address, chain_type, wallet_type, status, signing_policy_id,
                hsm_key_handle, public_key, max_tx_limit, daily_limit, daily_usage,
                last_activity_at, created_at, updated_at
            ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
            "#,
            wallet.id,
            wallet.address,
            wallet.chain_type.to_string(),
            wallet.wallet_type.to_string(),
            wallet.status.to_string(),
            wallet.signing_policy_id,
            wallet.hsm_key_handle,
            wallet.public_key,
            wallet.max_tx_limit,
            wallet.daily_limit,
            wallet.daily_usage,
            wallet.last_activity_at,
            wallet.created_at,
            wallet.updated_at
        )
        .execute(&self.pool)
        .await?;

        Ok(wallet.clone())
    }

    async fn update(&self, wallet: &Wallet) -> domain::DomainResult<Wallet> {
        sqlx::query!(
            r#"
            UPDATE wallets SET
                status = $1,
                max_tx_limit = $2,
                daily_limit = $3,
                daily_usage = $4,
                last_activity_at = $5,
                updated_at = $6
            WHERE id = $7
            "#,
            wallet.status.to_string(),
            wallet.max_tx_limit,
            wallet.daily_limit,
            wallet.daily_usage,
            wallet.last_activity_at,
            wallet.updated_at,
            wallet.id
        )
        .execute(&self.pool)
        .await?;

        Ok(wallet.clone())
    }

    async fn delete(&self, id: Uuid) -> domain::DomainResult<()> {
        sqlx::query!(
            "UPDATE wallets SET status = 'ARCHIVED', updated_at = NOW() WHERE id = $1",
            id
        )
        .execute(&self.pool)
        .await?;

        Ok(())
    }

    async fn count(&self) -> domain::DomainResult<u64> {
        let count: i64 = sqlx::query_scalar!(
            "SELECT COUNT(*) FROM wallets WHERE status != 'ARCHIVED'"
        )
        .fetch_one(&self.pool)
        .await?;

        Ok(count as u64)
    }
}

impl PostgresWalletRepository {
    fn row_to_wallet(&self, row: &sqlx::postgres::PgRow) -> domain::DomainResult<Wallet> {
        let status_str: String = row.try_get("status")?;
        let status = WalletStatus::from_str(&status_str)
            .map_err(|_| anyhow::anyhow!("Invalid status: {}", status_str))?;

        let wallet_type: String = row.try_get("wallet_type")?;
        let wallet_type = WalletType::from_str(&wallet_type)
            .map_err(|_| anyhow::anyhow!("Invalid wallet type: {}", wallet_type))?;

        let chain_type: String = row.try_get("chain_type")?;
        let chain_type = ChainType::from_str(&chain_type)
            .map_err(|_| anyhow::anyhow!("Invalid chain type: {}", chain_type))?;

        Ok(Wallet {
            id: row.try_get("id")?,
            address: row.try_get("address")?,
            chain_type,
            wallet_type,
            status,
            signing_policy_id: row.try_get("signing_policy_id")?,
            hsm_key_handle: row.try_get("hsm_key_handle")?,
            public_key: row.try_get("public_key")?,
            max_tx_limit: row.try_get("max_tx_limit")?,
            daily_limit: row.try_get("daily_limit")?,
            daily_usage: row.try_get("daily_usage")?,
            last_activity_at: row.try_get("last_activity_at")?,
            created_at: row.try_get("created_at")?,
            updated_at: row.try_get("updated_at")?,
        })
    }
}

/// PostgreSQL policy repository implementation
pub struct PostgresPolicyRepository {
    pool: PgPool,
}

impl PostgresPolicyRepository {
    /// Creates a new PostgreSQL policy repository
    pub fn new(pool: PgPool) -> Self {
        Self { pool }
    }
}

#[async_trait]
impl PolicyRepository for PostgresPolicyRepository {
    async fn find_by_id(&self, id: Uuid) -> domain::DomainResult<Option<Policy>> {
        let row = sqlx::query!(
            r#"
            SELECT id, name, description, version, is_active, priority, rules_json,
                   applies_to_wallet_types, applies_to_chains, effective_from,
                   effective_until, created_by, created_at, updated_at
            FROM policies
            WHERE id = $1
            "#,
            id
        )
        .fetch_optional(&self.pool)
        .await?;

        match row {
            Some(r) => Ok(Some(self.row_to_policy(&r)?)),
            None => Ok(None),
        }
    }

    async fn find_active(&self) -> domain::DomainResult<Vec<Policy>> {
        let rows = sqlx::query!(
            r#"
            SELECT id, name, description, version, is_active, priority, rules_json,
                   applies_to_wallet_types, applies_to_chains, effective_from,
                   effective_until, created_by, created_at, updated_at
            FROM policies
            WHERE is_active = true
              AND effective_from <= NOW()
              AND (effective_until IS NULL OR effective_until > NOW())
            ORDER BY priority DESC, created_at DESC
            "#,
        )
        .fetch_all(&self.pool)
        .await?;

        rows.iter().map(|r| self.row_to_policy(r)).collect()
    }

    async fn find_by_type(&self, wallet_type: WalletType, chain: ChainType) -> domain::DomainResult<Vec<Policy>> {
        let rows = sqlx::query!(
            r#"
            SELECT id, name, description, version, is_active, priority, rules_json,
                   applies_to_wallet_types, applies_to_chains, effective_from,
                   effective_until, created_by, created_at, updated_at
            FROM policies
            WHERE is_active = true
              AND effective_from <= NOW()
              AND (effective_until IS NULL OR effective_until > NOW())
              AND (applies_to_wallet_types @> $1::jsonb OR applies_to_wallet_types = '[]'::jsonb)
              AND (applies_to_chains @> $2::jsonb OR applies_to_chains = '[]'::jsonb)
            ORDER BY priority DESC
            "#,
            serde_json::json!([wallet_type.to_string()]),
            serde_json::json!([chain.to_string()])
        )
        .fetch_all(&self.pool)
        .await?;

        rows.iter().map(|r| self.row_to_policy(r)).collect()
    }

    async fn create(&self, policy: &Policy) -> domain::DomainResult<Policy> {
        sqlx::query!(
            r#"
            INSERT INTO policies (
                id, name, description, version, is_active, priority, rules_json,
                applies_to_wallet_types, applies_to_chains, effective_from,
                effective_until, created_by, created_at, updated_at
            ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
            "#,
            policy.id,
            policy.name,
            policy.description,
            policy.version as i32,
            policy.is_active,
            policy.priority,
            serde_json::to_string(&policy.rules)?,
            serde_json::to_string(&policy.applies_to_wallet_types.iter().map(|t| t.to_string()).collect::<Vec<_>>())?,
            serde_json::to_string(&policy.applies_to_chains.iter().map(|c| c.to_string()).collect::<Vec<_>>())?,
            policy.effective_from,
            policy.effective_until,
            policy.created_by,
            policy.created_at,
            policy.updated_at
        )
        .execute(&self.pool)
        .await?;

        Ok(policy.clone())
    }

    async fn update(&self, policy: &Policy) -> domain::DomainResult<Policy> {
        sqlx::query!(
            r#"
            UPDATE policies SET
                name = $1,
                description = $2,
                version = $3,
                is_active = $4,
                priority = $5,
                rules_json = $6,
                applies_to_wallet_types = $7,
                applies_to_chains = $8,
                effective_from = $9,
                effective_until = $10,
                updated_at = $11
            WHERE id = $12
            "#,
            policy.name,
            policy.description,
            policy.version as i32,
            policy.is_active,
            policy.priority,
            serde_json::to_string(&policy.rules)?,
            serde_json::to_string(&policy.applies_to_wallet_types.iter().map(|t| t.to_string()).collect::<Vec<_>>())?,
            serde_json::to_string(&policy.applies_to_chains.iter().map(|c| c.to_string()).collect::<Vec<_>>())?,
            policy.effective_from,
            policy.effective_until,
            policy.updated_at,
            policy.id
        )
        .execute(&self.pool)
        .await?;

        Ok(policy.clone())
    }

    async fn delete(&self, id: Uuid) -> domain::DomainResult<()> {
        sqlx::query!(
            "UPDATE policies SET is_active = false, updated_at = NOW() WHERE id = $1",
            id
        )
        .execute(&self.pool)
        .await?;

        Ok(())
    }
}

impl PostgresPolicyRepository {
    fn row_to_policy(&self, row: &sqlx::postgres::PgRow) -> domain::DomainResult<Policy> {
        let rules_json: String = row.try_get("rules_json")?;
        let rules: Vec<PolicyRule> = serde_json::from_str(&rules_json)
            .map_err(|e| anyhow::anyhow!("Failed to parse rules: {}", e))?;

        let wallet_types_json: String = row.try_get("applies_to_wallet_types")?;
        let wallet_types: Vec<String> = serde_json::from_str(&wallet_types_json)
            .unwrap_or_default();
        let applies_to_wallet_types: Vec<WalletType> = wallet_types
            .iter()
            .filter_map(|t| WalletType::from_str(t).ok())
            .collect();

        let chains_json: String = row.try_get("applies_to_chains")?;
        let chains: Vec<String> = serde_json::from_str(&chains_json)
            .unwrap_or_default();
        let applies_to_chains: Vec<ChainType> = chains
            .iter()
            .filter_map(|c| ChainType::from_str(c).ok())
            .collect();

        Ok(Policy {
            id: row.try_get("id")?,
            name: row.try_get("name")?,
            description: row.try_get("description")?,
            version: row.try_get("version")? as u32,
            is_active: row.try_get("is_active")?,
            priority: row.try_get("priority")?,
            rules,
            applies_to_wallet_types,
            applies_to_chains,
            effective_from: row.try_get("effective_from")?,
            effective_until: row.try_get("effective_until")?,
            created_by: row.try_get("created_by")?,
            created_at: row.try_get("created_at")?,
            updated_at: row.try_get("updated_at")?,
        })
    }
}
