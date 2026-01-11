// Shared Logger Package - Structured JSON logging for CSIC Platform
// Government-grade audit logging with structured output

package logger

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// LogLevel represents the severity level of a log entry
type LogLevel string

const (
	LevelDebug LogLevel = "DEBUG"
	LevelInfo  LogLevel = "INFO"
	LevelWarn  LogLevel = "WARN"
	LevelError LogLevel = "ERROR"
	LevelFatal LogLevel = "FATAL"
)

// AuditEventType represents the category of an audit event
type AuditEventType string

const (
	EventAuth          AuditEventType = "AUTHENTICATION"
	EventDataAccess    AuditEventType = "DATA_ACCESS"
	EventDataModify    AuditEventType = "DATA_MODIFICATION"
	EventSystem        AuditEventType = "SYSTEM_EVENT"
	EventSecurity      AuditEventType = "SECURITY_EVENT"
	EventCompliance    AuditEventType = "COMPLIANCE_EVENT"
	EventInvestigation AuditEventType = "INVESTIGATION_EVENT"
)

// LogEntry represents a structured log entry
type LogEntry struct {
	Timestamp   time.Time              `json:"timestamp"`
	Level       LogLevel               `json:"level"`
	Service     string                 `json:"service"`
	EventType   AuditEventType         `json:"event_type,omitempty"`
	Operation   string                 `json:"operation"`
	UserID      string                 `json:"user_id,omitempty"`
	SessionID   string                 `json:"session_id,omitempty"`
	Resource    string                 `json:"resource,omitempty"`
	Action      string                 `json:"action,omitempty"`
	Result      string                 `json:"result,omitempty"`
	Message     string                 `json:"message"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	Error       string                 `json:"error,omitempty"`
	TraceID     string                 `json:"trace_id,omitempty"`
	IPAddress   string                 `json:"ip_address,omitempty"`
	UserAgent   string                 `json:"user_agent,omitempty"`
}

// Logger provides structured logging capabilities
type Logger struct {
	*zap.Logger
	mu        sync.Mutex
	service   string
	auditSink AuditSink
}

// AuditSink defines the interface for audit log persistence
type AuditSink interface {
	WriteAuditLog(ctx context.Context, entry LogEntry) error
	Close() error
}

// ConsoleAuditSink writes audit logs to console
type ConsoleAuditSink struct{}

func (s *ConsoleAuditSink) WriteAuditLog(ctx context.Context, entry LogEntry) error {
	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("failed to marshal audit log: %w", err)
	}
	fmt.Printf("[AUDIT] %s\n", string(data))
	return nil
}

func (s *ConsoleAuditSink) Close() error {
	return nil
}

// FileAuditSink writes audit logs to a file
type FileAuditSink struct {
	file *os.File
	mu   sync.Mutex
}

func NewFileAuditSink(filepath string) (*FileAuditSink, error) {
	file, err := os.OpenFile(filepath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return nil, fmt.Errorf("failed to open audit log file: %w", err)
	}
	return &FileAuditSink{file: file}, nil
}

func (s *FileAuditSink) WriteAuditLog(ctx context.Context, entry LogEntry) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("failed to marshal audit log: %w", err)
	}
	_, err = s.file.WriteString(fmt.Sprintf("%s\n", string(data)))
	return err
}

func (s *FileAuditSink) Close() error {
	return s.file.Close()
}

// Config holds logger configuration
type Config struct {
	ServiceName   string
	LogLevel      string
	OutputPath    string
	AuditLogPath  string
	Development   bool
	JSONOutput    bool
}

// NewLogger creates a new Logger instance
func NewLogger(cfg Config) (*Logger, error) {
	var zapCfg zap.Config

	if cfg.Development {
		zapCfg = zap.NewDevelopmentConfig()
		zapCfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	} else {
		zapCfg = zap.NewProductionConfig()
		zapCfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	}

	// Set log level
	level, err := zapcore.ParseLevel(cfg.LogLevel)
	if err != nil {
		level = zapcore.InfoLevel
	}
	zapCfg.Level = zap.NewAtomicLevelAt(level)

	// Set output path
	if cfg.OutputPath != "" {
		zapCfg.OutputPaths = []string{cfg.OutputPath}
		zapCfg.ErrorOutputPaths = []string{cfg.OutputPath}
	}

	// Create the base logger
	baseLogger, err := zapCfg.Build()
	if err != nil {
		return nil, fmt.Errorf("failed to build logger: %w", err)
	}

	logger := &Logger{
		Logger:     baseLogger,
		service:    cfg.ServiceName,
		auditSink:  &ConsoleAuditSink{},
	}

	// Setup audit log sink if path provided
	if cfg.AuditLogPath != "" {
		fileSink, err := NewFileAuditSink(cfg.AuditLogPath)
		if err != nil {
			logger.Warn("failed to create audit log file, using console", zap.String("path", cfg.AuditLogPath))
		} else {
			logger.auditSink = fileSink
		}
	}

	return logger, nil
}

// WithService returns a new logger with the specified service name
func (l *Logger) WithService(service string) *Logger {
	newLogger := &Logger{
		Logger:     l.Logger.With(zap.String("service", service)),
		service:    service,
		auditSink:  l.auditSink,
	}
	return newLogger
}

// SetAuditSink replaces the audit sink
func (l *Logger) SetAuditSink(sink AuditSink) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.auditSink = sink
}

// Log creates a structured log entry
func (l *Logger) Log(ctx context.Context, entry LogEntry) {
	entry.Timestamp = time.Now().utc()
	entry.Service = l.service

	// Convert to zap fields
	fields := []zap.Field{
		zap.String("level", string(entry.Level)),
		zap.String("service", entry.Service),
		zap.String("operation", entry.Operation),
		zap.String("message", entry.Message),
		zap.Time("timestamp", entry.Timestamp),
	}

	if entry.EventType != "" {
		fields = append(fields, zap.String("event_type", string(entry.EventType)))
	}
	if entry.UserID != "" {
		fields = append(fields, zap.String("user_id", entry.UserID))
	}
	if entry.SessionID != "" {
		fields = append(fields, zap.String("session_id", entry.SessionID))
	}
	if entry.Resource != "" {
		fields = append(fields, zap.String("resource", entry.Resource))
	}
	if entry.Action != "" {
		fields = append(fields, zap.String("action", entry.Action))
	}
	if entry.Result != "" {
		fields = append(fields, zap.String("result", entry.Result))
	}
	if entry.Error != "" {
		fields = append(fields, zap.String("error", entry.Error))
	}
	if entry.TraceID != "" {
		fields = append(fields, zap.String("trace_id", entry.TraceID))
	}
	if entry.IPAddress != "" {
		fields = append(fields, zap.String("ip_address", entry.IPAddress))
	}
	if entry.UserAgent != "" {
		fields = append(fields, zap.String("user_agent", entry.UserAgent))
	}
	if entry.Metadata != nil {
		fields = append(fields, zap.Any("metadata", entry.Metadata))
	}

	// Log at appropriate level
	switch entry.Level {
	case LevelDebug:
		l.Logger.Debug(entry.Message, fields...)
	case LevelInfo:
		l.Logger.Info(entry.Message, fields...)
	case LevelWarn:
		l.Logger.Warn(entry.Message, fields...)
	case LevelError:
		l.Logger.Error(entry.Message, fields...)
	case LevelFatal:
		l.Logger.Fatal(entry.Message, fields...)
	}

	// Write to audit log for compliance events
	if entry.EventType != "" && entry.Level != LevelDebug {
		l.auditSink.WriteAuditLog(ctx, entry)
	}
}

// Audit writes a compliance audit log entry
func (l *Logger) Audit(ctx context.Context, eventType AuditEventType, operation, message string, metadata map[string]interface{}) {
	entry := LogEntry{
		Level:     LevelInfo,
		EventType: eventType,
		Operation: operation,
		Message:   message,
		Metadata:  metadata,
	}
	l.Log(ctx, entry)
}

// Info logs an info level message
func (l *Logger) Info(message string, fields ...zap.Field) {
	l.Logger.Info(message, fields...)
}

// Error logs an error level message
func (l *Logger) Error(message string, fields ...zap.Field) {
	l.Logger.Error(message, fields...)
}

// Debug logs a debug level message
func (l *Logger) Debug(message string, fields ...zap.Field) {
	l.Logger.Debug(message, fields...)
}

// Warn logs a warning level message
func (l *Logger) Warn(message string, fields ...zap.Field) {
	l.Logger.Warn(message, fields...)
}

// Fatal logs a fatal level message and exits
func (l *Logger) Fatal(message string, fields ...zap.Field) {
	l.Logger.Fatal(message, fields...)
}

// WithFields returns a new logger with additional fields
func (l *Logger) WithFields(fields ...zap.Field) *zap.Logger {
	return l.Logger.With(fields...)
}

// Sync flushes any buffered log entries
func (l *Logger) Sync() error {
	return l.Logger.Sync()
}

// Close closes the logger and any audit sinks
func (l *Logger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if closer, ok := l.auditSink.(interface{ Close() error }); ok {
		closer.Close()
	}
	return l.Sync()
}

// Helper function to ensure UTC timestamp
func (t time.Time) utc() time.Time {
	return t.UTC()
}
