use axum::{
    routing::{get, post, put, delete},
    Router,
    extract::{Path, Json, Query, State},
    response::IntoResponse,
    http::StatusCode,
};
use chrono::{DateTime, Utc};
use serde::{Deserialize, Serialize};
use uuid::Uuid;
use validator::Validate;

use crate::domain::{
    wallet::{WalletType, ChainType, FreezeReason, RecoveryMethod},
    policy::{PolicySeverity, BlacklistSource},
    audit::{AuditActor, ActorType},
};
use crate::application::{WalletService, PolicyService};
use crate::domain::DomainResult;

/// Application state
#[derive(Clone)]
pub struct AppState {
    pub wallet_service: WalletService,
    pub policy_service: PolicyService,
}

/// API response wrapper
#[derive(Debug, Serialize)]
pub struct ApiResponse<T> {
    pub success: bool,
    pub data: Option<T>,
    pub error: Option<String>,
    pub timestamp: DateTime<Utc>,
}

impl<T> ApiResponse<T> {
    pub fn success(data: T) -> Self {
        Self {
            success: true,
            data: Some(data),
            error: None,
            timestamp: Utc::now(),
        }
    }

    pub fn error(message: &str) -> Self {
        Self {
            success: false,
            data: None,
            error: Some(message.to_string()),
            timestamp: Utc::now(),
        }
    }
}

/// Pagination parameters
#[derive(Debug, Deserialize)]
pub struct PaginationParams {
    pub limit: Option<u32>,
    pub offset: Option<u32>,
}

/// Wallet creation request
#[derive(Debug, Deserialize, Validate)]
pub struct CreateWalletRequest {
    #[validate(length(min = 26, max = 70))]
    pub address: String,
    
    pub chain_type: String,
    
    pub wallet_type: String,
    
    #[validate(decimal_places(equals = 8))]
    pub max_tx_limit: Option<String>,
    
    #[validate(decimal_places(equals = 8))]
    pub daily_limit: Option<String>,
}

/// Wallet response
#[derive(Debug, Serialize)]
pub struct WalletResponse {
    pub id: String,
    pub address: String,
    pub chain_type: String,
    pub wallet_type: String,
    pub status: String,
    pub max_tx_limit: String,
    pub daily_limit: String,
    pub daily_usage: String,
    pub created_at: DateTime<Utc>,
}

/// Wallet freeze request
#[derive(Debug, Deserialize, Validate)]
pub struct FreezeWalletRequest {
    pub reason: String,
    #[validate(length(max = 500))]
    pub notes: Option<String>,
}

/// Wallet unfreeze request
#[derive(Debug, Deserialize, Validate)]
pub struct UnfreezeWalletRequest {
    #[validate(length(min = 10, max = 100))]
    pub authorization_ref: String,
}

/// Policy creation request
#[derive(Debug, Deserialize, Validate)]
pub struct CreatePolicyRequest {
    #[validate(length(min = 3, max = 100))]
    pub name: String,
    
    #[validate(length(max = 500))]
    pub description: Option<String>,
    
    pub severity: String,
    
    pub rules: Vec<serde_json::Value>,
    
    pub applies_to_wallet_types: Option<Vec<String>>,
    
    pub applies_to_chains: Option<Vec<String>>,
}

/// Blacklist entry request
#[derive(Debug, Deserialize, Validate)]
pub struct AddBlacklistRequest {
    #[validate(length(min = 26, max = 70))]
    pub address: String,
    
    pub chain_type: Option<String>,
    
    #[validate(length(min = 10, max = 500))]
    pub reason: String,
    
    pub source: String,
    
    pub risk_level: Option<u8>,
}

/// Audit log query request
#[derive(Debug, Deserialize)]
pub struct AuditQueryRequest {
    pub category: Option<String>,
    pub event_type: Option<String>,
    pub severity: Option<String>,
    pub actor_id: Option<String>,
    pub resource_id: Option<String>,
    pub start_time: Option<DateTime<Utc>>,
    pub end_time: Option<DateTime<Utc>>,
    pub success: Option<bool>,
    pub search_query: Option<String>,
    pub page: Option<u32>,
    pub page_size: Option<u32>,
}

/// Creates the application router
pub fn create_router(app_state: AppState) -> Router {
    Router::new()
        .route("/health", get(health_check))
        .route("/api/v1/wallets", get(list_wallets).post(create_wallet))
        .route("/api/v1/wallets/:id", get(get_wallet))
        .route("/api/v1/wallets/:id/freeze", post(freeze_wallet))
        .route("/api/v1/wallets/:id/unfreeze", post(unfreeze_wallet))
        .route("/api/v1/wallets/:id/recover", post(initiate_recovery))
        .route("/api/v1/policies", get(list_policies).post(create_policy))
        .route("/api/v1/policies/:id", get(get_policy).delete(delete_policy))
        .route("/api/v1/blacklist", get(list_blacklist).post(add_blacklist))
        .route("/api/v1/blacklist/:address", delete(remove_blacklist))
        .route("/api/v1/whitelist", get(list_whitelist).post(add_whitelist))
        .route("/api/v1/audit/logs", get(query_audit_logs))
        .route("/api/v1/emergency/freeze", post(emergency_freeze))
        .route("/api/v1/emergency/unfreeze", post(emergency_unfreeze))
        .with_state(app_state)
}

/// Health check endpoint
async fn health_check() -> impl IntoResponse {
    Json(ApiResponse::success("Service is healthy"))
}

/// List all wallets
async fn list_wallets(
    State(state): State<AppState>,
    Query(params): Query<PaginationParams>,
) -> impl IntoResponse {
    let limit = params.limit.unwrap_or(50);
    let offset = params.offset.unwrap_or(0);

    match state.wallet_service.list_wallets(limit, offset).await {
        Ok(wallets) => {
            let response: Vec<WalletResponse> = wallets
                .iter()
                .map(|w| WalletResponse {
                    id: w.id.to_string(),
                    address: w.address.clone(),
                    chain_type: w.chain_type.to_string(),
                    wallet_type: w.wallet_type.to_string(),
                    status: w.status.to_string(),
                    max_tx_limit: w.max_tx_limit.to_string(),
                    daily_limit: w.daily_limit.to_string(),
                    daily_usage: w.daily_usage.to_string(),
                    created_at: w.created_at,
                })
                .collect();
            Json(ApiResponse::success(response))
        }
        Err(e) => Json(ApiResponse::error(&e.to_string())),
    }
}

/// Get a specific wallet
async fn get_wallet(
    State(state): State<AppState>,
    Path(id): Path<String>,
) -> impl IntoResponse {
    let wallet_id = match Uuid::parse_str(&id) {
        Ok(uuid) => uuid,
        Err(_) => return Json(ApiResponse::error("Invalid wallet ID format")),
    };

    match state.wallet_service.get_wallet(wallet_id).await {
        Ok(Some(wallet)) => {
            let response = WalletResponse {
                id: wallet.id.to_string(),
                address: wallet.address.clone(),
                chain_type: wallet.chain_type.to_string(),
                wallet_type: wallet.wallet_type.to_string(),
                status: wallet.status.to_string(),
                max_tx_limit: wallet.max_tx_limit.to_string(),
                daily_limit: wallet.daily_limit.to_string(),
                daily_usage: wallet.daily_usage.to_string(),
                created_at: wallet.created_at,
            };
            Json(ApiResponse::success(response))
        }
        Ok(None) => Json(ApiResponse::error("Wallet not found")),
        Err(e) => Json(ApiResponse::error(&e.to_string())),
    }
}

/// Create a new wallet
async fn create_wallet(
    State(state): State<AppState>,
    Json(req): Json<CreateWalletRequest>,
) -> impl IntoResponse {
    if let Err(errors) = req.validate() {
        return Json(ApiResponse::error(&format!("Validation error: {}", errors)));
    }

    let chain_type = match ChainType::from_str(&req.chain_type) {
        Ok(ct) => ct,
        Err(_) => return Json(ApiResponse::error("Invalid chain type")),
    };

    let wallet_type = match WalletType::from_str(&req.wallet_type) {
        Ok(wt) => wt,
        Err(_) => return Json(ApiResponse::error("Invalid wallet type")),
    };

    let actor = AuditActor {
        id: "api_user".to_string(),
        actor_type: ActorType::Admin,
        label: "API User".to_string(),
        ip_address: None,
        user_agent: None,
        session_id: None,
    };

    match state.wallet_service.create_wallet(
        req.address,
        chain_type,
        wallet_type,
        None,
        None,
        actor,
    ).await {
        Ok(wallet) => {
            let response = WalletResponse {
                id: wallet.id.to_string(),
                address: wallet.address.clone(),
                chain_type: wallet.chain_type.to_string(),
                wallet_type: wallet.wallet_type.to_string(),
                status: wallet.status.to_string(),
                max_tx_limit: wallet.max_tx_limit.to_string(),
                daily_limit: wallet.daily_limit.to_string(),
                daily_usage: wallet.daily_usage.to_string(),
                created_at: wallet.created_at,
            };
            Json(ApiResponse::success(response))
        }
        Err(e) => Json(ApiResponse::error(&e.to_string())),
    }
}

/// Freeze a wallet
async fn freeze_wallet(
    State(state): State<AppState>,
    Path(id): Path<String>,
    Json(req): Json<FreezeWalletRequest>,
) -> impl IntoResponse {
    let wallet_id = match Uuid::parse_str(&id) {
        Ok(uuid) => uuid,
        Err(_) => return Json(ApiResponse::error("Invalid wallet ID format")),
    };

    let reason = match FreezeReason::from_str(&req.reason) {
        Ok(r) => r,
        Err(_) => return Json(ApiResponse::error("Invalid freeze reason")),
    };

    let actor = AuditActor {
        id: "api_user".to_string(),
        actor_type: ActorType::Admin,
        label: "API User".to_string(),
        ip_address: None,
        user_agent: None,
        session_id: None,
    };

    match state.wallet_service.freeze_wallet(wallet_id, reason, req.notes, actor).await {
        Ok(wallet) => {
            Json(ApiResponse::success(format!("Wallet {} frozen successfully", wallet.id)))
        }
        Err(e) => Json(ApiResponse::error(&e.to_string())),
    }
}

/// Unfreeze a wallet
async fn unfreeze_wallet(
    State(state): State<AppState>,
    Path(id): Path<String>,
    Json(req): Json<UnfreezeWalletRequest>,
) -> impl IntoResponse {
    let wallet_id = match Uuid::parse_str(&id) {
        Ok(uuid) => uuid,
        Err(_) => return Json(ApiResponse::error("Invalid wallet ID format")),
    };

    let actor = AuditActor {
        id: "api_user".to_string(),
        actor_type: ActorType::Admin,
        label: "API User".to_string(),
        ip_address: None,
        user_agent: None,
        session_id: None,
    };

    match state.wallet_service.unfreeze_wallet(wallet_id, req.authorization_ref, actor).await {
        Ok(wallet) => {
            Json(ApiResponse::success(format!("Wallet {} unfrozen successfully", wallet.id)))
        }
        Err(e) => Json(ApiResponse::error(&e.to_string())),
    }
}

/// Initiate wallet recovery
async fn initiate_recovery(
    State(state): State<AppState>,
    Path(id): Path<String>,
) -> impl IntoResponse {
    let wallet_id = match Uuid::parse_str(&id) {
        Ok(uuid) => uuid,
        Err(_) => return Json(ApiResponse::error("Invalid wallet ID format")),
    };

    let actor = AuditActor {
        id: "api_user".to_string(),
        actor_type: ActorType::Admin,
        label: "API User".to_string(),
        ip_address: None,
        user_agent: None,
        session_id: None,
    };

    match state.wallet_service.initiate_recovery(wallet_id, actor).await {
        Ok(wallet) => {
            Json(ApiResponse::success(format!(
                "Recovery initiated for wallet {}. New status: {}",
                wallet.id, wallet.status
            )))
        }
        Err(e) => Json(ApiResponse::error(&e.to_string())),
    }
}

/// List policies (placeholder)
async fn list_policies() -> impl IntoResponse {
    Json(ApiResponse::success(Vec::<serde::json::Value>::new()))
}

/// Get a policy (placeholder)
async fn get_policy(Path(id): Path<String>) -> impl IntoResponse {
    Json(ApiResponse::error("Not implemented"))
}

/// Create a policy (placeholder)
async fn create_policy(Json(_req): Json<CreatePolicyRequest>) -> impl IntoResponse {
    Json(ApiResponse::error("Not implemented"))
}

/// Delete a policy (placeholder)
async fn delete_policy(Path(_id): Path<String>) -> impl IntoResponse {
    Json(ApiResponse::error("Not implemented"))
}

/// List blacklist entries (placeholder)
async fn list_blacklist() -> impl IntoResponse {
    Json(ApiResponse::success(Vec::<serde::json::Value>::new()))
}

/// Add blacklist entry (placeholder)
async fn add_blacklist(Json(_req): Json<AddBlacklistRequest>) -> impl IntoResponse {
    Json(ApiResponse::error("Not implemented"))
}

/// Remove blacklist entry (placeholder)
async fn remove_blacklist(Path(_address): Path<String>) -> impl IntoResponse {
    Json(ApiResponse::error("Not implemented"))
}

/// List whitelist entries (placeholder)
async fn list_whitelist() -> impl IntoResponse {
    Json(ApiResponse::success(Vec::<serde::json::Value>::new()))
}

/// Add whitelist entry (placeholder)
async fn add_whitelist(Json(_req): Json<serde::json::Value>) -> impl IntoResponse {
    Json(ApiResponse::error("Not implemented"))
}

/// Query audit logs (placeholder)
async fn query_audit_logs(Json(_req): Json<AuditQueryRequest>) -> impl IntoResponse {
    Json(ApiResponse::error("Not implemented"))
}

/// Emergency freeze all wallets (placeholder)
async fn emergency_freeze(Json(_req): Json<serde::json::Value>) -> impl IntoResponse {
    Json(ApiResponse::error("Not implemented"))
}

/// Emergency unfreeze all wallets (placeholder)
async fn emergency_unfreeze(Json(_req): Json<serde::json::Value>) -> impl IntoResponse {
    Json(ApiResponse::error("Not implemented"))
}
