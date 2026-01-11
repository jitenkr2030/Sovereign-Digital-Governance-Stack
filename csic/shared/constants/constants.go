// Constants Package - System-wide constants for CSIC Platform
// Configuration values, limits, and enumerated types

package constants

import "time"

// Service Names
const (
	ServiceAPI Gateway     = "api-gateway"
	ServiceIdentity        = "identity-service"
	ServiceIngestion       = "ingestion-engine"
	ServiceRisk            = "risk-engine"
	ServiceCaseManager     = "case-manager"
	ServiceAuditLog        = "audit-log-service"
	ServiceHealthMonitor   = "health-monitor"
	ServiceControlLayer    = "control-layer"
)

// HTTP Status Codes
const (
	StatusOK                  = 200
	StatusCreated             = 201
	StatusAccepted            = 202
	StatusNoContent           = 204
	StatusBadRequest          = 400
	StatusUnauthorized        = 401
	StatusForbidden           = 403
	StatusNotFound            = 404
	StatusMethodNotAllowed    = 405
	StatusConflict            = 409
	StatusUnprocessableEntity = 422
	StatusTooManyRequests     = 429
	StatusInternalServerError = 500
	StatusServiceUnavailable  = 503
)

// Blockchain Constants
const (
	BitcoinNetwork  = "bitcoin"
	EthereumNetwork = "ethereum"
	BSCNetwork      = "bsc"
	PolygonNetwork  = "polygon"

	BitcoinConfirmations  = 6
	EthereumConfirmations = 12

	MaxTransactionSize     = 100000 // bytes
	MaxInputCount          = 100
	MaxOutputCount         = 100
)

// System Limits
const (
	MaxPageSize     = 100
	DefaultPageSize = 20
	MaxAPIRetries   = 3
	APITimeout      = 30 * time.Second

	SessionDuration    = 24 * time.Hour
	TokenExpiry        = 1 * time.Hour
	PasswordMinLength  = 12
	PasswordMaxLength  = 128
	MaxLoginAttempts   = 5
	LockoutDuration    = 15 * time.Minute

	RateLimitRequests = 100
	RateLimitWindow   = time.Minute

	MaxAlertPerPage    = 50
	MaxLogsPerPage     = 100
	MaxTransactionsPerQuery = 10000
)

// Risk Score Constants
const (
	RiskScoreLow      = 30
	RiskScoreMedium   = 50
	RiskScoreHigh     = 70
	RiskScoreCritical = 90

	RiskLevelLow      = "LOW"
	RiskLevelMedium   = "MEDIUM"
	RiskLevelHigh     = "HIGH"
	RiskLevelCritical = "CRITICAL"
)

// Alert Severity Levels
const (
	SeverityInfo     = "INFO"
	SeverityWarning  = "WARNING"
	SeverityCritical = "CRITICAL"
	SeverityEmergency = "EMERGENCY"
)

// Alert Status
const (
	AlertStatusActive      = "ACTIVE"
	AlertStatusAcknowledged = "ACKNOWLEDGED"
	AlertStatusResolved    = "RESOLVED"
	AlertStatusDismissed   = "DISMISSED"
)

// Case Status
const (
	CaseStatusOpen       = "OPEN"
	CaseStatusInProgress = "IN_PROGRESS"
	CaseStatusReview     = "UNDER_REVIEW"
	CaseStatusClosed     = "CLOSED"
	CaseStatusArchived   = "ARCHIVED"
)

// Compliance Status
const (
	ComplianceStatusCompliant    = "COMPLIANT"
	ComplianceStatusNonCompliant = "NON_COMPLIANT"
	ComplianceStatusUnderReview  = "UNDER_REVIEW"
	ComplianceStatusPending      = "PENDING"
	ComplianceStatusRemediated   = "REMEDIATED"
)

// Exchange Status
const (
	ExchangeStatusOperational = "OPERATIONAL"
	ExchangeStatusSuspended   = "SUSPENDED"
	ExchangeStatusRevoked     = "REVOKED"
	ExchangeStatusPending     = "PENDING"
	ExchangeStatusInactive    = "INACTIVE"
)

// Transaction Status
const (
	TxStatusPending    = "PENDING"
	TxStatusConfirmed  = "CONFIRMED"
	TxStatusFlagged    = "FLAGGED"
	TxStatusBlocked    = "BLOCKED"
	TxStatusFailed     = "FAILED"
)

// User Roles
const (
	RoleAdmin            = "ADMIN"
	RoleRegulator        = "REGULATOR"
	RoleAuditor          = "AUDITOR"
	RoleAnalyst          = "ANALYST"
	RoleInvestigator     = "INVESTIGATOR"
	RoleViewer           = "VIEWER"
	RoleSystem           = "SYSTEM"
)

// Permissions
const (
	PermissionReadAll    = "read:all"
	PermissionWriteAll   = "write:all"
	PermissionManageUsers = "manage:users"
	PermissionManageSettings = "manage:settings"
	PermissionViewAlerts = "view:alerts"
	PermissionManageAlerts = "manage:alerts"
	PermissionViewExchanges = "view:exchanges"
	PermissionManageExchanges = "manage:exchanges"
	PermissionViewCompliance = "view:compliance"
	PermissionManageCompliance = "manage:compliance"
	PermissionViewAuditLogs = "view:audit"
	PermissionExportData = "export:data"
	PermissionInvestigate = "investigate"
)

// Audit Event Types
const (
	EventTypeLogin              = "LOGIN"
	EventTypeLogout             = "LOGOUT"
	EventTypeLoginFailed        = "LOGIN_FAILED"
	EventTypePasswordChange     = "PASSWORD_CHANGE"
	EventTypeDataAccess         = "DATA_ACCESS"
	EventTypeDataCreate         = "DATA_CREATE"
	EventTypeDataUpdate         = "DATA_UPDATE"
	EventTypeDataDelete         = "DATA_DELETE"
	EventTypeExport             = "EXPORT"
	EventTypePermissionChange   = "PERMISSION_CHANGE"
	EventTypeConfigurationChange = "CONFIGURATION_CHANGE"
	EventTypeSystemEvent        = "SYSTEM_EVENT"
	EventTypeSecurityEvent      = "SECURITY_EVENT"
	EventTypeComplianceEvent    = "COMPLIANCE_EVENT"
)

// Sanctions List Sources
const (
	SanctionsSourceOFAC      = "OFAC"
	SanctionsSourceUN        = "UN_CONSOLIDATED"
	SanctionsSourceEU        = "EU_SANCTIONS"
	SanctionsSourceUK        = "UK_SANCTIONS"
	SanctionsSourceFinCEN    = "FINCEN"
	SanctionsSourceCustom    = "CUSTOM"
)

// Geographic Regions
const (
	RegionNorthAmerica = "NA"
	RegionSouthAmerica = "SA"
	RegionEurope       = "EU"
	RegionAsiaPacific  = "APAC"
	RegionMiddleEast   = "ME"
	RegionAfrica       = "AF"
	RegionGlobal       = "GLOBAL"
)

// Currency Codes (ISO 4217)
const (
	CurrencyUSD = "USD"
	CurrencyEUR = "EUR"
	CurrencyGBP = "GBP"
	CurrencyJPY = "JPY"
	CurrencyCNY = "CNY"
	CurrencyBTC = "BTC"
	CurrencyETH = "ETH"
	CurrencyUSDT = "USDT"
	CurrencyUSDC = "USDC"
)

// Time Formats
const (
	TimeFormatISO8601     = "2006-01-02T15:04:05Z07:00"
	TimeFormatDateOnly    = "2006-01-02"
	TimeFormatTimeOnly    = "15:04:05"
	TimeFormatLogFile     = "2006-01-02-15-04-05"
	TimeFormatDisplay     = "Jan 02, 2006 15:04"
)

// Datebase Table Names
const (
	TableUsers           = "users"
	TableSessions        = "sessions"
	TableAlerts          = "alerts"
	TableCases           = "cases"
	TableTransactions    = "transactions"
	TableWallets         = "wallets"
	TableExchanges       = "exchanges"
	TableCompliance      = "compliance_records"
	TableAuditLogs       = "audit_logs"
	TableRiskScores      = "risk_scores"
	TableSanctions       = "sanctions_list"
	TableBlacklist       = "blacklist"
	TableWhitelist       = "whitelist"
	TableMiners          = "miners"
	TableBlocks          = "blocks"
	TableMempool         = "mempool"
	TableRegulatoryFilings = "regulatory_filings"
	TableEnforcementActions = "enforcement_actions"
)

// Kafka Topic Names
const (
	TopicTransactions     = "csic.transactions"
	TopicAlerts           = "csic.alerts"
	TopicRiskScores       = "csic.risk.scores"
	TopicCompliance       = "csic.compliance"
	TopicAudit            = "csic.audit"
	TopicBlocks           = "csic.blocks"
	TopicNotifications    = "csic.notifications"
)

// Redis Key Prefixes
const (
	RedisKeyUserSession   = "csic:session:"
	RedisKeyUserProfile   = "csic:user:"
	RedisKeyRateLimit     = "csic:ratelimit:"
	RedisKeyCache         = "csic:cache:"
	RedisKeyLock          = "csic:lock:"
	RedisKeyConfig        = "csic:config:"
)

// Cache TTLs
const (
	CacheTTLShort    = 1 * time.Minute
	CacheTTLMedium   = 5 * time.Minute
	CacheTTLLong     = 30 * time.Minute
	CacheTTLVeryLong = 1 * time.Hour
	CacheTTLPersistent = 24 * time.Hour
)

// HTTP Headers
const (
	HeaderContentType     = "Content-Type"
	HeaderAuthorization   = "Authorization"
	HeaderXRequestID      = "X-Request-ID"
	HeaderXCorrelationID  = "X-Correlation-ID"
	HeaderXUserID         = "X-User-ID"
	HeaderXRole           = "X-Role"
	HeaderXAPIKey         = "X-API-Key"
	HeaderAcceptLanguage  = "Accept-Language"
	HeaderUserAgent       = "User-Agent"
	HeaderXForwardedFor   = "X-Forwarded-For"
	HeaderXRealIP         = "X-Real-IP"
)

// Content Types
const (
	ContentTypeJSON       = "application/json"
	ContentTypeXML        = "application/xml"
	ContentTypeCSV        = "text/csv"
	ContentTypePDF        = "application/pdf"
	ContentTypeFormURL    = "application/x-www-form-urlencoded"
	ContentTypeMultipart  = "multipart/form-data"
)

// Environment Names
const (
	EnvDevelopment = "development"
	EnvStaging     = "staging"
	EnvProduction  = "production"
	EnvTest        = "test"
)

// Deployment Types
const (
	DeploymentTypeOnPremise = "on_premise"
	DeploymentTypeCloud     = "cloud"
	DeploymentTypeHybrid    = "hybrid"
	DeploymentTypeAirGapped = "air_gapped"
)
