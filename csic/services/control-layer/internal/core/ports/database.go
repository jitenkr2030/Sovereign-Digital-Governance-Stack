package ports

import (
	"context"

	"github.com/csic-platform/internal/core/domain"
)

// Database defines the interface for database operations.
// This interface provides methods for managing database connections and executing queries.
type Database interface {
	// Connect establishes a database connection
	Connect(ctx context.Context, dsn string) error

	// Disconnect closes the database connection
	Disconnect(ctx context.Context) error

	// Ping checks if the database is reachable
	Ping(ctx context.Context) error

	// GetDB returns the underlying database connection
	GetDB() interface{}

	// GetDriver returns the database driver name
	GetDriver() string

	// GetDSN returns the connection string
	GetDSN() string

	// SetMaxOpenConns sets the maximum number of open connections
	SetMaxOpenConns(n int)

	// SetMaxIdleConns sets the maximum number of idle connections
	SetMaxIdleConns(n int)

	// SetConnMaxLifetime sets the maximum connection lifetime
	SetConnMaxLifetime(d time.Duration)

	// GetStats returns connection pool statistics
	GetStats() *domain.DBStats

	// Close closes the database connection
	Close() error
}

// DatabaseTransaction defines the interface for database transactions.
type DatabaseTransaction interface {
	// Begin starts a new transaction
	Begin(ctx context.Context) error

	// Commit commits the transaction
	Commit(ctx context.Context) error

	// Rollback rolls back the transaction
	Rollback(ctx context.Context) error

	// GetTx returns the underlying transaction
	GetTx() interface{}

	// IsActive checks if the transaction is active
	IsActive() bool
}

// DatabaseQueryBuilder defines the interface for building database queries.
type DatabaseQueryBuilder interface {
	// Select creates a select query
	Select(columns ...string) DatabaseQueryBuilder

	// From specifies the table
	From(table string) DatabaseQueryBuilder

	// Where adds where conditions
	Where(condition string, args ...interface{}) DatabaseQueryBuilder

	// AndWhere adds additional where conditions
	AndWhere(condition string, args ...interface{}) DatabaseQueryBuilder

	// OrWhere adds or where conditions
	OrWhere(condition string, args ...interface{}) DatabaseQueryBuilder

	// Join adds a join
	Join(table, condition string, args ...interface{}) DatabaseQueryBuilder

	// LeftJoin adds a left join
	LeftJoin(table, condition string, args ...interface{}) DatabaseQueryBuilder

	// RightJoin adds a right join
	RightJoin(table, condition string, args ...interface{}) DatabaseQueryBuilder

	// GroupBy adds group by clause
	GroupBy(columns ...string) DatabaseQueryBuilder

	// Having adds having clause
	Having(condition string, args ...interface{}) DatabaseQueryBuilder

	// OrderBy adds order by clause
	OrderBy(columns ...string) DatabaseQueryBuilder

	// Asc specifies ascending order
	Asc() DatabaseQueryBuilder

	// Desc specifies descending order
	Desc() DatabaseQueryBuilder

	// Limit sets the limit
	Limit(limit int) DatabaseQueryBuilder

	// Offset sets the offset
	Offset(offset int) DatabaseQueryBuilder

	// Build builds the query
	Build() (string, []interface{})

	// Scan scans the result
	Scan(ctx context.Context, dest interface{}) error

	// ScanRow scans a single row
	ScanRow(ctx context.Context, dest interface{}) error

	// Exec executes the query
	Exec(ctx context.Context) (int64, error)
}

// DatabaseMigrator defines the interface for database migrations.
type DatabaseMigrator interface {
	// Migrate runs all pending migrations
	Migrate(ctx context.Context) error

	// MigrateTo runs migrations up to a specific version
	MigrateTo(ctx context.Context, version string) error

	// Rollback rolls back the last migration
	Rollback(ctx context.Context) error

	// RollbackTo rolls back migrations to a specific version
	RollbackTo(ctx context.Context, version string) error

	// Create creates a new migration file
	Create(ctx context.Context, name string) error

	// GetMigrations returns all migrations
	GetMigrations(ctx context.Context) ([]*domain.Migration, error)

	// GetCurrentVersion returns the current migration version
	GetCurrentVersion(ctx context.Context) (string, error)

	// Force sets the migration version
	Force(ctx context.Context, version int) error
}

// DatabaseSchemaManager defines the interface for schema management.
type DatabaseSchemaManager interface {
	// CreateSchema creates a new schema
	CreateSchema(ctx context.Context, name string) error

	// DropSchema drops a schema
	DropSchema(ctx context.Context, name string) error

	// GetSchemas returns all schemas
	GetSchemas(ctx context.Context) ([]string, error)

	// SchemaExists checks if a schema exists
	SchemaExists(ctx context.Context, name string) (bool, error)

	// CreateTable creates a new table
	CreateTable(ctx context.Context, table *domain.TableDefinition) error

	// DropTable drops a table
	DropTable(ctx context.Context, name string) error

	// TruncateTable truncates a table
	TruncateTable(ctx context.Context, name string) error

	// RenameTable renames a table
	RenameTable(ctx context.Context, oldName, newName string) error

	// GetTables returns all tables
	GetTables(ctx context.Context) ([]*domain.TableInfo, error)

	// GetTable returns a table definition
	GetTable(ctx context.Context, name string) (*domain.TableInfo, error)

	// TableExists checks if a table exists
	TableExists(ctx context.Context, name string) (bool, error)

	// AddColumn adds a column to a table
	AddColumn(ctx context.Context, tableName string, column *domain.ColumnDefinition) error

	// DropColumn drops a column from a table
	DropColumn(ctx context.Context, tableName, columnName string) error

	// ModifyColumn modifies a column in a table
	ModifyColumn(ctx context.Context, tableName string, column *domain.ColumnDefinition) error

	// RenameColumn renames a column
	RenameColumn(ctx context.Context, tableName, oldName, newName string) error

	// GetColumns returns all columns of a table
	GetColumns(ctx context.Context, tableName string) ([]*domain.ColumnInfo, error)

	// GetColumn returns a column definition
	GetColumn(ctx context.Context, tableName, columnName string) (*domain.ColumnInfo, error)

	// AddIndex adds an index to a table
	AddIndex(ctx context.Context, tableName string, index *domain.IndexDefinition) error

	// DropIndex drops an index from a table
	DropIndex(ctx context.Context, tableName, indexName string) error

	// RenameIndex renames an index
	RenameIndex(ctx context.Context, tableName, oldName, newName string) error

	// GetIndexes returns all indexes of a table
	GetIndexes(ctx context.Context, tableName string) ([]*domain.IndexInfo, error)

	// AddForeignKey adds a foreign key constraint
	AddForeignKey(ctx context.Context, tableName string, fk *domain.ForeignKeyDefinition) error

	// DropForeignKey drops a foreign key constraint
	DropForeignKey(ctx context.Context, tableName, fkName string) error

	// GetForeignKeys returns all foreign keys of a table
	GetForeignKeys(ctx context.Context, tableName string) ([]*domain.ForeignKeyInfo, error)

	// AddPrimaryKey adds a primary key constraint
	AddPrimaryKey(ctx context.Context, tableName string, columns []string) error

	// DropPrimaryKey drops a primary key constraint
	DropPrimaryKey(ctx context.Context, tableName string) error

	// AddUniqueConstraint adds a unique constraint
	AddUniqueConstraint(ctx context.Context, tableName string, columns []string, constraintName string) error

	// DropUniqueConstraint drops a unique constraint
	DropUniqueConstraint(ctx context.Context, tableName, constraintName string) error
}

// DatabaseHealthChecker defines health check operations for databases.
type DatabaseHealthChecker interface {
	// HealthCheck performs a health check
	HealthCheck(ctx context.Context) (*domain.HealthStatus, error)

	// GetConnectionStats returns connection statistics
	GetConnectionStats(ctx context.Context) (*domain.DBConnectionStats, error)

	// GetReplicationStats returns replication statistics
	GetReplicationStats(ctx context.Context) (*domain.DBReplicationStats, error)

	// GetStorageStats returns storage statistics
	GetStorageStats(ctx context.Context) (*domain.DBStorageStats, error)

	// GetPerformanceStats returns performance statistics
	GetPerformanceStats(ctx context.Context) (*domain.DBPerformanceStats, error)
}

// ConnectionPoolManager defines connection pool management operations.
type ConnectionPoolManager interface {
	// GetPoolStatus returns the pool status
	GetPoolStatus() *domain.ConnectionPoolStatus

	// SetMaxSize sets the maximum pool size
	SetMaxSize(size int)

	// SetMinSize sets the minimum pool size
	SetMinSize(size int)

	// SetMaxLifetime sets the maximum connection lifetime
	SetMaxLifetime(lifetime time.Duration)

	// SetIdleTimeout sets the idle connection timeout
	SetIdleTimeout(timeout time.Duration)

	// SetMaxRetries sets the maximum retry count
	SetMaxRetries(retries int)

	// SetRetryDelay sets the retry delay
	SetRetryDelay(delay time.Duration)

	// Acquire acquires a connection from the pool
	Acquire(ctx context.Context) (interface{}, error)

	// Release releases a connection back to the pool
	Release(conn interface{})

	// Close closes all connections
	Close(ctx context.Context) error
}

// DatabaseReplication defines replication operations.
type DatabaseReplication interface {
	// IsMaster checks if the database is a master
	IsMaster(ctx context.Context) (bool, error)

	// IsReplica checks if the database is a replica
	IsReplica(ctx context.Context) (bool, error)

	// GetMasterInfo returns master information
	GetMasterInfo(ctx context.Context) (*domain.ReplicationInfo, error)

	// GetReplicaInfo returns replica information
	GetReplicaInfo(ctx context.Context) (*domain.ReplicationInfo, error)

	// GetReplicationLag returns the replication lag
	GetReplicationLag(ctx context.Context) (time.Duration, error)

	// Promote promotes a replica to master
	Promote(ctx context.Context) error

	// ConfigureReplication configures replication
	ConfigureReplication(ctx context.Context, config *domain.ReplicationConfig) error

	// GetReplicationStatus returns the replication status
	GetReplicationStatus(ctx context.Context) (*domain.ReplicationStatus, error)
}

// DatabaseBackup defines backup operations.
type DatabaseBackup interface {
	// CreateBackup creates a backup
	CreateBackup(ctx context.Context, opts *domain.BackupOptions) error

	// RestoreBackup restores a backup
	RestoreBackup(ctx context.Context, backupPath string, opts *domain.RestoreOptions) error

	// ListBackups lists available backups
	ListBackups(ctx context.Context) ([]*domain.BackupInfo, error)

	// GetBackupInfo returns backup information
	GetBackupInfo(ctx context.Context, backupID string) (*domain.BackupInfo, error)

	// DeleteBackup deletes a backup
	DeleteBackup(ctx context.Context, backupID string) error

	// VerifyBackup verifies a backup
	VerifyBackup(ctx context.Context, backupID string) error

	// ScheduleBackup schedules a backup
	ScheduleBackup(ctx context.Context, schedule *domain.BackupSchedule) error

	// CancelScheduledBackup cancels a scheduled backup
	CancelScheduledBackup(ctx context.Context, scheduleID string) error
}

// DatabaseEncryption defines encryption operations.
type DatabaseEncryption interface {
	// Encrypt encrypts data
	Encrypt(ctx context.Context, plaintext []byte) ([]byte, error)

	// Decrypt decrypts data
	Decrypt(ctx context.Context, ciphertext []byte) ([]byte, error)

	// EncryptColumn encrypts a column value
	EncryptColumn(ctx context.Context, table, column string, value []byte) ([]byte, error)

	// DecryptColumn decrypts a column value
	DecryptColumn(ctx context.Context, table, column string, value []byte) ([]byte, error)

	// GetKeyInfo returns the encryption key information
	GetKeyInfo(ctx context.Context) (*domain.EncryptionKeyInfo, error)

	// RotateKey rotates the encryption key
	RotateKey(ctx context.Context, newKeyID string) error

	// ReEncrypt re-encrypts all data with a new key
	ReEncrypt(ctx context.Context, oldKeyID, newKeyID string) error
}

// DatabaseLogger defines logging operations for database operations.
type DatabaseLogger interface {
	// LogQuery logs a query
	LogQuery(ctx context.Context, query string, args []interface{}, duration time.Duration, err error)

	// LogTransaction logs a transaction
	LogTransaction(ctx context.Context, txID string, operation string, duration time.Duration, err error)

	// LogSlowQuery logs a slow query
	LogSlowQuery(ctx context.Context, query string, args []interface{}, duration time.Duration, threshold time.Duration)

	// LogError logs a database error
	LogError(ctx context.Context, err error, query string)

	// GetQueryLog returns the query log
	GetQueryLog(ctx context.Context, filter *domain.QueryLogFilter) ([]*domain.QueryLogEntry, error)

	// GetSlowQueryLog returns the slow query log
	GetSlowQueryLog(ctx context.Context, threshold time.Duration) ([]*domain.QueryLogEntry, error)

	// ClearQueryLog clears the query log
	ClearQueryLog(ctx context.Context) error
}

// DatabaseObserver defines database observation operations.
type DatabaseObserver interface {
	// ObserveQuery observes a query execution
	ObserveQuery(ctx context.Context, query string, args []interface{}, duration time.Duration, err error)

	// ObserveTransaction observes a transaction
	ObserveTransaction(ctx context.Context, txID string, operation string, duration time.Duration, err error)

	// ObserveConnection observes a connection event
	ObserveConnection(ctx context.Context, event string, duration time.Duration, err error)

	// ObserveLock observes a lock event
	ObserveLock(ctx context.Context, lockType string, table string, duration time.Duration, err error)

	// GetObservations returns observations
	GetObservations(ctx context.Context, filter *domain.ObservationFilter) ([]*domain.Observation, error)

	// GetMetrics returns metrics
	GetMetrics(ctx context.Context) (*domain.DatabaseMetrics, error)
}

// DatabaseQueryCache defines query caching operations.
type DatabaseQueryCache interface {
	// Get gets a cached query result
	Get(ctx context.Context, key string) ([]byte, bool, error)

	// Set sets a query result in cache
	Set(ctx context.Context, key string, result []byte, ttl time.Duration) error

	// Invalidate invalidates cached query results
	Invalidate(ctx context.Context, pattern string) error

	// InvalidateTable invalidates all cached results for a table
	InvalidateTable(ctx context.Context, tableName string) error

	// InvalidateQuery invalidates a specific query
	InvalidateQuery(ctx context.Context, query string, args []interface{}) error

	// Clear clears the cache
	Clear(ctx context.Context) error

	// GetStats returns cache statistics
	GetStats(ctx context.Context) (*domain.QueryCacheStats, error)
}

// DatabaseQueryAnalyzer defines query analysis operations.
type DatabaseQueryAnalyzer interface {
	// AnalyzeQuery analyzes a query
	AnalyzeQuery(ctx context.Context, query string, args []interface{}) (*domain.QueryAnalysis, error)

	// ExplainQuery explains a query execution plan
	ExplainQuery(ctx context.Context, query string, args []interface{}) (*domain.QueryPlan, error)

	// SuggestIndex suggests indexes for a query
	SuggestIndex(ctx context.Context, query string, args []interface{}) ([]*domain.IndexSuggestion, error)

	// GetSlowQueries returns slow queries
	GetSlowQueries(ctx context.Context, threshold time.Duration) ([]*domain.SlowQuery, error)

	// GetQueryStats returns query statistics
	GetQueryStats(ctx context.Context) ([]*domain.QueryStats, error)

	// GetTableStats returns table statistics
	GetTableStats(ctx context.Context, tableName string) (*domain.TableStats, error)
}

// DatabaseLockManager defines distributed lock operations.
type DatabaseLockManager interface {
	// Acquire acquires a lock
	Acquire(ctx context.Context, name string, ttl time.Duration) (string, error)

	// AcquireWithRetry acquires a lock with retry
	AcquireWithRetry(ctx context.Context, name string, ttl time.Duration, maxRetries int, retryDelay time.Duration) (string, error)

	// Release releases a lock
	Release(ctx context.Context, name, token string) error

	// Extend extends a lock
	Extend(ctx context.Context, name, token string, ttl time.Duration) error

	// IsLocked checks if a lock is held
	IsLocked(ctx context.Context, name string) (bool, error)

	// GetLockerInfo returns lock information
	GetLockerInfo(ctx context.Context) ([]*domain.LockInfo, error)

	// ForceRelease forces the release of a lock
	ForceRelease(ctx context.Context, name string) error
}

// DatabaseEventListener defines database event listening operations.
type DatabaseEventListener interface {
	// ListenForNotifications listens for notifications
	ListenForNotifications(ctx context.Context, channel string, handler func(payload string)) error

	// Notify sends a notification
	Notify(ctx context.Context, channel, payload string) error

	// ListenForChanges listens for data changes
	ListenForChanges(ctx context.Context, table string, handler func(*domain.ChangeEvent)) error

	// ListenForTrigger listens for trigger events
	ListenForTrigger(ctx context.Context, triggerName string, handler func(*domain.TriggerEvent)) error
}
