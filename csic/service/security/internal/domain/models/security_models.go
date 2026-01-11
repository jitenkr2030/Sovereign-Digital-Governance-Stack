package domain

import (
	"time"

	"github.com/google/uuid"
)

// AuditLog represents a tamper-proof audit log entry with cryptographic chaining
type AuditLog struct {
	ID            uuid.UUID     `json:"id" db:"id"`
	Timestamp     time.Time     `json:"timestamp" db:"timestamp"`
	ActorID       string        `json:"actor_id" db:"actor_id"`
	ActorType     ActorType     `json:"actor_type" db:"actor_type"`
	Action        string        `json:"action" db:"action"`
	ActionType    ActionType    `json:"action_type" db:"action_type"`
	ResourceID    string        `json:"resource_id" db:"resource_id"`
	ResourceType  string        `json:"resource_type" db:"resource_type"`
	ClientIP      string        `json:"client_ip" db:"client_ip"`
	UserAgent     string        `json:"user_agent" db:"user_agent"`
	SessionID     string        `json:"session_id" db:"session_id"`
	Payload       AuditPayload  `json:"payload" db:"payload"`
	Sensitivity   Sensitivity   `json:"sensitivity" db:"sensitivity"`
	
	// Cryptographic chain fields for tamper-proofing
	PreviousHash  string        `json:"previous_hash" db:"previous_hash"`
	CurrentHash   string        `json:"current_hash" db:"current_hash"`
	Signature     []byte        `json:"-" db:"signature"`
	ChainIndex    uint64        `json:"chain_index" db:"chain_index"`
	
	// Metadata
	CorrelationID string        `json:"correlation_id" db:"correlation_id"`
	SourceService string        `json:"source_service" db:"source_service"`
	
	CreatedAt     time.Time     `json:"created_at" db:"created_at"`
}

// ActorType represents the type of actor performing the action
type ActorType string

const (
	ActorTypeUser        ActorType = "user"
	ActorTypeService     ActorType = "service"
	ActorTypeSystem      ActorType = "system"
	ActorTypeAdmin       ActorType = "admin"
	ActorTypeExternal    ActorType = "external"
)

// ActionType categorizes the type of action performed
type ActionType string

const (
	ActionTypeAuthentication ActionType = "authentication"
	ActionTypeAuthorization  ActionType = "authorization"
	ActionTypeDataAccess     ActionType = "data_access"
	ActionTypeDataModification ActionType = "data_modification"
	ActionTypeConfiguration  ActionType = "configuration"
	ActionTypeAdministrative ActionType = "administrative"
	ActionTypeSecurity       ActionType = "security"
	ActionTypeSystem         ActionType = "system"
)

// Sensitivity defines the sensitivity level of the audit entry
type Sensitivity string

const (
	SensitivityLow      Sensitivity = "low"
	SensitivityMedium   Sensitivity = "medium"
	SensitivityHigh     Sensitivity = "high"
	SensitivityCritical Sensitivity = "critical"
)

// AuditPayload contains the data associated with the audit entry
type AuditPayload struct {
	Request     map[string]interface{} `json:"request,omitempty"`
	Response    map[string]interface{} `json:"response,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	Reason      string                 `json:"reason,omitempty"`
	Result      string                 `json:"result,omitempty"`
	ErrorCode   string                 `json:"error_code,omitempty"`
	ErrorMessage string               `json:"error_message,omitempty"`
}

// AuditFilter represents filter criteria for querying audit logs
type AuditFilter struct {
	StartTime      *time.Time   `json:"start_time"`
	EndTime        *time.Time   `json:"end_time"`
	ActorIDs       []string     `json:"actor_ids"`
	ActorTypes     []ActorType  `json:"actor_types"`
	Actions        []string     `json:"actions"`
	ActionTypes    []ActionType `json:"action_types"`
	ResourceTypes  []string     `json:"resource_types"`
	Sensitivities  []Sensitivity `json:"sensitivities"`
	CorrelationIDs []string     `json:"correlation_ids"`
	SearchQuery    string       `json:"search_query"`
	
	// Pagination
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
	
	// Sorting
	SortBy    string `json:"sort_by"` // timestamp, chain_index
	SortOrder string `json:"sort_order"` // asc, desc
}

// IntegrityReport represents the result of an integrity verification
type IntegrityReport struct {
	ID              uuid.UUID          `json:"id"`
	ReportTime      time.Time          `json:"report_time"`
	StartChainIndex uint64             `json:"start_chain_index"`
	EndChainIndex   uint64             `json:"end_chain_index"`
	TotalRecords    uint64             `json:"total_records"`
	ValidRecords    uint64             `json:"valid_records"`
	InvalidRecords  uint64             `json:"invalid_records"`
	Status          IntegrityStatus    `json:"status"`
	Issues          []IntegrityIssue   `json:"issues,omitempty"`
	VerifiedBy      string             `json:"verified_by"`
	Duration        time.Duration      `json:"duration"`
}

// IntegrityStatus represents the status of integrity verification
type IntegrityStatus string

const (
	IntegrityStatusValid   IntegrityStatus = "valid"
	IntegrityStatusInvalid IntegrityStatus = "invalid"
	IntegrityStatusPartial IntegrityStatus = "partial"
	IntegrityStatusError   IntegrityStatus = "error"
)

// IntegrityIssue represents an issue found during integrity verification
type IntegrityIssue struct {
	ChainIndex   uint64    `json:"chain_index"`
	RecordID     uuid.UUID `json:"record_id"`
	IssueType    string    `json:"issue_type"`
	Description  string    `json:"description"`
	Timestamp    time.Time `json:"timestamp"`
}

// ChainState represents the current state of the hash chain
type ChainState struct {
	ID              uuid.UUID `json:"id"`
	CurrentIndex    uint64    `json:"current_index"`
	CurrentHash     string    `json:"current_hash"`
	LastTimestamp   time.Time `json:"last_timestamp"`
	ChainIdentifier string    `json:"chain_identifier"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// HSMConfig represents the configuration for Hardware Security Module
type HSMConfig struct {
	ID              uuid.UUID     `json:"id" db:"id"`
	Name            string        `json:"name" db:"name"`
	Vendor          HSMVendor     `json:"vendor" db:"vendor"`
	ConnectionType  HSMConnectionType `json:"connection_type" db:"connection_type"`
	Endpoint        string        `json:"endpoint" db:"endpoint"`
	Port            int           `json:"port" db:"port"`
	SlotID          int           `json:"slot_id" db:"slot_id"`
	LibraryPath     string        `json:"library_path" db:"library_path"`
	KeyLabel        string        `json:"key_label" db:"key_label"`
	TokenLabel      string        `json:"token_label" db:"token_label"`
	IsActive        bool          `json:"is_active" db:"is_active"`
	Priority        int           `json:"priority" db:"priority"` // For failover
	HealthCheckInterval int       `json:"health_check_interval" db:"health_check_interval"` // seconds
	LastHealthCheck  *time.Time   `json:"last_health_check" db:"last_health_check"`
	HealthStatus    HealthStatus  `json:"health_status" db:"health_status"`
	CreatedAt       time.Time     `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time     `json:"updated_at" db:"updated_at"`
}

// HSMVendor represents the HSM vendor
type HSMVendor string

const (
	HSMVendorThales       HSMVendor = "thales"
	HSMVendorUtimaco      HSMVendor = "utimaco"
	HSMVendorAWSCloudHSM  HSMVendor = "aws_cloudhsm"
	HSMVendorAzureKeyVault HSMVendor = "azure_keyvault"
	HSMVendorGoogleCloudKMS HSMVendor = "google_cloud_kms"
	HSMVendorSoftHSM      HSMVendor = "soft_hsm"
	HSMVendorHashicorpVault HSMVendor = "hashicorp_vault"
)

// HSMConnectionType represents how to connect to the HSM
type HSMConnectionType string

const (
	HSMConnectionPKCS11     HSMConnectionType = "pkcs11"
	HSMConnectionREST       HSMConnectionType = "rest_api"
	HSMConnectionGRPC       HSMConnectionType = "grpc"
	HSMConnectionCloudSDK   HSMConnectionType = "cloud_sdk"
)

// HealthStatus represents the health status
type HealthStatus string

const (
	HealthStatusHealthy   HealthStatus = "healthy"
	HealthStatusDegraded  HealthStatus = "degraded"
	HealthStatusUnhealthy HealthStatus = "unhealthy"
	HealthStatusUnknown   HealthStatus = "unknown"
)

// HSMSigningRequest represents a request to sign data with HSM
type HSMSigningRequest struct {
	Data        []byte         `json:"data"`
	KeyID       string         `json:"key_id"`
	Algorithm   SigningAlgorithm `json:"algorithm"`
	Context     string         `json:"context,omitempty"`
}

// HSMSigningResponse represents the response from HSM signing
type HSMSigningResponse struct {
	Signature   []byte       `json:"signature"`
	KeyID       string       `json:"key_id"`
	Algorithm   SigningAlgorithm `json:"algorithm"`
	Timestamp   time.Time    `json:"timestamp"`
}

// SigningAlgorithm represents the signing algorithm
type SigningAlgorithm string

const (
	SigningAlgorithmRS256 SigningAlgorithm = "RS256"
	SigningAlgorithmRS384 SigningAlgorithm = "RS384"
	SigningAlgorithmRS512 SigningAlgorithm = "RS512"
	SigningAlgorithmES256 SigningAlgorithm = "ES256"
	SigningAlgorithmES384 SigningAlgorithm = "ES384"
	SigningAlgorithmES512 SigningAlgorithm = "ES512"
	SigningAlgorithmHS256 SigningAlgorithm = "HS256"
)

// HSMOperation represents an operation performed on the HSM
type HSMOperation struct {
	ID            uuid.UUID       `json:"id"`
	OperationType string          `json:"operation_type"`
	KeyID         string          `json:"key_id"`
	Status        OperationStatus `json:"status"`
	Duration      time.Duration   `json:"duration"`
	ErrorMessage  string          `json:"error_message,omitempty"`
	Timestamp     time.Time       `json:"timestamp"`
}

// OperationStatus represents the status of an HSM operation
type OperationStatus string

const (
	OperationStatusSuccess OperationStatus = "success"
	OperationStatusFailed  OperationStatus = "failed"
	OperationStatusPending OperationStatus = "pending"
)

// SIEMConfig represents the configuration for SIEM integration
type SIEMConfig struct {
	ID              uuid.UUID        `json:"id" db:"id"`
	Name            string           `json:"name" db:"name"`
	Vendor          SIEMVendor       `json:"vendor" db:"vendor"`
	Endpoint        string           `json:"endpoint" db:"endpoint"`
	Port            int              `json:"port" db:"port"`
	Protocol        SIEMProtocol     `json:"protocol" db:"protocol"`
	AuthType        SIEMAuthType     `json:"auth_type" db:"auth_type"`
	APIKey          string           `json:"-" db:"api_key"` // Encrypted
	Certificate     string           `json:"certificate" db:"certificate"` // PEM encoded
	IsActive        bool             `json:"is_active" db:"is_active"`
	Priority        int              `json:"priority" db:"priority"`
	BatchSize       int              `json:"batch_size" db:"batch_size"`
	FlushInterval   int              `json:"flush_interval" db:"flush_interval"` // seconds
	RetryAttempts   int              `json:"retry_attempts" db:"retry_attempts"`
	RetryDelay      int              `json:"retry_delay" db:"retry_delay"` // seconds
	EventTypes      []string         `json:"event_types" db:"event_types"` // JSON array
	Filters         string           `json:"filters" db:"filters"` // JSON
	CreatedAt       time.Time        `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time        `json:"updated_at" db:"updated_at"`
}

// SIEMVendor represents the SIEM vendor
type SIEMVendor string

const (
	SIEMVendorSplunk        SIEMVendor = "splunk"
	SIEMVendorELK           SIEMVendor = "elk"
	SIEMVendorAzureSentinel SIEMVendor = "azure_sentinel"
	SIEMVendorAWSecurityHub SIEMVendor = "aws_security_hub"
	SIEMVendorQRadar        SIEMVendor = "qradar"
	SIEMVendorSyslog        SIEMVendor = "syslog"
	SIEMVendorGraylog       SIEMVendor = "graylog"
	SIEMVendorDatadog       SIEMVendor = "datadog"
)

// SIEMProtocol represents the communication protocol
type SIEMProtocol string

const (
	SIEMProtocolHTTPS     SIEMProtocol = "https"
	SIEMProtocolHTTP      SIEMProtocol = "http"
	SIEMProtocolSyslogTCP SIEMProtocol = "syslog_tcp"
	SIEMProtocolSyslogUDP SIEMProtocol = "syslog_udp"
	SIEMProtocolSyslogTLS SIEMProtocol = "syslog_tls"
)

// SIEMAuthType represents the authentication type
type SIEMAuthType string

const (
	SIEMAuthAPIKey    SIEMAuthType = "api_key"
	SIEMAuthBasic     SIEMAuthType = "basic"
	SIEMAuthOAuth     SIEMAuthType = "oauth"
	SIEMAuthTLSCert   SIEMAuthType = "tls_cert"
	SIEMAuthNone      SIEMAuthType = "none"
)

// SecurityEvent represents a security event for SIEM forwarding
type SecurityEvent struct {
	ID              uuid.UUID          `json:"id"`
	Timestamp       time.Time          `json:"timestamp"`
	EventType       string             `json:"event_type"`
	Severity        EventSeverity      `json:"severity"`
	SourceIP        string             `json:"source_ip"`
	DestinationIP   string             `json:"destination_ip"`
	UserID          string             `json:"user_id"`
	ServiceName     string             `json:"service_name"`
	EventName       string             `json:"event_name"`
	Message         string             `json:"message"`
	Description     string             `json:"description"`
	Tags            map[string]string  `json:"tags"`
	RawData         map[string]interface{} `json:"raw_data"`
	SIEMFormat      SIEMFormat         `json:"siem_format"`
	ForwardingStatus ForwardingStatus  `json:"forwarding_status"`
	ForwardedAt     *time.Time         `json:"forwarded_at,omitempty"`
	RetryCount      int                `json:"retry_count"`
	ErrorMessage    string             `json:"error_message,omitempty"`
}

// EventSeverity represents the severity of a security event
type EventSeverity string

const (
	SeverityInfo     EventSeverity = "INFO"
	SeverityLow      EventSeverity = "LOW"
	SeverityMedium   EventSeverity = "MEDIUM"
	SeverityHigh     EventSeverity = "HIGH"
	SeverityCritical EventSeverity = "CRITICAL"
)

// SIEMFormat represents the format for SIEM forwarding
type SIEMFormat string

const (
	SIEMFormatCEF   SIEMFormat = "CEF"
	SIEMFormatLEEF  SIEMFormat = "LEEF"
	SIEMFormatJSON  SIEMFormat = "JSON"
	SIEMFormatSyslog SIEMFormat = "SYSLOG"
)

// ForwardingStatus represents the status of SIEM forwarding
type ForwardingStatus string

const (
	ForwardingStatusPending   ForwardingStatus = "pending"
	ForwardingStatusForwarded ForwardingStatus = "forwarded"
	ForwardingStatusFailed    ForwardingStatus = "failed"
	ForwardingStatusSkipped   ForwardingStatus = "skipped"
)

// SIEMForwardingLog represents the log of SIEM forwarding attempts
type SIEMForwardingLog struct {
	ID              uuid.UUID        `json:"id"`
	SIEMConfigID    uuid.UUID        `json:"siem_config_id"`
	EventID         uuid.UUID        `json:"event_id"`
	Status          ForwardingStatus `json:"status"`
	AttemptCount    int              `json:"attempt_count"`
	FirstAttemptAt  time.Time        `json:"first_attempt_at"`
	LastAttemptAt   time.Time        `json:"last_attempt_at"`
	NextRetryAt     *time.Time       `json:"next_retry_at,omitempty"`
	ErrorMessage    string           `json:"error_message,omitempty"`
	ResponseCode    string           `json:"response_code,omitempty"`
	ResponseBody    string           `json:"response_body,omitempty"`
	CreatedAt       time.Time        `json:"created_at"`
}

// SecurityPolicy represents a security policy configuration
type SecurityPolicy struct {
	ID              uuid.UUID     `json:"id"`
	Name            string        `json:"name"`
	Description     string        `json:"description"`
	PolicyType      PolicyType    `json:"policy_type"`
	Rules           []PolicyRule  `json:"rules"`
	Severity        EventSeverity `json:"severity"`
	IsActive        bool          `json:"is_active"`
	Priority        int           `json:"priority"`
	CreatedAt       time.Time     `json:"created_at"`
	UpdatedAt       time.Time     `json:"updated_at"`
}

// PolicyType represents the type of security policy
type PolicyType string

const (
	PolicyTypeAccessControl    PolicyType = "access_control"
	PolicyTypeDataProtection   PolicyType = "data_protection"
	PolicyTypeKeyManagement    PolicyType = "key_management"
	PolicyTypeAuditRetention   PolicyType = "audit_retention"
	PolicyTypeIncidentResponse PolicyType = "incident_response"
)

// PolicyRule represents a rule within a security policy
type PolicyRule struct {
	ID          uuid.UUID              `json:"id"`
	Condition   string                 `json:"condition"`
	Action      string                 `json:"action"`
	Parameters  map[string]interface{} `json:"parameters"`
	Priority    int                    `json:"priority"`
	IsEnabled   bool                   `json:"is_enabled"`
}

// SecurityAlert represents a security alert generated by policies
type SecurityAlert struct {
	ID              uuid.UUID       `json:"id"`
	AlertType       string          `json:"alert_type"`
	Severity        EventSeverity   `json:"severity"`
	Title           string          `json:"title"`
	Description     string          `json:"description"`
	SourceEventID   uuid.UUID       `json:"source_event_id"`
	PolicyID        uuid.UUID       `json:"policy_id"`
	RuleID          uuid.UUID       `json:"rule_id"`
	Status          AlertStatus     `json:"status"`
	AssignedTo      string          `json:"assigned_to,omitempty"`
	ResolvedAt      *time.Time      `json:"resolved_at,omitempty"`
	Resolution      string          `json:"resolution,omitempty"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
}

// AlertStatus represents the status of a security alert
type AlertStatus string

const (
	AlertStatusOpen      AlertStatus = "open"
	AlertStatusInvestigating AlertStatus = "investigating"
	AlertStatusResolved  AlertStatus = "resolved"
	AlertStatusDismissed AlertStatus = "dismissed"
)

// KeyInfo represents information about a cryptographic key
type KeyInfo struct {
	ID              uuid.UUID     `json:"id"`
	KeyID           string        `json:"key_id"`
	KeyType         KeyType       `json:"key_type"`
	KeyUsage        KeyUsage      `json:"key_usage"`
	Algorithm       string        `json:"algorithm"`
	KeySize         int           `json:"key_size"`
	IsActive        bool          `json:"is_active"`
	HSMConfigID     uuid.UUID     `json:"hsm_config_id"`
	KeyLabel        string        `json:"key_label"`
	RotationPolicy  string        `json:"rotation_policy"`
	RotationPeriod  int           `json:"rotation_period"` // days
	LastRotatedAt   *time.Time    `json:"last_rotated_at"`
	NextRotationAt  *time.Time    `json:"next_rotation_at"`
	CreatedAt       time.Time     `json:"created_at"`
	ExpiresAt       *time.Time    `json:"expires_at,omitempty"`
}

// KeyType represents the type of cryptographic key
type KeyType string

const (
	KeyTypeRSA     KeyType = "RSA"
	KeyTypeECDSA   KeyType = "ECDSA"
	KeyTypeAES     KeyType = "AES"
	KeyTypeHMAC    KeyType = "HMAC"
	KeyTypeED25519 KeyType = "ED25519"
)

// KeyUsage represents the intended usage of the key
type KeyUsage string

const (
	KeyUsageSigning     KeyUsage = "signing"
	KeyUsageEncryption  KeyUsage = "encryption"
	KeyUsageDecryption  KeyUsage = "decryption"
	KeyUsageKeyExchange KeyUsage = "key_exchange"
	KeyUsageAuthentication KeyUsage = "authentication"
	KeyUsageVerification KeyUsage = "verification"
)
