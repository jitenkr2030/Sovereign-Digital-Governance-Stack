use std::net::SocketAddr;
use std::sync::Arc;

use axum::{Router, Server};
use sqlx::postgres::PgPool;
use tracing_subscriber::{layer::SubscriberExt, util::SubscriberInitExt};
use clap::Parser;

use crate::api::http::{create_router, AppState};
use crate::application::{WalletService, PolicyService};
use crate::infrastructure::persistence::PostgresWalletRepository;
use crate::infrastructure::hsm_kafka::{SoftHsmKeyManager, KafkaMessageBus, InMemoryAuditLog};

/// CLI arguments
#[derive(Parser, Debug)]
#[command(name = "csic-cwgs")]
#[command(author = "CSIC Platform")]
#[command(version = "1.0.0")]
#[command(about = "CSIC Custodial Wallet Governance System", long_about = None)]
struct Args {
    /// HTTP server port
    #[arg(short, long, default_value = "8081")]
    http_port: u16,

    /// gRPC server port
    #[arg(short, long, default_value = "50052")]
    grpc_port: u16,

    /// PostgreSQL connection string
    #[arg(short, long, default_value = "postgres://csic_admin:secure@localhost:5432/cwgs")]
    database_url: String,

    /// Kafka brokers
    #[arg(long, default_value = "localhost:9092")]
    kafka_brokers: String,

    /// HSM library path
    #[arg(long, default_value = "/usr/lib/softhsm/libsofthsm2.so")]
    hsm_library: String,

    /// HSM PIN
    #[arg(long, default_value = "12345678")]
    hsm_pin: String,

    /// Log level
    #[arg(long, default_value = "info")]
    log_level: String,
}

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    // Parse CLI arguments
    let args = Args::parse();

    // Initialize tracing
    tracing_subscriber::registry()
        .with(tracing_subscriber::EnvFilter::new(&args.log_level))
        .with(tracing_subscriber::fmt::layer())
        .init();

    tracing::info!("Starting CSIC Custodial Wallet Governance System");

    // Create PostgreSQL connection pool
    let pool = PgPool::connect(&args.database_url).await?;
    tracing::info!("Connected to PostgreSQL database");

    // Initialize repositories
    let wallet_repo = Arc::new(PostgresWalletRepository::new(pool.clone()));
    let policy_repo = Arc::new(PostgresPolicyRepository::new(pool.clone()));

    // Initialize key manager (SoftHSM for development)
    let key_manager: Arc<dyn crate::domain::KeyManager> = Arc::new(SoftHsmKeyManager::new());

    // Initialize audit log (in-memory for development)
    let audit_log: Arc<dyn crate::application::AuditLog> = Arc::new(InMemoryAuditLog::new());

    // Initialize message bus (Kafka simulator for development)
    let message_bus: Arc<dyn crate::application::MessageBus> = Arc::new(KafkaMessageBus::new());

    // Create application services
    let wallet_service = WalletService::new(
        Arc::clone(&wallet_repo),
        Arc::clone(&policy_repo),
        Arc::clone(&key_manager),
        Arc::clone(&audit_log),
        Arc::clone(&message_bus),
    );

    let policy_service = PolicyService::new(
        Arc::clone(&policy_repo),
        Arc::clone(&audit_log),
    );

    // Create application state
    let app_state = AppState {
        wallet_service,
        policy_service,
    };

    // Create router
    let router = create_router(app_state);

    // Create HTTP server address
    let addr = SocketAddr::from(([0, 0, 0, 0], args.http_port));

    // Start HTTP server
    tracing::info!("Starting HTTP server on port {}", args.http_port);
    Server::bind(&addr)
        .serve(router.into_make_service())
        .await?;

    Ok(())
}

/// Additional repository import for PostgreSQL policy repository
use crate::infrastructure::persistence::PostgresPolicyRepository;
