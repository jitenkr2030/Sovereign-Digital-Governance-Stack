use std::fmt;

use chrono::{DateTime, Utc};
use serde::{Deserialize, Serialize};
use uuid::Uuid;

/// Categories of audit events
#[derive(Debug, Clone, PartialEq, Eq, sqlx::Type, Serialize, Deserialize)]
#[sqlx(type_name = "audit_category")]
pub enum AuditCategory {
    /// Wallet lifecycle events
    Wallet,
    /// Policy changes
    Policy,
    /// Transaction signing
    Signing,
    /// Key management
    KeyManagement,
    /// Emergency actions
    Emergency,
    /// Authentication events
    Authentication,
    /// Administrative actions
    Administration,
    /// Compliance reports
    Compliance,
    /// System events
    System,
}

/// Audit event types
#[derive(Debug, Clone, PartialEq, Eq, sqlx::Type, Serialize, Deserialize)]
#[sqlx(type_name = "audit_event_type")]
pub enum AuditEventType {
    // Wallet events
    WalletCreated,
    WalletFrozen,
    WalletUnfrozen,
    WalletArchived,
    WalletDeleted,
    WalletRecoveryInitiated,
    WalletRecoveryCompleted,
    WalletLimitsUpdated,
    
    // Policy events
    PolicyCreated,
    PolicyUpdated,
    PolicyActivated,
    PolicyDeactivated,
    PolicyDeleted,
    BlacklistEntryAdded,
    BlacklistEntryRemoved,
    WhitelistEntryAdded,
    WhitelistEntryRemoved,
    
    // Signing events
    SigningRequestCreated,
    SigningRequestApproved,
    SigningRequestRejected,
    SigningRequestCompleted,
    SigningRequestTimedOut,
    SigningRequestCancelled,
    MultiSigThresholdReached,
    
    // Key management events
    KeyGenerated,
    KeyRotated,
    KeyCompromised,
    KeyRetired,
    KeyDestroyed,
    KeyAccessed,
    
    // Emergency events
    EmergencyFreezeInitiated,
    EmergencyFreezeCompleted,
    EmergencyUnfreezeInitiated,
    EmergencyUnfreezeCompleted,
    EmergencyAccessRequested,
    EmergencyAccessApproved,
    EmergencyAccessDenied,
    
    // Authentication events
    LoginSuccess,
    LoginFailed,
    Logout,
    MfaVerified,
    MfaFailed,
    SessionExpired,
    PasswordChanged,
    
    // Administrative events
    AdminCreated,
    AdminUpdated,
    AdminDeleted,
    PermissionsChanged,
    RoleAssigned,
    RoleRevoked,
    
    // Compliance events
    RegulatoryReportGenerated,
    AuditLogExport,
    DataAccessRequest,
    DataExportCompleted,
    
    // System events
    ServiceStarted,
    ServiceStopped,
    ConfigurationChanged,
    DatabaseMigration,
    IntegrationConnected,
    IntegrationDisconnected,
}

impl fmt::Display for AuditEventType {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            AuditEventType::WalletCreated => write!(f, "WALLET_CREATED"),
            AuditEventType::WalletFrozen => write!(f, "WALLET_FROZEN"),
            AuditEventType::WalletUnfrozen => write!(f, "WALLET_UNFROZEN"),
            AuditEventType::WalletArchived => write!(f, "WALLET_ARCHIVED"),
            AuditEventType::WalletDeleted => write!(f, "WALLET_DELETED"),
            AuditEventType::WalletRecoveryInitiated => write!(f, "WALLET_RECOVERY_INITIATED"),
            AuditEventType::WalletRecoveryCompleted => write!(f, "WALLET_RECOVERY_COMPLETED"),
            AuditEventType::WalletLimitsUpdated => write!(f, "WALLET_LIMITS_UPDATED"),
            AuditEventType::PolicyCreated => write!(f, "POLICY_CREATED"),
            AuditEventType::PolicyUpdated => write!(f, "POLICY_UPDATED"),
            AuditEventType::PolicyActivated => write!(f, "POLICY_ACTIVATED"),
            AuditEventType::PolicyDeactivated => write!(f, "POLICY_DEACTIVATED"),
            AuditEventType::PolicyDeleted => write!(f, "POLICY_DELETED"),
            AuditEventType::BlacklistEntryAdded => write!(f, "BLACKLIST_ENTRY_ADDED"),
            AuditEventType::BlacklistEntryRemoved => write!(f, "BLACKLIST_ENTRY_REMOVED"),
            AuditEventType::WhitelistEntryAdded => write!(f, "WHITELIST_ENTRY_ADDED"),
            AuditEventType::WhitelistEntryRemoved => write!(f, "WHITELIST_ENTRY_REMOVED"),
            AuditEventType::SigningRequestCreated => write!(f, "SIGNING_REQUEST_CREATED"),
            AuditEventType::SigningRequestApproved => write!(f, "SIGNING_REQUEST_APPROVED"),
            AuditEventType::SigningRequestRejected => write!(f, "SIGNING_REQUEST_REJECTED"),
            AuditEventType::SigningRequestCompleted => write!(f, "SIGNING_REQUEST_COMPLETED"),
            AuditEventType::SigningRequestTimedOut => write!(f, "SIGNING_REQUEST_TIMED_OUT"),
            AuditEventType::SigningRequestCancelled => write!(f, "SIGNING_REQUEST_CANCELLED"),
            AuditEventType::MultiSigThresholdReached => write!(f, "MULTI_SIG_THRESHOLD_REACHED"),
            AuditEventType::KeyGenerated => write!(f, "KEY_GENERATED"),
            AuditEventType::KeyRotated => write!(f, "KEY_ROTATED"),
            AuditEventType::KeyCompromised => write!(f, "KEY_COMPROMISED"),
            AuditEventType::KeyRetired => write!(f, "KEY_RETIRED"),
            AuditEventType::KeyDestroyed => write!(f, "KEY_DESTROYED"),
            AuditEventType::KeyAccessed => write!(f, "KEY_ACCESSED"),
            AuditEventType::EmergencyFreezeInitiated => write!(f, "EMERGENCY_FREEZE_INITIATED"),
            AuditEventType::EmergencyFreezeCompleted => write!(f, "EMERGENCY_FREEZE_COMPLETED"),
            AuditEventType::EmergencyUnfreezeInitiated => write!(f, "EMERGENCY_UNFREEZE_INITIATED"),
            AuditEventType::EmergencyUnfreezeCompleted => write!(f, "EMERGENCY_UNFREEZE_COMPLETED"),
            AuditEventType::EmergencyAccessRequested => write!(f, "EMERGENCY_ACCESS_REQUESTED"),
            AuditEventType::EmergencyAccessApproved => write!(f, "EMERGENCY_ACCESS_APPROVED"),
            AuditEventType::EmergencyAccessDenied => write!(f, "EMERGENCY_ACCESS_DENIED"),
            AuditEventType::LoginSuccess => write!(f, "LOGIN_SUCCESS"),
            AuditEventType::LoginFailed => write!(f, "LOGIN_FAILED"),
            AuditEventType::Logout => write!(f, "LOGOUT"),
            AuditEventType::MfaVerified => write!(f, "MFA_VERIFIED"),
            AuditEventType::MfaFailed => write!(f, "MFA_FAILED"),
            AuditEventType::SessionExpired => write!(f, "SESSION_EXPIRED"),
            AuditEventType::PasswordChanged => write!(f, "PASSWORD_CHANGED"),
            AuditEventType::AdminCreated => write!(f, "ADMIN_CREATED"),
            AuditEventType::AdminUpdated => write!(f, "ADMIN_UPDATED"),
            AuditEventType::AdminDeleted => write!(f, "ADMIN_DELETED"),
            AuditEventType::PermissionsChanged => write!(f, "PERMISSIONS_CHANGED"),
            AuditEventType::RoleAssigned => write!(f, "ROLE_ASSIGNED"),
            AuditEventType::RoleRevoked => write!(f, "ROLE_REVOKED"),
            AuditEventType::RegulatoryReportGenerated => write!(f, "REGULATORY_REPORT_GENERATED"),
            AuditEventType::AuditLogExport => write!(f, "AUDIT_LOG_EXPORT"),
            AuditEventType::DataAccessRequest => write!(f, "DATA_ACCESS_REQUEST"),
            AuditEventType::DataExportCompleted => write!(f, "DATA_EXPORT_COMPLETED"),
            AuditEventType::ServiceStarted => write!(f, "SERVICE_STARTED"),
            AuditEventType::ServiceStopped => write!(f, "SERVICE_STOPPED"),
            AuditEventType::ConfigurationChanged => write!(f, "CONFIGURATION_CHANGED"),
            AuditEventType::DatabaseMigration => write!(f, "DATABASE_MIGRATION"),
            AuditEventType::IntegrationConnected => write!(f, "INTEGRATION_CONNECTED"),
            AuditEventType::IntegrationDisconnected => write!(f, "INTEGRATION_DISCONNECTED"),
        }
    }
}

/// Severity level of audit events
#[derive(Debug, Clone, PartialEq, Eq, sqlx::Type, Serialize, Deserialize)]
#[sqlx(type_name = "audit_severity")]
pub enum AuditSeverity {
    /// Informational (default)
    Info,
    /// Low severity
    Low,
    /// Medium severity
    Medium,
    /// High severity
    High,
    /// Critical severity (requires immediate attention)
    Critical,
}

impl fmt::Display for AuditSeverity {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            AuditSeverity::Info => write!(f, "INFO"),
            AuditSeverity::Low => write!(f, "LOW"),
            AuditSeverity::Medium => write!(f, "MEDIUM"),
            AuditSeverity::High => write!(f, "HIGH"),
            AuditSeverity::Critical => write!(f, "CRITICAL"),
        }
    }
}

/// Actor who performed the action
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct AuditActor {
    /// Actor ID (user ID, system ID, etc.)
    pub id: String,
    /// Actor type
    pub actor_type: ActorType,
    /// Actor label/name
    pub label: String,
    /// IP address if applicable
    pub ip_address: Option<String>,
    /// User agent if applicable
    pub user_agent: Option<String>,
    /// Session ID if applicable
    pub session_id: Option<String>,
}

/// Types of actors
#[derive(Debug, Clone, PartialEq, Eq, Serialize, Deserialize)]
pub enum ActorType {
    /// Human user
    User,
    /// System/automated process
    System,
    /// Service account
    Service,
    /// Admin user
    Admin,
    /// External system
    External,
    /// HSM
    Hsm,
}

impl fmt::Display for ActorType {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            ActorType::User => write!(f, "USER"),
            ActorType::System => write!(f, "SYSTEM"),
            ActorType::Service => write!(f, "SERVICE"),
            ActorType::Admin => write!(f, "ADMIN"),
            ActorType::External => write!(f, "EXTERNAL"),
            ActorType::Hsm => write!(f, "HSM"),
        }
    }
}

/// Audit event aggregate
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct AuditEvent {
    /// Unique event ID
    pub id: Uuid,
    /// Event category
    pub category: AuditCategory,
    /// Event type
    pub event_type: AuditEventType,
    /// Severity level
    pub severity: AuditSeverity,
    /// Actor who performed the action
    pub actor: AuditActor,
    /// Target resource ID
    pub resource_id: Option<Uuid>,
    /// Target resource type
    pub resource_type: Option<String>,
    /// Event description
    pub description: String,
    /// Additional metadata
    pub metadata: serde_json::Value,
    /// Before state (for state changes)
    pub before_state: Option<serde_json::Value>,
    /// After state (for state changes)
    pub after_state: Option<serde_json::Value>,
    /// Whether the action was successful
    pub success: bool,
    /// Error message if action failed
    pub error_message: Option<String>,
    /// Request ID for correlation
    pub request_id: Option<Uuid>,
    /// Correlation ID for tracing
    pub correlation_id: Option<Uuid>,
    /// Event timestamp
    pub timestamp: DateTime<Utc>,
}

impl AuditEvent {
    /// Creates a new audit event
    pub fn new(
        category: AuditCategory,
        event_type: AuditEventType,
        severity: AuditSeverity,
        actor: AuditActor,
        description: String,
    ) -> Self {
        Self {
            id: Uuid::new_v4(),
            category,
            event_type,
            severity,
            actor,
            resource_id: None,
            resource_type: None,
            description,
            metadata: serde_json::json!({}),
            before_state: None,
            after_state: None,
            success: true,
            error_message: None,
            request_id: None,
            correlation_id: None,
            timestamp: Utc::now(),
        }
    }

    /// Sets the resource
    pub fn with_resource(mut self, resource_id: Uuid, resource_type: &str) -> Self {
        self.resource_id = Some(resource_id);
        self.resource_type = Some(resource_type.to_string());
        self
    }

    /// Sets metadata
    pub fn with_metadata(mut self, metadata: serde_json::Value) -> Self {
        self.metadata = metadata;
        self
    }

    /// Sets state changes
    pub fn with_state_changes(
        mut self,
        before: Option<serde_json::Value>,
        after: Option<serde_json::Value>,
    ) -> Self {
        self.before_state = before;
        self.after_state = after;
        self
    }

    /// Marks the event as failed
    pub fn with_failure(mut self, error_message: String) -> Self {
        self.success = false;
        self.error_message = Some(error_message);
        self
    }

    /// Sets correlation IDs
    pub fn with_correlation(mut self, request_id: Uuid, correlation_id: Uuid) -> Self {
        self.request_id = Some(request_id);
        self.correlation_id = Some(correlation_id);
        self
    }

    /// Determines severity based on event type and success
    pub fn determine_severity(event_type: &AuditEventType, success: bool) -> AuditSeverity {
        // Critical events
        let critical_events = [
            AuditEventType::KeyCompromised,
            AuditEventType::EmergencyFreezeInitiated,
            AuditEventType::EmergencyAccessRequested,
            AuditEventType::EmergencyAccessDenied,
        ];
        
        if critical_events.contains(event_type) {
            return AuditSeverity::Critical;
        }

        // High severity events
        let high_events = [
            AuditEventType::KeyDestroyed,
            AuditEventType::WalletFrozen,
            AuditEventType::SigningRequestRejected,
            AuditEventType::KeyRetired,
            AuditEventType::EmergencyFreezeCompleted,
            AuditEventType::LoginFailed,
            AuditEventType::MfaFailed,
        ];
        
        if high_events.contains(event_type) {
            return AuditSeverity::High;
        }

        // Failed operations
        if !success {
            return AuditSeverity::Medium;
        }

        AuditSeverity::Info
    }
}

/// Audit query parameters
#[derive(Debug, Clone, Default)]
pub struct AuditQuery {
    /// Filter by category
    pub categories: Vec<AuditCategory>,
    /// Filter by event types
    pub event_types: Vec<AuditEventType>,
    /// Filter by severity
    pub severities: Vec<AuditSeverity>,
    /// Filter by actor ID
    pub actor_id: Option<String>,
    /// Filter by resource ID
    pub resource_id: Option<Uuid>,
    /// Filter by resource type
    pub resource_type: Option<String>,
    /// Start timestamp
    pub start_time: Option<DateTime<Utc>>,
    /// End timestamp
    pub end_time: Option<DateTime<Utc>>,
    /// Filter by success status
    pub success: Option<bool>,
    /// Search query (searches description and metadata)
    pub search_query: Option<String>,
    /// Pagination: page number (1-indexed)
    pub page: u32,
    /// Pagination: page size
    pub page_size: u32,
    /// Sort field
    pub sort_by: AuditSortField,
    /// Sort direction
    pub sort_direction: SortDirection,
}

/// Fields to sort audit logs by
#[derive(Debug, Clone, PartialEq, Eq)]
pub enum AuditSortField {
    Timestamp,
    Category,
    EventType,
    Severity,
    ActorId,
    ResourceId,
}

impl Default for AuditSortField {
    fn default() -> Self {
        AuditSortField::Timestamp
    }
}

/// Sort direction
#[derive(Debug, Clone, PartialEq, Eq)]
pub enum SortDirection {
    Asc,
    Desc,
}

impl Default for SortDirection {
    fn default() -> Self {
        SortDirection::Desc
    }
}

/// Audit query result with pagination
#[derive(Debug, Clone)]
pub struct AuditQueryResult {
    /// List of audit events
    pub events: Vec<AuditEvent>,
    /// Total count matching query
    pub total_count: u64,
    /// Current page
    pub page: u32,
    /// Page size
    pub page_size: u32,
    /// Total pages
    pub total_pages: u32,
    /// Whether there are more pages
    pub has_next: bool,
    /// Whether there are previous pages
    pub has_previous: bool,
}

/// Compliance report request
#[derive(Debug, Clone)]
pub struct ComplianceReportRequest {
    /// Report type
    pub report_type: ComplianceReportType,
    /// Start of reporting period
    pub period_start: DateTime<Utc>,
    /// End of reporting period
    pub period_end: DateTime<Utc>,
    /// Target wallet IDs (None = all wallets)
    pub wallet_ids: Option<Vec<Uuid>>,
    /// Include detailed transaction logs
    pub include_transactions: bool,
    /// Include policy violations
    pub include_violations: bool,
    /// Include audit trail
    pub include_audit_trail: bool,
    /// Format of the report
    pub format: ReportFormat,
    /// Requesting authority
    pub requesting_authority: String,
}

/// Types of compliance reports
#[derive(Debug, Clone, PartialEq, Eq)]
pub enum ComplianceReportType {
    /// Daily activity report
    DailyActivity,
    /// Weekly summary
    WeeklySummary,
    /// Monthly comprehensive report
    MonthlyComprehensive,
    /// Quarterly regulatory report
    QuarterlyRegulatory,
    /// Annual audit report
    AnnualAudit,
    /// Suspicious activity report
    SuspiciousActivity,
    /// Large transaction report
    LargeTransactions,
    /// Geographic risk report
    GeographicRisk,
    /// Wallet freeze report
    WalletFreezeReport,
    /// Custom ad-hoc report
    AdHoc,
}

/// Report output formats
#[derive(Debug, Clone, PartialEq, Eq)]
pub enum ReportFormat {
    Json,
    Csv,
    Pdf,
    Xml,
}

impl fmt::Display for ReportFormat {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            ReportFormat::Json => write!(f, "JSON"),
            ReportFormat::Csv => write!(f, "CSV"),
            ReportFormat::Pdf => write!(f, "PDF"),
            ReportFormat::Xml => write!(f, "XML"),
        }
    }
}
